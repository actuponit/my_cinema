import { json } from "../_shared/cors.ts";
import { dispense } from "../_shared/mqtt.ts";
import { verifyChapaSignature } from "../_shared/signatures.ts";
import { getSupabaseAdmin } from "../_shared/supabaseAdmin.ts";

// Permanent problems (unknown tx, amount mismatch) answer 200 so Chapa stops retrying;
// they're logged with ERROR for follow-up. Transient ones answer 5xx so Chapa retries.
Deno.serve(async (req: Request) => {
  if (req.method !== "POST") return json({ code: "method_not_allowed" }, 405);

  const secret = Deno.env.get("CHAPA_WEBHOOK_SECRET");
  const chapaSecretKey = Deno.env.get("CHAPA_SECRET_KEY");
  if (!secret || !chapaSecretKey) {
    console.error("[chapa-webhook] Missing CHAPA_WEBHOOK_SECRET or CHAPA_SECRET_KEY env var");
    return json({ code: "misconfigured" }, 500);
  }

  // 1. Verify HMAC-SHA256 signature over the raw body
  const rawBody = await req.text();
  if (!(await verifyChapaSignature(rawBody, req.headers.get("x-chapa-signature") ?? "", secret))) {
    return json({ code: "bad_signature" }, 401);
  }

  let body;
  try {
    body = JSON.parse(rawBody);
  } catch {
    return json({ code: "bad_json" }, 400);
  }
  const txRef = body.tx_ref || body.trx_ref;
  if (!txRef) return json({ code: "missing_tx_ref" }, 400);

  try {
    // 2. Re-verify with Chapa: the webhook body is not trusted for status or amount
    const verifyRes = await fetch(`https://api.chapa.co/v1/transaction/verify/${encodeURIComponent(txRef)}`, {
      headers: { "Authorization": `Bearer ${chapaSecretKey}` },
    });
    if (!verifyRes.ok) {
      console.error(`[chapa-webhook] Chapa verify ${txRef} failed with HTTP ${verifyRes.status}`);
      return json({ code: "chapa_verify_failed" }, 502);
    }
    const verifyData = await verifyRes.json();
    if (verifyData.status !== "success" || verifyData.data?.status !== "success") {
      console.warn(`[chapa-webhook] ${txRef} is not successful on Chapa:`, verifyData);
      return json({ status: "not_successful" });
    }
    if (verifyData.data.currency !== "ETB") {
      console.error(`[chapa-webhook] ERROR ${txRef} currency is ${verifyData.data.currency}, expected ETB`);
      return json({ status: "invalid_currency" });
    }

    const admin = getSupabaseAdmin();

    // 3. Paid amount must equal the stored basket, compared in cents
    const { data: rows, error: fetchErr } = await admin
      .from("transactions")
      .select("price, amount")
      .eq("tex_ref", txRef);
    if (fetchErr) {
      console.error("[chapa-webhook] Fetch basket failed:", fetchErr);
      return json({ code: "database_error" }, 500);
    }
    if (!rows?.length) {
      console.error(`[chapa-webhook] ERROR no transactions for tex_ref ${txRef}`);
      return json({ status: "unknown_tx_ref" });
    }
    const basketCents = rows.reduce(
      (sum: number, r: { price: number | string; amount: number | null }) =>
        sum + Math.round(Number(r.price) * 100) * (r.amount ?? 1),
      0,
    );
    const paidCents = Math.round(Number(verifyData.data.amount) * 100);
    if (basketCents !== paidCents) {
      console.error(`[chapa-webhook] ERROR amount mismatch for ${txRef}: paid ${paidCents}c, basket ${basketCents}c`);
      return json({ status: "amount_mismatch" });
    }

    // 4. Mark paid exactly once (only status=false rows flip; a trigger decrements stock)
    const { data: paidRows, error: markErr } = await admin.rpc("chapa_mark_paid", { p_tex_ref: txRef });
    if (markErr) {
      console.error("[chapa-webhook] chapa_mark_paid failed:", markErr);
      return json({ code: "database_error" }, 500);
    }
    if (!paidRows?.length) {
      console.log(`[chapa-webhook] Duplicate webhook for ${txRef}, already paid.`);
      return json({ status: "already_processed" });
    }

    // 5. Dispense
    const { data: slots, error: slotErr } = await admin
      .from("vending_machine_to_items")
      .select("id, motor_code")
      .in("id", paidRows.map((r: { combination_id: number }) => r.combination_id));
    if (slotErr || !slots) {
      console.error(`[chapa-webhook] ERROR paid but motor code lookup failed for ${txRef}:`, slotErr);
      return json({ status: "success", dispensed: false });
    }
    const motorBySlot = new Map(slots.map((s: { id: number; motor_code: string }) => [s.id, s.motor_code]));
    const motorCodes = paidRows.flatMap((r: { combination_id: number; amount: number | null }) => {
      const code = motorBySlot.get(r.combination_id);
      return code ? Array(r.amount ?? 1).fill(code) : [];
    });

    if (motorCodes.length) {
      try {
        await dispense(motorCodes.join(","));
      } catch (mqttErr) {
        // ponytail: no automatic refund; add one if dispense failures show up in logs
        console.error(`[chapa-webhook] ERROR paid but dispense failed for ${txRef}:`, mqttErr);
      }
    }
    return json({ status: "success" });
  } catch (err) {
    console.error("[chapa-webhook] Exception:", err);
    return json({ code: "internal_error" }, 500);
  }
});
