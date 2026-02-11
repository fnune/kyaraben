import { Component, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: React.ErrorInfo | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({ errorInfo })
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  handleReload = () => {
    window.location.reload()
  }

  handleCopyError = () => {
    const { error, errorInfo } = this.state
    const errorText = [
      `Error: ${error?.message}`,
      '',
      'Stack:',
      error?.stack,
      '',
      'Component stack:',
      errorInfo?.componentStack,
    ].join('\n')

    navigator.clipboard.writeText(errorText)
  }

  render() {
    if (this.state.hasError) {
      const { error, errorInfo } = this.state

      return (
        <div className="min-h-screen bg-status-error/5 p-8">
          <div className="max-w-3xl mx-auto">
            <h1 className="text-2xl font-bold text-status-error mb-4">Something went wrong</h1>

            <div className="bg-surface border border-status-error/30 rounded-card p-4 mb-4">
              <h2 className="font-semibold text-status-error mb-2">Error message</h2>
              <pre className="text-sm text-status-error/80 whitespace-pre-wrap wrap-break-word">
                {error?.message}
              </pre>
            </div>

            {error?.stack && (
              <details className="bg-surface border border-status-error/30 rounded-card p-4 mb-4">
                <summary className="font-semibold text-status-error cursor-pointer">
                  Stack trace
                </summary>
                <pre className="mt-2 text-xs text-on-surface-muted whitespace-pre-wrap wrap-break-word overflow-auto max-h-64">
                  {error.stack}
                </pre>
              </details>
            )}

            {errorInfo?.componentStack && (
              <details className="bg-surface border border-status-error/30 rounded-card p-4 mb-4">
                <summary className="font-semibold text-status-error cursor-pointer">
                  Component stack
                </summary>
                <pre className="mt-2 text-xs text-on-surface-muted whitespace-pre-wrap wrap-break-word overflow-auto max-h-64">
                  {errorInfo.componentStack}
                </pre>
              </details>
            )}

            <div className="flex gap-3">
              <button
                type="button"
                onClick={this.handleReload}
                className="px-4 py-2 bg-status-error text-white rounded-control hover:bg-status-error/90"
              >
                Reload app
              </button>
              <button
                type="button"
                onClick={this.handleCopyError}
                className="px-4 py-2 border border-status-error/30 text-status-error rounded-control hover:bg-status-error/5"
              >
                Copy error details
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
