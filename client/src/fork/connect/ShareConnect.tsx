import { useEffect, useState, type SyntheticEvent } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { tell } from '@/state/dialogs'
import { useNeko } from '@/state/provider'
import { rememberedName, startShare, watchShare } from '@/state/viewonly'
import { Button } from '@/components/ui/button'
import logo from '@/assets/images/logo.svg'

const input =
  'my-[5px] h-10 rounded-button border border-glass-border bg-white/[0.04] px-3 leading-5 text-text-normal outline-hidden selection:bg-text-link placeholder:text-text-muted focus:border-mint/50'

// The login screen of a view-only share link (state/viewonly.ts): only a name, no password. On the
// first visit with a remembered name, and on every reload of the tab, it starts by itself: a reload
// resumes the tab's session, a new visit logs in with the link's token.
export function ShareConnect() {
  const { t } = useTranslation()
  const neko = useNeko()
  const { client, app } = neko
  const { status, authenticated } = useStore(
    client.store,
    useShallow((s) => ({ status: s.connection.status, authenticated: s.authenticated })),
  )
  const [name, setName] = useState(() => new URL(location.href).searchParams.get('usr') || rememberedName())
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    const start = startShare(neko)
    if (!start) return
    setBusy(true)
    start.finally(() => setBusy(false))
  }, [neko])

  const submit = (e: SyntheticEvent) => {
    e.preventDefault()
    if (!name.trim()) return tell(app, t('connect.error'), t('connect.empty_displayname'))
    setBusy(true)
    watchShare(neko, name.trim()).finally(() => setBusy(false))
  }

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-background-floating/80" data-testid="connect">
      <div className="w-[320px] rounded-outer border border-glass-border bg-background-secondary p-4 shadow-sheet">
        <div className="flex w-full items-center justify-center">
          <img src={logo} alt="" className="mr-2.5 h-[90px]" />
          <span className="text-[30px] leading-14 text-pure">
            <b className="font-black text-mint">n</b>.eko
          </span>
        </div>
        {status === 'connecting' || busy ? (
          <div className="relative mx-auto h-[90px] w-[90px]" role="status" aria-label={t('fork:share.connecting')}>
            <div className="absolute top-0 left-0 h-full w-full animate-loader rounded-full bg-style-primary opacity-60"></div>
            <div className="absolute top-0 left-0 h-full w-full animate-loader rounded-full bg-style-primary opacity-60 [animation-delay:-1s]"></div>
          </div>
        ) : authenticated ? (
          // watching, but the connection dropped (network, server restart)
          <form className="flex flex-col" onSubmit={(e) => (e.preventDefault(), client.connect())}>
            <span className="block text-center leading-7.5 uppercase">{t('connection.disconnected')}</span>
            <Button className="my-[5px]" type="submit">
              {t('connect.connect')}
            </Button>
          </form>
        ) : (
          <form className="flex flex-col" onSubmit={submit}>
            <span className="block text-center text-[15px] leading-7.5 font-semibold text-pure">
              {t('fork:share.title')}
            </span>
            <span className="mb-2 block text-center text-[12.5px] leading-snug text-text-muted">
              {t('fork:share.hint')}
            </span>
            <input
              className={input}
              type="text"
              aria-label={t('connect.displayname')}
              placeholder={t('connect.displayname')}
              autoComplete="nickname"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
            />
            <Button className="my-[5px]" type="submit" data-testid="share-watch">
              {t('fork:share.watch')}
            </Button>
          </form>
        )}
      </div>
    </div>
  )
}
