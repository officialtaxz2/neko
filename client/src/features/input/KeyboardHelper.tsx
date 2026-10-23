import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { useStore } from 'zustand'
import { selectControlling } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import {
  ArrowCounterClockwiseIcon,
  ArrowDownIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ArrowUpIcon,
  ArrowUUpLeftIcon,
  ClipboardTextIcon,
  CopyIcon,
  SelectionAllIcon,
  SlidersHorizontalIcon,
  XIcon,
} from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useNeko } from '@/state/provider'
import { MODIFIERS, specialKeys, XK, type Modifier, type Shortcut } from './keys'
import { useTouch } from './useTouch'

const key =
  'inline-flex h-8 min-w-11 flex-none cursor-pointer items-center justify-center gap-1 rounded-button border border-glass-border bg-glass px-2 text-[10px] font-semibold text-text-normal transition-all duration-300 ease-fluid active:scale-95'
const CAPS: Record<Modifier, string> = { ctrl: 'CTRL', alt: 'ALT', shift: 'SHIFT', super: 'WIN' }
const SHORTCUT_ICONS: Record<Shortcut, ReactNode> = {
  all: <SelectionAllIcon className="size-3" />,
  copy: <CopyIcon className="size-3" />,
  paste: <ClipboardTextIcon className="size-3" />,
  undo: <ArrowUUpLeftIcon className="size-3" />,
}

// a press on the bar keeps the focus where it is, so the on-screen keyboard stays open
const keepFocus = (e: { preventDefault: () => void }) => e.preventDefault()

// the visible part of the page: the bar sits on the on-screen keyboard, not behind it
function useViewport() {
  const read = () => {
    const vv = window.visualViewport
    if (!vv || window.self !== window.top) return { left: 0, bottom: 0, width: undefined }
    return { left: vv.offsetLeft, bottom: Math.max(0, innerHeight - (vv.offsetTop + vv.height)), width: vv.width }
  }
  const [box, setBox] = useState(read)
  useEffect(() => {
    const vv = window.visualViewport
    if (!vv) return
    const update = () => setBox(read())
    vv.addEventListener('resize', update)
    vv.addEventListener('scroll', update)
    return () => {
      vv.removeEventListener('resize', update)
      vv.removeEventListener('scroll', update)
    }
  }, [])
  return box
}

// the fork: the special keys a phone's keyboard lacks (modifiers held until tapped again, Esc,
// Tab, Ctrl shortcuts, arrows), in a bar over the on-screen keyboard. For the host on a touch
// device; opened from the picture's buttons, gone (and every held key let go) when control goes.
export function KeyboardHelper() {
  const { app, client } = useNeko()
  const touch = useTouch()
  const open = useStore(app, (s) => s.keyboardHelper)
  const hosting = useStore(client.store, selectControlling)
  useEffect(() => {
    if (!hosting) app.setState({ keyboardHelper: false })
  }, [hosting, app])
  return open && hosting && touch ? <Bar /> : null
}

function Bar() {
  const { t } = useTranslation()
  const { app, client } = useNeko()
  const keys = useMemo(() => specialKeys(client), [client])
  const [held, setHeld] = useState(keys.held)
  const box = useViewport()
  useEffect(() => () => keys.reset(), [keys])

  const toggle = (m: Modifier) => (keys.toggle(m), setHeld(keys.held()))
  const reset = () => (keys.reset(), setHeld(keys.held()))
  const close = () => app.setState({ keyboardHelper: false })
  const button = (label: string, run: () => void, child: ReactNode, cls?: string | false, pressed?: boolean) => (
    <button
      type="button"
      className={cn(key, cls)}
      aria-label={label}
      aria-pressed={pressed}
      onMouseDown={keepFocus}
      onClick={run}
    >
      {child}
    </button>
  )
  const divider = <span className="h-5 w-px flex-none self-center bg-glass-line" />

  return createPortal(
    <div
      className="fixed z-50 box-border flex max-w-[100vw] flex-col gap-2 border-t border-glass-line bg-bar-room/90 px-3.5 py-2.5 shadow-bar-up backdrop-blur-xl select-none"
      style={{ left: box.left, bottom: box.bottom, width: box.width ?? '100%' }}
      data-testid="keyboard-helper"
      role="toolbar"
      aria-label={t('fork:input.keyboard_helper')}
    >
      <div className="flex items-center justify-between border-b border-glass-border pb-1.5">
        <span className="flex items-center gap-1.5 text-[11px] font-semibold tracking-wide text-pure uppercase">
          <SlidersHorizontalIcon className="size-3.5 text-mint" />
          {t('fork:input.keyboard_helper')}
        </span>
        <div className="flex items-center gap-3">
          {button(
            t('fork:input.reset_keys'),
            reset,
            <ArrowCounterClockwiseIcon className="size-3.5" />,
            'min-w-0 border-0 bg-transparent px-1.5 text-text-muted hover:text-mint',
          )}
          {button(
            t('fork:input.close'),
            close,
            <XIcon className="size-3.5" />,
            'min-w-0 border-0 bg-transparent px-1.5 text-text-muted hover:text-pure',
          )}
        </div>
      </div>
      <div className="flex w-full items-center gap-1.5 overflow-x-auto py-0.5 [scrollbar-width:none]">
        {(Object.keys(MODIFIERS) as Modifier[]).map((m) => (
          <span key={m} className="contents">
            {button(
              CAPS[m],
              () => toggle(m),
              CAPS[m],
              held.has(m) && 'border-mint-active bg-mint text-mint-ink shadow-mint-glow',
              held.has(m),
            )}
          </span>
        ))}
        {divider}
        {button('ESC', () => keys.tap(XK.Escape), 'ESC', 'bg-white/5 text-pure')}
        {button('TAB', () => keys.tap(XK.Tab), 'TAB', 'bg-white/5 text-pure')}
      </div>
      <div className="flex w-full flex-wrap items-center gap-1.5 py-0.5">
        {(Object.keys(SHORTCUT_ICONS) as Shortcut[]).map((s) => (
          <span key={s} className="contents">
            {button(
              t(`fork:input.${s}`),
              () => keys.shortcut(s),
              <>
                {SHORTCUT_ICONS[s]}
                {t(`fork:input.${s}`)}
              </>,
              'border-mint/15 bg-mint/5 text-mint hover:border-mint hover:bg-mint/10',
            )}
          </span>
        ))}
        {/* on a narrow phone the arrows wrap onto a line of their own rather than scroll out of sight */}
        <div className="ml-auto flex flex-none items-center gap-1">
          {button(
            t('fork:input.left'),
            () => keys.tap(XK.Left),
            <ArrowLeftIcon className="size-3" />,
            'size-8 min-w-8 p-0',
          )}
          <div className="flex flex-col gap-0.5">
            {button(
              t('fork:input.up'),
              () => keys.tap(XK.Up),
              <ArrowUpIcon className="size-2.5" />,
              'h-[15px] w-8 min-w-8 rounded-b-none p-0',
            )}
            {button(
              t('fork:input.down'),
              () => keys.tap(XK.Down),
              <ArrowDownIcon className="size-2.5" />,
              'h-[15px] w-8 min-w-8 rounded-t-none p-0',
            )}
          </div>
          {button(
            t('fork:input.right'),
            () => keys.tap(XK.Right),
            <ArrowRightIcon className="size-3" />,
            'size-8 min-w-8 p-0',
          )}
        </div>
      </div>
    </div>,
    document.body,
  )
}
