import { json } from "../_shared/cors.ts";
import { verifyCheckoutSignature } from "../_shared/signatures.ts";
import { getSupabaseAdmin } from "../_shared/supabaseAdmin.ts";

Deno.serve(async (req: Request) => {
  if (req.method !== "POST") return json({ code: "method_not_allowed" }, 405);

  const secret = Deno.env.get("VERIFY_WEBHOOK_SECRET");
  const verifyApiKey = Deno.env.get("VERIFY_API_KEY");
  if (!secret || !verifyApiKey) {
    console.error("[verify-webhook] Missing VERIFY_WEBHOOK_SECRET or VERIFY_API_KEY env var");
    return json({ code: "misconfigured" }, 500);
  }

  // 1. Signature + timestamp freshness
  const rawBody = await req.text();
  const reason = await verifyCheckoutSignature(
    rawBody,
    req.headers.get("VerifyCheckout-Timestamp") ?? "",
    req.headers.get("VerifyCheckout-Signature") ?? "",
    secret,
  );
  if (reason) return json({ code: "bad_signature", message: reason }, 401);

  // 2. Find the deposit id. Anything we can't use is acknowledged so verify.et stops retrying.
  let event;
  try {
    event = JSON.parse(rawBody);
  } catch {
    console.error("[verify-webhook] Undecodable body");
    return json({ status: "ignored" });
  }
  if (event.type === "webhook.test") return json({ status: "ok" });

  const depositId = event?.data?.object?.id;
  if (!depositId) {
    console.log(`[verify-webhook] Event ${event.id} (${event.type}) has no deposit id`);
    return json({ status: "ignored" });
  }

  try {
    // 3. Fetch the authoritative deposit state; the event body is only a hint
    const verifyBaseUrl = (Deno.env.get("VERIFY_API_BASE_URL") || "https://checkoutapi.verify.et").replace(/\/+$/, "");
    const verifyRes = await fetch(`${verifyBaseUrl}/v1/deposits/${encodeURIComponent(depositId)}`, {
      headers: { "Authorization": `Bearer ${verifyApiKey}`, "VerifyCheckout-Version": "2026-06-01" },
    });
    if (!verifyRes.ok) {
      console.error(`[verify-webhook] Verify.et returned ${verifyRes.status} for deposit ${depositId}`);
      return json({ code: "provider_error" }, 502);
    }
    const deposit = (await verifyRes.json())?.data;
    if (!deposit?.id || !deposit.status || !deposit.merchant_customer_id) {
      console.error(`[verify-webhook] ERROR unusable deposit ${depositId} from Verify.et:`, deposit);
      return json({ code: "invalid_provider_data" }, 502);
    }

    // 4. Apply atomically (idempotent: a succeeded deposit is credited once)
    const { error } = await getSupabaseAdmin().rpc("wallet_apply_deposit", {
      p_deposit_id: deposit.id,
      p_user_id: deposit.merchant_customer_id,
      p_status: deposit.status,
      p_amount: deposit.status === "succeeded" ? deposit.amount : null,
    });
    if (error) {
      console.error(`[verify-webhook] wallet_apply_deposit failed for ${depositId}:`, error);
      return json({ code: "database_error" }, 500);
    }

    console.log(`[verify-webhook] Processed deposit ${depositId} (${deposit.status})`);
    return json({ status: "success", deposit_id: depositId });
  } catch (err) {
    console.error("[verify-webhook] Exception:", err);
    return json({ code: "internal_error" }, 500);
  }
});
