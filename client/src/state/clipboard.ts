// the fork: the clipboard's recent texts (fork/tools/ClipboardPanel), newest first, at most 10,
// for this page only: what the remote desktop copied and what was typed into the panel
import { createStore } from 'zustand/vanilla'
import { scoped, type NekoApp } from './app'

export const CLIPBOARD_HISTORY = 10

export const clipboardStore = scoped((neko: NekoApp) => {
  const store = createStore(() => ({ history: [] as string[] }))
  neko.client.store.subscribe(
    (s) => s.control.clipboard?.text ?? '',
    (text) => remember(neko, text),
  )
  neko.bus.on('logout', () => store.setState({ history: [] }))
  return store
})

export function remember(neko: NekoApp, text: string) {
  if (!text.trim()) return
  const store = clipboardStore(neko)
  const history = [text, ...store.getState().history.filter((t) => t !== text)].slice(0, CLIPBOARD_HISTORY)
  store.setState({ history })
}

export const clearClipboard = (neko: NekoApp) => clipboardStore(neko).setState({ history: [] })

// the remote clipboard gets the text (Ctrl+V on the remote pastes it), and the local one too where
// the browser allows it (https)
export async function copyClipboard(neko: NekoApp, text: string): Promise<boolean> {
  neko.client.send('clipboard/set', { text })
  remember(neko, text)
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}
