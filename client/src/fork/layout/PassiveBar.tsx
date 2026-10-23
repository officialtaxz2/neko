import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { PauseIcon, PlayIcon, SpeakerHighIcon, SpeakerSlashIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useClient } from '@/state/provider'
import { dockButton } from './MobileDock'

// The bar of a view-only viewer (a share link), under the picture on every screen: playback and
// sound, nothing that acts on the room.
export function PassiveBar() {
  const { t } = useTranslation()
  const client = useClient()
  const { playable, playing, muted, volume } = useStore(
    client.store,
    useShallow((s) => ({
      playable: s.video.playable,
      playing: s.video.playing,
      muted: s.video.muted,
      volume: s.video.volume,
    })),
  )
  return (
    <div
      className="z-10 flex flex-none items-center justify-center gap-2 border-t border-glass-border bg-bar-room px-4 pt-2 pb-[calc(8px+env(safe-area-inset-bottom))]"
      data-testid="passive-bar"
    >
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
      <input
        type="range"
        className="slider h-11 w-[150px] max-w-[40vw]"
        min="0"
        max="100"
        aria-label={t('fork:dock.volume')}
        value={Math.round(volume * 100)}
        onChange={(e) => client.setVolume(Number(e.target.value) / 100)}
      />
    </div>
  )
}
