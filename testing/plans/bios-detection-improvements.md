# BIOS detection improvements

## Status: complete

The new model has been implemented with expanded hash data.

## Issues addressed

From testing/README.md cross-cutting issues:

1. Case sensitivity: kyaraben expects exact filename match (e.g. `scph5501.bin` not `SCPH5501.BIN`)
2. Single hash: only one MD5 accepted per file, but multiple BIOS versions work
3. "Required" semantics wrong: for regional BIOS, need "at least one of these" not "this specific one required"

## Research findings

Analyzed how EmuDeck and RetroDECK handle BIOS detection (both GPL-3, using as reference only):

### EmuDeck approach
- Hash-only checking, filename doesn't matter
- Scans ALL files in bios directory, computes MD5 of each
- Checks against hardcoded array of known hashes
- Returns true if ANY valid hash found
- Hash counts: PSX (30), PS2 (71!), Saturn (8), Dreamcast (4), DS (3)

### RetroDECK approach
- JSON database (~430KB) with structured entries
- Each entry: filename, md5, system, description, required field, optional paths
- Matches by filename, then verifies MD5
- Each BIOS variant is separate entry
- "At least one BIOS file required" as descriptive text

### Key takeaways
1. Both support MANY hash alternatives per system
2. EmuDeck ignores filename - pure hash matching
3. RetroDECK matches filename then verifies hash
4. Neither is case-sensitive on filenames
5. "Required" means "at least one of these group" not "this specific file"

## Implemented design

### New provisions model

```go
type Provision struct {
    Kind        ProvisionKind
    Filename    string   // Canonical filename (for display and lookup)
    Description string   // Short label like "USA", "ARM7", "IPL"
    Hashes      []string // Valid MD5 hashes for this file
    ImportViaUI bool
}

type ProvisionGroup struct {
    Provisions  []Provision
    MinRequired int    // 0 = optional, 1+ = at least N required
    Message     string // Shown when requirement unsatisfied
}
```

Emulators now have `ProvisionGroups []ProvisionGroup` instead of `Provisions []Provision`.

### Checking algorithm

1. For each ProvisionGroup:
   - Scan bios directory for each provision's filename (case-insensitive)
   - If file found, compute MD5 and check against all Hashes
   - Count how many provisions in the group are satisfied
   - Compare against MinRequired
2. Group is satisfied if satisfied >= MinRequired
3. Doctor reports group-level satisfaction in UI

### Benefits over old model
- Case-insensitive filename matching (already had this)
- Multiple hash alternatives per BIOS file via Hashes []string
- "At least N of these" semantics via MinRequired
- Better reporting (show group satisfaction status)
- Cleaner distinction between optional and required

## BIOS hash sources

Compile our own list using these references (credit in code comments):
- EmuDeck checkBIOS.sh (GPL-3) - hash arrays
- RetroDECK bios.json (GPL-3) - structured data with descriptions
- Libretro documentation
- Emulator wikis and documentation

## Systems to cover

Priority (have provisions today):
- [ ] PSX - many regional variants
- [ ] PS2 - many regional variants
- [ ] Saturn
- [ ] Dreamcast
- [ ] NDS

Future:
- [ ] 3DO
- [ ] Sega CD
- [ ] PC Engine CD
- [ ] Neo Geo

## Implementation status

Done:
- [x] New ProvisionGroup model in `internal/model/provision.go`
- [x] Updated checking logic in `internal/store/provision.go`
- [x] Updated doctor in `internal/doctor/doctor.go`
- [x] Updated all emulator definitions to use ProvisionGroups
- [x] Updated UI components for group-based display
- [x] Updated tests

Remaining:
- [ ] Import comprehensive hash data from EmuDeck/RetroDECK into Hashes arrays

## Credits

Hash data compiled from public sources with reference to:
- EmuDeck (https://github.com/EmuDeck/EmuDeck) - GPL-3
- RetroDECK (https://github.com/RetroDECK/RetroDECK) - GPL-3
- Libretro documentation
