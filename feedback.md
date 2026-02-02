# Feedback

## No installed emulator version display (resolved)

We don't show which version of each emulator we've actually installed anywhere: not in the CLI (`kyaraben status`) nor in the electron app UI.

Fixed in 6de53ed (Show emulator version info in CLI and improve frontend UI) and ad97b7b (Fix RetroArch version lookup).

---

## Update eden to v0.1.1

eden has shipped v0.1.1, need to update our version.

---

## Web app does not show whether 'Apply' has run already or needs to run

If an installation is e.g. cancelled or just hasn't been done yet because no apply run has completed, the user has no information about this. How might we present that to them? Does the CLI do this at the moment? How might the web app do it?

---

## Output when opening AppImage on my host system

```
 ~/Development/kyaraben λ ./ui/release/Kyaraben-0.1.0-x86_64.AppImage
[302991:0201/200919.248966:ERROR:dbus/object_proxy.cc:573] Failed to call method: org.freedesktop.systemd1.Manager.StartTransientUnit: object_path= /org/freedesktop/systemd1: org.freedesktop.systemd1.UnitExists: Unit app-org.chromium.Chromium-302991.scope was already loaded or has a fragment file.
[302991:0201/200919.393848:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200919.393862:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200919.394003:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200919.394009:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200919.394018:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_management_surface.cc:63] Failed to get image description for color space.
[kyaraben] Checking: /tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/kyaraben-x86_64-unknown-linux-gnu
[kyaraben] Checking: /tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/kyaraben-x86_64-unknown-linux-gnu
[kyaraben] Checking: /tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/binaries/kyaraben-x86_64-unknown-linux-gnu
[kyaraben] Found sidecar at: /tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/binaries/kyaraben-x86_64-unknown-linux-gnu
[kyaraben] Starting daemon: /tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/binaries/kyaraben-x86_64-unknown-linux-gnu
[kyaraben] Daemon ready
Error occurred in handler for 'doctor': Error: decoding config: open /home/fausto/.config/kyaraben/config.toml: no such file or directory
    at resolve (/tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/app.asar/dist-electron/main.js:166:28)
    at Interface.<anonymous> (/tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/app.asar/dist-electron/main.js:116:17)
    at Interface.emit (node:events:508:28)
    at [_onLine] [as _onLine] (node:internal/readline/interface:465:12)
    at [_normalWrite] [as _normalWrite] (node:internal/readline/interface:647:22)
    at Socket.ondata (node:internal/readline/interface:263:23)
    at Socket.emit (node:events:508:28)
    at addChunk (node:internal/streams/readable:559:12)
    at readableAddChunkPushByteMode (node:internal/streams/readable:510:3)
    at Readable.push (node:internal/streams/readable:390:5)
Error occurred in handler for 'sync_status': Error: decoding config: open /home/fausto/.config/kyaraben/config.toml: no such file or directory
    at resolve (/tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/app.asar/dist-electron/main.js:166:28)
    at Interface.<anonymous> (/tmp/nix-shell.EHP7lV/.mount_KyarabjYeFcR/resources/app.asar/dist-electron/main.js:116:17)
    at Interface.emit (node:events:508:28)
    at [_onLine] [as _onLine] (node:internal/readline/interface:465:12)
    at [_normalWrite] [as _normalWrite] (node:internal/readline/interface:647:22)
    at Socket.ondata (node:internal/readline/interface:263:23)
    at Socket.emit (node:events:508:28)
    at addChunk (node:internal/streams/readable:559:12)
    at readableAddChunkPushByteMode (node:internal/streams/readable:510:3)
    at Readable.push (node:internal/streams/readable:390:5)
[302991:0201/200919.471136:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200919.471148:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200919.478250:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200919.478258:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200920.411162:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200920.411174:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200920.416652:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200920.416659:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
[302991:0201/200920.422821:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:296] Unable to set image transfer function.
[302991:0201/200920.422828:ERROR:ui/ozone/platform/wayland/host/wayland_wp_color_manager.cc:214] Failed to populate image description for color space {r:[0.6063, 0.3237], g:[0.2372, 0.5927], b:[0.1415, 0.0508], w:[0.3127, 0.3290]}, transfer:SRGB, matrix:RGB, range:FULL}
```

