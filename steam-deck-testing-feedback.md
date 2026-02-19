# Steam Deck testing feedback

## Emulation directory field UX issues

1. Applied with default `~/Emulation` path without realizing it was the default - easy to miss

1. Folder icon behavior is confusing: clicking it copies the path to clipboard, but expected behavior would be to open a file/folder picker

1. Missing context about storage location: should indicate whether the selected path is on the SD card or internal storage

1. Missing storage space info: should show available space on each storage device to help users make informed decisions about where to place their emulation folder

Solution idea: show the Steam Deck SD card slot if detected. If picked, we'll create an Emulation dir inside it. Second: show "internal storage", and pick the computer's main drive where the user's home directory is. If they pick that, we create `~/Emulation`. Third option: select a custom directory. This opens the system file picker. Whatever the user picks results in the final path (including the `Emulation` part or whatever name, or also `Emulation-<instance>`) in the existing input component below

## Text input issues

1. Text fields don't automatically open the Steam Deck virtual keyboard - requires manually pressing STEAM + X to bring it up

## Sync issues

1. Unpaired device shows diff for many folders (particularly ROMs folders) even though no other device is connected - there should be no diff to display on an unpaired device with no sync history

1. Need option to control whether syncing is active while a game is open in Game Mode - users may want to pause sync during gameplay to avoid I/O contention or battery drain

## Sync UX

It's very hard to read whether syncing is ongoing, and what's syncing. There is a long list and the folder colors change. I think we need to go for a simpler design, piggy-backing off of the Syncthing UI, which should ideally open in a browser and not within the Electron process as a child. This needs more though however, so we might do it later.

## Pairing issues

1. (Fixed) Replaced mDNS-based pairing with syncthing's native discovery
   - Primary shows its device ID and polls pending devices
   - Secondary enters primary's device ID and waits for connection
   - Uses syncthing local discovery (UDP 21027) which must be open for sync anyway

1. Firewall blocking is silent and hard to diagnose
   - If primary's syncthing port (22100) is blocked by firewall (e.g., UFW), secondary can't connect
   - Syncthing's "pending devices" only shows devices that successfully connect at TCP/TLS level
   - If connection fails (firewall), nothing shows up in pending - both sides poll silently forever
   - Root cause took a long time to find because no error was surfaced
   - Solution: add connectivity check during pairing that tests if peer's syncthing port is reachable
     - "Cannot reach primary device - connection refused"
     - "Connection timed out - check firewall settings"
   - Could also poll `/rest/events` or `/rest/system/error` to surface syncthing connection errors

1. Connection errors (some of which are only logged at DEBUG level) should be shown to users in the pairing UI

1. Pairing UI doesn't show errors - it continues to look like it's still trying even when errors occur

## UI issues

1. Long paths in provisions cards (e.g., "Open /USERDIR/saves/gamecube/EUR") don't ellipsize and overflow the card

1. Some application icons missing: ES-DE, RetroArch, RPCS3 - other emulator icons work

1. Discovered devices list should use stable sort by order of appearance, not alphabetical sort by device ID

## Missing support

1. Impossible to play on the Steam Deck without controllers configured automatically, so this is a must
