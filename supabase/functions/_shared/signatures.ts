// Webhook signature checks. Pure functions so signatures_test.ts can cover them.

export async function hmacSha256Hex(secret: string, data: string): Promise<string> {
  const enc = new TextEncoder();
  const key = await crypto.subtle.importKey("raw", enc.encode(secret), { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
  const sig = new Uint8Array(await crypto.subtle.sign("HMAC", key, enc.encode(data)));
  return Array.from(sig, (b) => b.toString(16).padStart(2, "0")).join("");
}

export function constantTimeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false;
  let mismatch = 0;
  for (let i = 0; i < a.length; i++) mismatch |= a.charCodeAt(i) ^ b.charCodeAt(i);
  return mismatch === 0;
}

// Chapa: x-chapa-signature = hex(HMAC-SHA256(secret, rawBody)).
export async function verifyChapaSignature(rawBody: string, signature: string, secret: string): Promise<boolean> {
  return constantTimeEqual(signature.toLowerCase(), await hmacSha256Hex(secret, rawBody));
}

// verify.et: VerifyCheckout-Signature = "v1=" + hex(HMAC-SHA256(secret, timestamp + "." + rawBody)),
// timestamp in seconds or milliseconds, within 5 minutes. Returns the failure reason, or null if valid.
export async function verifyCheckoutSignature(
  rawBody: string,
  timestamp: string,
  signature: string,
  secret: string,
  nowMs = Date.now(),
): Promise<string | null> {
  if (!/^\d{10,13}$/.test(timestamp)) return "malformed timestamp";
  const sentMs = timestamp.length === 13 ? Number(timestamp) : Number(timestamp) * 1000;
  if (Math.abs(nowMs - sentMs) > 5 * 60 * 1000) return "stale timestamp";
  if (!signature.startsWith("v1=")) return "unsupported signature version";
  const expected = await hmacSha256Hex(secret, `${timestamp}.${rawBody}`);
  return constantTimeEqual(signature.slice(3).toLowerCase(), expected) ? null : "signature mismatch";
}