That's a bit much no? Should we be throwing errors for missing configs? I think that's an expected state of the app for a first run. Same for doctor. Then what's all that about 'image transfer function' and 'image description'?

---

## Open emulation directory button does not work if the directory does not exist

The OS throws an error and then the app handler times out after 5s. We should probably only enable that button if the directory already exists.

---

## Cannot run installed emulators

Generally I feel like `nix-portable` is probably doing something weird with the paths. I'm not sure.

### Cannot run installed duckstation

I see this:

```
 ~/Development/kyaraben λ ~/.local/state/kyaraben/bin/duckstation
/usr/bin/fusermount3: mount failed: Operation not permitted

Cannot mount AppImage, please check your FUSE setup.
You might still be able to extract the contents of this AppImage
if you run it with the --appimage-extract option.
See https://github.com/AppImage/AppImageKit/wiki/FUSE
for more information
open dir error: No such file or directory
```

---

### Cannot run installed retroarch

```
 ~/Development/kyaraben λ ~/.local/state/kyaraben/bin/retroarch

GameMode ERROR: D-Bus error: Could not call method 'QueryStatus' on 'com.feralinteractive.GameMode': The name is not activatable
```

Note that if I install retroarch with `nix-shell` and run it then it also prints that error message, but it _does_ work in that case:

```
 ~/Development/kyaraben λ nix-shell --packages retroarch
warning: Nix search path entry '/home/fausto/.nix-defexpr/channels' does not exist, ignoring
this derivation will be built:
  /nix/store/3mwqf6m36ijah3lgww535xxaqfg0vlfg-retroarch-with-cores-1.21.0.drv
these 3 paths will be fetched (164.97 MiB download, 196.02 MiB unpacked):
  /nix/store/757mya7v9rnjizz1f1j37cjq68ispgp2-declarative-retroarch.cfg
  /nix/store/1aq76q3jk5i0wgbr5d6mkx5hr7vppm98-libretro-core-info-1.22.0
  /nix/store/8cns3szl0vlcij037g5cjsnn9mdlifya-retroarch-assets-1.22.0-unstable-2025-11-10
copying path '/nix/store/1aq76q3jk5i0wgbr5d6mkx5hr7vppm98-libretro-core-info-1.22.0' from 'https://cache.nixos.org'...
copying path '/nix/store/8cns3szl0vlcij037g5cjsnn9mdlifya-retroarch-assets-1.22.0-unstable-2025-11-10' from 'https://cache.nixos.org'...
copying path '/nix/store/757mya7v9rnjizz1f1j37cjq68ispgp2-declarative-retroarch.cfg' from 'https://cache.nixos.org'...
building '/nix/store/3mwqf6m36ijah3lgww535xxaqfg0vlfg-retroarch-with-cores-1.21.0.drv'...
Kyaraben development environment
Go version: go version go1.25.5 linux/amd64
 ~/Development/kyaraben λ which retroarch
/nix/store/12zwidf0dbw12n15lc786wkj279h28f2-retroarch-with-cores-1.21.0/bin/retroarch
 ~/Development/kyaraben λ retroarch
GameMode ERROR: D-Bus error: Could not call method 'QueryStatus' on 'com.feralinteractive.GameMode': The name is not activatable
```

### Cannot run installed eden

This one is very strange...

