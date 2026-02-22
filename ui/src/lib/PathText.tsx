import { useHomeDir } from '@/lib/HomeDirContext'
import { collapseTilde } from '@/lib/paths'

export function PathText({ children }: { readonly children: string }) {
  const homeDir = useHomeDir()
  const displayPath = collapseTilde(children, homeDir)

  return (
    <span className="font-mono" title={children}>
      {displayPath}
    </span>
  )
}
