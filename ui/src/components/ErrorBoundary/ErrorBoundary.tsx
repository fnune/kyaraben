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
        <div className="min-h-screen bg-red-50 p-8">
          <div className="max-w-3xl mx-auto">
            <h1 className="text-2xl font-bold text-red-800 mb-4">Something went wrong</h1>

            <div className="bg-white border border-red-200 rounded-lg p-4 mb-4">
              <h2 className="font-semibold text-red-700 mb-2">Error message</h2>
              <pre className="text-sm text-red-600 whitespace-pre-wrap break-words">
                {error?.message}
              </pre>
            </div>

            {error?.stack && (
              <details className="bg-white border border-red-200 rounded-lg p-4 mb-4">
                <summary className="font-semibold text-red-700 cursor-pointer">Stack trace</summary>
                <pre className="mt-2 text-xs text-gray-700 whitespace-pre-wrap break-words overflow-auto max-h-64">
                  {error.stack}
                </pre>
              </details>
            )}

            {errorInfo?.componentStack && (
              <details className="bg-white border border-red-200 rounded-lg p-4 mb-4">
                <summary className="font-semibold text-red-700 cursor-pointer">
                  Component stack
                </summary>
                <pre className="mt-2 text-xs text-gray-700 whitespace-pre-wrap break-words overflow-auto max-h-64">
                  {errorInfo.componentStack}
                </pre>
              </details>
            )}

            <div className="flex gap-3">
              <button
                type="button"
                onClick={this.handleReload}
                className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
              >
                Reload app
              </button>
              <button
                type="button"
                onClick={this.handleCopyError}
                className="px-4 py-2 border border-red-300 text-red-700 rounded-md hover:bg-red-100"
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
