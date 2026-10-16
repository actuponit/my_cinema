import { handleCors, json } from "../_shared/cors.ts";
import { getSupabaseAdmin, getSupabaseUserClient } from "../_shared/supabaseAdmin.ts";

const TERMINAL_STATUSES = new Set(["succeeded", "failed", "expired", "cancelled"]);

Deno.serve(async (req: Request) => {
  const corsResponse = handleCors(req);
  if (corsResponse) return corsResponse;

  if (req.method !== "GET") {
    return json({ code: "method_not_allowed" }, 405);
  }

  try {
    const authHeader = req.headers.get("Authorization");
    if (!authHeader) {
      return json({ code: "unauthorized", message: "Missing authorization header" }, 401);
    }

    const userClient = getSupabaseUserClient(authHeader);
    const { data: { user }, error: authErr } = await userClient.auth.getUser();

    if (authErr || !user) {
      return json({ code: "unauthorized", message: "Invalid or expired session" }, 401);
    }

    const url = new URL(req.url);
    const topupId = url.searchParams.get("id");
    if (!topupId) {
      return json({ code: "invalid_input", message: "Missing topup id query parameter" }, 400);
    }

    const admin = getSupabaseAdmin();

    const { data: topup, error: topupErr } = await admin
      .from("wallet_topups")
      .select("id, user_id, deposit_id, requested_amount, credited_amount, status, checkout_url, expires_at")
      .eq("id", topupId)
      .maybeSingle();

    if (topupErr || !topup || topup.user_id !== user.id) {
      return json({ code: "not_found", message: "Topup not found" }, 404);
    }

    let currentTopup = topup;

    // Refresh from Verify.et if not yet in a terminal state
    if (!TERMINAL_STATUSES.has(topup.status)) {
      const verifyBaseUrl = (Deno.env.get("VERIFY_API_BASE_URL") || "https://checkoutapi.verify.et").replace(/\/+$/, "");
      const verifyApiKey = Deno.env.get("VERIFY_API_KEY");

      if (verifyApiKey) {
        try {
          const verifyRes = await fetch(`${verifyBaseUrl}/v1/deposits/${topup.deposit_id}`, {
            headers: {
              "Authorization": `Bearer ${verifyApiKey}`,
              "VerifyCheckout-Version": "2026-06-01",
            },
          });

          if (verifyRes.ok) {
            const verifyData = await verifyRes.json();
            const d = verifyData?.data;
            if (d && d.status) {
              const amount = d.status === "succeeded" && d.amount ? parseFloat(d.amount) : null;
              const { data: updatedRows } = await admin.rpc("wallet_apply_deposit", {
                p_deposit_id: d.id,
                p_user_id: user.id,
                p_status: d.status,
                p_amount: amount,
              });

              if (updatedRows && updatedRows.length > 0) {
                currentTopup = updatedRows[0];
              }
            }
          }
        } catch (fetchErr) {
          console.error("[wallet-topup-status] Warning: failed to refresh from verify.et:", fetchErr);
        }
      }
    }

    const { data: wallet } = await admin
      .from("wallets")
      .select("balance")
      .eq("user_id", user.id)
      .maybeSingle();

    return json({
      topup_id: currentTopup.id,
      status: currentTopup.status,
      requested_amount: currentTopup.requested_amount != null ? String(currentTopup.requested_amount) : null,
      credited_amount: currentTopup.credited_amount != null ? String(currentTopup.credited_amount) : null,
      balance: wallet ? String(wallet.balance) : "0.00",
    });
  } catch (err: any) {
    console.error("[wallet-topup-status] Exception:", err);
    return json({ code: "internal_error", message: err.message || "Internal error" }, 500);
  }
});