```
 ~/Development/kyaraben λ cat ~/.local/share/applications/eden.desktop
─────┬────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
     │ File: /home/fausto/.local/share/applications/eden.desktop
─────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
   1 │ [Desktop Entry]
   2 │ Type=Application
   3 │ Name=Eden
   4 │ GenericName=Nintendo Switch Emulator
   5 │ Exec=/tmp/nix-shell.dmfnVT/TestGenerateDesktopFiles3625339760/001/kyaraben/bin/eden %f
   6 │ Icon=eden
   7 │ Categories=Game;Emulator;
─────┴────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 ~/Development/kyaraben λ /tmp/nix-shell.dmfnVT/TestGenerateDesktopFiles3625339760/001/kyaraben/bin/eden
zsh: no such file or directory: /tmp/nix-shell.dmfnVT/TestGenerateDesktopFiles3625339760/001/kyaraben/bin/eden
```

## CLI output on apply

It interleaves 'Installing emulators' all the time. This is probably not necessary in the CLI?

```
  /nix/store/2v59zbb6i773c1b0mwwdqhw3nghfm6d9-curl-8.6.0
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
  /nix/store/vcxkk3l25hlix62aivl2prfhpczrvymp-expat-2.6.2
  /nix/store/c2yb135iv4maadia5f760b3xhbh6jh61-gcc-13.2.0-lib
  /nix/store/a2fqd5k8ymb2cadb5awp5b90a5m8acln-gcc-13.2.0-libgcc
  /nix/store/qjb0xvlkzdbrxxasmajw93phx3wh8vq6-gettext-0.21.1
  /nix/store/bv5kc5im1r2b99x7vni3cwphnzgmd8ck-git-minimal-2.44.0
  /nix/store/ddwyrxif62r8n6xclvskjyy6szdhvj60-glibc-2.39-5
  /nix/store/n32h02pn52pk38m0p00wh7f2ri8xrh8h-gmp-with-cxx-6.3.0
  /nix/store/avqi5nnx7qydr078ssgifc2hgzqipqgx-gnugrep-3.11
  /nix/store/237dff1igc3v09p9r23a37yw8dr04bv6-gnused-4.9
  /nix/store/d5wvzzmqx3dkmp36r1vzz69gak6x5bkx-keyutils-1.6.3-lib
  /nix/store/s32cldbh9pfzd9z82izi12mdlrw0yf8q-libidn2-2.3.7
  /nix/store/li8plf2qixrlrlny7qhw5ylgq01h3z7q-libkrb5-1.21.2
  /nix/store/kci440kzdmyi7b1axs3w6nlmswk3881j-libpsl-0.21.5
  /nix/store/gqrbbhxahk4mayblnc0sfpksgph197bb-libssh2-1.11.0
  /nix/store/7n0mbqydcipkpbxm24fab066lxk68aqk-libunistring-1.1
  /nix/store/dvwbmkf5gqwly9ysp6sld4c6iwmqijm3-nghttp2-1.60.0-lib
  /nix/store/p25ghy7y53lyc834xnw5mrhfq096wa4x-openssl-3.0.13
  /nix/store/5sqdrc4jpr4vjiiqycyw8q4v3zchpdka-pcre2-10.43
  /nix/store/7ararm009ri4jrg1rgz2n1bhdzhln5s2-publicsuffix-list-0-unstable-2024-01-07
  /nix/store/rxganm4ibf31qngal3j3psp20mak37yy-xgcc-13.2.0-libgcc
  /nix/store/zph9xw0drmq3rl2ik5slg0n2frw9lw5m-zlib-1.3.1
  /nix/store/ss6gh67xv5jw6jh0l7dwmyx9823wvb60-zstd-1.5.5
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/a2fqd5k8ymb2cadb5awp5b90a5m8acln-gcc-13.2.0-libgcc' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/rxganm4ibf31qngal3j3psp20mak37yy-xgcc-13.2.0-libgcc' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/7n0mbqydcipkpbxm24fab066lxk68aqk-libunistring-1.1' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/7ararm009ri4jrg1rgz2n1bhdzhln5s2-publicsuffix-list-0-unstable-2024-01-07' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/s32cldbh9pfzd9z82izi12mdlrw0yf8q-libidn2-2.3.7' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/ddwyrxif62r8n6xclvskjyy6szdhvj60-glibc-2.39-5' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/7i768y3q30fx0qgajwdp9m7bzj13xiyg-attr-2.5.2' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/s279kslfwqlnx79df9ygj9f758x3skda-brotli-1.1.0-lib' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/a1s263pmsci9zykm5xcdf7x9rv26w6d5-bash-5.2p26' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/vcxkk3l25hlix62aivl2prfhpczrvymp-expat-2.6.2' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/kci440kzdmyi7b1axs3w6nlmswk3881j-libpsl-0.21.5' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/p25ghy7y53lyc834xnw5mrhfq096wa4x-openssl-3.0.13' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/237dff1igc3v09p9r23a37yw8dr04bv6-gnused-4.9' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/d5wvzzmqx3dkmp36r1vzz69gak6x5bkx-keyutils-1.6.3-lib' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/dvwbmkf5gqwly9ysp6sld4c6iwmqijm3-nghttp2-1.60.0-lib' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/c2yb135iv4maadia5f760b3xhbh6jh61-gcc-13.2.0-lib' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/5sqdrc4jpr4vjiiqycyw8q4v3zchpdka-pcre2-10.43' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/zph9xw0drmq3rl2ik5slg0n2frw9lw5m-zlib-1.3.1' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/7d4h6k9rmh0gy39s024ggkbkasspsb4n-acl-2.3.2' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/li8plf2qixrlrlny7qhw5ylgq01h3z7q-libkrb5-1.21.2' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/avqi5nnx7qydr078ssgifc2hgzqipqgx-gnugrep-3.11' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/gqrbbhxahk4mayblnc0sfpksgph197bb-libssh2-1.11.0' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/qjb0xvlkzdbrxxasmajw93phx3wh8vq6-gettext-0.21.1' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/n32h02pn52pk38m0p00wh7f2ri8xrh8h-gmp-with-cxx-6.3.0' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/ss6gh67xv5jw6jh0l7dwmyx9823wvb60-zstd-1.5.5' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/ifzwv2xqwdnv1gz87rxkizi67py5p3vj-coreutils-9.4' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/2v59zbb6i773c1b0mwwdqhw3nghfm6d9-curl-8.6.0' from 'https://cache.nixos.org'...
Installing emulators (this may take a while on first run)...
copying path '/nix/store/bv5kc5im1r2b99x7vni3cwphnzgmd8ck-git-minimal-2.44.0' from 'https://cache.nixos.org'...
```

