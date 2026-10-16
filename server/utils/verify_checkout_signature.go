package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var verifyTimestampPattern = regexp.MustCompile(`^\d{10,13}$`)

// VerifyCheckoutSignature checks a verify.et webhook: header "v1=<hex>" must equal
// HMAC-SHA256(secret, timestamp + "." + rawBody), and the timestamp must be within 5 minutes.
func VerifyCheckoutSignature(raw []byte, timestamp, signature, secret string, now time.Time) error {
	if !verifyTimestampPattern.MatchString(timestamp) {
		return errors.New("malformed timestamp")
	}
	ts, _ := strconv.ParseInt(timestamp, 10, 64)
	sent := time.Unix(ts, 0)
	if len(timestamp) == 13 {
		sent = time.UnixMilli(ts)
	}
	if age := now.Sub(sent); age > 5*time.Minute || age < -5*time.Minute {
		return errors.New("stale timestamp")
	}

	hexSig, ok := strings.CutPrefix(signature, "v1=")
	if !ok {
		return errors.New("unsupported signature version")
	}
	received, err := hex.DecodeString(hexSig)
	if err != nil {
		return errors.New("malformed signature")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(raw)
	if !hmac.Equal(mac.Sum(nil), received) {
		return errors.New("signature mismatch")
	}
	return nil
}
