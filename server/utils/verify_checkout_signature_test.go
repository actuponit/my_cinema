package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func signVerify(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyCheckoutSignature(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	ts := strconv.FormatInt(now.Unix(), 10)
	body := []byte(`{"id":"evt_1","type":"deposit.succeeded","data":{"object":{"id":"dep_1"}}}`)
	good := signVerify("whsec_test", ts, body)

	cases := []struct {
		name         string
		body         []byte
		ts, sig, sec string
		now          time.Time
		wantErr      bool
	}{
		{"valid", body, ts, good, "whsec_test", now, false},
		{"valid millisecond timestamp", body, ts + "000", signVerify("whsec_test", ts+"000", body), "whsec_test", now, false},
		{"tampered body", []byte(`{"id":"evt_2"}`), ts, good, "whsec_test", now, true},
		{"wrong secret", body, ts, good, "whsec_other", now, true},
		{"stale timestamp", body, ts, good, "whsec_test", now.Add(6 * time.Minute), true},
		{"future timestamp", body, ts, good, "whsec_test", now.Add(-6 * time.Minute), true},
		{"missing version prefix", body, ts, good[3:], "whsec_test", now, true},
		{"malformed hex", body, ts, "v1=zz", "whsec_test", now, true},
		{"malformed timestamp", body, "abc", good, "whsec_test", now, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyCheckoutSignature(tc.body, tc.ts, tc.sig, tc.sec, tc.now)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
