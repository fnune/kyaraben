import * as http from 'node:http'

export interface Device {
  deviceID: string
  name: string
  addresses?: string[]
  compression?: string
  autoAcceptFolders?: boolean
  paused?: boolean
}

export interface Folder {
  id: string
  path: string
  type?: string
  devices?: Array<{ deviceID: string }>
}

export interface LocalChange {
  action: string
  type: string
  name: string
  modified: string
  size: number
}

export interface PendingDevice {
  deviceID: string
  name: string
  address: string
  time: string
}

interface FolderState {
  state: 'idle' | 'scanning' | 'syncing' | 'error'
  error: string
  globalBytes: number
  needBytes: number
  globalFiles: number
  needFiles: number
  localBytes: number
  localFiles: number
  inSyncBytes: number
  inSyncFiles: number
  receiveOnlyTotalItems: number
  receiveOnlyChangedBytes: number
  pullErrors: number
}

interface ConnectionState {
  connected: boolean
  address: string
  paused: boolean
  type?: string
  isLocal?: boolean
}

interface DeviceCompletion {
  completion: number
  globalBytes: number
  needBytes: number
}

interface ConnectionServiceInfo {
  error: string
  lanAddresses: string[]
  wanAddresses: string[]
}

interface ServerState {
  myID: string
  devices: Map<string, Device>
  folders: Map<string, Folder>
  folderStates: Map<string, FolderState>
  connections: Map<string, ConnectionState>
  localChanges: Map<string, LocalChange[]>
  pendingDevices: Map<string, PendingDevice>
  deviceCompletions: Map<string, DeviceCompletion>
  discoveredAddresses: Map<string, string[]>
  connectionServiceStatus: Map<string, ConnectionServiceInfo>
}

export class FakeSyncthingController {
  private state: ServerState

  constructor(myID: string) {
    this.state = {
      myID,
      devices: new Map(),
      folders: new Map(),
      folderStates: new Map(),
      connections: new Map(),
      localChanges: new Map(),
      pendingDevices: new Map(),
      deviceCompletions: new Map(),
      discoveredAddresses: new Map(),
      connectionServiceStatus: new Map(),
    }
  }

  addDevice(device: Device): void {
    const existing = this.state.devices.get(device.deviceID)
    this.state.devices.set(device.deviceID, {
      addresses: ['dynamic'],
      compression: 'metadata',
      autoAcceptFolders: false,
      paused: false,
      ...device,
      name: device.name || existing?.name || '',
    })
  }

  removeDevice(deviceID: string): void {
    this.state.devices.delete(deviceID)
    this.state.connections.delete(deviceID)
  }

  addFolder(folder: Folder): void {
    this.state.folders.set(folder.id, {
      type: 'sendreceive',
      devices: [],
      ...folder,
    })
    if (!this.state.folderStates.has(folder.id)) {
      this.state.folderStates.set(folder.id, {
        state: 'idle',
        error: '',
        globalBytes: 0,
        needBytes: 0,
        globalFiles: 0,
        needFiles: 0,
        localBytes: 0,
        localFiles: 0,
        inSyncBytes: 0,
        inSyncFiles: 0,
        receiveOnlyTotalItems: 0,
        receiveOnlyChangedBytes: 0,
        pullErrors: 0,
      })
    }
  }

  setConnected(deviceID: string, connected: boolean): void {
    const existing = this.state.connections.get(deviceID)
    this.state.connections.set(deviceID, {
      connected,
      address: existing?.address ?? '192.168.1.100:22100',
      paused: existing?.paused ?? false,
    })
  }

  setFolderState(
    folderID: string,
    state: 'idle' | 'scanning' | 'syncing' | 'error',
    error = '',
  ): void {
    const existing = this.state.folderStates.get(folderID)
    if (existing) {
      existing.state = state
      existing.error = error
    }
  }

  setFolderProgress(folderID: string, needBytes: number, globalBytes: number): void {
    const existing = this.state.folderStates.get(folderID)
    if (existing) {
      existing.needBytes = needBytes
      existing.globalBytes = globalBytes
      existing.inSyncBytes = globalBytes - needBytes
      if (needBytes > 0) {
        existing.state = 'syncing'
      }
    }
  }

