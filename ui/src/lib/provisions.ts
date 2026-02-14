import type { DoctorResponse } from '@/types/daemon'

export function getNewlyFoundProvisions(
  oldProvisions: DoctorResponse,
  newProvisions: DoctorResponse,
): string[] {
  const found: string[] = []
  for (const [emulatorId, newResults] of Object.entries(newProvisions)) {
    const oldResults = oldProvisions[emulatorId] ?? []
    for (const newResult of newResults) {
      if (newResult.status === 'found') {
        const oldResult = oldResults.find((r) => r.filename === newResult.filename)
        if (!oldResult || oldResult.status !== 'found') {
          found.push(newResult.filename)
        }
      }
    }
  }
  return found
}
