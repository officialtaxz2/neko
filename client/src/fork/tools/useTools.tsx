import type { ReactNode } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, selectSession } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import {
  ArrowsOutIcon,
  BroadcastIcon,
  ClipboardTextIcon,
  KeyboardIcon,
  LockIcon,
  LockOpenIcon,
  MicrophoneIcon,
  MicrophoneSlashIcon,
  MonitorIcon,
  PictureInPictureIcon,
  SlidersHorizontalIcon,
} from '@phosphor-icons/react'
import { useNeko } from '@/state/provider'
import { micAllowed, micStore } from '@/state/microphone'
import { mediaStore } from '@/state/media'
import { requestFullscreen } from '@/state/stage'
import { remember } from '@/state/storage'
import { useTouch } from '@/features/input/useTouch'
import { canPip } from '@/features/video/usePip'

export interface Tool {
  id: string
  icon: ReactNode
  label: string
  on?: boolean // a switch that is on (input paused, microphone on)
  run: () => void
  // the screen sizes open from the tile itself (a menu) instead of a run
  menu?: 'resolution'
}

// The tools that used to float on the picture, in one place, as the Vue client had them: input
// (on-screen keyboard and special keys on touch, input pause, microphone, clipboard) and picture &
// stream (fullscreen, picture in picture, screen size, the stream's path). The phone's sheet shows
// them all; the desktop's More popover leaves out what its bar already has (fullscreen, microphone,
// input pause).
export function useTools(sheet: boolean): { input: Tool[]; view: Tool[] } {
  const { t } = useTranslation()
  const neko = useNeko()
  const { client, overlay, app } = neko
  const touch = useTouch()
  const { hosting, canHost, admin, locked, viewOnly } = useStore(
    client.store,
    useShallow((s) => ({
      hosting: selectControlling(s),
      canHost: !!selectSession(s)?.profile.can_host,
      admin: selectIsAdmin(s),
      locked: s.control.locked,
      viewOnly: !!selectSession(s)?.profile.is_view_only,
    })),
  )
  const mic = micStore(neko)
  const micOn = useStore(mic, (s) => s.on)
  const micOk = useStore(client.store, () => micAllowed(neko))
  const backend = useStore(mediaStore(neko), (s) => s.selection.backend)
  const stream = backend === 'll-hls' ? 'LL-HLS' : backend === 'hls' ? 'HLS' : 'WebRTC'

  const input: Tool[] = []
  if (canHost && hosting) {
    if (touch)
      input.push(
        {
          id: 'keyboard',
          icon: <KeyboardIcon />,
          label: t('fork:input.keyboard'),
          run: () => overlay.mobileKeyboardToggle(),
        },
        {
          id: 'keyboard-helper',
          icon: <SlidersHorizontalIcon />,
          label: t('fork:input.keyboard_helper'),
          run: () => app.setState((s) => ({ keyboardHelper: !s.keyboardHelper })),
        },
      )
    if (sheet)
      input.push({
        id: 'input-pause',
        icon: locked ? <LockIcon /> : <LockOpenIcon />,
        label: t(locked ? 'fork:controls.resume_input' : 'fork:controls.pause_input'),
        on: locked,
        run: () => (locked ? client.unlock() : client.lock()),
      })
    if (sheet && micOk)
      input.push({
        id: 'microphone',
        icon: micOn ? <MicrophoneIcon /> : <MicrophoneSlashIcon />,
        label: t('fork:controls.microphone'),
        on: micOn,
        run: () => void mic.toggle(),
      })
    input.push({
      id: 'clipboard',
      icon: <ClipboardTextIcon />,
      label: t('fork:controls.clipboard'),
      run: () => app.setState((s) => ({ clipboardPanel: !s.clipboardPanel })),
    })
  }

  const view: Tool[] = []
  if (sheet)
    view.push({
      id: 'fullscreen',
      icon: <ArrowsOutIcon />,
      label: t('fork:controls.fullscreen'),
      run: () => requestFullscreen(neko),
    })
  if (canPip)
    view.push({
      id: 'pip',
      icon: <PictureInPictureIcon />,
      label: t('fork:controls.pip'),
      run: () => {
        const el = client.transport.element
        if (el instanceof HTMLVideoElement) el.requestPictureInPicture().catch(() => {})
      },
    })
  if (admin)
    view.push({
      id: 'resolution',
      icon: <MonitorIcon />,
      label: t('fork:controls.resolution'),
      run: () => {},
      menu: 'resolution',
    })
  if (!viewOnly)
    view.push({
      id: 'stream',
      icon: <BroadcastIcon />,
      label: t('fork:controls.stream', { name: stream }),
      run: () => {
        app.setState({ side: true, tab: 'settings' })
        remember('side', true)
        remember('tab', 'settings')
      },
    })
  return { input, view }
}