## Uninstall script can't remove nix store

```
 ~/Development/kyaraben λ ./ui/binaries/kyaraben-x86_64-unknown-linux-gnu uninstall
This will remove:

  /home/fausto/.local/state/kyaraben (nix store, manifest, state)

  Desktop entries:
    /home/fausto/.local/share/applications/com.libretro.RetroArch.desktop
    /home/fausto/.local/share/applications/tic80.desktop
    /home/fausto/.local/share/applications/com.libretro.RetroArch.desktop
    /home/fausto/.local/share/applications/tic80.desktop
    /home/fausto/.local/share/applications/duckstation.desktop
    /home/fausto/.local/share/applications/eden.desktop

  Icons:
    /home/fausto/.local/share/icons/hicolor/scalable/apps/duckstation.svg
    /home/fausto/.local/share/icons/hicolor/scalable/apps/eden.svg

  Managed config files:
    /home/fausto/.config/retroarch/retroarch.cfg
    /home/fausto/.config/retroarch/config/mgba_libretro/mgba_libretro.cfg
    /home/fausto/.config/duckstation/settings.ini
    /home/fausto/.config/retroarch/config/bsnes_libretro/bsnes_libretro.cfg
    /home/fausto/.config/eden/qt-config.ini

This will NOT remove:
  ~/Emulation (your ROMs, saves, BIOS)
  /home/fausto/.config/kyaraben (your kyaraben config)

Proceed? [y/N] y

Removing kyaraben files...
  Removed: /home/fausto/.config/retroarch/retroarch.cfg
  Removed: /home/fausto/.config/retroarch/config/mgba_libretro/mgba_libretro.cfg
  Removed: /home/fausto/.config/duckstation/settings.ini
  Removed: /home/fausto/.config/retroarch/config/bsnes_libretro/bsnes_libretro.cfg
  Removed: /home/fausto/.config/eden/qt-config.ini
  Removed: /home/fausto/.local/share/applications/com.libretro.RetroArch.desktop
  Removed: /home/fausto/.local/share/applications/tic80.desktop
  Removed: /home/fausto/.local/share/applications/duckstation.desktop
  Removed: /home/fausto/.local/share/applications/eden.desktop
  Removed: /home/fausto/.local/share/icons/hicolor/scalable/apps/duckstation.svg
  Removed: /home/fausto/.local/share/icons/hicolor/scalable/apps/eden.svg
  Warning: could not remove /home/fausto/.local/state/kyaraben: unlinkat /home/fausto/.local/state/kyaraben/build/nix/.nix-portable/nix/store/a2fqd5k8ymb2cadb5awp5b90a5m8acln-gcc-13.2.0-libgcc/lib/libgcc_s.so: permission denied

Done. Kyaraben files have been removed.

To fully uninstall, also remove:
  /home/fausto/.config/kyaraben (your config)
  The kyaraben binary itself
 ~/Development/kyaraben λ ls -la ~/.local/state/kyaraben/
drwxr-xr-x - fausto  1 Feb 20:27 󱧼 build
```

