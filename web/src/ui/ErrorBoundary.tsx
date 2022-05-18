import React from 'react';

type ErrorBoundaryWithRetryFallback = (
  error: Error,
  retry: () => void
) => React.ReactNode;

interface ErrorBoundaryWithRetryProps {
  fallback: React.ReactNode | ErrorBoundaryWithRetryFallback;
}

interface ErrorBoundaryWithRetryState {
  error: Error | null;
}

export class ErrorBoundaryWithRetry extends React.Component<
  ErrorBoundaryWithRetryProps,
  ErrorBoundaryWithRetryState
> {
  state: ErrorBoundaryWithRetryState = {error: null};

  static getDerivedStateFromError(error: Error) {
    return {error: error};
  }

  // componentDidCatch(error: Error, errorInfo: ErrorInfo) {
  //     // You can also log the error to an error reporting service
  // }

  _retry = () => {
    this.setState({error: null});
  };

  render() {
    const {children, fallback} = this.props;
    const {error} = this.state;
    if (error) {
      if (typeof fallback === 'function') {
        return fallback(error, this._retry);
      }
      return fallback;
    }

    return children;
  }
}
