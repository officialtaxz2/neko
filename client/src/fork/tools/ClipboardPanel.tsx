import { useEffect, useRef, useState } from 'react'
import * as PopoverPrimitive from '@radix-ui/react-popover'
import { useStore } from 'zustand'
import { selectControlling } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { CheckIcon, ClockCounterClockwiseIcon, CopyIcon, InfoIcon, TrashIcon, XIcon } from '@phosphor-icons/react'
import { useNeko } from '@/state/provider'
import { clearClipboard, clipboardStore, copyClipboard, remember } from '@/state/clipboard'

const SEND_AFTER_MS = 500 // typing: the remote clipboard gets the text once the typing pauses
const truncate = (text: string, n = 45) => (text.length <= n ? text : text.slice(0, n) + '…')

// The clipboard's panel (More → Clipboard), for the host: the remote clipboard as a text field to
// read and edit, and the recent texts, each copied back to both clipboards with one tap. Above the
// room bar's right end. Modal, so the picture's input cannot take the focus (and the keys) while the
// mouse crosses it; Escape, a click elsewhere or losing control closes it.
export function ClipboardPanel() {
  const { app, client } = useNeko()
  const open = useStore(app, (s) => s.clipboardPanel)
  const hosting = useStore(client.store, selectControlling)
  useEffect(() => {
    if (!hosting) app.setState({ clipboardPanel: false })
  }, [hosting, app])
  return (
    <PopoverPrimitive.Root open={open && hosting} onOpenChange={(o) => app.setState({ clipboardPanel: o })} modal>
      <PopoverPrimitive.Anchor className="pointer-events-none fixed right-5 bottom-[78px] size-0 max-[768px]:right-3 max-[768px]:bottom-[68px]" />
      {open && hosting && <Panel />}
    </PopoverPrimitive.Root>
  )
}

function Panel() {
  const { t } = useTranslation()
  const neko = useNeko()
  const { app, client } = neko
  const remote = useStore(client.store, (s) => s.control.clipboard?.text ?? '')
  const history = useStore(clipboardStore(neko), (s) => s.history)
  const [text, setText] = useState(remote)
  const [copied, setCopied] = useState('')
  useEffect(() => setText(remote), [remote]) // follows what the remote desktop copies
  const typing = useRef(0)
  const field = useRef<HTMLTextAreaElement>(null)
  const close = () => app.setState({ clipboardPanel: false })
  useEffect(() => () => clearTimeout(typing.current), [])

  const copy = async (item: string) => {
    await copyClipboard(neko, item)
    setCopied(item)
    setTimeout(() => setCopied((c) => (c === item ? '' : c)), 1500)
  }

  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content
        side="top"
        align="end"
        sideOffset={10}
        collisionPadding={12}
        className="z-[1300] flex max-h-[min(420px,calc(100vh-120px))] w-[340px] max-w-[calc(100vw-24px)] flex-col gap-2.5 rounded-outer border border-glass-line bg-sheet p-3 shadow-sheet outline-hidden"
        aria-label={t('fork:clipboard.title')}
        data-testid="clipboard-panel"
        onOpenAutoFocus={(e) => (e.preventDefault(), field.current?.focus())}
      >
        <div className="flex items-center justify-between border-b border-glass-border pb-2">
          <span className="flex items-center gap-1.5 text-[12px] font-semibold tracking-wide text-pure uppercase">
            <ClockCounterClockwiseIcon className="size-4 text-mint" />
            {t('fork:clipboard.title')}
          </span>
          <button
            type="button"
            className="cursor-pointer rounded-button p-1 text-text-muted hover:text-pure"
            aria-label={t('fork:controls.close')}
            onClick={close}
          >
            <XIcon className="size-4" />
          </button>
        </div>
        <textarea
          ref={field}
          className="h-[84px] w-full flex-none resize-none rounded-button border border-glass-border bg-white/[0.03] p-2 text-[13px] text-text-normal outline-hidden placeholder:text-text-muted focus:border-mint/50"
          aria-label={t('fork:controls.clipboard')}
          placeholder={t('fork:clipboard.placeholder')}
          value={text}
          data-testid="clipboard-text"
          onFocus={(e) => e.target.select()}
          onChange={(e) => {
            const value = e.target.value
            setText(value)
            clearTimeout(typing.current)
            typing.current = window.setTimeout(() => {
              client.send('clipboard/set', { text: value })
              remember(neko, value)
            }, SEND_AFTER_MS)
          }}
        />
        <div className="flex min-h-0 flex-col gap-1.5">
          {history.length > 0 && (
            <div className="flex items-center justify-between text-[11.5px] text-text-muted">
              <span>{t('fork:clipboard.items', { count: history.length })}</span>
              <button
                type="button"
                className="inline-flex cursor-pointer items-center gap-1 rounded-button px-1.5 py-0.5 hover:text-badge"
                data-testid="clipboard-clear"
                onClick={() => clearClipboard(neko)}
              >
                <TrashIcon className="size-3.5" />
                {t('fork:clipboard.clear')}
              </button>
            </div>
          )}
          {history.length === 0 ? (
            <p className="py-3 text-center text-[12.5px] text-text-muted">{t('fork:clipboard.empty')}</p>
          ) : (
            <ul className="flex min-h-0 flex-col gap-1 overflow-y-auto" data-testid="clipboard-history">
              {history.map((item) => (
                <li
                  key={item}
                  className="flex items-center gap-2 rounded-button border border-glass-border bg-glass py-1 pr-1 pl-2.5"
                >
                  <span className="min-w-0 flex-1 truncate text-[12.5px] text-text-normal" title={item}>
                    {truncate(item)}
                  </span>
                  <button
                    type="button"
                    className="inline-flex size-8 flex-none cursor-pointer items-center justify-center rounded-button text-text-muted hover:bg-glass-hover hover:text-pure"
                    aria-label={t('fork:clipboard.copy')}
                    title={copied === item ? t('fork:clipboard.copied') : t('fork:clipboard.copy')}
                    onClick={() => void copy(item)}
                  >
                    {copied === item ? <CheckIcon className="size-4 text-mint" /> : <CopyIcon className="size-4" />}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <p className="flex gap-1.5 border-t border-glass-border pt-2 text-[11.5px] leading-snug text-text-muted">
          <InfoIcon className="mt-px size-3.5 flex-none" />
          {t('fork:clipboard.hint')}
        </p>
      </PopoverPrimitive.Content>
    </PopoverPrimitive.Portal>
  )
}
