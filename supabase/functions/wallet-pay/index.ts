import * as bcrypt from "npm:bcryptjs@2.4.3";
import { handleCors, json } from "../_shared/cors.ts";
import { parseItems, UUID_REGEX } from "../_shared/items.ts";
import { closeMQTT, connectMQTT, publishDispense } from "../_shared/mqtt.ts";
import { constantTimeEqual } from "../_shared/signatures.ts";
import { getSupabaseAdmin } from "../_shared/supabaseAdmin.ts";

// Kiosk-only. POST { machine_id, username, password, items: [{ item_id: <slot id>, amount }] }
// with header x-kiosk-secret. Business outcomes are always HTTP 200 with a status the kiosk shows.

const CREDENTIAL_REGEX = /^[A-Za-z0-9]{4,6}$/;
// Valid bcrypt hash, compared against when the username is unknown so timing doesn't leak it.
const DUMMY_HASH = "$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGDMVr5yUP1KUOYTa";

const result = (status: string, message: string, extra: Record<string, unknown> = {}) =>
  json({ status, message, ...extra });
const notCharged = () => result("ERROR", "Something went wrong. You were not charged.");

Deno.serve(async (req: Request) => {
  const corsResponse = handleCors(req);
  if (corsResponse) return corsResponse;
  if (req.method !== "POST") return json({ code: "method_not_allowed" }, 405);

  // ponytail: one shared kiosk secret (extractable from the app build); per-machine keys if a kiosk leaks it.
  const kioskSecret = Deno.env.get("KIOSK_SECRET");
  if (!kioskSecret) {
    console.error("[wallet-pay] Missing KIOSK_SECRET env var");
    return json({ code: "misconfigured" }, 500);
  }
  if (!constantTimeEqual(req.headers.get("x-kiosk-secret") ?? "", kioskSecret)) {
    return json({ code: "unauthorized" }, 401);
  }

  try {
    const body = await req.json().catch(() => null);
    const { machine_id, username, password } = body ?? {};

    // 1. Validate inputs
    const items = parseItems(body?.items);
    if (!items || typeof machine_id !== "string" || !UUID_REGEX.test(machine_id)) {
      return result("INVALID_ITEMS", "These items are not available on this machine.");
    }
    if (!CREDENTIAL_REGEX.test(String(username ?? "")) || !CREDENTIAL_REGEX.test(String(password ?? ""))) {
      return result("INVALID_CREDENTIALS", "Wrong username or password.");
    }

    const supabase = getSupabaseAdmin();

    // 2. Global throttle on wrong credentials (30/min across all kiosks)
    const { data: throttled, error: throttleErr } = await supabase.rpc("wallet_pay_throttled");
    if (throttleErr) {
      console.error("[wallet-pay] wallet_pay_throttled failed:", throttleErr);
      return notCharged();
    }
    if (throttled) return result("TRY_LATER", "Too many attempts right now. Please wait a minute.");

    const recordFailure = async () => {
      const { error } = await supabase.rpc("wallet_pay_record_failure");
      if (error) console.error("[wallet-pay] wallet_pay_record_failure failed:", error);
    };

    // 3. Credentials
    const { data: cred, error: credErr } = await supabase
      .from("machine_credentials")
      .select("user_id, password_hash, locked_until")
      .eq("username", username)
      .maybeSingle();
    if (credErr) {
      console.error("[wallet-pay] Fetch credentials failed:", credErr);
      return notCharged();
    }
    if (!cred) {
      await bcrypt.compare(password, DUMMY_HASH);
      await recordFailure();
      return result("INVALID_CREDENTIALS", "Wrong username or password.");
    }

    const lockedMs = cred.locked_until ? new Date(cred.locked_until).getTime() - Date.now() : 0;
    if (lockedMs > 0) {
      return result("LOCKED", `Too many wrong attempts. Try again in ${Math.ceil(lockedMs / 60000)} minutes.`);
    }

    if (!(await bcrypt.compare(password, cred.password_hash))) {
      const { data: failed, error: failErr } = await supabase.rpc("machine_login_failed", { p_user_id: cred.user_id });
      if (failErr) console.error("[wallet-pay] machine_login_failed failed:", failErr);
      await recordFailure();
      const nowLocked = failed?.[0]?.locked_until && new Date(failed[0].locked_until).getTime() > Date.now();
      return nowLocked
        ? result("LOCKED", "Too many wrong attempts. Try again in 15 minutes.")
        : result("INVALID_CREDENTIALS", "Wrong username or password.");
    }

    // 4. Reach the machine before charging, so an offline broker never takes money
    let mqtt;
    try {
      mqtt = await connectMQTT();
    } catch (err) {
      console.error("[wallet-pay] MQTT connect failed:", err);
      return result("ERROR", "Machine is offline. You were not charged.");
    }

    try {
      // 5. Charge atomically (balance, stock, ledger, transactions)
      const texRef = crypto.randomUUID();
      const { data: payRows, error: payErr } = await supabase.rpc("wallet_pay", {
        p_user_id: cred.user_id,
        p_machine_id: machine_id,
        p_items: items,
        p_tex_ref: texRef,
      });
      if (payErr || !payRows?.length) {
        console.error(`[wallet-pay] ERROR wallet_pay tex_ref=${texRef} user=${cred.user_id}:`, payErr);
        return notCharged();
      }

      const { status, total, balance, motor_codes } = payRows[0];
      switch (status) {
        case "SUCCESS":
          break;
        case "INSUFFICIENT_BALANCE":
          return result(status, `Not enough balance. You have ${balance} ETB, this costs ${total} ETB. Top up in the AutoSuq Wallet app.`, {
            total: String(total),
            balance: String(balance),
          });
        case "OUT_OF_STOCK":
          return result(status, "Some items just sold out. Please update your cart.");
        case "INVALID_ITEMS":
          return result(status, "These items are not available on this machine.");
        default: // NO_WALLET can't happen: a credential requires a wallet.
          console.error(`[wallet-pay] ERROR tex_ref=${texRef} user=${cred.user_id}: unexpected status ${status}`);
          return notCharged();
      }

      // 6. Dispense
      if (motor_codes) {
        try {
          await publishDispense(mqtt, motor_codes);
        } catch (err) {
          // ponytail: no automatic refund (D9); add wallet_refund(tex_ref) if dispense failures show up in logs
          console.error(`[wallet-pay] ERROR dispense failed tex_ref=${texRef} user=${cred.user_id}:`, err);
        }
      }
      return result("SUCCESS", "Payment successful. Dispensing your items.", {
        txRef: texRef,
        total: String(total),
        balance: String(balance),
      });
    } finally {
      await closeMQTT(mqtt);
    }
  } catch (err) {
    console.error("[wallet-pay] Unhandled exception:", err);
    return notCharged();
  }
});
