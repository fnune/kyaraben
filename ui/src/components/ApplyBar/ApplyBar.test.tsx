import type { ConfigFileDiff, PreflightResponse } from '@shared/daemon'
import type { ApplyStatus } from '@shared/ui'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, type Mock, vi } from 'vitest'
import { useApply } from '@/lib/ApplyContext'
import { useConfig } from '@/lib/ConfigContext'
import { useHomeDir } from '@/lib/HomeDirContext'
import { useOpenLog } from '@/lib/useOpenLog'
import { ApplyBar } from './ApplyBar'

vi.mock('@/lib/ApplyContext')
vi.mock('@/lib/ConfigContext')
vi.mock('@/lib/HomeDirContext')
vi.mock('@/lib/useOpenLog')
vi.mock('@/lib/BottomBarSlot', () => ({
  BottomBarPortal: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}))

const mockUseApply = useApply as Mock
const mockUseConfig = useConfig as Mock
const mockUseHomeDir = useHomeDir as Mock
const mockUseOpenLog = useOpenLog as Mock

function createDiff(overrides: Partial<ConfigFileDiff> = {}): ConfigFileDiff {
  return {
    path: '/test/path',
    isNewFile: false,
    hasChanges: false,
    userModified: false,
    kyarabenChanged: false,
    ...overrides,
  }
}

function createPreflightData(overrides: Partial<PreflightResponse> = {}): PreflightResponse {
  return {
    diffs: [],
    filesToBackup: [],
    ...overrides,
  }
}

function createApplyMock(overrides: Partial<ReturnType<typeof useApply>> = {}) {
  return {
    status: 'idle' as ApplyStatus,
    progressSteps: [],
    error: null,
    preflightData: null,
    syncPendingData: null,
    logPosition: null,
    apply: vi.fn(),
    confirmApply: vi.fn(),
    confirmSyncPending: vi.fn(),
    cancel: vi.fn(),
    reset: vi.fn(),
    onCompleteRef: { current: null },
    ...overrides,
  }
}

function createChanges(overrides: Partial<ReturnType<typeof useConfig>['changes']> = {}) {
  return {
    total: 0,
    installs: 0,
    removes: 0,
    upgrades: 0,
    downgrades: 0,
    hasConfigChanges: false,
    downloadBytes: 0,
    freeBytes: 0,
    configChanges: [] as readonly string[],
    installItems: [] as readonly { id: string; name: string }[],
    removeItems: [] as readonly { id: string; name: string }[],
    upgradeItems: [] as readonly { id: string; name: string }[],
    downgradeItems: [] as readonly { id: string; name: string }[],
    ...overrides,
  }
}

function createConfigMock(overrides: Partial<ReturnType<typeof useConfig>> = {}) {
  return {
    changes: createChanges(),
    apply: vi.fn(),
    reapply: vi.fn(),
    discard: vi.fn(),
    upgradeAvailable: false,
    ...overrides,
  }
}

describe('ApplyBar', () => {
  beforeEach(() => {
    mockUseHomeDir.mockReturnValue('/home/user')
    mockUseOpenLog.mockReturnValue(vi.fn())
  })

  describe('idle state', () => {
    it('renders nothing when no changes', () => {
      mockUseApply.mockReturnValue(createApplyMock({ status: 'idle' }))
      mockUseConfig.mockReturnValue(createConfigMock())

      const { container } = render(<ApplyBar />)

      expect(container).toBeEmptyDOMElement()
    })

    it('renders apply bar when there are config changes', () => {
      mockUseApply.mockReturnValue(createApplyMock({ status: 'idle' }))
      mockUseConfig.mockReturnValue(
        createConfigMock({
          changes: createChanges({
            hasConfigChanges: true,
            configChanges: ['Hotkey settings'],
          }),
        }),
      )

      render(<ApplyBar />)

      expect(screen.getByRole('button', { name: 'Apply' })).toBeInTheDocument()
      expect(screen.getByText(/hotkey settings changed/i)).toBeInTheDocument()
    })
  })

  describe('reviewing state', () => {
    it('renders nothing when no preflightData', () => {
      mockUseApply.mockReturnValue(createApplyMock({ status: 'reviewing', preflightData: null }))
      mockUseConfig.mockReturnValue(createConfigMock())

      const { container } = render(<ApplyBar />)

      expect(container).toBeEmptyDOMElement()
    })

    it('renders Continue button when no user conflicts', () => {
      mockUseApply.mockReturnValue(
        createApplyMock({
          status: 'reviewing',
          preflightData: createPreflightData({
            diffs: [createDiff({ kyarabenChanged: true })],
          }),
        }),
      )
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)

      expect(screen.getByRole('button', { name: 'Continue' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    })

    it('renders "Continue and override" when user conflicts exist', () => {
      mockUseApply.mockReturnValue(
        createApplyMock({
          status: 'reviewing',
          preflightData: createPreflightData({
            diffs: [
              createDiff({
                userModified: true,
                hasChanges: true,
                userChanges: [{ key: 'foo', writtenValue: 'bar', currentValue: 'baz' }],
              }),
            ],
          }),
        }),
      )
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)

      expect(screen.getByRole('button', { name: 'Continue and override' })).toBeInTheDocument()
    })

    it('calls confirmApply when Continue is clicked', async () => {
      const confirmApply = vi.fn()
      mockUseApply.mockReturnValue(
        createApplyMock({
          status: 'reviewing',
          preflightData: createPreflightData(),
          confirmApply,
        }),
      )
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)
      await userEvent.click(screen.getByRole('button', { name: 'Continue' }))

      expect(confirmApply).toHaveBeenCalled()
    })

    it('calls reset when Cancel is clicked', async () => {
      const reset = vi.fn()
      mockUseApply.mockReturnValue(
        createApplyMock({
          status: 'reviewing',
          preflightData: createPreflightData(),
          reset,
        }),
      )
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)
      await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))

      expect(reset).toHaveBeenCalled()
    })
  })

  describe('confirming_sync state', () => {
    it('renders nothing', () => {
      mockUseApply.mockReturnValue(createApplyMock({ status: 'confirming_sync' }))
      mockUseConfig.mockReturnValue(createConfigMock())

      const { container } = render(<ApplyBar />)

      expect(container).toBeEmptyDOMElement()
    })
  })

  describe('done states', () => {
    it.each([
      'success',
      'error',
      'cancelled',
    ] as const)('renders Done button when status is %s', (status) => {
      mockUseApply.mockReturnValue(createApplyMock({ status }))
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)

      expect(screen.getByRole('button', { name: 'Done' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Open log in terminal' })).toBeInTheDocument()
    })

    it('calls reset when Done is clicked', async () => {
      const reset = vi.fn()
      mockUseApply.mockReturnValue(createApplyMock({ status: 'success', reset }))
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)
      await userEvent.click(screen.getByRole('button', { name: 'Done' }))

      expect(reset).toHaveBeenCalled()
    })

    it('calls openLog when Open log is clicked', async () => {
      const openLog = vi.fn()
      mockUseOpenLog.mockReturnValue(openLog)
      mockUseApply.mockReturnValue(createApplyMock({ status: 'success', logPosition: 42 }))
      mockUseConfig.mockReturnValue(createConfigMock())

      render(<ApplyBar />)
      await userEvent.click(screen.getByRole('button', { name: 'Open log in terminal' }))

      expect(openLog).toHaveBeenCalledWith(42)
    })
  })
})
