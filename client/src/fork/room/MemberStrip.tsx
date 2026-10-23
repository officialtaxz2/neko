import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { ShieldIcon, UserCircleIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { useNeko } from '@/state/provider'
import { Avatar } from '@/components/Avatar'

// a member: a 38px ring around the avatar, mint and glowing for the host
const ring =
  'relative flex size-[38px] items-center justify-center rounded-full border-[1.5px] border-glass-line bg-white/5 p-0.5 transition-all duration-300 ease-fluid'
const host = 'border-mint shadow-mint-glow'
// the badge at the top right: you, or an admin
const badge =
  'absolute -top-1 -right-1 z-[2] size-4 rounded-full border-[1.5px] border-glass-border bg-background-secondary p-px shadow-[0_2px_6px_rgba(0,0,0,0.3)]'

// The members in the room bar, in the fork's design: you first, then the others; a click or the
// context menu on another member opens the member menu.
export function MemberStrip() {
  const { t } = useTranslation()
  const { client, app } = useNeko()
  const actions = useActions()
  const { hostId, sessions, sessionId } = useStore(
    client.store,
    useShallow((s) => ({ hostId: s.control.host_id, sessions: s.sessions, sessionId: s.session_id })),
  )
  const me = sessionId ? sessions[sessionId] : undefined
  const others = Object.values(sessions).filter((m) => m.id !== sessionId && m.state.is_connected)

  return (
    <div className="flex min-h-12 min-w-0 flex-1 items-center overflow-x-auto overflow-y-hidden [scrollbar-color:rgba(255,255,255,0.15)_transparent] [scrollbar-width:thin]">
      <ul className="flex items-center gap-1.5 p-1 whitespace-nowrap" aria-label={t('fork:room.members')}>
        {me && (
          <li className="relative">
            <div
              className={cn(ring, me.id === hostId && host)}
              title={me.profile.name}
              data-testid="member"
              data-self
              data-host={me.id === hostId || undefined}
            >
              <UserCircleIcon className={cn(badge, 'text-mint')} aria-hidden />
              <Avatar seed={me.profile.name} avatar={me.profile.avatar} size={30} />
            </div>
          </li>
        )}
        {others.map((m) => (
          <li key={m.id} className="relative">
            <button
              type="button"
              className={cn(
                ring,
                'cursor-pointer hover:-translate-y-0.5 hover:border-white/30 hover:shadow-[0_4px_12px_rgba(0,0,0,0.3)]',
                m.id === hostId && host,
              )}
              title={m.profile.name}
              aria-label={m.profile.name}
              aria-haspopup="menu"
              data-testid="member"
              data-host={m.id === hostId || undefined}
              onClick={(e) => {
                if (e.detail !== 0) return actions.openMenu(e, m.id)
                // Enter or Space: under the member, not at the corner of the page
                const r = e.currentTarget.getBoundingClientRect()
                app.setState({ menu: { x: r.left, y: r.bottom, id: m.id } })
              }}
              onContextMenu={(e) => actions.openMenu(e, m.id)}
            >
              {m.profile.is_admin && <ShieldIcon className={cn(badge, 'text-admin')} aria-hidden />}
              <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={30} />
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
