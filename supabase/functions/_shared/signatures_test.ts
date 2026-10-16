import { hmacSha256Hex, verifyChapaSignature, verifyCheckoutSignature } from "./signatures.ts";

function assert(cond: boolean, msg: string) {
  if (!cond) throw new Error(msg);
}

Deno.test("chapa signature", async () => {
  const body = '{"tx_ref":"abc","status":"success"}';
  const sig = await hmacSha256Hex("s3cret", body);
  assert(await verifyChapaSignature(body, sig, "s3cret"), "valid signature rejected");
  assert(await verifyChapaSignature(body, sig.toUpperCase(), "s3cret"), "uppercase hex rejected");
  assert(!(await verifyChapaSignature(body + " ", sig, "s3cret")), "tampered body accepted");
  assert(!(await verifyChapaSignature(body, sig, "other")), "wrong secret accepted");
  assert(!(await verifyChapaSignature(body, "", "s3cret")), "empty signature accepted");
});

Deno.test("verify.et signature", async () => {
  const body = '{"type":"deposit.succeeded","data":{"object":{"id":"dep_1"}}}';
  const now = 1_760_000_000_000;
  const ts = String(now / 1000);
  const sig = "v1=" + (await hmacSha256Hex("whsec", `${ts}.${body}`));

  assert((await verifyCheckoutSignature(body, ts, sig, "whsec", now)) === null, "valid signature rejected");
  const tsMs = String(now);
  const sigMs = "v1=" + (await hmacSha256Hex("whsec", `${tsMs}.${body}`));
  assert((await verifyCheckoutSignature(body, tsMs, sigMs, "whsec", now)) === null, "ms timestamp rejected");

  assert((await verifyCheckoutSignature(body, ts, sig, "whsec", now + 6 * 60 * 1000)) === "stale timestamp", "stale accepted");
  assert((await verifyCheckoutSignature(body, "12ab", sig, "whsec", now)) === "malformed timestamp", "bad ts accepted");
  assert((await verifyCheckoutSignature(body, ts, sig.slice(3), "whsec", now)) === "unsupported signature version", "no v1= accepted");
  assert((await verifyCheckoutSignature(body + "x", ts, sig, "whsec", now)) === "signature mismatch", "tampered body accepted");
});
