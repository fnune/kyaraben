import { type ChildProcess, spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import * as os from 'node:os'
import * as path from 'node:path'
import * as readline from 'node:readline'
import { app, BrowserWindow, ipcMain, shell } from 'electron'
import type { InvokeChannel } from './channels'
import { checkForUpdates, downloadUpdate } from './updater'

function getInstanceName(): string | null {
  for (let i = 0; i < process.argv.length; i++) {
    const arg = process.argv[i]
    if (arg.startsWith('--instance=')) {
      return arg.split('=')[1]
    }
    if (arg === '--instance' && i + 1 < process.argv.length) {
      return process.argv[i + 1]
    }
  }
  return null
}

const instanceName = getInstanceName()
const kyarabenDirName = instanceName ? `kyaraben-${instanceName}` : 'kyaraben'

// Set userData to XDG state directory instead of config
const stateDir = process.env.XDG_STATE_HOME || path.join(os.homedir(), '.local', 'state')
const kyarabenStateDir = path.join(stateDir, kyarabenDirName)
app.setPath('userData', path.join(kyarabenStateDir, 'ui'))

// Protocol types for daemon communication.
// Source of truth: internal/daemon/protocol.go (ProgressEvent)
// Generated types: src/types/daemon.gen.ts
// Keep these in sync when modifying the protocol.
interface DaemonCommand {
  type: string
  id?: string
  data?: unknown
}

interface DaemonEvent {
  type: string
  id?: string
  data: unknown
}

interface ProgressEvent {
  step: string
  message: string
  output?: string
  buildPhase?: string
  packageName?: string
  progressPercent?: number
  bytesDownloaded?: number
  bytesTotal?: number
  bytesPerSecond?: number
  logPosition?: number
}

// Daemon process handle
let daemon: {
  process: ChildProcess
  rl: readline.Interface
  pending: Map<string, { resolve: (value: DaemonEvent) => void; reject: (err: Error) => void }>
} | null = null

function getSidecarName(): string {
  const arch = process.arch === 'x64' ? 'x86_64' : process.arch === 'arm64' ? 'aarch64' : 'unknown'

  if (process.platform === 'linux') {
    return `kyaraben-${arch}-unknown-linux-gnu`
  }
  if (process.platform === 'darwin') {
    return `kyaraben-${arch}-apple-darwin`
  }
  if (process.platform === 'win32') {
    return 'kyaraben-x86_64-pc-windows-msvc.exe'
  }
  return 'kyaraben'
}

function findSidecarPath(): string {
  const sidecarName = getSidecarName()
  const searchPaths: string[] = []

  // 1. Next to the current executable (AppImage/installed)
  const exeDir = path.dirname(app.getPath('exe'))
  searchPaths.push(path.join(exeDir, sidecarName))

  // 2. In resources (packaged app)
  searchPaths.push(path.join(process.resourcesPath, sidecarName))
  searchPaths.push(path.join(process.resourcesPath, 'binaries', sidecarName))

  // 3. Development/testing: check relative to app directory
  // app.getAppPath() returns dist-electron/ when running main.js directly
  const appPath = app.getAppPath()
  searchPaths.push(path.join(appPath, '..', 'binaries', sidecarName))
  searchPaths.push(path.join(appPath, 'binaries', sidecarName))

  // 4. Check relative to __dirname (dist-electron/)
  searchPaths.push(path.join(__dirname, '..', 'binaries', sidecarName))

  // 5. Check APPDIR for AppImage
  const appdir = process.env.APPDIR
  if (appdir) {
    searchPaths.push(path.join(appdir, 'usr', 'bin', sidecarName))
    searchPaths.push(path.join(appdir, sidecarName))
  }

  const fs = require('node:fs')
  for (const searchPath of searchPaths) {
    console.error(`[kyaraben] Checking: ${searchPath}`)
    if (fs.existsSync(searchPath)) {
      console.error(`[kyaraben] Found sidecar at: ${searchPath}`)
      return searchPath
    }
  }

  throw new Error(`Sidecar binary '${sidecarName}' not found. Searched: ${searchPaths.join(', ')}`)
}

function spawnInTerminal(command: string): { success: boolean; error?: string; command?: string } {
  const { spawn: spawnProcess, execSync } = require('node:child_process')

  function trySpawn(cmd: string, args: string[]): boolean {
    try {
      execSync(`which ${cmd}`, { stdio: 'ignore' })
      spawnProcess(cmd, args, { detached: true, stdio: 'ignore' }).unref()
      return true
    } catch {
      return false
    }
  }

  const userTerminal = process.env.TERMINAL
  if (userTerminal && trySpawn(userTerminal, ['-e', 'sh', '-c', command])) {
    console.error(`[kyaraben] Opened terminal: ${userTerminal} (from $TERMINAL): ${command}`)
    return { success: true }
  }

  if (trySpawn('xdg-terminal-exec', [command])) {
    console.error(`[kyaraben] Opened terminal: xdg-terminal-exec (freedesktop default): ${command}`)
    return { success: true }
  }

  const desktop = process.env.XDG_CURRENT_DESKTOP?.toLowerCase() ?? ''

  if (desktop.includes('gnome')) {
    try {
      const gsettingsOutput = execSync(
        'gsettings get org.gnome.desktop.default-applications.terminal exec',
        { encoding: 'utf8' },
      ).trim()
      const gnomeTerminal = gsettingsOutput.replace(/^'|'$/g, '')
      if (gnomeTerminal && trySpawn(gnomeTerminal, ['--', 'sh', '-c', command])) {
        console.error(
          `[kyaraben] Opened terminal: ${gnomeTerminal} (GNOME default from gsettings): ${command}`,
        )
        return { success: true }
      }
    } catch {
      /* gsettings unavailable */
    }
  }

  if (desktop.includes('kde') && trySpawn('konsole', ['-e', 'sh', '-c', command])) {
    console.error(`[kyaraben] Opened terminal: konsole (KDE default): ${command}`)
    return { success: true }
  }

  const fallbacks = [
    { cmd: 'x-terminal-emulator', args: ['-e', command] },
    { cmd: 'kitty', args: ['sh', '-c', command] },
    { cmd: 'alacritty', args: ['-e', 'sh', '-c', command] },
    { cmd: 'wezterm', args: ['start', '--', 'sh', '-c', command] },
    { cmd: 'foot', args: ['sh', '-c', command] },
    { cmd: 'gnome-terminal', args: ['--', 'sh', '-c', command] },
    { cmd: 'konsole', args: ['-e', 'sh', '-c', command] },
    { cmd: 'tilix', args: ['-e', command] },
    { cmd: 'terminator', args: ['-e', command] },
    { cmd: 'xfce4-terminal', args: ['-e', command] },
    { cmd: 'mate-terminal', args: ['-e', command] },
    { cmd: 'lxterminal', args: ['-e', command] },
    { cmd: 'sakura', args: ['-e', command] },
    { cmd: 'terminology', args: ['-e', command] },
    { cmd: 'urxvt', args: ['-e', 'sh', '-c', command] },
    { cmd: 'st', args: ['-e', 'sh', '-c', command] },
    { cmd: 'xterm', args: ['-e', command] },
  ]

  for (const term of fallbacks) {
    if (trySpawn(term.cmd, term.args)) {
      console.error(`[kyaraben] Opened terminal: ${term.cmd} (fallback): ${command}`)
      return { success: true }
    }
  }

  console.error('[kyaraben] No terminal emulator found')
  return { success: false, error: 'No terminal emulator found', command }
}

async function ensureDaemon(): Promise<void> {
  if (daemon) return

  const sidecarPath = findSidecarPath()
  const daemonArgs = ['daemon']
  if (instanceName) {
    daemonArgs.push('--instance', instanceName)
  }
  console.error(`[kyaraben] Starting daemon: ${sidecarPath} ${daemonArgs.join(' ')}`)

  const child = spawn(sidecarPath, daemonArgs, {
    stdio: ['pipe', 'pipe', 'inherit'],
    env: process.env,
  })

  if (!child.stdout) {
    throw new Error('Failed to get daemon stdout')
  }

  const rl = readline.createInterface({
    input: child.stdout,
    crlfDelay: Number.POSITIVE_INFINITY,
  })

  daemon = {
    process: child,
    rl,
    pending: new Map(),
  }

  // Handle daemon events
  rl.on('line', (line: string) => {
    try {
      const event: DaemonEvent = JSON.parse(line)

      if (event.type === 'ready') {
        const handler = daemon?.pending.get('__ready__')
        if (handler) {
          daemon?.pending.delete('__ready__')
          handler.resolve(event)
        }
        return
      }

      if (event.id) {
        const handler = daemon?.pending.get(event.id)
        if (handler) {
          handler.resolve(event)
        }
      }
    } catch (err) {
      console.error(`[kyaraben] Failed to parse daemon event: ${err}`)
    }
  })

  child.on('exit', (code) => {
    console.error(`[kyaraben] Daemon exited with code ${code}`)
    daemon = null
  })

  // Wait for ready event
  const currentDaemon = daemon
  await new Promise<DaemonEvent>((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('Daemon startup timeout')), 10000)
    currentDaemon.pending.set('__ready__', {
      resolve: (event) => {
        clearTimeout(timeout)
        resolve(event)
      },
      reject,
    })
  })

  console.error('[kyaraben] Daemon ready')
}

