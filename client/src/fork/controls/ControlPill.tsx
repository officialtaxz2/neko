import { useEffect, useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { CheckIcon, CursorIcon, HandIcon, LockIcon, LockOpenIcon, PauseIcon } from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { useNeko } from '@/state/provider'
import { Avatar } from '@/components/Avatar'
import { pillState, pillVisible } from './pill'

// the pill's action buttons: a round chip, mint for the one to press
const act =
  'inline-flex h-8 flex-none cursor-pointer items-center gap-1.5 rounded-full bg-white/8 px-3.5 text-[13px] font-semibold text-inherit transition-colors duration-300 ease-fluid hover:bg-white/14 disabled:cursor-default disabled:opacity-60 max-[600px]:px-[11px]'
const primary = 'bg-mint text-mint-ink hover:bg-mint-active'

// The control pill of the fork's design: who controls the desktop, in words, with the one action
// that fits (take, request, take over, give up) and, for the host, the pause of their own input.
// `compact` is the phone's dock: no second line, the input pause lives in the More sheet.
export function ControlPill({ compact = false, separator = false }: { compact?: boolean; separator?: boolean }) {
  const { t } = useTranslation()
  const { client } = useNeko()
  const actions = useActions()
  const { state, visible, admin, implicit, inputLocked, playable, host } = useStore(
    client.store,
    useShallow((s) => {
      const id = s.control.host_id
      return {
        state: pillState(s),
        visible: pillVisible(s, client.receiveOnly),
        admin: selectIsAdmin(s),
        implicit: s.settings.implicit_hosting,
        inputLocked: s.control.locked,
        playable: s.video.playable,
        host: id ? s.sessions[id]?.profile : undefined,
      }
    }),
  )
  const [requested, setRequested] = useState(false)
  const [shake, setShake] = useState(false)
  const timer = useRef(0)

  // a request is answered by the host, or not: "requested" for a minute, or until the host changes
  const hostId = useStore(client.store, (s) => s.control.host_id)
  useEffect(() => {
    setRequested(false)
    clearTimeout(timer.current)
  }, [hostId])
  useEffect(() => () => clearTimeout(timer.current), [])

  // a click into the picture without control nudges the pill
  useEffect(() => {
    let t = 0
    const off = client.events.on('overlay.click', () => {
      if (client.controlling || client.state.settings.implicit_hosting) return
      setShake(true)
      clearTimeout(t)
      t = window.setTimeout(() => setShake(false), 1100)
    })
    return () => (off(), clearTimeout(t))
  }, [client])

  if (!visible) return null

  const request = () => {
    if (!playable) return
    client.request()
    setRequested(true)
    clearTimeout(timer.current)
    timer.current = window.setTimeout(() => setRequested(false), 60000)
  }
  const hostName = host?.name ?? ''
  const title = {
    mine: t('fork:controls.you_control'),
    other: t('fork:controls.other_controls', { name: hostName }),
    locked: t('fork:controls.locked_by_admin'),
    implicit: t('fork:controls.click_to_control'),
    free: t('fork:controls.take'),
  }[state]
  const subtitle = {
    mine: t(inputLocked ? 'fork:controls.input_locked' : 'fork:controls.input_active'),
    other: requested ? t('fork:controls.waiting_for', { name: hostName }) : t('fork:controls.watching'),
    locked: t('fork:controls.admin_only'),
    implicit: '',
    free: '',
  }[state]
  const lockLabel = t(inputLocked ? 'fork:controls.unlock_input' : 'fork:controls.lock_input')

  return (
    <>
      <div
        className={cn(
          'inline-flex h-11 min-w-0 items-center gap-2.5 rounded-full border border-glass-line bg-white/[0.04] pr-1.5 pl-[7px] font-medium whitespace-nowrap text-pure transition-all duration-300 ease-fluid',
          compact && 'gap-1.5',
          state === 'free' && 'border-mint/50 bg-mint/12 px-1 hover:bg-mint/18',
          state === 'mine' && 'border-transparent bg-mint text-mint-ink shadow-mint-glow',
          shake && 'animate-pill-nudge border-mint',
        )}
        role="group"
        aria-label={title}
        data-testid="control-pill"
        data-state={state}
      >
        {state === 'free' ? (
          <button
            type="button"
            className="inline-flex h-full cursor-pointer items-center gap-2 pr-3 pl-[9px] text-[13.5px] font-semibold text-mint disabled:cursor-default disabled:opacity-55"
            disabled={!playable}
            data-testid="control-take"
            onClick={() => playable && actions.toggleControl()}
          >
            <CursorIcon className="size-4" />
            <span>{title}</span>
          </button>
        ) : (
          <>
            <span
              className={cn(
                'inline-flex min-w-[30px] flex-none items-center justify-center text-[15px]',
                state === 'mine' && 'ml-1',
                state === 'locked' && 'text-warn',
                state === 'implicit' && 'text-mint',
              )}
              aria-hidden
            >
              {state === 'other' && host ? (
                <Avatar seed={host.name} size={compact ? 28 : 30} />
              ) : state === 'locked' ? (
                <LockIcon className="size-4" />
              ) : compact && inputLocked && state === 'mine' ? (
                <PauseIcon className="size-4" />
              ) : (
                <CursorIcon className="size-4" />
              )}
            </span>
            <span className="flex min-w-0 flex-col leading-[1.15]">
              <span className="truncate text-[14px]">{title}</span>
              {subtitle && !compact && (
                <small
                  className={cn('text-[11px] font-medium text-text-muted', state === 'mine' && 'text-mint-ink/72')}
                >
                  {subtitle}
                </small>
              )}
            </span>
            {state === 'mine' && (
              <>
                {!compact && (
                  <button
                    type="button"
                    className={cn(
                      act,
                      'w-8 justify-center bg-mint-ink/14 px-0 hover:bg-mint-ink/24',
                      inputLocked && 'bg-mint-ink text-mint hover:bg-mint-ink',
                    )}
                    aria-pressed={inputLocked}
                    aria-label={lockLabel}
                    title={lockLabel}
                    data-testid="input-lock"
                    onClick={() => (inputLocked ? client.unlock() : client.lock())}
                  >
                    {inputLocked ? <LockIcon className="size-3.5" /> : <LockOpenIcon className="size-3.5" />}
                  </button>
                )}
                {!implicit && (
                  <button
                    type="button"
                    className={cn(act, 'bg-mint-ink/14 hover:bg-mint-ink/24')}
                    data-testid="control-release"
                    onClick={() => client.release()}
                  >
                    {t('fork:controls.release')}
                  </button>
                )}
              </>
            )}
            {state === 'other' &&
              (admin ? (
                <button type="button" className={cn(act, primary)} data-testid="control-force" onClick={actions.take}>
                  {t('fork:controls.force_take')}
                </button>
              ) : requested ? (
                <button type="button" className={act} disabled>
                  <CheckIcon className="size-3.5" />
                  {t('fork:controls.requested')}
                </button>
              ) : (
                <button
                  type="button"
                  className={cn(act, primary)}
                  disabled={!playable}
                  data-testid="control-request"
                  onClick={request}
                >
                  <HandIcon className="size-3.5" />
                  {t('fork:controls.request')}
                </button>
              ))}
          </>
        )}
      </div>
      {separator && <span className="h-7 w-px flex-none bg-glass-line" aria-hidden />}
    </>
  )
}
