import type { BackendApp, WailsRuntime } from './types'

export function appApi(): BackendApp {
  const app = window.go?.main?.App
  if (!app) {
    throw new Error('EasyRenamer backend is not ready')
  }
  return app
}

export function runtimeApi(): WailsRuntime {
  const runtime = window.runtime
  if (!runtime) {
    throw new Error('Wails runtime is not ready')
  }
  return runtime
}
