export function expandTilde(path: string, homeDir: string): string {
  if (!homeDir) return path
  if (path.startsWith('~/')) {
    return homeDir + path.slice(1)
  }
  if (path === '~') {
    return homeDir
  }
  return path
}

export function collapseTilde(path: string, homeDir: string): string {
  if (!homeDir) return path
  if (path.startsWith(`${homeDir}/`)) {
    return `~${path.slice(homeDir.length)}`
  }
  if (path === homeDir) {
    return '~'
  }
  return path
}
