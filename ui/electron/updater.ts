import * as fs from 'node:fs'
import * as http from 'node:http'
import * as https from 'node:https'
import * as os from 'node:os'
import * as path from 'node:path'
import { app } from 'electron'
import * as semver from 'semver'

export interface UpdateInfo {
  available: boolean
  currentVersion: string
  latestVersion: string
  downloadUrl: string
  releaseNotes?: string
}

interface GitHubRelease {
  tag_name: string
  body?: string
  assets: Array<{
    name: string
    browser_download_url: string
  }>
}

function getReleasesUrl(): string {
  return (
    process.env.KYARABEN_RELEASES_URL ||
    'https://api.github.com/repos/fnune/kyaraben/releases/latest'
  )
}

function fetch(url: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https') ? https : http
    const options = {
      headers: {
        'User-Agent': 'kyaraben-updater',
        Accept: 'application/vnd.github.v3+json',
      },
    }

    const req = protocol.get(url, options, (res) => {
      if (res.statusCode === 301 || res.statusCode === 302) {
        const location = res.headers.location
        if (location) {
          fetch(location).then(resolve).catch(reject)
          return
        }
      }

      if (res.statusCode !== 200) {
        reject(new Error(`HTTP ${res.statusCode}`))
        return
      }

      let data = ''
      res.on('data', (chunk) => {
        data += chunk
      })
      res.on('end', () => resolve(data))
    })

    req.on('error', reject)
    req.setTimeout(30000, () => {
      req.destroy()
      reject(new Error('Request timeout'))
    })
  })
}

function getExpectedAssetName(): string {
  const arch = process.arch === 'x64' ? 'x86_64' : process.arch === 'arm64' ? 'aarch64' : 'unknown'
  return `Kyaraben-${arch}.AppImage`
}

export async function checkForUpdates(): Promise<UpdateInfo> {
  const currentVersion = app.getVersion()
  const releasesUrl = getReleasesUrl()

  try {
    const responseText = await fetch(releasesUrl)
    const release: GitHubRelease = JSON.parse(responseText)

    const latestVersion = release.tag_name.replace(/^v/, '')
    const available = semver.valid(latestVersion) && semver.gt(latestVersion, currentVersion)

    const expectedAsset = getExpectedAssetName()
    const asset = release.assets.find((a) => a.name === expectedAsset)
    const downloadUrl = asset?.browser_download_url || ''

    return {
      available: Boolean(available && downloadUrl),
      currentVersion,
      latestVersion,
      downloadUrl,
      releaseNotes: release.body,
    }
  } catch (error) {
    const msg = error instanceof Error ? error.message : String(error)
    console.error(`[kyaraben] Update check failed: ${msg}`)
    return {
      available: false,
      currentVersion,
      latestVersion: currentVersion,
      downloadUrl: '',
    }
  }
}

export function downloadUpdate(
  url: string,
  onProgress: (percent: number) => void,
): Promise<string> {
  return new Promise((resolve, reject) => {
    const tempDir = path.join(os.tmpdir(), 'kyaraben-update')
    fs.mkdirSync(tempDir, { recursive: true })
    const tempPath = path.join(tempDir, `kyaraben-update-${Date.now()}.AppImage`)

    const makeRequest = (requestUrl: string) => {
      const protocol = requestUrl.startsWith('https') ? https : http

      const req = protocol.get(requestUrl, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          const location = res.headers.location
          if (location) {
            makeRequest(location)
            return
          }
        }

        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode}`))
          return
        }

        const file = fs.createWriteStream(tempPath)
        const totalSize = Number.parseInt(res.headers['content-length'] || '0', 10)
        let downloaded = 0

        res.on('data', (chunk: Buffer) => {
          downloaded += chunk.length
          if (totalSize > 0) {
            onProgress(Math.round((downloaded / totalSize) * 100))
          }
        })

        res.pipe(file)

        file.on('finish', () => {
          file.close()
          fs.chmodSync(tempPath, 0o755)
          resolve(tempPath)
        })

        file.on('error', (err) => {
          file.close()
          try {
            fs.unlinkSync(tempPath)
          } catch {
            // Ignore cleanup errors
          }
          reject(err)
        })
      })

      req.on('error', (err) => {
        try {
          fs.unlinkSync(tempPath)
        } catch {
          // Ignore cleanup errors
        }
        reject(err)
      })

      req.setTimeout(600000, () => {
        req.destroy()
        try {
          fs.unlinkSync(tempPath)
        } catch {
          // Ignore cleanup errors
        }
        reject(new Error('Download timeout'))
      })
    }

    makeRequest(url)
  })
}