async function sendCommand(cmd: DaemonCommand): Promise<DaemonEvent> {
  await ensureDaemon()

  if (!daemon || !daemon.process.stdin) {
    throw new Error('Daemon not running')
  }

  const currentDaemon = daemon
  const stdin = daemon.process.stdin
  const requestId = randomUUID()
  const commandWithId = { ...cmd, id: requestId }
  const json = `${JSON.stringify(commandWithId)}\n`
  stdin.write(json)

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      currentDaemon.pending.delete(requestId)
      reject(new Error('Command timeout'))
    }, 600000) // 10 minute timeout for long operations

    currentDaemon.pending.set(requestId, {
      resolve: (event) => {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        if (event.type === 'error') {
          reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
        } else {
          resolve(event)
        }
      },
      reject: (err) => {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        reject(err)
      },
    })
  })
}

// Apply is special because it streams multiple progress events
async function applyCommand(): Promise<{ messages: string[]; cancelled: boolean }> {
  await ensureDaemon()

  if (!daemon || !daemon.process.stdin) {
    throw new Error('Daemon not running')
  }

  const currentDaemon = daemon
  const stdin = daemon.process.stdin
  const messages: string[] = []
  const requestId = randomUUID()
  const json = `${JSON.stringify({ type: 'apply', id: requestId })}\n`
  stdin.write(json)

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(
      () => {
        currentDaemon.pending.delete(requestId)
        reject(new Error('Apply timeout'))
      },
      15 * 60 * 1000,
    )

    const handleEvent = (event: DaemonEvent) => {
      if (event.type === 'progress') {
        const data = event.data as ProgressEvent | undefined
        const msg = data?.message
        if (msg) messages.push(msg)

        // Stream progress to renderer
        if (mainWindow && !mainWindow.isDestroyed()) {
          try {
            mainWindow.webContents.send('apply:progress', {
              step: data?.step ?? 'unknown',
              message: msg ?? '',
              output: data?.output,
              buildPhase: data?.buildPhase,
              packageName: data?.packageName,
              progressPercent: data?.progressPercent,
              bytesDownloaded: data?.bytesDownloaded,
              bytesTotal: data?.bytesTotal,
              bytesPerSecond: data?.bytesPerSecond,
              logPosition: data?.logPosition,
            })
          } catch (sendErr) {
            console.error(
              `[kyaraben] Failed to send progress: ${sendErr instanceof Error ? sendErr.message : String(sendErr)}`,
            )
          }
        }
      } else if (event.type === 'result') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        messages.push('Apply completed successfully')
        resolve({ messages, cancelled: false })
      } else if (event.type === 'cancelled') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        resolve({ messages, cancelled: true })
      } else if (event.type === 'error') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
      }
    }

    currentDaemon.pending.set(requestId, { resolve: handleEvent, reject })
  })
}

