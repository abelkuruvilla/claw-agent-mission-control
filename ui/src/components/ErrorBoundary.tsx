'use client';
import { Component, ReactNode } from 'react';

interface Props { children: ReactNode; }
interface State { hasError: boolean; error: Error | null; }

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: { componentStack: string }) {
    console.error('[ErrorBoundary] Caught:', error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-[#0d1117] text-white p-6 text-center">
          <h1 className="text-2xl font-bold text-red-400 mb-2">Something went wrong</h1>
          <p className="text-slate-400 mb-4">A client-side error occurred. Please refresh the page.</p>
          <pre className="text-xs text-slate-500 bg-[#161b22] p-3 rounded max-w-lg overflow-auto text-left">
            {this.state.error?.message}
          </pre>
          <button
            onClick={() => window.location.reload()}
            className="mt-6 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded text-white text-sm"
          >
            Reload Page
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
