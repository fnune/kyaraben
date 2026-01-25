export interface SettingsProps {
  readonly userStore: string
  readonly onUserStoreChange: (value: string) => void
}

export function Settings({ userStore, onUserStoreChange }: SettingsProps) {
  return (
    <div className="p-4 bg-gray-50 rounded-lg mb-6">
      <label className="block">
        <span className="text-sm font-medium text-gray-700">Emulation folder</span>
        <input
          type="text"
          value={userStore}
          onChange={(e) => onUserStoreChange(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 px-3 py-2 border"
          placeholder="~/Emulation"
        />
      </label>
    </div>
  )
}
