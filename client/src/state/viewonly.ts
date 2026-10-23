// View-only share links (docs/VIEW_ONLY_SHARING.md): `https://host/#/<16-character token>` logs in a
// passive session, with the token as the password of POST /api/login (in the body: the fragment
// never reaches a request line, a proxy log or a Referer). The server enforces what the session may
// do; the GUI only shows the picture and the sound.
//
// Its session token is kept apart from a member's: in sessionStorage under its own key, so a member's
// saved session in the same browser stays untouched, and a reload of the tab resumes the session
// instead of logging in again. Leaving the page logs the session out (a request that outlives the
// page), so ten reloads leave one session in the room, not ten.
import { ApiError, type NekoClient } from '@m1k1o/neko'
import { t } from '@/i18n'
import { actions } from './actions'
import { scoped, type NekoApp } from './app'
import { tell } from './dialogs'
import { get, set } from './storage'

export const SHARE_TOKEN_KEY = 'neko_view_session'

// the token of a share link in this URL, or undefined
export function shareToken(href: string = location.href): string | undefined {
  try {
    return /^#\/([A-Za-z0-9_-]{16})$/.exec(new URL(href).hash)?.[1]
  } catch {
    return undefined
  }
}

// sessionStorage where the browser allows it; a private window or blocked storage keeps the token in
// memory for this page only. One per page: the client writes it, the login screen asks it whether
// there is a session to resume
let tab: Pick<Storage, 'getItem' | 'setItem' | 'removeItem'> | undefined
export const tabStorage = () => (tab ??= makeTabStorage())
// the share session this tab can resume, if any
export const savedShareSession = () => !!tabStorage().getItem(SHARE_TOKEN_KEY)

function makeTabStorage(): Pick<Storage, 'getItem' | 'setItem' | 'removeItem'> {
  const memory = new Map<string, string>()
  const store = (() => {
    try {
      return globalThis.sessionStorage
    } catch {
      return undefined
    }
  })()
  return {
    getItem: (k) => {
      try {
        return store ? store.getItem(k) : (memory.get(k) ?? null)
      } catch {
        return memory.get(k) ?? null
      }
    },
    setItem: (k, v) => {
      memory.set(k, v)
      try {
        store?.setItem(k, v)
      } catch {}
    },
    removeItem: (k) => {
      memory.delete(k)
      try {
        store?.removeItem(k)
      } catch {}
    },
  }
}

type PageTarget = Pick<EventTarget, 'addEventListener' | 'removeEventListener'>

// a share session ends with the page: closing the tab, navigating away or reloading logs it out; a
// page kept in the back/forward cache only disconnects (lifecycle.ts) and resumes the same session
export function bindShareLeave(client: NekoClient, target: PageTarget | undefined = globalThis): () => void {
  if (typeof target?.addEventListener !== 'function') return () => {}
  const hide = (e: Event) => {
    if ((e as PageTransitionEvent).persisted || !client.state.authenticated) return
    client.logout({ keepalive: true }).catch(() => {})
  }
  target.addEventListener('pagehide', hide)
  return () => target.removeEventListener('pagehide', hide)
}

// the name a viewer watched under, offered (and used for the automatic login) on the next visit; not
// `displayname`, which start-up clears as the Vue client's stored login
export const rememberedName = () => get<string>('view_name', '')
export const rememberName = (name: string) => set('view_name', name)

// log in as a viewer under this name with the link's token; a refused link is said as such
export function watchShare(neko: NekoApp, name: string) {
  rememberName(name)
  return actions(neko)
    .login(name, neko.share!)
    .catch((err) => {
      // shown, not awaited: the login screen is usable again while the message is up
      void tell(
        neko.app,
        t('connect.error'),
        err instanceof ApiError && err.status === 401 ? t('fork:share.invalid') : err.message,
      )
    })
}

// the automatic start happens once per page: after a kick or a logout the viewer decides
const started = scoped(() => ({ done: false }))

// the start of a share link page: resume the tab's session (a reload), else log in under the name
// from ?usr= or the last visit; without a name the login screen asks. Only the tab's own share
// session is resumed, never a member's session from a cookie. Null when there is nothing to start
// (it ran already on this page, or the viewer is logged in).
export function startShare(neko: NekoApp): Promise<void> | null {
  const once = started(neko)
  if (once.done || neko.client.state.authenticated) return null
  once.done = true
  const login = async () => {
    const name = new URL(location.href).searchParams.get('usr') || rememberedName()
    if (name) await watchShare(neko, name)
  }
  if (!savedShareSession()) return login()
  return neko.client.authenticate().then(() => neko.client.connect(), login)
}
