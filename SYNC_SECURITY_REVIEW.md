# Syncthing security review

Reviewed: 2026-02-23

Scope: all Syncthing configuration generation, relay server, pairing flows,
systemd unit, and client API interaction in kyaraben.

Issues are ordered by importance (highest first).

---

## 1. Relay pairing codes have low entropy and no authentication

**Severity: high**

Pairing codes are 6 characters from base32 (32 symbols), giving approximately
31 bits of entropy (`32^6 = ~10^9`). The relay server has no authentication on
any endpoint. Anyone who guesses or brute-forces a valid code can retrieve the
primary device's Syncthing device ID via `GET /pair/{code}`, then submit their
own device ID via `POST /pair/{code}/response` to complete the pairing.

The rate limiter helps (10 creates/min, 30 gets/min per IP), but an attacker
with multiple IPs or patience can enumerate codes. Sessions expire after 5
minutes, which limits the window, but the combination of short codes, no
authentication, and publicly accessible endpoints makes this the most
significant risk.

~~Additionally, `DELETE /pair/{code}` has no rate limiting at all
(`relay/internal/server/server.go:48`), allowing an attacker to delete
legitimate sessions by guessing codes.~~ (fixed: DELETE is now rate-limited)

**Syncthing default comparison:** Syncthing itself uses 56-character device IDs
(224 bits of entropy) and requires mutual acceptance for pairing. Kyaraben's
relay bypasses this by trading security for UX convenience.

**Suggestions:**
- Add authentication or a shared secret to the relay (e.g., HMAC-signed
  sessions).
- Consider requiring confirmation of the paired device before completing
  pairing (the current flow auto-accepts any device that submits a response).

## 2. ~~Global discovery enabled by default~~ (fixed)

**Severity: medium**

Fixed: global discovery is now disabled by default. A toggle has been added to
the sync settings UI so users can enable it when they need cross-network sync.
Local discovery remains enabled for LAN pairing.

## 3. Syncthing listen port binds to 0.0.0.0

**Severity: medium**

`internal/sync/config.go:170-173`:
```go
ListenAddresses: []string{
    fmt.Sprintf("tcp://0.0.0.0:%d", g.syncConfig.Syncthing.ListenPort),
    fmt.Sprintf("quic://0.0.0.0:%d", g.syncConfig.Syncthing.ListenPort),
},
```

The Syncthing sync protocol listener binds to all interfaces. This is the
Syncthing default and is required for devices to connect to each other.
However, combined with global discovery, it means the Syncthing port (default
22100) is open to inbound connections from any network.

Syncthing's BEP (Block Exchange Protocol) is authenticated and encrypted using
TLS with device ID verification, so unauthorized devices cannot exchange data.
The risk is limited to exposing the port to port scanners and potential
denial-of-service.

**Syncthing default comparison:** same behavior (Syncthing defaults to
`tcp://0.0.0.0:22000`). Kyaraben uses a non-standard port (22100) which
provides marginal obscurity.

## 4. GUI communicates over plain HTTP

**Severity: medium**

`internal/sync/client.go:47`:
```go
return fmt.Sprintf("http://127.0.0.1:%d", c.config.Syncthing.GUIPort)
```

`internal/sync/status.go:119`:
```go
GUIURL: fmt.Sprintf("http://127.0.0.1:%d", c.config.Syncthing.GUIPort),
```

The Syncthing GUI API is accessed over plain HTTP. Since it's bound to
127.0.0.1, this is only exploitable by local processes. The API key is sent in
the `X-API-Key` header on every request. Any local process can sniff loopback
traffic or make requests to the GUI port.

**Syncthing default comparison:** Syncthing defaults to HTTPS for the GUI with
auto-generated TLS certificates. Kyaraben does not enable GUI TLS.

**Suggestion:** Consider enabling GUI TLS, though the risk is low since it's
localhost-only and the threat model is limited to malicious local processes.

## 5. ~~API key passed on the systemd command line~~ (fixed)

**Severity: medium**

Fixed: removed `--gui-apikey` from the systemd unit template. Syncthing reads
the API key from `config.xml` at startup. The `APIKey` field was also removed
from `UnitParams`.

## 6. Relay server CORS allows all origins

**Severity: low-medium**

`relay/internal/server/server.go:112`:
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

