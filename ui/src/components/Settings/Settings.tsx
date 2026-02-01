export interface SettingsProps {
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
}

export function Settings({ userStore, onUserStoreChange }: SettingsProps) {
  const handleOpenFolder = () => {
    window.electron.invoke('open_path', userStore)
  }

  return (
    <div className="p-4 bg-gray-50 rounded-lg mb-6">
      <label className="block">
        <span className="text-sm font-medium text-gray-700">Emulation folder</span>
        <div className="mt-1 flex gap-2">
          <input
            type="text"
            value={userStore}
            onChange={(e) => onUserStoreChange(e.target.value)}
            className="block flex-1 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
            placeholder="~/Emulation"
          />
          <button
            type="button"
            onClick={handleOpenFolder}
            className="px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-100"
            title="Open folder"
          >
            <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z" />
            </svg>
          </button>
        </div>
      </label>
    </div>
  )
}
