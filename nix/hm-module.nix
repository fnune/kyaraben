# Home-manager module for kyaraben
{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.programs.kyaraben;

  # System and emulator definitions matching the Go code
  systemEmulators = {
    snes = {
      default = "retroarch:bsnes";
      emulators = {
        "retroarch:bsnes" = {
          package = pkgs.retroarch.override { cores = with pkgs.libretro; [ bsnes ]; };
          configGenerator = "retroarch";
        };
      };
    };
    psx = {
      default = "duckstation";
      emulators = {
        duckstation = {
          package = pkgs.duckstation;
          configGenerator = "duckstation";
        };
      };
    };
    tic80 = {
      default = "tic80";
      emulators = {
        tic80 = {
          package = pkgs.tic-80;
          configGenerator = null;
        };
      };
    };
  };

  enabledSystems = filterAttrs (name: value: value.enable) cfg.systems;

  # Generate the emulator packages to install
  emulatorPackages = mapAttrsToList (sysName: sysCfg:
    let
      sysInfo = systemEmulators.${sysName};
      emuId = if sysCfg.emulator != null then sysCfg.emulator else sysInfo.default;
      emuInfo = sysInfo.emulators.${emuId};
    in
    emuInfo.package
  ) enabledSystems;

  # Generate directory structure
  userStoreDirs = [
    "${cfg.userStore}/roms"
    "${cfg.userStore}/bios"
    "${cfg.userStore}/saves"
    "${cfg.userStore}/states"
    "${cfg.userStore}/screenshots"
  ] ++ (concatLists (mapAttrsToList (sysName: _: [
    "${cfg.userStore}/roms/${sysName}"
    "${cfg.userStore}/bios/${sysName}"
    "${cfg.userStore}/saves/${sysName}"
    "${cfg.userStore}/states/${sysName}"
    "${cfg.userStore}/screenshots/${sysName}"
  ]) enabledSystems));

  # RetroArch config generation
  retroarchConfig = let
    snesEnabled = hasAttr "snes" enabledSystems;
  in
  optionalString snesEnabled ''
    # Managed by kyaraben home-manager module
    system_directory = "${cfg.userStore}/bios"
    savefile_directory = "${cfg.userStore}/saves/snes"
    savestate_directory = "${cfg.userStore}/states/snes"
    screenshot_directory = "${cfg.userStore}/screenshots/snes"
    rgui_browser_directory = "${cfg.userStore}/roms/snes"
    sort_savefiles_enable = "false"
    sort_savestates_enable = "false"
    sort_screenshots_enable = "false"
  '';

  # DuckStation config generation
  duckstationConfig = let
    psxEnabled = hasAttr "psx" enabledSystems;
  in
  optionalString psxEnabled ''
    ; Managed by kyaraben home-manager module

    [BIOS]
    SearchDirectory = ${cfg.userStore}/bios/psx

    [MemoryCards]
    Directory = ${cfg.userStore}/saves/psx

    [Folders]
    SaveStates = ${cfg.userStore}/states/psx
    Screenshots = ${cfg.userStore}/screenshots/psx

    [GameList]
    RecursivePaths = ${cfg.userStore}/roms/psx
  '';

in
{
  options.programs.kyaraben = {
    enable = mkEnableOption "kyaraben emulation manager";

    userStore = mkOption {
      type = types.str;
      default = "~/Emulation";
      example = "/home/user/Games/Emulation";
      description = "Path to the emulation directory containing ROMs, saves, etc.";
    };

    systems = mkOption {
      type = types.attrsOf (types.submodule {
        options = {
          enable = mkEnableOption "this system";
          emulator = mkOption {
            type = types.nullOr types.str;
            default = null;
            description = "Emulator to use. If null, uses the default for this system.";
          };
        };
      });
      default = {};
      example = literalExpression ''
        {
          snes.enable = true;
          psx.enable = true;
        }
      '';
      description = "Systems to enable for emulation.";
    };
  };

  config = mkIf cfg.enable {
    # Install emulator packages
    home.packages = emulatorPackages;

    # Create directory structure
    home.activation.kyarabenDirs = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      ${concatMapStringsSep "\n" (dir: ''
        $DRY_RUN_CMD mkdir -p "${dir}"
      '') userStoreDirs}
    '';

    # Generate RetroArch config if needed
    xdg.configFile."retroarch/retroarch.cfg" = mkIf (retroarchConfig != "") {
      text = retroarchConfig;
    };

    # Generate DuckStation config if needed
    xdg.configFile."duckstation/settings.ini" = mkIf (duckstationConfig != "") {
      text = duckstationConfig;
    };
  };
}
