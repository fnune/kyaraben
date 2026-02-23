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

Additionally, `DELETE /pair/{code}` has no rate limiting at all
(`relay/internal/server/server.go:48`), allowing an attacker to delete
legitimate sessions by guessing codes.

**Syncthing default comparison:** Syncthing itself uses 56-character device IDs
(224 bits of entropy) and requires mutual acceptance for pairing. Kyaraben's
relay bypasses this by trading security for UX convenience.

**Suggestions:**
- Increase code length (8-10 characters) or add a secondary verification step.
- Add authentication or a shared secret to the relay (e.g., HMAC-signed
  sessions).
- Add rate limiting to `DELETE /pair/{code}`.
- Consider requiring the primary to confirm the secondary's device ID before
  completing pairing (the current flow auto-accepts any device that submits a
  response).

## 2. Global discovery and local announce are enabled

**Severity: medium**

`internal/sync/config.go:174-175`:
```go
GlobalAnnounceEnabled: true,
LocalAnnounceEnabled:  true,
```

These match Syncthing defaults but expose the device to the global discovery
network and local network broadcast discovery. In kyaraben's context, this means:

- The device's Syncthing ID, IP address, and listen port are announced to
  Syncthing's global discovery servers, making it discoverable by anyone who
  knows the device ID.
- Local announce broadcasts on the LAN can reveal the Syncthing instance to
  other devices on the same network.

This is the expected behavior for Syncthing's peer-to-peer model, but it means
kyaraben devices are visible on the public discovery network.

**Suggestions:**
- Consider whether global announce is needed given that the relay server already
  handles device discovery for pairing. Disabling it would reduce the device's
  public footprint.
- At minimum, document this behavior so users understand their device is
  announced publicly.

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

## 5. API key passed on the systemd command line

**Severity: medium**

`internal/sync/systemd.go:22`:
```
ExecStart={{.BinaryPath}} serve --no-browser --config={{.ConfigDir}} --data={{.DataDir}} --gui-address=127.0.0.1:{{.GUIPort}} --gui-apikey={{.APIKey}}
```

The API key appears in the systemd unit file and in the process command line.
This means any user who can read `/proc/<pid>/cmdline` or the systemd unit file
can obtain the API key. On a single-user system (the typical kyaraben use case)
this is low risk, but on shared systems it leaks the API key to other users.

**Syncthing default comparison:** Syncthing stores the API key only in
`config.xml` (which kyaraben does set with 0600 permissions). Passing it on the
command line is an additional exposure vector.

**Suggestion:** Remove `--gui-apikey` from the command line. Syncthing reads
the API key from `config.xml` at startup; the command-line flag is redundant
since kyaraben already writes it to `config.xml`.

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

## 7. X-Forwarded-For uses last entry instead of first

**Severity: low-medium**

`relay/internal/server/ratelimit.go:103-106`:
```go
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
    ips := strings.Split(xff, ",")
    return strings.TrimSpace(ips[len(ips)-1])
}
```

The rate limiter uses the *last* IP in `X-Forwarded-For`. The standard
convention is that the first IP is the client's real IP and subsequent entries
are added by each proxy. Using the last entry means behind a reverse proxy
(like Koyeb's), the rate limiter may use the proxy's IP rather than the
client's IP, potentially rate-limiting all clients together or not at all,
depending on the proxy chain.

More critically, if the relay is directly exposed (no proxy), an attacker can
forge `X-Forwarded-For` headers to bypass rate limiting entirely by appending a
different IP.

**Suggestion:** If deployed behind a trusted proxy, use the first IP (or better,
`X-Real-IP` set by the trusted proxy). If not behind a proxy, ignore
`X-Forwarded-For` entirely and use `RemoteAddr`.

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

## 10. NAT traversal enabled by default

**Severity: low**

`natEnabled` is not explicitly set in the generated config, so Syncthing falls
back to its default of `true`. This enables UPnP and NAT-PMP port mapping,
which can automatically open ports on the user's router without their knowledge.

**Syncthing default comparison:** same default (`natEnabled: true`). This is
intentional for peer-to-peer connectivity but may surprise users who expect
kyaraben not to modify their router configuration.

**Suggestion:** Either explicitly set `natEnabled: false` (since the relay
handles connectivity) or document that UPnP port mapping is enabled.

## 11. Crash reporting not explicitly disabled

**Severity: low**

`crashReportingEnabled` is not set in the generated config, defaulting to
`true`. This sends crash data (including public IP, Syncthing version, and
build hostname) to Syncthing's crash reporting server. This is inconsistent
with the explicit opt-out of usage reporting (`urAccepted: -1`).

**Suggestion:** Add `crashReportingEnabled: false` to the options for
consistency with the usage reporting opt-out.

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
- **Rate limiting** is applied to relay endpoints (except DELETE).
- **Relay sessions expire after 5 minutes** and are capped at 5 per IP and
  10,000 total.
- **`--no-browser`** flag prevents Syncthing from opening a browser window.
- **Staggered versioning** on save/state folders provides recovery from
  accidental overwrites.
