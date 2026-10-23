import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { SpeakerHighIcon, SpeakerSlashIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useClient } from '@/state/provider'
import { volumeAdjustable } from '@/fork/controls/styles'
import { MoreTools } from './MoreTools'

// The phone's More sheet, up from the bottom: the sound, then every tool, and Close. Not a modal
// focus trap: the on-screen keyboard's tool hands the focus to the picture's input as it closes the
// sheet. Escape and a tap beside it close it, the focus goes back to the More button.
export function MoreSheet({ onClose, opener }: { onClose: () => void; opener: HTMLElement | null }) {
  const { t } = useTranslation()
  const client = useClient()
  const { muted, volume } = useStore(
    client.store,
    useShallow((s) => ({ muted: s.video.muted || s.video.volume === 0, volume: s.video.volume })),
  )
  const first = useRef<HTMLButtonElement>(null)
  useEffect(() => first.current?.focus(), [])
  const close = () => (onClose(), opener?.focus())

  return createPortal(
    <div
      className="fixed inset-0 z-[1400] flex items-end bg-black/45"
      onClick={(e) => e.target === e.currentTarget && close()}
      onKeyDown={(e) => e.key === 'Escape' && (e.stopPropagation(), close())}
    >
      <div
        className="animate-sheet-up flex max-h-[85vh] w-full flex-col gap-3.5 overflow-y-auto rounded-t-outer border-t border-glass-line bg-sheet px-3.5 pt-2 pb-[calc(14px+env(safe-area-inset-bottom))] shadow-sheet"
        role="dialog"
        aria-modal="true"
        aria-label={t('fork:controls.more')}
        data-testid="more-sheet"
      >
        <div className="mx-auto h-1 w-9 flex-none rounded-full bg-white/20" aria-hidden />
        <div className="flex items-center gap-3">
          <button
            ref={first}
            type="button"
            className={cn(
              'inline-flex size-11 flex-none cursor-pointer items-center justify-center rounded-full border border-glass-border bg-glass text-text-normal',
              muted && 'border-warn/40 text-warn',
            )}
            aria-label={muted ? t('fork:dock.unmute') : t('fork:dock.mute')}
            aria-pressed={muted}
            data-testid="sheet-mute"
            onClick={() => (muted ? client.unmute() : client.mute())}
          >
            {muted ? <SpeakerSlashIcon className="size-4" /> : <SpeakerHighIcon className="size-4" />}
          </button>
          {volumeAdjustable ? (
            <input
              type="range"
              className="slider h-6 min-w-0 flex-1"
              min="0"
              max="100"
              aria-label={t('fork:dock.volume')}
              value={Math.round(volume * 100)}
              onChange={(e) => client.setVolume(Number(e.target.value) / 100)}
            />
          ) : (
            <span className="text-[13px] text-text-muted">{muted ? t('fork:dock.unmute') : t('fork:dock.mute')}</span>
          )}
        </div>
        <MoreTools sheet done={onClose} />
        <button
          type="button"
          className="h-11 flex-none cursor-pointer rounded-inner border border-glass-border bg-glass text-[14px] font-semibold text-pure hover:bg-glass-hover"
          onClick={close}
        >
          {t('fork:controls.close')}
        </button>
      </div>
    </div>,
    document.body,
  )
}
