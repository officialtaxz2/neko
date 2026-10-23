import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { useNeko } from '@/state/provider'
import { setSetting } from '@/state/settings'
import { Switch } from '@/components/Switch'
import { useTouch } from './useTouch'

type Key = 'force_touch' | 'trackpad_mode' | 'trackpad_cursor_hidden'

// the fork: the settings of touch input, as rows of the settings' list. "Force touch" on any
// device; the trackpad and its pointer only where touch input applies.
export function InputSetting({ row, label }: { row: string; label: string }) {
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
  const item = (key: Key, hint?: string) => (
    <li className={row} key={key}>
      <span className={label}>
        {t(`fork:input.${key}`)}
        {hint && <small>{hint}</small>}
      </span>
      <Switch
        label={t(`fork:input.${key}`)}
        data-testid={`setting-${key}`}
        checked={s[key]}
        onChange={(e) => setSetting(neko, key, e.target.checked)}
      />
    </li>
  )
  return (
    <>
      {item('force_touch', t('fork:input.force_touch_hint'))}
      {touch && item('trackpad_mode', t('fork:input.trackpad_hint'))}
      {touch && s.trackpad_mode && item('trackpad_cursor_hidden')}
    </>
  )
}
