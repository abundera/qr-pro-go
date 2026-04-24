package qrpro

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const defaultToleranceSeconds = 300

// VerifyWebhookSignature verifies a Stripe-compatible webhook signature of
// the form: "t=<unix_ts>,v1=<hex>". It recomputes HMAC-SHA256 of
// "{t}.{body}" with secret and compares in constant time. Returns nil on
// success or an error on mismatch / timestamp skew > toleranceSeconds.
//
// Pass 0 for toleranceSeconds to use the default (300s).
func VerifyWebhookSignature(signature string, body []byte, secret string, toleranceSeconds int) error {
	if toleranceSeconds <= 0 {
		toleranceSeconds = defaultToleranceSeconds
	}
	parts := map[string]string{}
	for _, p := range strings.Split(signature, ",") {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}
	t, ok1 := parts["t"]
	v1, ok2 := parts["v1"]
	if !ok1 || !ok2 {
		return errors.New("invalid signature header")
	}
	ts, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}
	now := time.Now().Unix()
	skew := now - ts
	if skew < 0 {
		skew = -skew
	}
	if skew > int64(toleranceSeconds) {
		return errors.New("signature timestamp outside tolerance")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(t))
	mac.Write([]byte("."))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	a, err := hex.DecodeString(expected)
	if err != nil {
		return fmt.Errorf("hex decode: %w", err)
	}
	b, err := hex.DecodeString(v1)
	if err != nil {
		return errors.New("invalid signature hex")
	}
	if !hmac.Equal(a, b) {
		return errors.New("signature mismatch")
	}
	return nil
}
