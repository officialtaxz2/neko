import { useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { ChatCenteredTextIcon, LockIcon, ShieldIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { isLocked, type LockResource } from '@/state/app'
import { useApp, useNeko } from '@/state/provider'
import { remember } from '@/state/storage'
import { useSlot } from '@/plugins'
import { EachHook } from '@/components/EachHook'
import logo from '@/assets/images/logo.svg'

const zero = () => 0
const always = () => true
const resources: LockResource[] = ['control', 'login']

// a status chip: a pill with an icon and a text that narrow screens drop
const chip =
  'inline-flex h-7 flex-none items-center gap-1.5 rounded-full border border-glass-line bg-white/[0.03] px-2.5 text-[12.5px] font-medium whitespace-nowrap text-text-normal max-[600px]:px-2'
const chipText = 'max-[600px]:hidden'

// The header of the fork's design: the logo (opens About), the room's status as chips (locks, admin,
// who is in the room) and the side panel's toggle with an unread badge. Locks are switched in the
// settings; here they are only shown.
export function Header() {
  const { t } = useTranslation()
  const { client, app } = useNeko()
  const { admin, settings, count } = useStore(
    client.store,
    useShallow((s) => ({
      admin: selectIsAdmin(s),
      settings: s.settings,
      count: Object.values(s.sessions).filter((m) => m.state.is_connected).length,
    })),
  )
  const side = useStore(app, (s) => s.side)
  // a plugin's room lock (file transfer) is a chip here and a switch in the settings, so its own
  // header button is left out
  const roomLocks = useSlot('room.lock')
  const items = useSlot('header.item').filter((i) => !roomLocks.some((l) => l.replaces === i.id))
  const tabs = useSlot('side.tab')
  const locks: string[] = resources.filter((r) => isLocked(r, settings))

  return (
    <div className="flex min-w-0 flex-1 items-center gap-2.5 pr-3 pl-4 max-[600px]:gap-1.5 max-[600px]:pr-2 max-[600px]:pl-2.5">
      <button
        type="button"
        className="inline-flex flex-none cursor-pointer items-center gap-[9px] rounded-button px-1 py-0.5 text-pure"
        aria-label={t('fork:header.about')}
        data-testid="about-open"
        onClick={() => app.setState({ about: true })}
      >
        <img src={logo} alt="" className="block h-7" />
        <span className="text-[24px] leading-7 font-medium max-[600px]:text-[20px]">
          <b className="font-black text-mint">n</b>.eko
        </span>
      </button>
      <div className="flex min-w-0 flex-1 items-center justify-end gap-2 overflow-hidden" role="status">
        <EachHook items={roomLocks} use={(l) => l.useAvailable ?? always}>
          {(available) =>
            [...locks, ...roomLocks.filter((l, i) => available[i] && l.locked(settings)).map((l) => l.id)].map((r) => (
              <span
                key={r}
                className={cn(chip, 'border-warn/40 bg-warn/[0.07] text-warn')}
                title={t(`fork:chip.${r}`)}
                data-testid={`lock-${r}`}
                data-locked
              >
                <LockIcon className="size-3" aria-hidden />
                <span className={chipText}>{t(`fork:chip.${r}`)}</span>
              </span>
            ))
          }
        </EachHook>
        {admin && (
          <span className={chip} title={t('admin_loggedin')} data-testid="admin-badge">
            <ShieldIcon className="size-3 text-admin" aria-hidden />
            <span className={chipText}>{t('fork:header.admin')}</span>
          </span>
        )}
        {count > 0 && (
          <span className={chip} title={t('fork:header.in_room', { count })}>
            <span className="size-[7px] rounded-full bg-mint shadow-[0_0_8px_var(--color-mint)]" aria-hidden />
            <span className={chipText}>{t('fork:header.in_room', { count })}</span>
            <span className="hidden max-[600px]:inline" aria-hidden>
              {count}
            </span>
          </span>
        )}
      </div>
      {items.length > 0 && (
        <ul className="flex flex-none items-center whitespace-nowrap">
          {items.map(({ id, component: Item }) => (
            <Item key={id} />
          ))}
        </ul>
      )}
      {/* what the side panel has to show, for the badge on its toggle */}
      <EachHook items={tabs} use={(tab) => tab.useBadge ?? zero}>
        {(counts) => <Toggle side={side} unread={counts.reduce((n, c) => n + c, 0)} />}
      </EachHook>
    </div>
  )
}

// the side panel's toggle, with a badge when the unread count changed while the panel was closed
function Toggle({ side, unread }: { side: boolean; unread: number }) {
  const { t } = useTranslation()
  const app = useApp()
  const [read, setRead] = useState(unread)
  const badge = !side && read !== unread
  return (
    <button
      type="button"
      className={cn(
        'relative inline-flex h-[38px] flex-none cursor-pointer items-center gap-2 rounded-control border border-glass-border bg-glass px-3.5 text-[13.5px] font-semibold text-text-normal transition-all duration-300 ease-fluid hover:bg-glass-hover hover:text-pure max-[600px]:w-10 max-[600px]:justify-center max-[600px]:px-0',
        side && 'border-mint/35 bg-mint/10 text-mint hover:bg-mint/15 hover:text-mint',
      )}
      aria-expanded={side}
      aria-label={t(side ? 'fork:header.close_panel' : 'fork:header.open_panel')}
      data-testid="side-toggle"
      onClick={() => {
        app.setState({ side: !side })
        remember('side', !side)
        setRead(unread)
      }}
    >
      <ChatCenteredTextIcon className="size-4" aria-hidden />
      <span className="max-[600px]:hidden">{t('fork:header.chat')}</span>
      {badge && (
        <span
          className="absolute top-[5px] right-[5px] size-2 rounded-full bg-badge shadow-[0_0_0_2px_var(--color-background-tertiary)]"
          title={t('fork:header.new_messages')}
        />
      )}
    </button>
  )
}