// Pairing streams progress events like apply
async function pairingCommand(): Promise<{
  success: boolean
  peerDeviceId?: string
  peerName?: string
}> {
  await ensureDaemon()

  if (!daemon || !daemon.process.stdin) {
    throw new Error('Daemon not running')
  }

  const currentDaemon = daemon
  const stdin = daemon.process.stdin
  const requestId = randomUUID()
  const json = `${JSON.stringify({ type: 'sync_start_pairing', id: requestId })}\n`
  stdin.write(json)

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(
      () => {
        currentDaemon.pending.delete(requestId)
        reject(new Error('Pairing timeout'))
      },
      6 * 60 * 1000,
    )

    const handleEvent = (event: DaemonEvent) => {
      if (event.type === 'progress') {
        const data = event.data as { message?: string } | undefined
        if (mainWindow && !mainWindow.isDestroyed()) {
          try {
            mainWindow.webContents.send('pairing:progress', {
              message: data?.message ?? '',
            })
          } catch (sendErr) {
            console.error(
              `[sync] Failed to send pairing progress: ${sendErr instanceof Error ? sendErr.message : String(sendErr)}`,
            )
          }
        }
      } else if (event.type === 'result') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        const data = event.data as { success?: boolean; peerDeviceId?: string; peerName?: string }
        resolve({
          success: data?.success ?? true,
          peerDeviceId: data?.peerDeviceId,
          peerName: data?.peerName,
        })
      } else if (event.type === 'cancelled') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        resolve({ success: false })
      } else if (event.type === 'error') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
      }
    }

    currentDaemon.pending.set(requestId, { resolve: handleEvent, reject })
  })
}

