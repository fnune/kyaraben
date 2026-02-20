import { createContext, type ReactNode, useContext, useEffect, useState } from 'react'

const HomeDirContext = createContext<string>('')

export function HomeDirProvider({ children }: { readonly children: ReactNode }) {
  const [homeDir, setHomeDir] = useState('')

  useEffect(() => {
    window.electron.invoke<string>('get_home_dir').then(setHomeDir)
  }, [])

  return <HomeDirContext.Provider value={homeDir}>{children}</HomeDirContext.Provider>
}

export function useHomeDir(): string {
  return useContext(HomeDirContext)
}
