import type { DoctorResponse } from '@/types/daemon'

export interface FoundProvision {
  id: string
  emulatorId: string
  filename: string
  displayName: string
  foundPath: string | undefined
  expectedPath: string | undefined
}

export function getNewlyFoundProvisions(
  oldProvisions: DoctorResponse,
  newProvisions: DoctorResponse,
): FoundProvision[] {
  const found: FoundProvision[] = []
  for (const [emulatorId, newResults] of Object.entries(newProvisions)) {
    const oldResults = oldProvisions[emulatorId] ?? []
    for (const newResult of newResults) {
      if (newResult.status === 'found') {
        const oldResult = oldResults.find((r) => r.filename === newResult.filename)
        if (!oldResult || oldResult.status !== 'found') {
          const displayName =
            newResult.displayName ||
            newResult.description ||
            newResult.filename
          found.push({
            id: `${emulatorId}:${newResult.filename}`,
            emulatorId,
            filename: newResult.filename,
            displayName,
            foundPath: newResult.foundPath,
            expectedPath: newResult.expectedPath,
          })
        }
      }
    }
  }
  return found
}
