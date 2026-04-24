// Minimal quickstart: create, update, analytics, webhook verification.
//
// Run: ABUNDERA_API_KEY=... go run ./examples/quickstart
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	qrpro "github.com/abundera/qr-pro-go"
)

func main() {
	apiKey := os.Getenv("ABUNDERA_API_KEY")
	if apiKey == "" {
		log.Fatal("ABUNDERA_API_KEY not set")
	}
	c, err := qrpro.New(apiKey)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	code, err := c.CreateCode(ctx, qrpro.CodeCreate{
		DestinationURL: "https://example.com/launch",
		Label:          "Quickstart example",
		Tags:           []string{"example"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created:", code.ShortURL)

	newDest := "https://example.com/launch/v2"
	updated, _ := c.UpdateCode(ctx, code.ID, qrpro.CodePatch{DestinationURL: &newDest})
	fmt.Println("updated:", updated.DestinationURL)

	stats, _ := c.GetAnalytics(ctx, code.ID, nil)
	fmt.Println("scans:", stats.TotalScans)

	secret := "whsec_example"
	body := []byte(fmt.Sprintf(`{"event":"code.scanned","code_id":"%s"}`, code.ID))
	ts := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	sig := "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	if err := qrpro.VerifyWebhookSignature(sig, body, secret, 0); err != nil {
		log.Fatal(err)
	}
	fmt.Println("webhook signature verified")

	_ = c.DeleteCode(ctx, code.ID)
}
