// One instance of the GUI: the core client with its input overlay, the GUI store and the bus, wired
// (event lines, toasts, the settings, also from the URL). The application makes one (main.tsx),
// adds the plugins (initPlugins), points it at the server (client.setUrl) and renders under
// <NekoProvider>; two instances share nothing.
import { NekoClient, Overlay, type NekoClientOptions } from '@m1k1o/neko'
import { createAppStore, type NekoApp } from './app'
import { createBus } from './bus'
import { wireEvents } from './events'
import { initSettings } from './settings'
import { bindPageLifecycle } from './lifecycle'
import { bindShareLeave, shareToken, SHARE_TOKEN_KEY, tabStorage } from './viewonly'
import { bindMedia, mediaTransport, readMedia, resolveMedia } from './media'
import { exposeDiagnostics } from './diagnostics'
import { captureErrors, debugEnabled } from './debug'
import { clipboardStore } from './clipboard'

// share: the token of a view-only link; read from the page's URL (#/<token>) unless given
export type NekoAppOptions = Pick<NekoClientOptions, 'transport'> & { share?: string }

export function createNekoApp(options: NekoAppOptions = {}): NekoApp {
  if (debugEnabled()) captureErrors() // the fork: page errors from start-up on, for ?debug
  const app = createAppStore()
  const share = 'share' in options ? options.share : shareToken()
  // the fork: WebRTC, or HLS when chosen (media.ts)
  const media = resolveMedia(new URL(location.href).search, readMedia())
  const autoplay = () => app.getState().settings.autoplay
  const client = new NekoClient({
    autologin: true,
    // a share link resumes or logs in from its own login screen (fork/connect/ShareConnect.tsx)
    autoconnect: !share,
    ...(share && { tokenStorage: tabStorage(), tokenKey: SHARE_TOKEN_KEY }),
    // read when a track arrives, so the setting applies without a reload
    get autoplay() {
      return autoplay()
    },
    transport: options.transport ?? mediaTransport(media, autoplay),
  })
  const neko: NekoApp = { client, overlay: new Overlay(client), app, bus: createBus(), share }
  wireEvents(neko)
  initSettings(neko)
  bindPageLifecycle(client)
  if (share) bindShareLeave(client)
  bindMedia(neko, media)
  exposeDiagnostics(neko)
  clipboardStore(neko) // the fork: records the clipboard's recent texts from the start
  return neko
}
