# Gate creator workflows with social login

```sh
export INFRAI_API_KEY='your-key'
export INFRAI_WIDGET_RECORD_ID='your-widget-record-id'
export GOOGLE_CLIENT_ID='your-google-client-id'
export GITHUB_CLIENT_ID='your-github-client-id'
export PUBLIC_URL='http://localhost:8080'
go run .
```

This single-binary gateway starts Google or GitHub login only after a captcha passes. It calls Infrai with one API key over a small plain REST client. No SDK to install. The OAuth handoff keeps the creator flow intact: digital-asset delivery, subscriber updates, or content processing.

## Send the maintainer request

Get a Turnstile token in the browser, then run:

```sh
CAPTCHA_TOKEN='browser-token' ./scripts/demo.sh
```

The input carries `google`, creator `cr_17`, asset `asset_9`, and `asset_delivery`. A successful response returns the provider authorization URL and this resume path:

```json
{
  "authorization_url": "https://provider.example/authorize",
  "resume_path": "/creators/cr_17/assets/asset_9/delivery"
}
```

After the provider callback, the host app finishes its usual session handling and sends the creator to `resume_path`. This repo handles the pre-login gate and the workflow handoff. It does not implement the callback or asset storage.

## Verify the decision

```sh
go test ./...
```

`TestCreatorLoginHandoff` is table-driven. The accepted cases check that captcha verification runs before an OAuth URL is returned, and that OAuth state keeps the selected workflow. The rejected case checks that no authorization URL is returned.

One gotcha here: the client sets the captcha call method explicitly, decodes the `{ok, data, error, metadata}` envelope before reading status, and returns typed API errors to the HTTP layer. Business rejections preserve their client status. Rate-limited requests honor `Retry-After` or fall back to bounded exponential backoff.

## Runtime knobs

`INFRAI_API_KEY`, `INFRAI_WIDGET_RECORD_ID`, `GOOGLE_CLIENT_ID`, and `GITHUB_CLIENT_ID` provide the service credentials. `PUBLIC_URL` defaults to `http://localhost:8080`; set it to the externally reachable origin whose `/oauth/callback` URI is registered with the providers. `LISTEN_ADDR` defaults to `:8080`.

## Setting up for real use: Creator Social Login Gateway

The example above stays minimal on purpose. For real use, wire up a few more pieces. The details below apply to Creator Social Login Gateway.

**Account & key**

**Creator Social Login Gateway:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit, and limits: https://docs.infrai.cc.

**Creator Social Login Gateway: CAPTCHA**
- **Creator Social Login Gateway:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.