async function joinPrimaryCommand(code: string): Promise<{
  success: boolean
  peerDeviceId?: string
  peerName?: string
}> {
  await ensureDaemon()

  if (!daemon || !daemon.process.stdin) {
    throw new Error('Daemon not running')
  }

  const currentDaemon = daemon
  const stdin = daemon.process.stdin
  const requestId = randomUUID()
  const json = `${JSON.stringify({ type: 'sync_join_primary', id: requestId, data: { code, pairingAddr: '' } })}\n`
  stdin.write(json)

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(
      () => {
        currentDaemon.pending.delete(requestId)
        reject(new Error('Join primary timeout'))
      },
      6 * 60 * 1000,
    )

    const handleEvent = (event: DaemonEvent) => {
      if (event.type === 'progress') {
        const data = event.data as { message?: string } | undefined
        if (mainWindow && !mainWindow.isDestroyed()) {
          try {
            mainWindow.webContents.send('pairing:progress', {
              message: data?.message ?? '',
            })
          } catch (sendErr) {
            console.error(
              `[sync] Failed to send join progress: ${sendErr instanceof Error ? sendErr.message : String(sendErr)}`,
            )
          }
        }
      } else if (event.type === 'result') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        const data = event.data as { success?: boolean; peerDeviceId?: string; peerName?: string }
        resolve({
          success: data?.success ?? true,
          peerDeviceId: data?.peerDeviceId,
          peerName: data?.peerName,
        })
      } else if (event.type === 'cancelled') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        resolve({ success: false })
      } else if (event.type === 'error') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
      }
    }

    currentDaemon.pending.set(requestId, { resolve: handleEvent, reject })
  })
}