---

## DuckStation onboarding wizard

DuckStation runs an onboarding wizard on first launch that wants to create a config file and set up autoupdates. This is not good for kyaraben's managed experience. We need a better default config that prevents this wizard from appearing.

---

## Missing icons for nixpkgs emulators (retroarch, tic80)

Emulators from nixpkgs (NixStoreDesktop entries) don't have icons showing up in the desktop environment. The desktop files reference icons like `Icon=com.libretro.RetroArch` and `Icon=tic80`, but these icons are in the nix store's share/icons directory which isn't in the system's icon search path.

Options:
1. Copy icons from nix store to ~/.local/share/icons during desktop file generation
2. Embed SVG icons for these emulators like we do for eden/duckstation
3. Add nix store's share directory to XDG_DATA_DIRS (may have other side effects)

---

## TIC-80 runs in CLI mode instead of GUI

TIC-80 falls back to console/CLI mode instead of showing a GUI window. Running it shows:

```
 TIC-80 tiny computer
 version 1.2.3042-dev ()
 https://tic80.com (C) 2017-2025

 hello! type help for help

>
```

This is likely due to nix-portable's environment not properly exposing the graphics stack (SDL2/Wayland/X11) to the binary. The SDL libraries are found but display initialization seems to fail silently.

Possible causes:
- nix-portable sandbox interfering with display server access
- Missing environment variables for SDL/Wayland
- Need to use a different package or wrapper approach for GUI apps from nixpkgs

---

## RetroArch fails with EGL/graphics initialization errors

RetroArch installed via nix-portable fails to launch entirely because it cannot access the host's graphics stack. Running it produces:

