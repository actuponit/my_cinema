import { handleCors, json } from "../_shared/cors.ts";
import { getSupabaseAdmin, getSupabaseUserClient } from "../_shared/supabaseAdmin.ts";

Deno.serve(async (req: Request) => {
  const corsResponse = handleCors(req);
  if (corsResponse) return corsResponse;

  if (req.method !== "POST") {
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

    const admin = getSupabaseAdmin();

    // Verify user is a customer with an active wallet
    const { data: wallet, error: walletErr } = await admin
      .from("wallets")
      .select("user_id, balance")
      .eq("user_id", user.id)
      .maybeSingle();

    if (walletErr || !wallet) {
      return json({ code: "not_customer", message: "Only customer accounts can use the wallet" }, 403);
    }

    let requestedAmount: string | null = null;
    try {
      const body = await req.json();
      if (body && body.requested_amount) {
        const amtStr = String(body.requested_amount).trim();
        const num = parseFloat(amtStr);
        if (isNaN(num) || num <= 0 || num > 10000 || !/^\d+(\.\d{1,2})?$/.test(amtStr)) {
          return json({ code: "invalid_input", message: "Invalid requested amount (must be 0.01 - 10000.00)" }, 400);
        }
        requestedAmount = amtStr;
      }
    } catch (_) {
      // Body is optional
    }

    const verifyBaseUrl = (Deno.env.get("VERIFY_API_BASE_URL") || "https://checkoutapi.verify.et").replace(/\/+$/, "");
    const verifyApiKey = Deno.env.get("VERIFY_API_KEY");
    const walletWebUrl = (Deno.env.get("WALLET_WEB_URL") || "https://wallet.autosuq.com").replace(/\/+$/, "");

    if (!verifyApiKey) {
      console.error("[wallet-topup] Missing VERIFY_API_KEY env var");
      return json({ code: "internal_error", message: "Topup service misconfigured" }, 500);
    }

    const topupId = crypto.randomUUID();
    const returnUrl = `${walletWebUrl}/topup/return?ref=${topupId}`;

    const verifyRes = await fetch(`${verifyBaseUrl}/v1/deposits`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${verifyApiKey}`,
        "Idempotency-Key": topupId,
        "VerifyCheckout-Version": "2026-06-01",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        merchant_customer_id: user.id,
        currency: "ETB",
        return_url: returnUrl,
      }),
    });

    const verifyData = await verifyRes.json();
    if (!verifyRes.ok || !verifyData?.data?.id) {
      console.error("[wallet-topup] Verify.et error:", verifyData);
      return json({ code: "provider_error", message: "Failed to initiate topup with payment provider" }, 502);
    }

    const deposit = verifyData.data;

    const { error: insertErr } = await admin.from("wallet_topups").insert({
      id: topupId,
      user_id: user.id,
      deposit_id: deposit.id,
      requested_amount: requestedAmount ? parseFloat(requestedAmount) : null,
      status: deposit.status || "awaiting_transfer",
      checkout_url: deposit.checkout_url || null,
      expires_at: deposit.expires_at || null,
    });

    if (insertErr) {
      console.error("[wallet-topup] DB insert error:", insertErr);
      return json({ code: "internal_error", message: "Failed to record topup session" }, 500);
    }

    return json({
      topup_id: topupId,
      checkout_url: deposit.checkout_url,
      expires_at: deposit.expires_at,
    }, 201);
  } catch (err: any) {
    console.error("[wallet-topup] Exception:", err);
    return json({ code: "internal_error", message: err.message || "Internal error" }, 500);
  }
});
