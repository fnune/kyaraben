import { collapseTilde, expandTilde } from '@/lib/paths'

export function PathText({ children }: { readonly children: string }) {
  const homeDir = window.electron.homeDir
  const displayPath = collapseTilde(children, homeDir)
  const fullPath = expandTilde(children, homeDir)

  return (
    <span className="font-mono" title={fullPath}>
      {displayPath}
    </span>
  )
}