```
libGL error: pci id for fd 28: 1002:67df, driver unknown
libGL error: failed to load driver: unknown
libGL error: pci id for fd 28: 1002:67df, driver unknown
libGL error: failed to load driver: unknown
[INFO] [X11/GLX] Suspend screensaver.
[INFO] [X11] X/Y exts: w: 3840 h: 2160 x: 0 y: 0.
[INFO] [X11] Monitor axis: 3840 2160.
[INFO] Trying to find shared context driver for: gl
[INFO] Found a shared context driver: gl
[INFO] [X11] XDisplay selected: 0x7fc9a8001440
[INFO] [GLX] glXSwapInterval = 0
[ERROR] [EGL] Could not get EGL display.
[INFO] [EGL]: Quitting EGL.
[ERROR] [X11/EGL]: EGL context creation failed.
[ERROR] [Video]: Driver gl failed, falling back to next driver: sdl2
[INFO] [X11] X/Y exts: w: 3840 h: 2160 x: 0 y: 0.
[INFO] [X11] Monitor axis: 3840 2160.
[INFO] [SDL2]: SDL anthropic.2.28.4 x11 Video Context.
[INFO] [SDL] Quitting SDL.
[ERROR] [SDL2]: Failed to initialize window.
[ERROR] [Video]: Cannot initialize video driver "sdl2".
[ERROR] [Video]: Cannot find video driver "sdl2".
[FATAL] Failed to init any video driver. Aborting ...
[INFO] Application has been terminated!
```

Root cause: nix-portable's sandboxed environment can't access the host system's EGL/Mesa/Vulkan drivers. The libraries exist but can't find the actual GPU drivers.