async function syncEnableCommand(mode: string): Promise<{ success: boolean }> {
  await ensureDaemon()

  if (!daemon || !daemon.process.stdin) {
    throw new Error('Daemon not running')
  }

  const currentDaemon = daemon
  const stdin = daemon.process.stdin
  const requestId = randomUUID()
  const json = `${JSON.stringify({ type: 'sync_enable', id: requestId, data: { mode } })}\n`
  stdin.write(json)

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(
      () => {
        currentDaemon.pending.delete(requestId)
        reject(new Error('Sync enable timeout'))
      },
      10 * 60 * 1000,
    )

    const handleEvent = (event: DaemonEvent) => {
      if (event.type === 'progress') {
        const data = event.data as
          | { phase?: string; message?: string; percent?: number }
          | undefined
        if (mainWindow && !mainWindow.isDestroyed()) {
          try {
            mainWindow.webContents.send('sync_enable:progress', {
              phase: data?.phase ?? '',
              message: data?.message ?? '',
              percent: data?.percent ?? 0,
            })
          } catch (sendErr) {
            console.error(
              `[sync] Failed to send enable progress: ${sendErr instanceof Error ? sendErr.message : String(sendErr)}`,
            )
          }
        }
      } else if (event.type === 'result') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        resolve({ success: true })
      } else if (event.type === 'error') {
        clearTimeout(timeout)
        currentDaemon.pending.delete(requestId)
        reject(new Error((event.data as { error?: string })?.error || 'Unknown error'))
      }
    }

    currentDaemon.pending.set(requestId, { resolve: handleEvent, reject })
  })
}

