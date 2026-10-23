import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { useNeko } from '@/state/provider'
import { setSetting } from '@/state/settings'
import { useTouch } from './useTouch'

type Key = 'force_touch' | 'trackpad_mode' | 'trackpad_cursor_hidden'

// the fork: the settings of touch input. "Force touch" on any device; the trackpad and its
// pointer only where touch input applies.
export function InputSetting() {
  const { t } = useTranslation()
  const neko = useNeko()
  const touch = useTouch()
  const s = useStore(
    neko.app,
    useShallow((a) => ({
      force_touch: a.settings.force_touch,
      trackpad_mode: a.settings.trackpad_mode,
      trackpad_cursor_hidden: a.settings.trackpad_cursor_hidden,
    })),
  )
  const row = (key: Key, hint?: string) => (
    <li className="flex flex-row items-start gap-3 border-b border-background-secondary py-[5px]" key={key}>
      <span className="mr-auto flex min-w-0 flex-col leading-6">
        {t(`fork:input.${key}`)}
        {hint && <small className="pb-1 text-[12px] leading-snug whitespace-normal text-text-muted">{hint}</small>}
      </span>
      <label className="relative h-6 w-[42px] flex-none">
        <input
          type="checkbox"
          className="peer h-0 w-0 opacity-0"
          aria-label={t(`fork:input.${key}`)}
          data-testid={`setting-${key}`}
          checked={s[key]}
          onChange={(e) => setSetting(neko, key, e.target.checked)}
        />
        <span className="absolute inset-0 cursor-pointer rounded-[34px] bg-background-tertiary transition-all duration-200 peer-checked:bg-style-primary peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-style-primary" />
        <span className="absolute bottom-[3px] left-[3px] h-[18px] w-[18px] cursor-pointer rounded-full bg-white shadow-[0_2px_4px_rgba(0,0,0,0.3)] transition-all duration-300 peer-checked:translate-x-[18px]" />
      </label>
    </li>
  )
  return (
    <>
      {row('force_touch', t('fork:input.force_touch_hint'))}
      {touch && row('trackpad_mode', t('fork:input.trackpad_hint'))}
      {touch && s.trackpad_mode && row('trackpad_cursor_hidden')}
    </>
  )
}
