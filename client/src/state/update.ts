// The fork: a tab left open across a deploy keeps running the client it loaded, against the new
// server, and the files it loads later (a language, the HLS player) are gone. The page compares the
// hashed names of its built files with those of the index.html the server has now; when they differ
// it loads again where that interrupts nothing (no connection, or one that just came back after an
// interruption, such as the deploy's restart) and otherwise offers it in a toast. A new tab or a
// reload needs none of this: the server has every load revalidate index.html (no-cache, ETag).
import { toast as sonner } from 'sonner'
import { HLSTransport } from '@m1k1o/neko'
import { t } from '@/i18n'
import type { NekoApp } from './app'

// a version that turns up this soon after the connection was interrupted reloads even when the
// connection is back already (a share link and a kept session reconnect by themselves)
export const RECENT_MS = 60_000
// while the connection is being re-established, how often the server's index.html is asked for
export const RETRY_MS = 5_000
// coming back to the tab asks at most this often
export const VISIBLE_MS = 60_000
// the version a reload was made for: reaching it again means the reload did not bring it (a stale
// copy on the way), so only the toast is offered then
export const RELOAD_KEY = 'update_reload'

// the built files an HTML page names (entry script, preloads, stylesheet): the paths under assets/,
// whose names carry their content's hash; empty when the page names none
export function builtFiles(urls: Iterable<string>, base: string): string[] {
  const paths = [...urls].map((u) => new URL(u, base).pathname).filter((p) => p.includes('/assets/'))
  return [...new Set(paths)].sort()
}

// the src and href of an index.html's script and link tags
export const htmlURLs = (html: string) =>
  [...html.matchAll(/<(?:script|link)\b[^>]*?\b(?:src|href)=["']([^"']+)["']/gi)].map((m) => m[1])

// the page's own: read before anything loads lazily, since a lazy import adds its own preload links
export const pageFiles = (doc: Document = document) =>
  builtFiles(
    [...doc.querySelectorAll('script[src], link[href]')].map(
      (el) => el.getAttribute('src') ?? el.getAttribute('href')!,
    ),
    doc.baseURI,
  )

export const sameFiles = (a: string[], b: string[]) => a.length === b.length && a.every((f, i) => f === b[i])

// what the server serves at the page's address now; undefined for anything that is not this
// client's index.html (an error, a redirect such as OAuth's login, a page without built files)
export async function servedFiles(url: string, fetcher: typeof fetch = fetch): Promise<string[] | undefined> {
  const res = await fetcher(url, { cache: 'no-cache', credentials: 'same-origin' })
  if (res.status !== 200 || res.redirected || !res.headers.get('content-type')?.includes('text/html')) return
  const files = builtFiles(htmlURLs(await res.text()), url)
  return files.length ? files : undefined
}

export interface Moment {
  status: 'connected' | 'connecting' | 'disconnected'
  // when the connection was last not connected (0: never)
  interruptedAt: number
  now: number
  // someone is writing in a field: the login's name, a chat message
  typing: boolean
  // a file of the old build could not be loaded: what needed it does not work
  loadFailed: boolean
}

// what a new version does: load it, or offer it
export function onNewVersion(m: Moment): 'reload' | 'offer' {
  if (m.typing) return 'offer'
  if (m.loadFailed || m.status !== 'connected') return 'reload'
  return m.now - m.interruptedAt < RECENT_MS ? 'reload' : 'offer'
}

type Target = Pick<EventTarget, 'addEventListener'>
export interface UpdateEnv {
  url: string
  fetch: typeof fetch
  reload: () => void
  storage: Pick<Storage, 'getItem' | 'setItem'> | undefined
  window: Target
  document: Target & { visibilityState: DocumentVisibilityState; activeElement: Element | null }
  now: () => number
  setInterval: (fn: () => void, ms: number) => number
  clearInterval: (id: number) => void
}

