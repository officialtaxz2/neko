// modal dialogs (legacy SweetAlert look, components/Dialog) and toasts (sonner, components/ui/sonner)
import { toast as sonner } from 'sonner'
import type { AppStore } from './app'

export interface Dialog {
  title: string
  text?: string
  icon: 'warning' | 'error' | 'info' | 'question'
  cancel: boolean
  // the fork: the confirm button's own word ("Log out") instead of "Yes"
  confirm?: string
  resolve: (ok: boolean) => void
}

// ask() resolves true on confirm, tell() when dismissed
function dialog(app: AppStore, d: Omit<Dialog, 'resolve'>) {
  app.getState().dialog?.resolve(false)
  return new Promise<boolean>((resolve) => app.setState({ dialog: { ...d, resolve } }))
}
export const ask = (app: AppStore, title: string, text?: string, confirm?: string) =>
  dialog(app, { title, text, icon: confirm ? 'question' : 'warning', cancel: true, confirm })
export const tell = (app: AppStore, title: string, text?: string, icon: Dialog['icon'] = 'error') =>
  dialog(app, { title, text, icon, cancel: false })

export type ToastKind = 'info' | 'success' | 'warning' | 'error'
// a toast of a kind for 5 s, like the legacy client's
export function toast(title: string, text?: string, kind: ToastKind = 'info') {
  sonner[kind](title, { description: text, duration: 5000 })
}
// the fork: the offer of a new version (state/update.ts) stays until it is used; the connection's
// own toasts clear the rest when it is back
export const UPDATE_TOAST = 'update'
export const dismissToasts = () => sonner.getToasts().forEach((t) => t.id !== UPDATE_TOAST && sonner.dismiss(t.id))
