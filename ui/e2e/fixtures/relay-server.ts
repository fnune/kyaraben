import { type ChildProcess, spawn } from 'node:child_process'
import * as fs from 'node:fs'

export interface RelayServer {
  port: number
  url: string
  process: ChildProcess
  close: () => void
}

let nextRelayPort = 19600

function getRelayBinaryPath(): string {
  const envPath = process.env.KYARABEN_RELAY_BINARY
  if (envPath && fs.existsSync(envPath)) {
    return envPath
  }
  throw new Error(
    'KYARABEN_RELAY_BINARY environment variable must be set to the relay binary path. Run "just ui-e2e" to set this up automatically.',
  )
}

export async function startRelayServer(): Promise<RelayServer> {
  const binaryPath = getRelayBinaryPath()
  const maxAttempts = 5

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    const port = nextRelayPort++

    const proc = spawn(binaryPath, ['-addr', `:${port}`, '-no-rate-limit'], {
      stdio: ['ignore', 'pipe', 'pipe'],
    })

    proc.stderr?.on('data', (data) => {
      console.error(`relay stderr: ${data}`)
    })

    try {
      await waitForServer(`http://localhost:${port}/health`, 3000)
      return {
        port,
        url: `http://localhost:${port}`,
        process: proc,
        close: () => {
          proc.kill('SIGTERM')
        },
      }
    } catch {
      proc.kill('SIGKILL')
    }
  }

  throw new Error(`Failed to start relay server after ${maxAttempts} attempts`)
}

async function waitForServer(url: string, timeoutMs: number): Promise<void> {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    try {
      const response = await fetch(url)
      if (response.ok) {
        return
      }
    } catch {
      // Server not ready yet
    }
    await sleep(100)
  }
  throw new Error(`Relay server did not start within ${timeoutMs}ms`)
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
