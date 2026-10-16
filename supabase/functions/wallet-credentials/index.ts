import { handleCors, json } from "../_shared/cors.ts";
import { getSupabaseAdmin, getSupabaseUserClient } from "../_shared/supabaseAdmin.ts";
import * as bcrypt from "npm:bcryptjs@2.4.3";

const CREDENTIAL_REGEX = /^[A-Za-z0-9]{4,6}$/;

Deno.serve(async (req: Request) => {
  const corsResponse = handleCors(req);
  if (corsResponse) return corsResponse;

  if (req.method !== "PUT" && req.method !== "POST") {
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
      .select("user_id")
      .eq("user_id", user.id)
      .maybeSingle();

    if (walletErr || !wallet) {
      return json({ code: "not_customer", message: "Only customer accounts can set kiosk credentials" }, 403);
    }

    const body = await req.json();
    const { username, password } = body;

    if (!username || !password || !CREDENTIAL_REGEX.test(username) || !CREDENTIAL_REGEX.test(password)) {
      return json({
        code: "invalid_input",
        message: "Username and PIN/password must be 4 to 6 alphanumeric characters",
      }, 400);
    }

    // Check if username is already taken by someone else
    const { data: existing, error: existErr } = await admin
      .from("machine_credentials")
      .select("user_id")
      .eq("username", username)
      .maybeSingle();

    if (existErr) {
      console.error("[wallet-credentials] DB check error:", existErr);
    }

    if (existing && existing.user_id !== user.id) {
      return json({ code: "username_taken", message: "This username is already in use. Please choose another." }, 409);
    }

    const salt = await bcrypt.genSalt(10);
    const passwordHash = await bcrypt.hash(password, salt);

    const { error: upsertErr } = await admin.from("machine_credentials").upsert({
      user_id: user.id,
      username,
      password_hash: passwordHash,
      failed_attempts: 0,
      locked_until: null,
      updated_at: new Date().toISOString(),
    }, {
      onConflict: "user_id",
    });

    if (upsertErr) {
      if (upsertErr.code === "23505" || upsertErr.message?.includes("unique")) {
        return json({ code: "username_taken", message: "This username is already in use" }, 409);
      }
      console.error("[wallet-credentials] DB upsert error:", upsertErr);
      return json({ code: "internal_error", message: "Failed to update machine credentials" }, 500);
    }

    return json({ username }, 200);
  } catch (err: any) {
    console.error("[wallet-credentials] Exception:", err);
    return json({ code: "internal_error", message: err.message || "Internal error" }, 500);
  }
});
