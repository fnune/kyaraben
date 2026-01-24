import { type ChildProcess, spawn } from 'node:child_process'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import type { Options } from '@wdio/types'

const __dirname = dirname(fileURLToPath(import.meta.url))

// Path to the built Tauri app binary
const appPath = resolve(__dirname, 'src-tauri/target/release/kyaraben-ui')

let tauriDriver: ChildProcess | null = null

export const config: Options.Testrunner = {
  runner: 'local',
  tsConfigPath: './tsconfig.json',
  specs: ['./e2e/**/*.spec.ts'],
  maxInstances: 1,
  capabilities: [
    {
      // @ts-expect-error tauri:options is valid for tauri-driver
      'tauri:options': {
        application: appPath,
      },
    },
  ],
  reporters: ['spec'],
  framework: 'mocha',
  mochaOpts: {
    ui: 'bdd',
    // Nix builds can take several minutes, especially first run downloading from cache.nixos.org
    timeout: 900000, // 15 minutes
  },

  // Start tauri-driver before tests
  onPrepare: async () => {
    tauriDriver = spawn('tauri-driver', [], {
      stdio: ['ignore', 'pipe', 'pipe'],
    })

    tauriDriver.stdout?.on('data', (data: Buffer) => {
      console.log(`[tauri-driver] ${data.toString()}`)
    })

    tauriDriver.stderr?.on('data', (data: Buffer) => {
      console.error(`[tauri-driver] ${data.toString()}`)
    })

    // Wait for tauri-driver to start
    await new Promise((resolve) => setTimeout(resolve, 2000))
  },

  // Stop tauri-driver after tests
  onComplete: () => {
    if (tauriDriver) {
      tauriDriver.kill()
    }
  },

  // WebDriver configuration for tauri-driver
  hostname: '127.0.0.1',
  port: 4444,
  connectionRetryCount: 3,
  connectionRetryTimeout: 60000,
}
