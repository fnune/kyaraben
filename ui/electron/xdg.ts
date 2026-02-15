import { spawn } from 'node:child_process'
import * as os from 'node:os'
import * as path from 'node:path'

// On systems with Flatpak apps (like Steam Deck), xdg-open resolves .desktop files
// by searching directories in XDG_DATA_DIRS. Flatpak apps install their .desktop files
// to export directories (e.g., /var/lib/flatpak/exports/share/applications), but some
// distros also have stub .desktop files in /usr/share/applications that redirect to
// app stores. If XDG_DATA_DIRS doesn't prioritize Flatpak exports, xdg-open may find
// the stub first and open the app store instead of the actual app.
export function buildXdgDataDirs(currentXdgDataDirs: string | undefined): string {
  const flatpakDataDirs = [
    path.join(os.homedir(), '.local/share/flatpak/exports/share'),
    '/var/lib/flatpak/exports/share',
  ]
  const defaultDataDirs = ['/usr/local/share', '/usr/share']
  const parsedDirs = currentXdgDataDirs?.split(':').filter(Boolean)
  const currentDirs = parsedDirs?.length ? parsedDirs : defaultDataDirs
  const seen = new Set<string>()
  const result: string[] = []
  for (const dir of [...flatpakDataDirs, ...currentDirs]) {
    if (!seen.has(dir)) {
      seen.add(dir)
      result.push(dir)
    }
  }
  return result.join(':')
}

export function xdgOpen(target: string): void {
  const child = spawn('xdg-open', [target], {
    detached: true,
    stdio: 'ignore',
    env: { ...process.env, XDG_DATA_DIRS: buildXdgDataDirs(process.env.XDG_DATA_DIRS) },
  })
  child.unref()
}
