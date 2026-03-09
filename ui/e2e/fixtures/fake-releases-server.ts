import * as fs from 'node:fs'
import * as http from 'node:http'

export interface FakeReleasesServerOptions {
  version: string
  appImagePath?: string
  simulateRedirect?: boolean
}

export function startFakeReleasesServer(
  port: number,
  options: FakeReleasesServerOptions,
): http.Server {
  const { version, appImagePath, simulateRedirect } = options

  const server = http.createServer((req, res) => {
    const url = new URL(req.url || '/', `http://localhost:${port}`)

    if (url.pathname === '/releases/latest') {
      const arch = process.arch === 'x64' ? 'x86_64' : 'aarch64'
      const assetName = `Kyaraben-${arch}.AppImage`
      const downloadPath = simulateRedirect ? '/redirect' : `/download/${assetName}`

      const release = {
        tag_name: `v${version}`,
        body: 'Test release notes',
        assets: [
          {
            name: assetName,
            browser_download_url: `http://localhost:${port}${downloadPath}`,
          },
        ],
      }

      res.writeHead(200, { 'Content-Type': 'application/json' })
      res.end(JSON.stringify(release))
      return
    }

    if (url.pathname === '/redirect' && appImagePath) {
      const arch = process.arch === 'x64' ? 'x86_64' : 'aarch64'
      const assetName = `Kyaraben-${arch}.AppImage`
      res.writeHead(302, { Location: `http://localhost:${port}/download/${assetName}` })
      res.end()
      return
    }

    if (url.pathname.startsWith('/download/') && appImagePath) {
      const stat = fs.statSync(appImagePath)
      res.writeHead(200, {
        'Content-Type': 'application/octet-stream',
        'Content-Length': stat.size,
      })
      fs.createReadStream(appImagePath).pipe(res)
      return
    }

    res.writeHead(404)
    res.end('Not found')
  })

  server.listen(port)
  return server
}
