import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { PauseIcon, PlayIcon, SpeakerHighIcon, SpeakerSlashIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useClient } from '@/state/provider'
import { Emotes } from '@/features/emotes/Emotes'

// a dock button: 44px, glass
export const dockButton =
  'relative inline-flex size-11 flex-none cursor-pointer items-center justify-center rounded-control border border-glass-border bg-glass text-text-normal transition-all duration-300 ease-fluid hover:bg-glass-hover hover:text-pure disabled:cursor-default disabled:opacity-50'

// The phone's dock right under the picture, in place of the room bar: playback and sound for now;
// the control pill and the More sheet join it later (phase 5).
export function MobileDock() {
  const { t } = useTranslation()
  const client = useClient()
  const { playable, playing, muted } = useStore(
    client.store,
    useShallow((s) => ({ playable: s.video.playable, playing: s.video.playing, muted: s.video.muted })),
  )
  return (
    <div
      className="phone:flex hidden flex-none items-center gap-2 border-t border-glass-border bg-bar-room px-2.5 pt-2 pb-[calc(8px+env(safe-area-inset-bottom))]"
      data-testid="mobile-dock"
    >
      <div className="ml-auto flex flex-none items-center gap-1.5">
        <button
          type="button"
          className={dockButton}
          disabled={!playable}
          aria-label={playing ? t('fork:dock.pause') : t('fork:dock.play')}
          onClick={() => (playing ? client.pause() : client.play().catch(() => {}))}
        >
          {playing ? <PauseIcon className="size-4" /> : <PlayIcon className="size-4" />}
        </button>
        <button
          type="button"
          className={cn(dockButton, muted && 'text-warn hover:text-warn')}
          aria-label={muted ? t('fork:dock.unmute') : t('fork:dock.mute')}
          aria-pressed={muted}
          onClick={() => (muted ? client.unmute() : client.mute())}
        >
          {muted ? <SpeakerSlashIcon className="size-4" /> : <SpeakerHighIcon className="size-4" />}
        </button>
        <Emotes />
      </div>
    </div>
  )
}
