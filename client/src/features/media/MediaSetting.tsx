import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useNeko } from '@/state/provider'
import { chooseMedia, clearMediaOverride, isHLS, type MediaBackend } from '@/state/media'
import { Button } from '@/components/ui/button'
import { modeLabel, useMedia } from './useMedia'

// The settings' choice of the stream's path: WebRTC (the default), then the HLS modes the server
// offers (LL-HLS before HLS). A stored HLS choice the server no longer offers stays visible,
// disabled. A choice takes effect with a fresh page load.
export function MediaSetting() {
  const { t } = useTranslation()
  const neko = useNeko()
  const selection = useMedia((s) => s.selection)
  const modes = useMedia((s) => s.modes)
  const hls = useMedia((s) => s.hls)
  const { backend } = selection

  const options: { value: MediaBackend; name: string; description: string; disabled: boolean }[] = [
    { value: 'webrtc', name: 'WebRTC', description: t('fork:media.webrtc_desc'), disabled: false },
  ]
  for (const mode of ['ll-hls', 'hls'] as const) {
    const available = modes.includes(mode)
    if (!available && backend !== mode) continue
    options.push({
      value: mode,
      name: modeLabel(mode),
      description: available
        ? t(`fork:media.${mode === 'll-hls' ? 'll_hls' : 'hls'}_desc`)
        : t('fork:media.unavailable'),
      disabled: !available,
    })
  }
  const status =
    isHLS(backend) && hls
      ? `${modeLabel(backend)}: ${t(`fork:media.status.${hls.status}`)}${hls.detail ? ` (${hls.detail})` : ''}`
      : `${t('fork:media.current')}: WebRTC`

  return (
    <li className="flex flex-col gap-2 pt-0.5 pb-2.5" data-testid="media-setting">
      {/* named by the settings' Stream heading above it; the name stays for screen readers */}
      <span className="sr-only" id="media-label">
        {t('fork:media.label')}
      </span>
      <div role="radiogroup" aria-labelledby="media-label" className="flex flex-col gap-1.5">
        {options.map((o) => (
          <label
            key={o.value}
            className={cn(
              'flex cursor-pointer items-start gap-2.5 rounded-inner border border-glass-border bg-glass px-3 py-2 transition-colors duration-200 hover:bg-glass-hover',
              backend === o.value && 'border-mint/40 bg-mint/[0.07]',
              o.disabled && 'cursor-not-allowed opacity-60 hover:bg-glass',
            )}
          >
            <input
              type="radio"
              name="media-backend"
              className="mt-[3px] accent-[var(--color-mint)]"
              value={o.value}
              checked={backend === o.value}
              disabled={o.disabled}
              data-testid={`media-${o.value}`}
              onChange={() => chooseMedia(neko, o.value)}
            />
            <span className="flex min-w-0 flex-col whitespace-normal">
              <span className="text-[13.5px] font-semibold text-text-normal">
                {o.name}
                {o.value === 'webrtc' && (
                  <em className="ml-1.5 text-[11.5px] font-normal text-text-muted not-italic">
                    {t('fork:media.default')}
                  </em>
                )}
              </span>
              <small className="text-[12px] text-text-muted">{o.description}</small>
            </span>
          </label>
        ))}
      </div>
      <small className="text-[12px] whitespace-normal text-text-muted" role="status">
        {status}
      </small>
      {selection.overridden && (
        <div className="flex flex-col gap-1.5">
          <small className="text-[12px] whitespace-normal text-warn">
            {t(selection.invalid ? 'fork:media.invalid_override' : 'fork:media.override')}
          </small>
          <Button variant="outline" className="h-8 text-[12px]" onClick={() => clearMediaOverride(neko)}>
            {t('fork:media.clear_override')}
          </Button>
        </div>
      )}
      {isHLS(backend) && (
        <small className="text-[12px] whitespace-normal text-text-muted">{t('fork:media.limitations')}</small>
      )}
    </li>
  )
}
