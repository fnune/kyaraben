import { useCallback, useState } from 'react'
import type {
  CommandType,
  ConfigResponse,
  DoctorResponse,
  InstallStatus,
  SetConfigRequest,
  StatusResponse,
  System,
} from '@/types/daemon'

type DaemonError = {
  readonly message: string
}

type DaemonResult<T> =
  | { readonly ok: true; readonly data: T }
  | { readonly ok: false; readonly error: DaemonError }

async function invoke<T>(command: CommandType, data?: SetConfigRequest): Promise<DaemonResult<T>> {
  try {
    const result = (await window.electron.invoke(command, data)) as T
    return { ok: true, data: result }
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return { ok: false, error: { message } }
  }
}

export interface UseDaemonReturn {
  readonly loading: boolean
  readonly error: string | null
  readonly getSystems: () => Promise<DaemonResult<readonly System[]>>
  readonly getConfig: () => Promise<DaemonResult<ConfigResponse>>
  readonly setConfig: (config: SetConfigRequest) => Promise<DaemonResult<{ success: boolean }>>
  readonly getStatus: () => Promise<DaemonResult<StatusResponse>>
  readonly runDoctor: () => Promise<DaemonResult<DoctorResponse>>
  readonly apply: () => Promise<DaemonResult<readonly string[]>>
  readonly getInstallStatus: () => Promise<DaemonResult<InstallStatus>>
  readonly installApp: () => Promise<DaemonResult<void>>
  readonly uninstallApp: () => Promise<DaemonResult<void>>
}

export function useDaemon(): UseDaemonReturn {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const withLoading = useCallback(
    async <T>(fn: () => Promise<DaemonResult<T>>): Promise<DaemonResult<T>> => {
      setLoading(true)
      setError(null)
      try {
        const result = await fn()
        if (!result.ok) {
          setError(result.error.message)
        }
        return result
      } finally {
        setLoading(false)
      }
    },
    [],
  )

  const getSystems = useCallback(
    () => withLoading(() => invoke<readonly System[]>('get_systems')),
    [withLoading],
  )

  const getConfig = useCallback(
    () => withLoading(() => invoke<ConfigResponse>('get_config')),
    [withLoading],
  )

  const setConfig = useCallback(
    (config: SetConfigRequest) =>
      withLoading(() => invoke<{ success: boolean }>('set_config', config)),
    [withLoading],
  )

  const getStatus = useCallback(
    () => withLoading(() => invoke<StatusResponse>('status')),
    [withLoading],
  )

  const runDoctor = useCallback(
    () => withLoading(() => invoke<DoctorResponse>('doctor')),
    [withLoading],
  )

  const apply = useCallback(
    () => withLoading(() => invoke<readonly string[]>('apply')),
    [withLoading],
  )

  const getInstallStatus = useCallback(
    () => withLoading(() => invoke<InstallStatus>('get_install_status')),
    [withLoading],
  )

  const installApp = useCallback(
    () => withLoading(() => invoke<void>('install_app')),
    [withLoading],
  )

  const uninstallApp = useCallback(
    () => withLoading(() => invoke<void>('uninstall_app')),
    [withLoading],
  )

  return {
    loading,
    error,
    getSystems,
    getConfig,
    setConfig,
    getStatus,
    runDoctor,
    apply,
    getInstallStatus,
    installApp,
    uninstallApp,
  }
}
