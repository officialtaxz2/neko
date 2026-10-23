import { useEffect, useState, type ReactNode } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { CaretDownIcon, PlayIcon, SignOutIcon, SquareIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { isLocked, type LockResource, type ViewerSettings } from '@/state/app'
import { ask } from '@/state/dialogs'
import { useNeko } from '@/state/provider'
import { setSetting } from '@/state/settings'
import { htmlLang, langs, setLang, type Lang } from '@/i18n'
import { Switch } from '@/components/Switch'
import { MediaSetting } from '@/features/media/MediaSetting'
import { InputSetting } from '@/features/input/InputSetting'

// a row: the name (with a hint below it) on the left, the control on the right
const row = 'flex min-h-12 items-center justify-between gap-4 border-t border-white/[0.04] py-1.5 first:border-t-0'
const label =
  'flex min-w-0 flex-col text-[14px] leading-[1.3] font-medium text-text-normal [&_small]:mt-0.5 [&_small]:text-[12px] [&_small]:font-normal [&_small]:whitespace-normal [&_small]:text-text-muted'
const select =
  'h-9 w-[160px] max-w-full cursor-pointer appearance-none truncate rounded-button border border-glass-border bg-white/[0.04] pr-8 pl-3 text-[13px] text-text-normal hover:border-glass-line [&_option]:bg-background-tertiary'
const button =
  'inline-flex h-[34px] flex-none cursor-pointer items-center justify-center gap-2 rounded-button px-3.5 text-[13px] font-semibold transition-colors duration-300 ease-fluid'

// a language in its own words (Deutsch, 日本語), the code where the browser cannot name it
const langName = (l: Lang) => {
  try {
    const tag = htmlLang(l)
    return new Intl.DisplayNames([tag], { type: 'language' }).of(tag) ?? l
  } catch {
    return l
  }
}

function Section({ title, badge, children }: { title: string; badge?: string; children: ReactNode }) {
  return (
    <section className="border-b border-white/[0.06] pt-3 pb-1 last:border-b-0">
      <h3 className="mb-2 flex items-center gap-2 text-[11.5px] font-bold tracking-[0.08em] text-text-muted uppercase">
        {title}
        {badge && (
          <span className="rounded-full border border-admin/30 bg-admin/12 px-[7px] py-0.5 text-[10px] tracking-[0.04em] text-admin">
            {badge}
          </span>
        )}
      </h3>
      <ul>{children}</ul>
    </section>
  )
}

function Select({ children, ...props }: React.ComponentProps<'select'>) {
  return (
    <span className="relative inline-flex flex-none">
      <select className={select} {...props}>
        {children}
      </select>
      <CaretDownIcon className="pointer-events-none absolute top-1/2 right-2.5 size-3.5 -translate-y-1/2 text-text-muted" />
    </span>
  )
}

// The settings of the fork's design, in groups as the Vue client had them: Stream, Input, Chat,
// Room (admins: the locks and the live broadcast) and Account (language, logging out after asking).
// The rows are upstream's settings (features/settings/Settings.tsx, which stays as it is).
export function Settings() {
  const { t, i18n } = useTranslation()
  const neko = useNeko()
  const { client, app } = neko
  const actions = useActions()
  const {
    settings: s,
    broadcast,
    openInApp,
    keyboardLayouts,
  } = useStore(
    app,
    useShallow((a) => ({
      settings: a.settings,
      broadcast: a.broadcast,
      openInApp: a.openInApp,
      keyboardLayouts: a.keyboardLayouts,
    })),
  )
  const admin = useStore(client.store, selectIsAdmin)
  const roomSettings = useStore(client.store, (st) => st.settings)
  const [url, setUrl] = useState(broadcast.url)
  useEffect(() => setUrl(broadcast.url), [broadcast.url])

  const toggle = (
    key: keyof ViewerSettings & ('scroll_invert' | 'autoplay' | 'ignore_emotes' | 'chat_sound' | 'links_in_app'),
    name: string,
  ) => (
    <li className={row}>
      <span className={label}>{name}</span>
      <Switch label={name} checked={s[key]} onChange={(e) => setSetting(neko, key, e.target.checked)} />
    </li>
  )
  const lock = (r: LockResource) => (
    <li className={row}>
      <span className={label}>
        {t(`locks.${r}.locked`)}
        <small>{t(`fork:settings.lock_${r}_hint`)}</small>
      </span>
      <Switch
        label={t(`locks.${r}.locked`)}
        data-testid={`lock-switch-${r}`}
        checked={isLocked(r, roomSettings)}
        onChange={() => actions.toggleLock(r)}
      />
    </li>
  )
  const logout = async () => {
    if (await ask(app, t('fork:settings.logout_title'), t('fork:settings.logout_text'), t('logout'))) actions.logout()
  }

  return (
    <div className="flex flex-1 flex-col overflow-x-hidden px-5 pb-5" data-testid="settings">
      <Section title={t('fork:settings.stream')}>
        {/* WebRTC or HLS / LL-HLS (state/media.ts) */}
        <MediaSetting />
        {toggle('autoplay', t('setting.autoplay'))}
      </Section>

      <Section title={t('fork:settings.input')}>
        <li className={row}>
          <span className={label}>{t('setting.scroll')}</span>
          <input
            type="range"
            className="slider h-6 w-[120px] flex-none"
            min="-5"
            max="5"
            step="1"
            aria-label={t('setting.scroll')}
            value={s.scroll_sensitivity}
            onChange={(e) => setSetting(neko, 'scroll_sensitivity', Number(e.target.value))}
          />
        </li>
        {toggle('scroll_invert', t('setting.scroll_invert'))}
        <li className={row}>
          <span className={label}>{t('setting.keyboard_layout')}</span>
          <Select
            aria-label={t('setting.keyboard_layout')}
            value={s.keyboard_layout}
            onChange={(e) => setSetting(neko, 'keyboard_layout', e.target.value)}
          >
            {Object.entries(keyboardLayouts).map(([code, name]) => (
              <option key={code} value={code}>
                {name}
              </option>
            ))}
          </Select>
        </li>
        {/* force touch, the trackpad (features/input) */}
        <InputSetting row={row} label={label} />
      </Section>

      <Section title={t('fork:settings.chat')}>
        {toggle('chat_sound', t('setting.chat_sound'))}
        {toggle('ignore_emotes', t('setting.ignore_emotes'))}
        {openInApp && toggle('links_in_app', t('setting.links_in_app'))}
      </Section>

      {admin && (
        <Section title={t('fork:settings.room')} badge={t('fork:header.admin')}>
          {lock('control')}
          {lock('login')}
          <li className={cn(row, 'flex-col items-stretch gap-2 py-2.5')}>
            <div className="flex items-center justify-between gap-4">
              <span className={label}>
                {t('setting.broadcast_title')}
                <small>{t('fork:settings.broadcast_hint')}</small>
              </span>
              {!broadcast.active ? (
                <button
                  type="button"
                  className={cn(button, 'bg-mint text-mint-ink hover:bg-mint-active')}
                  data-testid="broadcast-start"
                  onClick={() => actions.broadcastStart(url)}
                >
                  <PlayIcon className="size-3.5" />
                  {t('fork:settings.broadcast_start')}
                </button>
              ) : (
                <button
                  type="button"
                  className={cn(button, 'bg-[#a62626] text-white hover:bg-[#bf2d2d]')}
                  data-testid="broadcast-stop"
                  onClick={actions.broadcastStop}
                >
                  <SquareIcon className="size-3.5" />
                  {t('fork:settings.broadcast_stop')}
                </button>
              )}
            </div>
            <input
              className="h-9 w-full rounded-button border border-glass-border bg-white/[0.03] px-3 text-[13px] text-text-normal outline-hidden placeholder:text-text-muted focus:border-mint/50 disabled:opacity-60"
              aria-label={t('setting.broadcast_title')}
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={broadcast.active}
              placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
            />
          </li>
        </Section>
      )}

      <Section title={t('fork:settings.account')}>
        <li className={row}>
          <span className={label}>{t('fork:setting.language')}</span>
          <Select
            data-testid="lang"
            aria-label={t('fork:setting.language')}
            value={i18n.language}
            onChange={(e) => setLang(e.target.value as Lang)}
          >
            {langs.map((l) => (
              <option key={l} value={l}>
                {langName(l)}
              </option>
            ))}
          </Select>
        </li>
        <li className="pt-3 pb-1">
          <button
            type="button"
            className={cn(button, 'h-10 w-full border border-badge/45 bg-transparent text-[#ff7b85] hover:bg-badge/10')}
            data-testid="logout"
            onClick={logout}
          >
            <SignOutIcon className="size-4" />
            {t('logout')}
          </button>
        </li>
      </Section>
    </div>
  )
}
