// the fork: the host's microphone into the remote desktop, one per instance, so the room bar and the
// phone's More sheet switch the same one (it was upstream's Controls' own state)
import { createStore } from 'zustand/vanilla'
import { selectControlling, selectSession } from '@m1k1o/neko'
import { scoped, type NekoApp } from './app'
import { tell } from './dialogs'
import { t } from '@/i18n'

export const micStore = scoped((neko: NekoApp) => {
  const store = createStore(() => ({ on: false }))
  let mic: { track: MediaStreamTrack; stop: () => void } | null = null
  const off = () => {
    if (!mic) return
    mic.stop()
    mic.track.stop()
    mic = null
    store.setState({ on: false })
  }
  // dropped when control goes (the next host gets the audio input) and whenever the connection is
  // made anew: the track belonged to the old peer connection
  neko.client.store.subscribe(
    (s) => micAllowed(neko) && s.connection.status === 'connected',
    (keep) => !keep && off(),
  )
  const toggle = async () => {
    if (mic) return off()
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      mic = { track: stream.getAudioTracks()[0], stop: neko.client.shareMedia(stream) }
      store.setState({ on: true })
    } catch (err) {
      tell(neko.app, t('controls.mic_error'), (err as Error).message)
    }
  }
  return Object.assign(store, { toggle, off })
})

// the host whose profile may share media, on a page that can reach the microphone (not plain http)
export const micAllowed = ({ client }: NekoApp) =>
  selectControlling(client.state) &&
  !!selectSession(client.state)?.profile.can_share_media &&
  typeof navigator !== 'undefined' &&
  !!navigator.mediaDevices
