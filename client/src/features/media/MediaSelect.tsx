import { useTranslation } from 'react-i18next'
import { useNeko } from '@/state/provider'
import { chooseMedia, normalizeMedia } from '@/state/media'
import { modeLabel, useMedia } from './useMedia'

// the short form of the choice for a view-only viewer's bar: shown only when the server offers HLS
// (or HLS is what this page runs)
export function MediaSelect() {
  const { t } = useTranslation()
  const neko = useNeko()
  const backend = useMedia((s) => s.selection.backend)
  const modes = useMedia((s) => s.modes)
  if (!modes.length && backend === 'webrtc') return null
  const unavailable = backend !== 'webrtc' && !modes.includes(backend)
  return (
    <select
      className="h-11 max-w-[130px] cursor-pointer rounded-control border border-glass-border bg-glass px-2.5 text-[13px] text-text-normal hover:bg-glass-hover [&_option]:bg-background-tertiary"
      aria-label={t('fork:media.label')}
      data-testid="media-select"
      value={backend}
      onChange={(e) => chooseMedia(neko, normalizeMedia(e.target.value))}
    >
      <option value="webrtc">WebRTC</option>
      {(['ll-hls', 'hls'] as const)
        .filter((m) => modes.includes(m))
        .map((m) => (
          <option key={m} value={m}>
            {modeLabel(m)}
          </option>
        ))}
      {unavailable && (
        <option value={backend} disabled>
          {modeLabel(backend)} ({t('fork:media.unavailable')})
        </option>
      )}
    </select>
  )
}
