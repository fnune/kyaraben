"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
const node_child_process_1 = require("node:child_process");
const path = __importStar(require("node:path"));
const readline = __importStar(require("node:readline"));
const electron_1 = require("electron");
// Daemon process handle
let daemon = null;
function getSidecarName() {
    const arch = process.arch === 'x64' ? 'x86_64' : process.arch === 'arm64' ? 'aarch64' : 'unknown';
    if (process.platform === 'linux') {
        return `kyaraben-${arch}-unknown-linux-gnu`;
    }
    if (process.platform === 'darwin') {
        return `kyaraben-${arch}-apple-darwin`;
    }
    if (process.platform === 'win32') {
        return 'kyaraben-x86_64-pc-windows-msvc.exe';
    }
    return 'kyaraben';
}
function findSidecarPath() {
    const sidecarName = getSidecarName();
    const searchPaths = [];
    // 1. Next to the current executable (AppImage/installed)
    const exeDir = path.dirname(electron_1.app.getPath('exe'));
    searchPaths.push(path.join(exeDir, sidecarName));
    // 2. In resources (packaged app)
    searchPaths.push(path.join(process.resourcesPath, sidecarName));
    searchPaths.push(path.join(process.resourcesPath, 'binaries', sidecarName));
    // 3. Development: check relative to app directory
    const appPath = electron_1.app.getAppPath();
    searchPaths.push(path.join(appPath, 'src-tauri', 'binaries', sidecarName));
    searchPaths.push(path.join(appPath, '..', 'binaries', sidecarName));
    // 4. Check APPDIR for AppImage
    const appdir = process.env.APPDIR;
    if (appdir) {
        searchPaths.push(path.join(appdir, 'usr', 'bin', sidecarName));
        searchPaths.push(path.join(appdir, sidecarName));
    }
    const fs = require('node:fs');
    for (const searchPath of searchPaths) {
        console.error(`[kyaraben] Checking: ${searchPath}`);
        if (fs.existsSync(searchPath)) {
            console.error(`[kyaraben] Found sidecar at: ${searchPath}`);
            return searchPath;
        }
    }
    throw new Error(`Sidecar binary '${sidecarName}' not found. Searched: ${searchPaths.join(', ')}`);
}
async function ensureDaemon() {
    if (daemon)
        return;
    const sidecarPath = findSidecarPath();
    console.error(`[kyaraben] Starting daemon: ${sidecarPath}`);
    const child = (0, node_child_process_1.spawn)(sidecarPath, ['daemon'], {
        stdio: ['pipe', 'pipe', 'inherit'],
    });
    if (!child.stdout) {
        throw new Error('Failed to get daemon stdout');
    }
    const rl = readline.createInterface({
        input: child.stdout,
        crlfDelay: Number.POSITIVE_INFINITY,
    });
    daemon = {
        process: child,
        rl,
        pending: new Map(),
        requestId: 0,
    };
    // Handle daemon events
    rl.on('line', (line) => {
        try {
            const event = JSON.parse(line);
            // For now, we handle responses sequentially since the daemon
            // doesn't support request IDs yet
            const firstPending = daemon?.pending.entries().next().value;
            if (firstPending) {
                const [id, { resolve }] = firstPending;
                daemon?.pending.delete(id);
                resolve(event);
            }
        }
        catch (err) {
            console.error(`[kyaraben] Failed to parse daemon event: ${err}`);
        }
    });
    child.on('exit', (code) => {
        console.error(`[kyaraben] Daemon exited with code ${code}`);
        daemon = null;
    });
    // Wait for ready event
    const currentDaemon = daemon;
    await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => reject(new Error('Daemon startup timeout')), 10000);
        const id = currentDaemon.requestId++;
        currentDaemon.pending.set(id, {
            resolve: (event) => {
                clearTimeout(timeout);
                if (event.type === 'ready') {
                    resolve(event);
                }
                else {
                    reject(new Error(`Expected ready event, got: ${event.type}`));
                }
            },
            reject,
        });
    });
    console.error('[kyaraben] Daemon ready');
}
async function sendCommand(cmd) {
    await ensureDaemon();
    if (!daemon || !daemon.process.stdin) {
        throw new Error('Daemon not running');
    }
    const currentDaemon = daemon;
    const stdin = daemon.process.stdin;
    const json = `${JSON.stringify(cmd)}\n`;
    stdin.write(json);
    return new Promise((resolve, reject) => {
        const id = currentDaemon.requestId++;
        const timeout = setTimeout(() => {
            currentDaemon.pending.delete(id);
            reject(new Error('Command timeout'));
        }, 600000); // 10 minute timeout for long operations
        currentDaemon.pending.set(id, {
            resolve: (event) => {
                clearTimeout(timeout);
                if (event.type === 'error') {
                    reject(new Error(event.data?.error || 'Unknown error'));
                }
                else {
                    resolve(event);
                }
            },
            reject: (err) => {
                clearTimeout(timeout);
                reject(err);
            },
        });
    });
}
// Apply is special because it returns multiple progress events
async function applyCommand() {
    await ensureDaemon();
    if (!daemon || !daemon.process.stdin) {
        throw new Error('Daemon not running');
    }
    const currentDaemon = daemon;
    const stdin = daemon.process.stdin;
    const messages = [];
    const json = `${JSON.stringify({ type: 'apply' })}\n`;
    stdin.write(json);
    return new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
            reject(new Error('Apply timeout'));
        }, 15 * 60 * 1000); // 15 minute timeout
        const handleEvent = (event) => {
            if (event.type === 'progress') {
                const msg = event.data?.message;
                if (msg)
                    messages.push(msg);
                // Continue listening for more events
                const id = currentDaemon.requestId++;
                currentDaemon.pending.set(id, { resolve: handleEvent, reject });
            }
            else if (event.type === 'result') {
                clearTimeout(timeout);
                messages.push('Apply completed successfully');
                resolve(messages);
            }
            else if (event.type === 'error') {
                clearTimeout(timeout);
                reject(new Error(event.data?.error || 'Unknown error'));
            }
        };
        const id = currentDaemon.requestId++;
        currentDaemon.pending.set(id, { resolve: handleEvent, reject });
    });
}
// IPC handlers
function setupIpcHandlers() {
    electron_1.ipcMain.handle('get_systems', async () => {
        const event = await sendCommand({ type: 'get_systems' });
        return event.data;
    });
    electron_1.ipcMain.handle('get_config', async () => {
        const event = await sendCommand({ type: 'get_config' });
        return event.data;
    });
    electron_1.ipcMain.handle('set_config', async (_, data) => {
        await sendCommand({ type: 'set_config', data });
    });
    electron_1.ipcMain.handle('status', async () => {
        const event = await sendCommand({ type: 'status' });
        return event.data;
    });
    electron_1.ipcMain.handle('doctor', async () => {
        const event = await sendCommand({ type: 'doctor' });
        return event.data;
    });
    electron_1.ipcMain.handle('apply', async () => {
        return await applyCommand();
    });
    electron_1.ipcMain.handle('get_install_status', async () => {
        const os = require('node:os');
        const fs = require('node:fs');
        const homeDir = os.homedir();
        const binDir = path.join(homeDir, '.local', 'bin');
        const appsDir = path.join(homeDir, '.local', 'share', 'applications');
        const appPath = path.join(binDir, 'kyaraben.AppImage');
        const desktopPath = path.join(appsDir, 'kyaraben.desktop');
        const cliPath = path.join(binDir, 'kyaraben');
        const installed = fs.existsSync(appPath) && fs.existsSync(desktopPath);
        return {
            installed,
            appPath: fs.existsSync(appPath) ? appPath : null,
            desktopPath: fs.existsSync(desktopPath) ? desktopPath : null,
            cliPath: fs.existsSync(cliPath) ? cliPath : null,
        };
    });
    electron_1.ipcMain.handle('install_app', async () => {
        const os = require('node:os');
        const fs = require('node:fs/promises');
        const homeDir = os.homedir();
        const binDir = path.join(homeDir, '.local', 'bin');
        const appsDir = path.join(homeDir, '.local', 'share', 'applications');
        const appPath = path.join(binDir, 'kyaraben.AppImage');
        const desktopPath = path.join(appsDir, 'kyaraben.desktop');
        const cliPath = path.join(binDir, 'kyaraben');
        // Create directories
        await fs.mkdir(binDir, { recursive: true });
        await fs.mkdir(appsDir, { recursive: true });
        // Copy AppImage
        const currentExe = electron_1.app.getPath('exe');
        await fs.copyFile(currentExe, appPath);
        await fs.chmod(appPath, 0o755);
        // Copy CLI binary
        try {
            const sidecarPath = findSidecarPath();
            await fs.copyFile(sidecarPath, cliPath);
            await fs.chmod(cliPath, 0o755);
        }
        catch {
            // CLI copy is optional
        }
        // Create .desktop file
        const desktopContent = `[Desktop Entry]
Name=Kyaraben
Comment=Declarative emulation manager
Exec=${appPath}
Icon=applications-games
Terminal=false
Type=Application
Categories=Game;Emulator;
`;
        await fs.writeFile(desktopPath, desktopContent);
    });
    electron_1.ipcMain.handle('uninstall_app', async () => {
        const os = require('node:os');
        const fs = require('node:fs/promises');
        const homeDir = os.homedir();
        const binDir = path.join(homeDir, '.local', 'bin');
        const appsDir = path.join(homeDir, '.local', 'share', 'applications');
        const appPath = path.join(binDir, 'kyaraben.AppImage');
        const desktopPath = path.join(appsDir, 'kyaraben.desktop');
        const cliPath = path.join(binDir, 'kyaraben');
        // Remove files (ignore errors if they don't exist)
        await fs.unlink(appPath).catch(() => undefined);
        await fs.unlink(desktopPath).catch(() => undefined);
        await fs.unlink(cliPath).catch(() => undefined);
    });
}
// Window creation
let mainWindow = null;
function createWindow() {
    mainWindow = new electron_1.BrowserWindow({
        width: 1200,
        height: 800,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            contextIsolation: true,
            nodeIntegration: false,
        },
    });
    // Load the app
    if (process.env.VITE_DEV_SERVER_URL) {
        mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL);
    }
    else {
        mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'));
    }
    mainWindow.on('closed', () => {
        mainWindow = null;
    });
}
// App lifecycle
electron_1.app.whenReady().then(() => {
    setupIpcHandlers();
    createWindow();
    electron_1.app.on('activate', () => {
        if (electron_1.BrowserWindow.getAllWindows().length === 0) {
            createWindow();
        }
    });
});
electron_1.app.on('window-all-closed', () => {
    // Kill daemon on exit
    if (daemon) {
        daemon.process.kill();
        daemon = null;
    }
    if (process.platform !== 'darwin') {
        electron_1.app.quit();
    }
});
