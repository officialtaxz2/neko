import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import type { Session } from '@m1k1o/neko'
import { CrownIcon, EyeSlashIcon, MicrophoneSlashIcon, ShieldIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { selectMuted } from '@/state/app'
import { useNeko } from '@/state/provider'
import { Avatar } from '@/components/Avatar'

const item =
  'flex min-h-14 w-full items-center gap-3 rounded-inner border border-transparent px-2.5 py-2 text-left transition-colors duration-200 ease-fluid'
const badge = 'inline-flex size-6 flex-none items-center justify-center rounded-full bg-white/[0.05] [&_svg]:size-3.5'

// The side panel's Members tab, as the Vue client had it: everyone in the room with their roles
// (host, admin, muted, ignored by you), you first. A tap, Enter or the context menu on another member
// opens the member menu, so a phone, where the room bar with its members is hidden, still reaches it.
export function MemberList() {
  const { t } = useTranslation()
  const { client, app } = useNeko()
  const { hostId, sessions, sessionId } = useStore(
    client.store,
    useShallow((s) => ({ hostId: s.control.host_id, sessions: s.sessions, sessionId: s.session_id })),
  )
  const muted = useStore(
    client.store,
    useShallow((s) => Object.keys(s.sessions).filter((id) => selectMuted(s, id))),
  )
  const ignored = useStore(app, (s) => s.ignored)
  const me = sessionId ? sessions[sessionId] : undefined
  const others = Object.values(sessions).filter((m) => m.id !== sessionId && m.state.is_connected)
  const count = others.length + (me ? 1 : 0)

  const badges = (m: Session) => (
    <span className="flex flex-none items-center gap-1">
      {m.id === hostId && (
        <span className={cn(badge, 'text-mint')} title={t('fork:room.host')} aria-label={t('fork:room.host')}>
          <CrownIcon />
        </span>
      )}
      {m.profile.is_admin && (
        <span className={cn(badge, 'text-admin')} title={t('fork:header.admin')} aria-label={t('fork:header.admin')}>
          <ShieldIcon />
        </span>
      )}
      {muted.includes(m.id) && (
        <span className={cn(badge, 'text-badge')} title={t('fork:room.muted')} aria-label={t('fork:room.muted')}>
          <MicrophoneSlashIcon />
        </span>
      )}
      {ignored[m.id] && (
        <span
          className={cn(badge, 'text-text-muted')}
          title={t('fork:room.ignored')}
          aria-label={t('fork:room.ignored')}
        >
          <EyeSlashIcon />
        </span>
      )}
    </span>
  )
  const avatar = (m: Session) => (
    <span
      className={cn(
        'flex size-[38px] flex-none items-center justify-center rounded-full border-[1.5px] border-glass-line p-0.5',
        m.id === hostId && 'border-mint shadow-mint-glow',
      )}
    >
      <Avatar seed={m.profile.name} avatar={m.profile.avatar} size={30} />
    </span>
  )
  const open = (e: React.MouseEvent<HTMLButtonElement>, id: string) => {
    e.preventDefault()
    // a tap or click at its point; Enter or Space (no pointer) under the member
    const r = e.currentTarget.getBoundingClientRect()
    const at = e.detail === 0 && e.type === 'click' ? { x: r.left, y: r.bottom } : { x: e.clientX, y: e.clientY }
    app.setState({ menu: { ...at, id } })
  }

  return (
    <div className="flex min-w-0 flex-1 flex-col px-3 pb-3" data-testid="member-list">
      <h3 className="mx-2 mt-3 mb-2 text-[11.5px] font-bold tracking-[0.08em] text-text-muted uppercase">
        {t('fork:room.members')} ({count})
      </h3>
      <ul className="flex flex-col gap-1">
        {me && (
          <li>
            <div className={cn(item, 'bg-white/[0.02]')} data-testid="member-row" data-self>
              {avatar(me)}
              <span className="min-w-0 flex-1 truncate text-[14px] font-medium text-text-normal">
                {me.profile.name} <span className="text-text-muted">({t('you')})</span>
              </span>
              {badges(me)}
            </div>
          </li>
        )}
        {others.map((m) => (
          <li key={m.id}>
            <button
              type="button"
              className={cn(item, 'cursor-pointer hover:border-glass-border hover:bg-glass-hover')}
              aria-haspopup="menu"
              data-testid="member-row"
              onClick={(e) => open(e, m.id)}
              onContextMenu={(e) => open(e, m.id)}
            >
              {avatar(m)}
              <span className="min-w-0 flex-1 truncate text-[14px] font-medium text-text-normal">{m.profile.name}</span>
              {badges(m)}
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