const browserEnv = (): UpdateEnv => ({
  url: new URL('./', document.baseURI).href,
  fetch: (...args) => fetch(...args),
  reload: () => location.reload(),
  storage: (() => {
    try {
      return sessionStorage
    } catch {
      return undefined
    }
  })(),
  window,
  document,
  now: Date.now,
  setInterval: (fn, ms) => window.setInterval(fn, ms),
  clearInterval: (id) => window.clearInterval(id),
})

const TEXT_INPUTS = new Set(['text', 'password', 'search', 'email', 'url', 'tel', 'number'])
// a focused field with text in it; the login focuses its empty name field by itself
const isTyping = (el: Element | null) => {
  if (!el) return false
  if (el.tagName === 'TEXTAREA') return (el as HTMLTextAreaElement).value !== ''
  if (el.tagName === 'INPUT')
    return TEXT_INPUTS.has((el as HTMLInputElement).type) && (el as HTMLInputElement).value !== ''
  return (el as HTMLElement).isContentEditable === true && !!el.textContent
}

// start-up (main.tsx, built pages only): `own` is pageFiles() from before the first lazy import
export function bindUpdates(neko: NekoApp, own: string[], env: UpdateEnv = browserEnv()) {
  // a page without built files has nothing to compare (every answer would differ)
  if (!own.length) return
  const { client } = neko
  let interruptedAt = 0
  let lastVisible = env.now()
  let busy = false
  // a check asked for while one runs: it runs once that one is done (with its loadFailed)
  let queued: boolean | undefined
  let retry: number | undefined

  const reload = () => {
    client.disconnect()
    env.reload()
  }
  const offer = () => {
    sonner.info(t('fork:update.title'), {
      id: 'update',
      duration: Infinity,
      action: { label: t('fork:update.reload'), onClick: reload },
    })
    return false
  }

  const act = (files: string[], loadFailed: boolean) => {
    const moment = {
      status: client.state.connection.status,
      interruptedAt,
      now: env.now(),
      typing: isTyping(env.document.activeElement),
      loadFailed,
    }
    if (onNewVersion(moment) === 'offer') return offer()
    const version = files.join(' ')
    // without the storage a reload could repeat for ever on a stale copy: offer it then
    try {
      if (!env.storage || env.storage.getItem(RELOAD_KEY) === version) return offer()
      env.storage.setItem(RELOAD_KEY, version)
    } catch {
      return offer()
    }
    reload()
    return true
  }

  const check = async (loadFailed = false): Promise<void> => {
    if (busy) {
      queued = !!queued || loadFailed
      return
    }
    busy = true
    let reloading = false
    try {
      const files = await servedFiles(env.url, env.fetch)
      if (files && !sameFiles(files, own)) reloading = act(files, loadFailed)
    } catch {
      // the server is not there (the deploy is still running): the next trigger asks again
    } finally {
      busy = false
    }
    const next = queued
    queued = undefined
    if (next !== undefined && !reloading) return check(next)
  }

  client.events.on('connection.status', (status) => {
    if (status !== 'connected') interruptedAt = env.now()
    if (retry !== undefined) env.clearInterval(retry)
    retry = status === 'connecting' ? env.setInterval(() => void check(), RETRY_MS) : undefined
    void check()
  })
  env.document.addEventListener('visibilitychange', () => {
    if (env.document.visibilityState !== 'visible' || env.now() - lastVisible < VISIBLE_MS) return
    lastVisible = env.now()
    void check()
  })
  env.window.addEventListener('pageshow', (e) => (e as PageTransitionEvent).persisted && void check())
  // Vite's lazy imports (a language, the HLS player); the error itself goes on as before
  env.window.addEventListener('vite:preloadError', () => void check(true))
  // the HLS worker is loaded by its URL, outside Vite's helper: a player that ends in failure asks too
  const transport = client.transport
  if (transport instanceof HLSTransport) transport.on('hls', (hls) => hls.status === 'terminal' && void check(true))
}
