import { type ChildProcess, spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import * as path from 'node:path'
import * as readline from 'node:readline'
import { app, BrowserWindow, ipcMain, Menu, shell } from 'electron'

// Protocol types for daemon communication.
// Source of truth: internal/daemon/types.go
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

  // 4. Check APPDIR for AppImage
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

async function ensureDaemon(): Promise<void> {
  if (daemon) return

  const sidecarPath = findSidecarPath()
  console.error(`[kyaraben] Starting daemon: ${sidecarPath}`)

  const child = spawn(sidecarPath, ['daemon'], {
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
        const data = event.data as { step?: string; message?: string; output?: string }
        const msg = data?.message
        if (msg) messages.push(msg)

        // Stream progress to renderer
        if (mainWindow && !mainWindow.isDestroyed()) {
          try {
            mainWindow.webContents.send('apply:progress', {
              step: data?.step ?? 'unknown',
              message: msg ?? '',
              output: data?.output,
            })
          } catch (sendErr) {
            console.error('[kyaraben] Failed to send progress:', sendErr)
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

// IPC handlers
function setupIpcHandlers(): void {
  ipcMain.handle('get_systems', async () => {
    const event = await sendCommand({ type: 'get_systems' })
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

  ipcMain.handle('apply', async () => {
    try {
      return await applyCommand()
    } catch (err) {
      console.error('[kyaraben] Apply failed:', err)
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

  ipcMain.handle('uninstall_preview', async () => {
    const event = await sendCommand({ type: 'uninstall_preview' })
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

  ipcMain.handle('launch_emulator', (_, execLine: string) => {
    const { spawn } = require('node:child_process')
    spawn(execLine, [], {
      detached: true,
      stdio: 'ignore',
      shell: true,
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
  Menu.setApplicationMenu(null)
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
