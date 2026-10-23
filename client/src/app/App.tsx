import { useStore } from 'zustand'
import { IconContext } from '@phosphor-icons/react'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { useNeko } from '@/state/provider'
import { useSlot } from '@/plugins'
import { selectSession, WebRTCTransport } from '@m1k1o/neko'
import { cn } from '@/lib/utils'
import logo from '@/assets/images/logo.svg'
import { Dialog } from '@/components/Dialog'
import { Toaster } from '@/fork/layout/Toaster'
import { Header } from '@/fork/layout/Header'
import { Side } from '@/fork/layout/Side'
import { RoomBar } from '@/fork/layout/RoomBar'
import { MobileDock } from '@/fork/layout/MobileDock'
import { PassiveBar } from '@/fork/layout/PassiveBar'
import { ShareConnect } from '@/fork/connect/ShareConnect'
import { Stage } from '@/features/video/Stage'
import { MemberMenu } from '@/features/members/MemberMenu'
import { Connect } from '@/features/connect/Connect'
import { About } from '@/features/about/About'

const params = new URL(location.href).searchParams
const cast = !!params.get('cast')
const videoOnly = cast || !!params.get('embed')

export function App() {
  const { t } = useTranslation()
  const neko = useNeko()
  const { client, app } = neko
  // the fork: a view-only share link, or a session the server made view-only, sees only the picture
  // with a playback bar (fork/layout/PassiveBar.tsx): no header, side panel, room bar or toasts
  const passive = useStore(client.store, (s) => !!neko.share || !!selectSession(s)?.profile.is_view_only)
  const bare = videoOnly || passive
  const connected = useStore(client.store, (s) => s.connection.status === 'connected')
  const { side, about } = useStore(
    app,
    useShallow((s) => ({ side: s.side, about: s.about })),
  )
  const memberMenu = useSlot('member.menu')

  if (!WebRTCTransport.supported()) {
    return (
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="w-[320px] rounded-[5px] bg-background-secondary p-2.5">
          <div className="flex w-full items-center justify-center">
            <img src={logo} alt="n.eko" className="mr-2.5 h-[90px]" />
            <span className="text-[30px] leading-14">
              <b className="font-black">n</b>.eko
            </span>
          </div>
          <div className="flex flex-col">
            <span className="block text-center leading-7.5 uppercase">{t('unsupported')}</span>
          </div>
        </div>
      </div>
    )
  }

  const expanded = !bare && side
  return (
    // every icon filled, as in the v2 client (phosphor defaults to the outline `regular` weight)
    <IconContext.Provider value={{ weight: 'fill' }}>
      <div
        id="neko"
        className="tablet:relative tablet:max-h-none tablet:flex-col absolute inset-0 flex max-h-[100vh] max-w-[100vw]"
      >
        <main
          className={cn(
            'flex min-w-[360px] max-w-full grow flex-col overflow-auto',
            expanded ? 'tablet:portrait:h-[40vh] tablet:landscape:h-screen phone:portrait:h-auto' : 'tablet:h-screen',
          )}
        >
          {!bare && (
            // the fork's header: glass over the page, above the picture
            <div
              className="z-[100] flex h-menu shrink-0 border-b border-glass-border bg-bar-header shadow-bar-down backdrop-blur-md"
              data-testid="header"
            >
              <Header />
            </div>
          )}
          {/* .video-container: index.html sizes it before the app renders (iOS) */}
          {/* the fork: the picture sits in a soft vignette; on a portrait phone it keeps its own
              height and the dock follows right below it */}
          <div className="video-container phone:portrait:h-[min(62.5vw,60vh)]! phone:portrait:flex-none relative flex max-w-full grow overflow-hidden bg-[radial-gradient(circle_at_50%_50%,rgba(12,13,18,0.2)_0%,rgba(7,8,12,0.7)_100%)]">
            <Stage hideControls={cast} extraControls={videoOnly} />
          </div>
          {!bare && <RoomBar />}
          {!bare && <MobileDock />}
          {passive && !cast && <PassiveBar />}
        </main>
        {!bare && <MemberMenu items={memberMenu} />}
        {expanded && <Side />}
        {!connected && (neko.share ? <ShareConnect /> : <Connect />)}
        {!bare && <Toaster />}
        {about && <About />}
        <Dialog />
      </div>
    </IconContext.Provider>
  )
}
