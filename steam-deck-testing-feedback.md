# Steam Deck testing feedback

## Emulation directory field UX issues

1. Applied with default `~/Emulation` path without realizing it was the default - easy to miss

2. Folder icon behavior is confusing: clicking it copies the path to clipboard, but expected behavior would be to open a file/folder picker

3. Missing context about storage location: should indicate whether the selected path is on the SD card or internal storage

4. Missing storage space info: should show available space on each storage device to help users make informed decisions about where to place their emulation folder

## Text input issues

1. Text fields don't automatically open the Steam Deck virtual keyboard - requires manually pressing STEAM + X to bring it up

## Sync issues

1. Unpaired device shows diff for many folders (particularly ROMs folders) even though no other device is connected - there should be no diff to display on an unpaired device with no sync history

## Pairing issues

1. (Fixed) Replaced mDNS-based pairing with syncthing's native discovery
   - Primary shows its device ID and polls pending devices
   - Secondary enters primary's device ID and waits for connection
   - Uses syncthing local discovery (UDP 21027) which must be open for sync anyway

2. Firewall blocking is silent and hard to diagnose
   - If primary's syncthing port (22100) is blocked by firewall (e.g., UFW), secondary can't connect
   - Syncthing's "pending devices" only shows devices that successfully connect at TCP/TLS level
   - If connection fails (firewall), nothing shows up in pending - both sides poll silently forever
   - Root cause took a long time to find because no error was surfaced
   - Solution: add connectivity check during pairing that tests if peer's syncthing port is reachable
     - "Cannot reach primary device - connection refused"
     - "Connection timed out - check firewall settings"
   - Could also poll `/rest/events` or `/rest/system/error` to surface syncthing connection errors

3. Connection errors are only logged at DEBUG level but should be shown to users in the pairing UI

4. Pairing UI doesn't show errors - it continues to look like it's still trying even when errors occur

5. (Needs UI update) Secondary flow now requires device ID input instead of pairing code

6. Primary auto-accept loop doesn't stop after successfully pairing a device - keeps polling

## UI issues

1. Long paths in provisions cards (e.g., "Open /USERDIR/saves/gamecube/EUR") don't ellipsize and overflow the card
