# qr-pro-go

Official Go SDK for [Abundera QR Pro](https://pro.qr.abundera.ai).

Dynamic QR codes that you actually own. Scan analytics. Webhooks. Privacy-safe by default.

## Install

```bash
go get github.com/abundera/qr-pro-go
```

Requires Go 1.21+.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    qrpro "github.com/abundera/qr-pro-go"
)

func main() {
    c, err := qrpro.New(os.Getenv("ABUNDERA_API_KEY"))
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()

    code, err := c.CreateCode(ctx, qrpro.CodeCreate{
        DestinationURL: "https://example.com/launch",
        Label:          "Spring launch poster",
        Tags:           []string{"print", "q2"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(code.ShortURL)

    // Later: change where it points
    newDest := "https://example.com/launch/v2"
    _, _ = c.UpdateCode(ctx, code.ID, qrpro.CodePatch{DestinationURL: &newDest})

    // Pull analytics
    stats, _ := c.GetAnalytics(ctx, code.ID, &qrpro.AnalyticsParams{From: "2026-04-01"})
    fmt.Println(stats.TotalScans)
}
```

## Webhook verification

```go
import qrpro "github.com/abundera/qr-pro-go"

func handler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    sig := r.Header.Get("X-Abundera-Signature")

    if err := qrpro.VerifyWebhookSignature(sig, body, os.Getenv("ABUNDERA_WEBHOOK_SECRET"), 0); err != nil {
        http.Error(w, "bad signature", http.StatusBadRequest)
        return
    }
    // safe to parse and act on body
}
```

## Configuration

```go
c, _ := qrpro.New(apiKey,
    qrpro.WithBaseURL("https://pro.qr.abundera.ai"),
    qrpro.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
    qrpro.WithMaxRetries(5),
    qrpro.WithUserAgent("myapp/1.2.3"),
)
```

## Error handling

Non-2xx responses (after retries) return a `*qrpro.Error`:

```go
code, err := c.CreateCode(ctx, qrpro.CodeCreate{DestinationURL: "not a url"})
if err != nil {
    var apiErr *qrpro.Error
    if errors.As(err, &apiErr) {
        log.Printf("status=%d code=%s msg=%s req=%s", apiErr.Status, apiErr.Code, apiErr.Message, apiErr.RequestID)
    }
}
```

## Coverage

Supported surfaces: codes (CRUD + slug check), analytics (JSON + CSV), groups, webhooks, user. See [API docs](https://pro.qr.abundera.ai/docs/) for the full OpenAPI 3.1 spec.

## License

MIT © Abundera, Inc.
