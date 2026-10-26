import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { initI18n } from './i18n'
import { plugins, initPlugins } from './plugins'
import { createNekoApp } from './state/neko'
import { NekoProvider } from './state/provider'
import { App } from './app/App'
import { ErrorBoundary } from './app/ErrorBoundary'
import { bindUpdates, pageFiles } from './state/update'

// start-up, in the order the pieces need: the strings of the active language (the plugins'
// namespaces included), the instance (the client with its event wiring that turns server events
// into event lines and toasts, and the settings, also from the URL), the plugins (slots, event
// dispatch, their own wiring), then, once the strings are there (event lines, toasts and the first
// render use them), the connection: resume a saved session (autologin) and connect
// the fork: the built files this page runs, read before the first lazy import (state/update.ts)
const build = pageFiles()
// `fork`: the strings of the fork's own GUI (src/fork)
const ready = initI18n(['fork', ...plugins.map((p) => p.ns)])
const neko = createNekoApp()
initPlugins(neko)
// the fork: a deploy reaches an open tab too (not in `npm run dev`, which has no built files)
if (import.meta.env.PROD) bindUpdates(neko, build)
ready.then(() => {
  neko.client.setUrl(location.href)
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ErrorBoundary>
        <NekoProvider neko={neko}>
          <App />
        </NekoProvider>
      </ErrorBoundary>
    </StrictMode>,
  )
})