// IPC handlers
function setupIpcHandlers(): void {
  ipcMain.handle('get_systems', async () => {
    const event = await sendCommand({ type: 'get_systems' })
    return event.data
  })

  ipcMain.handle('get_frontends', async () => {
    const event = await sendCommand({ type: 'get_frontends' })
    return event.data
  })

  ipcMain.handle('get_config', async () => {
    const event = await sendCommand({ type: 'get_config' })
    return event.data
  })

  ipcMain.handle(
    'set_config',
    async (_, data: { userStore: string; systems: Record<string, string> }) => {
      const event = await sendCommand({ type: 'set_config', data })
      return event.data
    },
  )

  ipcMain.handle('status', async () => {
    const event = await sendCommand({ type: 'status' })
    return event.data
  })

  ipcMain.handle('doctor', async () => {
    const event = await sendCommand({ type: 'doctor' })
    return event.data
  })

  ipcMain.handle('preflight', async () => {
    const event = await sendCommand({ type: 'preflight' })
    return event.data
  })

  ipcMain.handle('apply', async () => {
    try {
      return await applyCommand()
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      console.error(`[kyaraben] Apply failed: ${msg}`)
      throw err
    }
  })

  ipcMain.handle('cancel_apply', async () => {
    const event = await sendCommand({ type: 'cancel_apply' })
    return event.data
  })

  ipcMain.handle('get_install_status', async () => {
    const event = await sendCommand({ type: 'install_status' })
    return event.data
  })

  ipcMain.handle('install_app', async () => {
    const appImagePath = process.env.APPIMAGE || ''
    const sidecarPath = findSidecarPath()
    const event = await sendCommand({
      type: 'install_kyaraben',
      data: { appImagePath, sidecarPath },
    })
    return event.data
  })

  ipcMain.handle('sync_status', async () => {
    const event = await sendCommand({ type: 'sync_status' })
    return event.data
  })

  ipcMain.handle('sync_add_device', async (_, data: { deviceId: string; name?: string }) => {
    const event = await sendCommand({ type: 'sync_add_device', data })
    return event.data
  })

  ipcMain.handle('sync_remove_device', async (_, data: { deviceId: string }) => {
    const event = await sendCommand({ type: 'sync_remove_device', data })
    return event.data
  })

  ipcMain.handle('sync_start_pairing', async () => {
    try {
      return await pairingCommand()
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      console.error(`[sync] Pairing failed: ${msg}`)
      const cleanError = new Error(msg)
      cleanError.stack = undefined
      throw cleanError
    }
  })

  ipcMain.handle('sync_join_primary', async (_, data: { code: string; pairingAddr: string }) => {
    return await joinPrimaryCommand(data.code)
  })

  ipcMain.handle('sync_cancel_pairing', async () => {
    const event = await sendCommand({ type: 'sync_cancel_pairing' })
    return event.data
  })

  ipcMain.handle('sync_revert_folder', async (_, data: { folderId: string }) => {
    const event = await sendCommand({ type: 'sync_revert_folder', data })
    return event.data
  })

  ipcMain.handle('sync_local_changes', async (_, data: { folderId: string }) => {
    const event = await sendCommand({ type: 'sync_local_changes', data })
    return event.data
  })

  ipcMain.handle('sync_pause', async () => {
    console.log('[sync] Pause requested')
    try {
      const event = await sendCommand({ type: 'sync_pause' })
      console.log('[sync] Pause result:', JSON.stringify(event.data))
      return event.data
    } catch (err) {
      console.error(`[sync] Pause failed: ${err instanceof Error ? err.message : String(err)}`)
      throw err
    }
  })

  ipcMain.handle('sync_resume', async () => {
    console.log('[sync] Resume requested')
    try {
      const event = await sendCommand({ type: 'sync_resume' })
      console.log('[sync] Resume result:', JSON.stringify(event.data))
      return event.data
    } catch (err) {
      console.error(`[sync] Resume failed: ${err instanceof Error ? err.message : String(err)}`)
      throw err
    }
  })

  ipcMain.handle('sync_pending', async () => {
    const event = await sendCommand({ type: 'sync_pending' })
    return event.data
  })

  ipcMain.handle('sync_enable', async (_, data: { mode: string }) => {
    try {
      return await syncEnableCommand(data.mode)
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      console.error(`[sync] Enable failed: ${msg}`)
      throw err
    }
  })

  ipcMain.handle('sync_reset', async () => {
    const event = await sendCommand({ type: 'sync_reset' })
    return event.data
  })

  ipcMain.handle('uninstall_preview', async () => {
    const event = await sendCommand({ type: 'uninstall_preview' })
    return event.data
  })

  ipcMain.handle('refresh_icon_caches', async () => {
    const event = await sendCommand({ type: 'refresh_icon_caches' })
    return event.data
  })

  ipcMain.handle('open_path', (_, pathToOpen: string) => {
    const expandedPath = pathToOpen.startsWith('~')
      ? pathToOpen.replace('~', app.getPath('home'))
      : pathToOpen
    shell.openPath(expandedPath)
    return ''
  })

  ipcMain.handle('path_exists', async (_, pathToCheck: string) => {
    const fs = require('node:fs')
    const expandedPath = pathToCheck.startsWith('~')
      ? pathToCheck.replace('~', app.getPath('home'))
      : pathToCheck
    return fs.existsSync(expandedPath)
  })

  ipcMain.handle('read_file', async (_, filePath: string) => {
    const fs = require('node:fs')
    const expandedPath = filePath.startsWith('~')
      ? filePath.replace('~', app.getPath('home'))
      : filePath
    return fs.readFileSync(expandedPath, 'utf-8')
  })

  ipcMain.handle('launch_emulator', (_, execLine: string) => {
    const { spawn } = require('node:child_process')
    spawn(execLine, [], {
      detached: true,
      stdio: 'ignore',
      shell: true,
    }).unref()
    return { success: true }
  })

  ipcMain.handle('open_log_tail', (_event, position?: number) => {
    const logPath = path.join(kyarabenStateDir, 'kyaraben.log')
    const cmd =
      position !== undefined ? `tail -c +${position} -f "${logPath}"` : `tail -f "${logPath}"`
    return spawnInTerminal(cmd)
  })

  ipcMain.handle('launch_cli_uninstall', () => {
    const { spawn: spawnProcess } = require('node:child_process')
    const sidecarPath = findSidecarPath()
    const pid = process.pid
    spawnProcess(sidecarPath, ['uninstall', '--force', `--wait-pid=${pid}`, '--notify'], {
      detached: true,
      stdio: 'ignore',
      env: process.env,
    }).unref()
    return { success: true }
  })

  ipcMain.handle('get_bug_report_info', async () => {
    const os = require('node:os')
    const fs = require('node:fs')
    const stateDir = path.join(os.homedir(), '.local', 'state', 'kyaraben')

    const stateInfo = {
      exists: false,
      manifestExists: false,
      flakeExists: false,
      brokenSymlinks: [] as string[],
    }

    if (fs.existsSync(stateDir)) {
      stateInfo.exists = true
      stateInfo.manifestExists = fs.existsSync(path.join(stateDir, 'build', 'manifest.json'))
      stateInfo.flakeExists = fs.existsSync(path.join(stateDir, 'build', 'flake'))

      for (const subdir of ['bin', 'desktop', 'icons']) {
        const dir = path.join(stateDir, subdir)
        if (fs.existsSync(dir)) {
          for (const entry of fs.readdirSync(dir)) {
            const fullPath = path.join(dir, entry)
            try {
              fs.statSync(fullPath)
            } catch {
              stateInfo.brokenSymlinks.push(path.join(subdir, entry))
            }
          }
        }
      }
    }

    return {
      appVersion: app.getVersion(),
      platform: process.platform,
      arch: process.arch,
      osRelease: os.release(),
      stateDir: stateInfo,
    }
  })

  ipcMain.handle('check_for_updates', async () => {
    return await checkForUpdates()
  })

  ipcMain.handle('download_update', async (_, url: string) => {
    try {
      const tempPath = await downloadUpdate(url, (percent) => {
        if (mainWindow && !mainWindow.isDestroyed()) {
          mainWindow.webContents.send('update:progress', { percent })
        }
      })
      return { success: true, path: tempPath }
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : String(error) }
    }
  })

  ipcMain.handle('apply_update', async (_, tempPath: string) => {
    const sidecarPath = findSidecarPath()

    const event = await sendCommand({
      type: 'install_kyaraben',
      data: { appImagePath: tempPath, sidecarPath },
    })

    if (event.type === 'error') {
      return {
        success: false,
        error: (event.data as { error?: string })?.error || 'Install failed',
      }
    }

    app.relaunch()
    app.exit(0)
    return { success: true }
  })

  // Compile-time check: ensure all INVOKE_CHANNELS have handlers registered above.
  // If this errors, add the missing handler to this function.
  const _dependencies = [
    'get_systems',
    'get_frontends',
    'get_config',
    'set_config',
    'status',
    'doctor',
    'preflight',
    'apply',
    'cancel_apply',
    'sync_status',
    'sync_remove_device',
    'sync_start_pairing',
    'sync_join_primary',
    'sync_cancel_pairing',
    'sync_pending',
    'sync_enable',
    'sync_revert_folder',
    'sync_local_changes',
    'sync_reset',
    'uninstall_preview',
    'refresh_icon_caches',
    'get_install_status',
    'install_app',
    'open_path',
    'path_exists',
    'read_file',
    'get_bug_report_info',
    'launch_emulator',
    'open_log_tail',
    'launch_cli_uninstall',
    'check_for_updates',
    'download_update',
    'apply_update',
  ] as const
  type HandledChannels = (typeof _dependencies)[number]
  type _AssertAllChannelsHandled = InvokeChannel extends HandledChannels ? true : never
  const _check: _AssertAllChannelsHandled = true
  void _check
}

// Window creation
let mainWindow: BrowserWindow | null = null

function createWindow(): void {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  })

  // Load the app
  if (process.env.VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'))
  }

  // Log renderer crashes
  mainWindow.webContents.on('render-process-gone', (_event, details) => {
    console.error('[kyaraben] Renderer process gone:', details)
  })

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

// Global error handlers
process.on('uncaughtException', (error) => {
  console.error('[kyaraben] Uncaught exception:', error)
})

process.on('unhandledRejection', (reason) => {
  console.error('[kyaraben] Unhandled rejection:', reason)
})

// App lifecycle
app.whenReady().then(() => {
  setupIpcHandlers()
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  // Kill daemon on exit
  if (daemon) {
    daemon.process.kill()
    daemon = null
  }

  if (process.platform !== 'darwin') {
    app.quit()
  }
})