  addLocalChanges(folderID: string, files: LocalChange[]): void {
    const existing = this.state.localChanges.get(folderID) ?? []
    this.state.localChanges.set(folderID, [...existing, ...files])
    const folderState = this.state.folderStates.get(folderID)
    if (folderState) {
      folderState.receiveOnlyTotalItems = this.state.localChanges.get(folderID)?.length ?? 0
    }
  }

  clearLocalChanges(folderID: string): void {
    this.state.localChanges.set(folderID, [])
    const folderState = this.state.folderStates.get(folderID)
    if (folderState) {
      folderState.receiveOnlyTotalItems = 0
    }
  }

  addPendingDevice(device: PendingDevice): void {
    this.state.pendingDevices.set(device.deviceID, device)
  }

  removePendingDevice(deviceID: string): void {
    this.state.pendingDevices.delete(deviceID)
  }

  setDeviceCompletion(deviceID: string, completion: number, needBytes = 0, globalBytes = 0): void {
    this.state.deviceCompletions.set(deviceID, {
      completion,
      needBytes,
      globalBytes,
    })
  }

  setDiscoveredAddresses(deviceID: string, addresses: string[]): void {
    this.state.discoveredAddresses.set(deviceID, addresses)
  }

  setConnectionServiceStatus(
    name: string,
    info: { error?: string; lanAddresses?: string[]; wanAddresses?: string[] },
  ): void {
    this.state.connectionServiceStatus.set(name, {
      error: info.error ?? '',
      lanAddresses: info.lanAddresses ?? [],
      wanAddresses: info.wanAddresses ?? [],
    })
  }

  getState(): ServerState {
    return this.state
  }
}

