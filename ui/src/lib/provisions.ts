import type { DoctorResponse } from '@/types/daemon'

export interface FoundProvision {
  id: string
  emulatorId: string
  filename: string
  displayName: string
  description: string | undefined
  foundPath: string | undefined
  expectedPath: string | undefined
}

function provisionKey(emulatorId: string, result: DoctorResponse[string][number]) {
  const description = result.description ?? ''
  const expectedPath = result.expectedPath ?? ''
  return `${emulatorId}:${result.filename}:${description}:${expectedPath}`
}

export function getNewlyFoundProvisions(
  oldProvisions: DoctorResponse,
  newProvisions: DoctorResponse,
): FoundProvision[] {
  const found: FoundProvision[] = []
  const oldFoundKeys = new Set<string>()

  for (const [emulatorId, oldResults] of Object.entries(oldProvisions)) {
    for (const oldResult of oldResults ?? []) {
      if (oldResult.status === 'found') {
        oldFoundKeys.add(provisionKey(emulatorId, oldResult))
      }
    }
  }

  for (const [emulatorId, newResults] of Object.entries(newProvisions)) {
    for (const newResult of newResults) {
      if (newResult.status !== 'found') {
        continue
      }

      const key = provisionKey(emulatorId, newResult)
      if (oldFoundKeys.has(key)) {
        continue
      }

      found.push({
        id: key,
        emulatorId,
        filename: newResult.filename,
        displayName: newResult.displayName || newResult.filename,
        description: newResult.description || undefined,
        foundPath: newResult.foundPath,
        expectedPath: newResult.expectedPath,
      })
    }
  }

  return found
}