This is a fundamental limitation of running GUI applications through nix-portable. Possible solutions:
1. Use AppImages for graphics-heavy applications (like we do for duckstation/eden)
2. Find or create AppImages for retroarch/tic80
3. Investigate nix-portable configuration for exposing host graphics
4. Consider proot or other sandboxing approaches that allow GPU access
5. Hybrid approach: download RetroArch AppImage directly but still fetch cores via nix (cores are just shared libraries, don't need GPU access)

---

## Type generator is fragile and incomplete

The `scripts/generate-types` tool only generates command/event types from Go to TypeScript. Adding a new system or emulator currently requires manual updates in many places:

- `internal/model/system.go` and `internal/model/emulator.go` (Go constants)
- `ui/src/types/daemon.ts` (SystemID and EmulatorID union types)
- `ui/src/types/ui.ts` (SYSTEM_MANUFACTURERS mapping)
- `ui/src/components/SystemIcon/SystemIcon.tsx` (SYSTEM_LABELS mapping)

All of these are stringly-typed and easy to get out of sync. The registry already contains most of this information (system definitions know their names, emulators know which systems they support), but it's not being leveraged.

Ideally, the Go registry would be the single source of truth, and we'd generate:
- TypeScript types for SystemID and EmulatorID
- Manufacturer mappings (could be added to system definitions)
- Icon/label mappings (could be added to system definitions)
- Any other UI metadata

This would make adding new systems/emulators a single-file change instead of a multi-file hunt.

Note: CONTRIBUTING.md documents the Go-side steps for adding emulators but is missing all the UI-side updates, making it an incomplete guide.

---

## Removing emulator support breaks existing configs

If we remove support for an emulator (e.g., TIC-80) but the user's config still references it, the apply fails hard:

```
generating flake: unknown emulator: tic80
```

Instead, we should:
1. Skip unknown emulators during flake generation (don't fail)
2. Show a warning in the UI: "Emulator 'tic80' is no longer supported and will be removed from your config"
3. Optionally auto-clean the config to remove stale entries

This makes upgrades smoother when we deprecate emulators.

## Version tracking for emulators

It would be useful to have a programmatic, non-LLM way to check if any emulator supported by kyaraben has new versions available that we're not offering. This would:

1. Automatically query GitHub releases (or other sources) for each emulator
2. Compare against versions.toml to identify:
   - New versions available that we don't have
   - Old versions we're tracking that could be removed
3. Provide a quick way to bump versions and remove old versions

This should be a development script (e.g., `scripts/check-versions.go` or a make target) or a CI job that runs periodically and creates PRs when updates are available.

---

## False "update on apply" messages for already-installed versions (resolved)

Both the CLI (`kyaraben status`) and the Electron frontend display "update to vX.Y.Z on apply" for emulators that already have that version installed. Example CLI output:

```
Managed emulators:
  Eden                 latest (update to v0.1.1 on apply)
  Flycast              latest (update to 2.6 on apply)
  mGBA                 latest (update to 0.10.5 on apply)
  RetroArch (bsnes)    latest
  RPCS3                latest (update to 0.0.18-12817-fff0c96b on apply)
  Azahar               latest (update to 2124.3 on apply)
  DuckStation          latest (update to v0.1-10655 on apply)
  melonDS              latest (update to 1.1 on apply)
  PCSX2                latest (update to v2.6.3 on apply)
  PPSSPP               latest (update to v1.19.3 on apply)
  Vita3K               latest (update to 3912 on apply)
  Cemu                 latest (update to 2.4 on apply)
  Dolphin              latest (update to 2512 on apply)
```

All these emulators already have the indicated versions installed. The comparison logic is likely comparing the wrong values, or the installed version isn't being read correctly.

Additionally, displaying "latest" is not useful. The output should show the actual installed version and whether it's pinned. Since most users will use auto-updating, pinned should be the special case:
- `v0.1.1` for auto-updating emulators (the common case)
- `v0.1.1 (pinned)` for pinned versions

So the output should look more like:
```
Managed emulators:
  Eden                 v0.1.1
  Flycast              2.6
  mGBA                 0.10.5 (pinned)
  ...
```

Fixed in 6de53ed (Show emulator version info in CLI and improve frontend UI) and ad97b7b (Fix RetroArch version lookup using package name instead of emulator ID).

---

## Standardize version mode vocabulary across UI and CLI

The UI and CLI should use consistent vocabulary for version modes:
- Show the actual installed version by default (no "latest" label)
- Only annotate pinned versions with "(pinned)" since auto-updating is the common case
- This convention should be applied everywhere: CLI status output, Electron app emulator list, any version-related messaging

---

## Dolphin prompts for autoupdates on launch

Dolphin tries to enable its built-in autoupdate mechanism when it first launches. Since Kyaraben manages emulator updates, we need to preconfigure Dolphin to disable this prompt/feature.

---

## Reconsider flake generations and lock file approach

Each `kyaraben apply` creates a new flake generation directory and a new lock file:

```
warning: creating lock file '/home/fausto/.local/state/kyaraben/build/flake/generations/2026-02-02T07-55-32/flake.lock'
```

Questions to consider:
- Are we keeping these lock files around? It seems like we're just throwing them away.
- The warning is a bit ugly for users who don't care about nix internals.
- Should we reconsider the generations system for flake versions?

---

## UI feels frozen during nix flake lock creation

In both the UI and CLI, there's a gap between when "Installing emulators" appears and when nix actually starts producing output (the `warning: creating lock file` message). During this time the interface appears frozen with no indication that work is happening.

The delay occurs because nix is evaluating the flake and creating/updating the lock file before it starts downloading or building anything. This can take several seconds on first run.

We should emit a log message before this phase starts, something like "Resolving package versions..." or "Preparing nix environment...". This would thread through naturally to the UI via the existing log streaming.

---

## Emulator health check for `kyaraben doctor`

Similar to the version checking script, it would be useful to have a way to quickly verify that installed emulators are actually working. This could be part of `kyaraben doctor`.

Challenges:
- Not all emulators support `--version` or `--help` flags (e.g., flycast, eden just launch the GUI)
- Spawning a bunch of GUI windows on the user's machine is disruptive

Possible approaches:
1. For emulators with `--version`/`--help` flags, use those (retroarch, duckstation, pcsx2, ppsspp, mgba, melonds, azahar, dolphin, cemu, vita3k, rpcs3)
2. For others, check that the binary exists and is executable
3. Could also verify the wrapper script points to a valid store path
4. Run `file` on the binary to ensure it's a valid ELF executable
5. For AppImages, could use `--appimage-extract --appimage-offset` to verify integrity without mounting

The goal is to catch issues like:
- Broken wrapper scripts (the melonds/vita3k `nix shell` issue)
- Missing binaries
- Corrupted downloads
- FUSE/permissions issues

This would complement the existing version check script by verifying runtime health rather than just version currency.
