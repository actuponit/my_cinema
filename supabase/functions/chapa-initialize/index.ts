import { handleCors, json } from "../_shared/cors.ts";
import { parseItems, UUID_REGEX } from "../_shared/items.ts";
import { getSupabaseAdmin } from "../_shared/supabaseAdmin.ts";

const UNAVAILABLE = "Payment is unavailable right now. Please try again.";

// POST { machine_id, items: [{ item_id: <slot id>, amount }] } → { tx_ref, checkout_url }
Deno.serve(async (req: Request) => {
  const corsResponse = handleCors(req);
  if (corsResponse) return corsResponse;
  if (req.method !== "POST") return json({ code: "method_not_allowed" }, 405);

  try {
    const chapaSecretKey = Deno.env.get("CHAPA_SECRET_KEY");
    if (!chapaSecretKey) {
      console.error("[chapa-initialize] Missing CHAPA_SECRET_KEY env var");
      return json({ code: "misconfigured", message: UNAVAILABLE }, 500);
    }

    const body = await req.json().catch(() => null);
    const items = parseItems(body?.items);
    if (!items || typeof body?.machine_id !== "string" || !UUID_REGEX.test(body.machine_id)) {
      return json({ code: "invalid_items", message: "These items are not available on this machine." }, 400);
    }

    // Price and store the basket first: the webhook can only dispense what is stored here.
    // If Chapa then fails, the unpaid rows are inert (stock only moves when status flips to paid).
    const txRef = crypto.randomUUID();
    const { data: rows, error } = await getSupabaseAdmin().rpc("chapa_create_basket", {
      p_machine_id: body.machine_id,
      p_items: items,
      p_tex_ref: txRef,
    });
    if (error || !rows?.length) {
      console.error("[chapa-initialize] chapa_create_basket failed:", error);
      return json({ code: "internal_error", message: UNAVAILABLE }, 500);
    }
    const { status, total } = rows[0];
    if (status === "OUT_OF_STOCK") {
      return json({ code: "out_of_stock", message: "Some items just sold out. Please update your cart." }, 409);
    }
    if (status !== "OK") {
      return json({ code: "invalid_items", message: "These items are not available on this machine." }, 400);
    }

    const chapaRes = await fetch("https://api.chapa.co/v1/transaction/initialize", {
      method: "POST",
      headers: { "Content-Type": "application/json", "Authorization": `Bearer ${chapaSecretKey}` },
      body: JSON.stringify({
        amount: Number(total).toFixed(2),
        currency: "ETB",
        tx_ref: txRef,
        "customization[title]": "AutoSuq Vending Machine",
        "customization[description]": "Item purchase at kiosk",
      }),
    });
    const chapaData = await chapaRes.json().catch(() => null);
    const checkoutUrl = chapaData?.data?.checkout_url;
    if (!chapaRes.ok || !checkoutUrl) {
      console.error(`[chapa-initialize] Chapa error for ${txRef}:`, chapaRes.status, chapaData);
      return json({ code: "provider_error", message: "Failed to initiate payment" }, 502);
    }

    return json({ tx_ref: txRef, checkout_url: checkoutUrl });
  } catch (err) {
    console.error("[chapa-initialize] Exception:", err);
    return json({ code: "internal_error", message: UNAVAILABLE }, 500);
  }
});
