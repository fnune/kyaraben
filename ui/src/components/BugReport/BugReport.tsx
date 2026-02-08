import { useCallback, useEffect, useState } from 'react'
import { getBugReportInfo, getStatus, runDoctor } from '@/lib/daemon'
import { Modal } from '@/lib/Modal'
import type { BugReportInfo, DoctorResponse, StatusResponse } from '@/types/daemon'

export interface BugReportProps {
  readonly open: boolean
  readonly onClose: () => void
}

function generateMarkdown(
  description: string,
  info: BugReportInfo | null,
  status: StatusResponse | null,
  doctor: DoctorResponse | null,
): string {
  const lines: string[] = []

  lines.push('## Description')
  lines.push(description || '[Describe the issue here]')
  lines.push('')

  lines.push('## Environment')
  if (info) {
    lines.push(`- App version: ${info.appVersion}`)
    lines.push(`- Platform: ${info.platform}`)
    lines.push(`- Architecture: ${info.arch}`)
    lines.push(`- OS release: ${info.osRelease}`)
  } else {
    lines.push('- Unable to fetch environment info')
  }
  lines.push('')

  lines.push('## Configuration')
  if (status) {
    lines.push(`- User store: ${status.userStore || '(not set)'}`)
    lines.push(`- Last applied: ${status.lastApplied || 'never'}`)
    lines.push(
      `- Enabled systems: ${status.enabledSystems.length > 0 ? status.enabledSystems.join(', ') : 'none'}`,
    )
    const emulators = (status.installedEmulators ?? []).map((e) => `${e.id} (${e.version})`)
    lines.push(`- Installed emulators: ${emulators.length > 0 ? emulators.join(', ') : 'none'}`)
  } else {
    lines.push('- Unable to fetch configuration')
  }
  lines.push('')

  lines.push('## State directory (~/.local/state/kyaraben)')
  if (info?.stateDir) {
    const sd = info.stateDir
    if (!sd.exists) {
      lines.push('- Directory does not exist')
    } else {
      lines.push(`- build/manifest.json: ${sd.manifestExists ? 'exists' : 'missing'}`)
      lines.push(`- build/flake: ${sd.flakeExists ? 'exists' : 'missing'}`)
      lines.push(
        `- Broken symlinks: ${sd.brokenSymlinks.length > 0 ? sd.brokenSymlinks.join(', ') : 'none'}`,
      )
    }
  } else {
    lines.push('- Unable to check state directory')
  }
  lines.push('')

  lines.push('## Provisions (doctor)')
  if (doctor) {
    const entries = Object.entries(doctor)
    if (entries.length === 0) {
      lines.push('- No provisions checked')
    } else {
      for (const [systemId, results] of entries) {
        const found = results.filter((r) => r.status === 'found').length
        const missing = results.filter((r) => r.status === 'missing')
        const hasUnsatisfiedRequired = missing.some((r) => r.groupRequired && !r.groupSatisfied)
        if (results.length === 0) {
          lines.push(`- ${systemId}: OK (no provisions required)`)
        } else if (missing.length === 0) {
          lines.push(`- ${systemId}: ${found} found, 0 missing`)
        } else {
          const reqNote = hasUnsatisfiedRequired ? ' (required)' : ''
          lines.push(`- ${systemId}: ${found} found, ${missing.length} missing${reqNote}`)
        }
      }
    }
  } else {
    lines.push('- Unable to run doctor')
  }

  return lines.join('\n')
}

export function BugReport({ open, onClose }: BugReportProps) {
  const [description, setDescription] = useState('')
  const [info, setInfo] = useState<BugReportInfo | null>(null)
  const [status, setStatus] = useState<StatusResponse | null>(null)
  const [doctor, setDoctor] = useState<DoctorResponse | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!open) return
    setDescription('')
    setCopied(false)

    getBugReportInfo().then((result) => {
      if (result.ok) setInfo(result.data)
    })
    getStatus().then((result) => {
      if (result.ok) setStatus(result.data)
    })
    runDoctor().then((result) => {
      if (result.ok) setDoctor(result.data)
    })
  }, [open])

  const markdown = generateMarkdown(description, info, status, doctor)

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(markdown).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [markdown])

  return (
    <Modal open={open} onClose={onClose} title="Report a problem">
      <div className="space-y-4">
        <div>
          <label
            htmlFor="bug-description"
            className="block text-sm font-medium text-on-surface-secondary mb-1"
          >
            Describe the issue
          </label>
          <textarea
            id="bug-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What happened? What did you expect?"
            className="w-full h-24 px-3 py-2 border border-outline-strong bg-surface-raised text-on-surface placeholder-on-surface-dim rounded-control text-sm resize-none focus:outline-hidden focus:ring-2 focus:ring-accent"
          />
        </div>
        <div>
          <label
            htmlFor="bug-report"
            className="block text-sm font-medium text-on-surface-secondary mb-1"
          >
            Generated report
          </label>
          <textarea
            id="bug-report"
            readOnly
            value={markdown}
            className="w-full h-48 px-3 py-2 border border-outline-strong rounded-control text-xs font-mono bg-surface text-on-surface-secondary resize-none"
          />
        </div>
        <div className="flex justify-end">
          <button
            type="button"
            onClick={handleCopy}
            className="px-4 py-2 text-sm font-medium text-white bg-accent rounded-control hover:bg-accent-hover focus:outline-hidden focus:ring-2 focus:ring-accent"
          >
            {copied ? 'Copied!' : 'Copy to clipboard'}
          </button>
        </div>
      </div>
    </Modal>
  )
}