The relay server allows requests from any origin. Combined with the lack of
authentication, this means any website can create pairing sessions, poll for
responses, and delete sessions on behalf of the user.

**Suggestion:** Restrict CORS to known kyaraben origins, or add
authentication/CSRF protection.

## 7. ~~X-Forwarded-For uses last entry instead of first~~ (fixed)

**Severity: low-medium**

Fixed: `getClientIP` now uses the first entry in `X-Forwarded-For` (the
client's real IP) instead of the last (which was the proxy's IP).

## 8. Relay server runs plain HTTP

**Severity: low**

`relay/internal/server/server.go:76`:
```go
errCh <- s.server.ListenAndServe()
```

The relay server itself runs HTTP, not HTTPS. In production, Koyeb likely
terminates TLS, so traffic between the client and Koyeb's edge is encrypted.
The production URL uses `https://` (`internal/sync/relayclient.go:18`). This is
fine if TLS termination is handled by the hosting platform.

**Note:** Device IDs are transmitted through the relay. While device IDs are
not secrets in Syncthing's threat model (knowing an ID does not grant access
without mutual acceptance), they are sensitive identifiers.

## 9. All folders use sendreceive mode

**Severity: low**

`internal/sync/config.go:17`:
```go
FolderTypeSendReceive FolderType = "sendreceive"
```

All synced folders use bidirectional `sendreceive` mode. This means a
compromised paired device can push arbitrary files into the user's emulation
directories.

**Syncthing default comparison:** same default. Syncthing's `sendonly` or
`receiveonly` modes could limit damage from a compromised peer, but they would
break kyaraben's bidirectional sync model.

**Mitigation already in place:** `ignoreDelete` is set for ROM and BIOS folders,
preventing a malicious peer from deleting those files.

## 10. ~~NAT traversal enabled by default~~ (fixed)

**Severity: low**

Fixed: `NATEnabled` is now explicitly set to `false` in the generated config,
preventing Syncthing from opening ports on the user's router via UPnP/NAT-PMP.

## 11. ~~Crash reporting not explicitly disabled~~ (fixed)

**Severity: low**

Fixed: `CrashReportingEnabled` is now explicitly set to `false` in the
generated config, consistent with the usage reporting opt-out.

## 12. GUI URL exposed to the frontend UI

**Severity: low**

The Syncthing GUI URL (e.g., `http://127.0.0.1:8484`) is sent to the Electron
frontend via the daemon protocol (`SyncStatusResponse.GUIURL`) and rendered as
a clickable link. The API key is not sent to the frontend, so the UI can only
open the URL in a browser; it cannot make authenticated API calls to Syncthing
directly.

This is informational. If Syncthing's GUI has no password set (which is the
case here since kyaraben relies on API key auth and localhost binding), anyone
with browser access to `127.0.0.1:8484` can access the Syncthing web UI.

**Syncthing default comparison:** Syncthing shows a warning when the GUI has no
password and is accessible. Since kyaraben binds to localhost only, this is
acceptable for single-user systems.

---

## Positive security findings

Several things are done well:

- **GUI binds to 127.0.0.1 only** (`config.go:165`), preventing remote access
  to the Syncthing admin interface.
- **API key is 64 hex characters** (32 bytes of entropy, generated from
  `crypto/rand`) (`setup.go:238-243`), which is strong.
- **API key file permissions** are 0600, config directory is 0700
  (`setup.go:227-231`).
- **Auto-upgrades are disabled** (`AutoUpgradeIntervalH: 0`) and
  `STNOUPGRADE=1` is set, preventing Syncthing from downloading and running
  arbitrary binaries.
- **Usage reporting is declined** (`URAccepted: -1`).
- **No default folder** is created (`STNODEFAULTFOLDER=1`), preventing
  accidental exposure of home directory contents.
- **`autoAcceptFolders` is explicitly false** for all devices, preventing a
  paired device from creating arbitrary shared folders.
- **Config.xml is written with 0600 permissions** (`config.go:363`).
- **Rate limiting** is applied to all relay endpoints.
- **Relay sessions expire after 5 minutes** and are capped at 5 per IP and
  10,000 total.
- **`--no-browser`** flag prevents Syncthing from opening a browser window.
- **Staggered versioning** on save/state folders provides recovery from
  accidental overwrites.
