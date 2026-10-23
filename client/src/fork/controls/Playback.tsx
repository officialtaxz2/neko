import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import {
  MicrophoneIcon,
  MicrophoneSlashIcon,
  PauseIcon,
  PlayIcon,
  SpeakerHighIcon,
  SpeakerSlashIcon,
} from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useNeko } from '@/state/provider'
import { micAllowed, micStore } from '@/state/microphone'
import { barButton, volumeAdjustable } from './styles'

// The room bar's playback: play/pause, the sound with its slider, and the host's microphone.
export function Playback() {
  const { t } = useTranslation()
  const neko = useNeko()
  const { client } = neko
  const { playable, playing, muted, volume } = useStore(
    client.store,
    useShallow((s) => ({
      playable: s.video.playable,
      playing: s.video.playing,
      muted: s.video.muted || s.video.volume === 0,
      volume: s.video.volume,
    })),
  )
  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        className={barButton}
        disabled={!playable}
        aria-label={playing ? t('fork:dock.pause') : t('fork:dock.play')}
        title={playing ? t('fork:dock.pause') : t('fork:dock.play')}
        data-testid="play"
        onClick={() => (playing ? client.pause() : client.play().catch(() => {}))}
      >
        {playing ? <PauseIcon className="size-4" /> : <PlayIcon className="size-4" />}
      </button>
      <div className="flex items-center gap-1.5">
        <button
          type="button"
          className={cn(barButton, 'border-transparent bg-transparent', muted && 'text-warn hover:text-warn')}
          aria-label={muted ? t('fork:dock.unmute') : t('fork:dock.mute')}
          aria-pressed={muted}
          title={muted ? t('fork:dock.unmute') : t('fork:dock.mute')}
          data-testid="mute"
          onClick={() => (muted ? client.unmute() : client.mute())}
        >
          {muted ? <SpeakerSlashIcon className="size-4" /> : <SpeakerHighIcon className="size-4" />}
        </button>
        {volumeAdjustable && (
          <input
            type="range"
            className="slider h-5 w-[110px]"
            min="0"
            max="100"
            aria-label={t('fork:dock.volume')}
            value={Math.round(volume * 100)}
            onChange={(e) => client.setVolume(Number(e.target.value) / 100)}
          />
        )}
      </div>
      <Microphone />
    </div>
  )
}

// the host's microphone, where the profile and the page allow it
export function Microphone({ className }: { className?: string }) {
  const { t } = useTranslation()
  const neko = useNeko()
  const mic = micStore(neko)
  const on = useStore(mic, (s) => s.on)
  const allowed = useStore(neko.client.store, () => micAllowed(neko))
  if (!allowed) return null
  const label = t(on ? 'controls.mic_off' : 'controls.mic_on')
  return (
    <button
      type="button"
      className={cn(barButton, on ? 'border-mint/40 text-mint shadow-mint-glow' : 'text-text-muted', className)}
      aria-pressed={on}
      aria-label={label}
      title={label}
      data-testid="mic"
      data-on={on || undefined}
      onClick={mic.toggle}
    >
      {on ? <MicrophoneIcon className="size-4" /> : <MicrophoneSlashIcon className="size-4" />}
    </button>
  )
}
