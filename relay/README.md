# Kyaraben relay

A lightweight relay server that enables kyaraben devices to pair using short 6-character codes instead of 56-character Syncthing device IDs.

## How it works

```mermaid
sequenceDiagram
    participant P as Primary device
    participant R as Relay server
    participant S as Secondary device

    Note over P: User clicks "Start pairing"
    P->>R: POST /pair {deviceId: "ABC..."}
    R-->>P: {code: "X4K9M2"}
    Note over P: Display code to user

    Note over S: User enters code "X4K9M2"
    S->>R: GET /pair/X4K9M2
    R-->>S: {deviceId: "ABC..."}
    S->>R: POST /pair/X4K9M2/response {deviceId: "DEF..."}

    loop Poll for response
        P->>R: GET /pair/X4K9M2/response
        R-->>P: {ready: true, deviceId: "DEF..."}
    end

    Note over P,S: Both devices now have each other's<br/>device ID and can connect via Syncthing
```

## Running locally

```sh
go run ./cmd/relay
```

The server listens on `:8080` by default. Override with `-addr` flag or `PORT` environment variable.

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | /pair | Create pairing session, returns 6-char code |
| GET | /pair/:code | Get primary's device ID |
| POST | /pair/:code/response | Submit secondary's device ID |
| GET | /pair/:code/response | Poll for secondary's response |
| DELETE | /pair/:code | Cancel session |
| GET | /health | Health check |

Sessions expire after 5 minutes. Rate limits apply per IP.

## Deployment

The relay is deployed to Koyeb using Pulumi. Deployment runs automatically on merge to main when relay code changes.

### Manual deployment

```sh
cd pulumi
pulumi up
```

Requires `KOYEB_TOKEN` environment variable.

### CI/CD setup

The GitHub Actions workflow uses Pulumi ESC for secrets:

1. Configure OIDC trust in Pulumi Cloud:
   - Settings → Access Management → OIDC Issuers
   - Add issuer: `https://token.actions.githubusercontent.com`
   - Subject filter: `repo:fnune/kyaraben:*`

2. Create ESC environment `kyaraben/relay-deploy`:
   ```yaml
   values:
     koyeb:
       token:
         fn::secret: "your-koyeb-api-token"
     environmentVariables:
       KOYEB_TOKEN: ${koyeb.token}
   ```

3. Create Pulumi stack: `pulumi stack init prod`