export function startFakeSyncthingServer(
  port: number,
  controller: FakeSyncthingController,
): http.Server {
  const server = http.createServer((req, res) => {
    const url = new URL(req.url ?? '/', `http://localhost:${port}`)
    const state = controller.getState()

    res.setHeader('Content-Type', 'application/json')

    if (req.method === 'GET' && url.pathname === '/rest/system/ping') {
      res.writeHead(200)
      res.end(JSON.stringify({ ping: 'pong' }))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/system/status') {
      const connectionServiceStatus: Record<
        string,
        { error: string; lanAddresses: string[]; wanAddresses: string[] }
      > = {}
      for (const [name, info] of state.connectionServiceStatus) {
        connectionServiceStatus[name] = info
      }
      res.writeHead(200)
      res.end(
        JSON.stringify({
          myID: state.myID,
          uptime: 60,
          connectionServiceStatus:
            Object.keys(connectionServiceStatus).length > 0 ? connectionServiceStatus : undefined,
        }),
      )
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/config/devices') {
      const devices = Array.from(state.devices.values())
      res.writeHead(200)
      res.end(JSON.stringify(devices))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/config/folders') {
      const folders = Array.from(state.folders.values())
      res.writeHead(200)
      res.end(JSON.stringify(folders))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/db/status') {
      const folderID = url.searchParams.get('folder')
      if (!folderID) {
        res.writeHead(400)
        res.end(JSON.stringify({ error: 'folder parameter required' }))
        return
      }
      const folderState = state.folderStates.get(folderID)
      if (!folderState) {
        res.writeHead(404)
        res.end(JSON.stringify({ error: 'folder not found' }))
        return
      }
      res.writeHead(200)
      res.end(
        JSON.stringify({
          state: folderState.state,
          error: folderState.error,
          globalFiles: folderState.globalFiles,
          globalBytes: folderState.globalBytes,
          localFiles: folderState.localFiles,
          localBytes: folderState.localBytes,
          needFiles: folderState.needFiles,
          needBytes: folderState.needBytes,
          inSyncFiles: folderState.inSyncFiles,
          inSyncBytes: folderState.inSyncBytes,
          pullErrors: folderState.pullErrors,
          receiveOnlyTotalItems: folderState.receiveOnlyTotalItems,
          receiveOnlyChangedBytes: folderState.receiveOnlyChangedBytes,
        }),
      )
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/system/connections') {
      const connections: Record<string, ConnectionState> = {}
      for (const [deviceID, conn] of state.connections) {
        connections[deviceID] = conn
      }
      res.writeHead(200)
      res.end(JSON.stringify({ connections }))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/db/completion') {
      const deviceID = url.searchParams.get('device')
      if (!deviceID) {
        res.writeHead(400)
        res.end(JSON.stringify({ error: 'device parameter required' }))
        return
      }
      const completion = state.deviceCompletions.get(deviceID) ?? {
        completion: 100,
        needBytes: 0,
        globalBytes: 0,
      }
      res.writeHead(200)
      res.end(
        JSON.stringify({
          completion: completion.completion,
          globalBytes: completion.globalBytes,
          needBytes: completion.needBytes,
          globalItems: 0,
          needItems: 0,
          needDeletes: 0,
        }),
      )
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/db/localchanged') {
      const folderID = url.searchParams.get('folder')
      if (!folderID) {
        res.writeHead(400)
        res.end(JSON.stringify({ error: 'folder parameter required' }))
        return
      }
      const changes = state.localChanges.get(folderID) ?? []
      res.writeHead(200)
      res.end(JSON.stringify({ files: changes }))
      return
    }

    if (req.method === 'POST' && url.pathname === '/rest/db/revert') {
      const folderID = url.searchParams.get('folder')
      if (!folderID) {
        res.writeHead(400)
        res.end(JSON.stringify({ error: 'folder parameter required' }))
        return
      }
      controller.clearLocalChanges(folderID)
      res.writeHead(200)
      res.end(JSON.stringify({ status: 'ok' }))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/cluster/pending/devices') {
      const pending: Record<string, { name: string; address: string; time: string }> = {}
      for (const [deviceID, device] of state.pendingDevices) {
        pending[deviceID] = {
          name: device.name,
          address: device.address,
          time: device.time,
        }
      }
      res.writeHead(200)
      res.end(JSON.stringify(pending))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/cluster/pending/folders') {
      res.writeHead(200)
      res.end(JSON.stringify({}))
      return
    }

    if (req.method === 'GET' && url.pathname === '/rest/system/discovery') {
      const discovery: Record<string, { addresses: string[] }> = {}
      for (const [deviceID, addresses] of state.discoveredAddresses) {
        discovery[deviceID] = { addresses }
      }
      res.writeHead(200)
      res.end(JSON.stringify(discovery))
      return
    }

    if (req.method === 'PUT' && url.pathname.startsWith('/rest/config/devices/')) {
      let body = ''
      req.on('data', (chunk) => {
        body += chunk
      })
      req.on('end', () => {
        try {
          const device = JSON.parse(body) as Device
          controller.addDevice(device)
          res.writeHead(200)
          res.end(JSON.stringify({ status: 'ok' }))
        } catch {
          res.writeHead(400)
          res.end(JSON.stringify({ error: 'invalid JSON' }))
        }
      })
      return
    }

    if (req.method === 'DELETE' && url.pathname.startsWith('/rest/config/devices/')) {
      const deviceID = url.pathname.split('/').pop()
      if (deviceID) {
        controller.removeDevice(deviceID)
      }
      res.writeHead(200)
      res.end(JSON.stringify({ status: 'ok' }))
      return
    }

    if (req.method === 'PUT' && url.pathname === '/rest/config/folders') {
      let body = ''
      req.on('data', (chunk) => {
        body += chunk
      })
      req.on('end', () => {
        try {
          const folders = JSON.parse(body) as Folder[]
          for (const folder of folders) {
            controller.addFolder(folder)
          }
          res.writeHead(200)
          res.end(JSON.stringify({ status: 'ok' }))
        } catch {
          res.writeHead(400)
          res.end(JSON.stringify({ error: 'invalid JSON' }))
        }
      })
      return
    }

    if (req.method === 'PUT' && url.pathname.startsWith('/rest/config/folders/')) {
      let body = ''
      req.on('data', (chunk) => {
        body += chunk
      })
      req.on('end', () => {
        try {
          const folder = JSON.parse(body) as Folder
          controller.addFolder(folder)
          res.writeHead(200)
          res.end(JSON.stringify({ status: 'ok' }))
        } catch {
          res.writeHead(400)
          res.end(JSON.stringify({ error: 'invalid JSON' }))
        }
      })
      return
    }

    if (req.method === 'POST' && url.pathname === '/rest/system/restart') {
      res.writeHead(200)
      res.end(JSON.stringify({ status: 'ok' }))
      return
    }

    res.writeHead(404)
    res.end(JSON.stringify({ error: 'not found', path: url.pathname }))
  })

  server.listen(port)
  return server
}
