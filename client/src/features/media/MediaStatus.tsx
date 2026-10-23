import { useTranslation } from 'react-i18next'
import { CircleNotchIcon, LockIcon, WarningCircleIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useNeko } from '@/state/provider'
import { chooseMedia, isHLS, retryHLS } from '@/state/media'
import { Button } from '@/components/ui/button'
import { modeLabel, useMedia } from './useMedia'

// HLS on the picture: a small chip while it plays; otherwise what it is doing (negotiating, waiting
// for media, paused by the room's private mode) and, after a failure, Retry or back to WebRTC. HLS
// never falls back on its own.
export function MediaStatus() {
  const { t } = useTranslation()
  const neko = useNeko()
  const backend = useMedia((s) => s.selection.backend)
  const hls = useMedia((s) => s.hls)
  if (!isHLS(backend) || !hls || hls.status === 'idle') return null
  const label = modeLabel(backend)

  if (hls.status === 'streaming')
    return (
      <div
        className="pointer-events-none absolute top-3 left-3 z-10 inline-flex h-6 items-center gap-1.5 rounded-full border border-glass-line bg-bar-header px-2.5 text-[11.5px] font-semibold text-text-normal backdrop-blur-md"
        role="status"
        data-testid="media-status"
        data-status={hls.status}
      >
        <span className="size-1.5 rounded-full bg-mint shadow-[0_0_6px_var(--color-mint)]" aria-hidden />
        {label}
      </div>
    )

  const terminal = hls.status === 'terminal'
  const Icon = terminal ? WarningCircleIcon : hls.status === 'paused' ? LockIcon : CircleNotchIcon
  return (
    <div
      className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-4 max-[600px]:px-14 max-[600px]:py-2"
      role="status"
      data-testid="media-status"
      data-status={hls.status}
    >
      <div
        className={cn(
          // on a phone the picture is short: a tighter card that scrolls rather than spills, clear of
          // the picture's buttons on the right
          'flex max-h-full max-w-[340px] flex-col items-center gap-1.5 overflow-y-auto rounded-outer border border-glass-line bg-bar-room/90 px-5 py-4 text-center shadow-sheet backdrop-blur-md max-[600px]:gap-1 max-[600px]:rounded-inner max-[600px]:px-3.5 max-[600px]:py-2.5',
          terminal && 'pointer-events-auto border-style-error/40',
        )}
      >
        <Icon
          className={cn(
            'size-6 shrink-0 max-[600px]:size-5',
            terminal ? 'text-style-error' : 'text-mint',
            hls.status !== 'paused' && !terminal && 'animate-spin',
          )}
          aria-hidden
        />
        <strong className="text-[15px] text-pure">{label}</strong>
        <span className="text-[13.5px] text-text-normal">{t(`fork:media.status.${hls.status}`)}</span>
        {terminal && hls.detail && <small className="text-[12px] text-text-muted">{hls.detail}</small>}
        <small className="text-[12px] text-text-muted max-[600px]:hidden">{t('fork:media.limitations')}</small>
        {terminal && (
          <div className="mt-2 flex flex-wrap justify-center gap-2 max-[600px]:mt-1">
            <Button
              className="h-9 px-3 text-[12px] max-[600px]:h-8 max-[600px]:px-2.5 max-[600px]:text-[11px]"
              data-testid="media-retry"
              onClick={() => retryHLS(neko)}
            >
              {t('fork:media.retry')}
            </Button>
            <Button
              variant="outline"
              className="h-9 px-3 text-[12px] max-[600px]:h-8 max-[600px]:px-2.5 max-[600px]:text-[11px]"
              data-testid="media-webrtc"
              onClick={() => chooseMedia(neko, 'webrtc')}
            >
              {t('fork:media.use_webrtc')}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
