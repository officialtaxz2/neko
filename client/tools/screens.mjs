// Screenshots of the GUI against a simulated server, for the visual checks of the fork's design
// (src/fork/README.md): `node tools/screens.mjs [out-dir]` (default ./screens-out).
// Vite serves the app; Playwright answers /api/* and the /api/ws event socket, and a fake
// RTCPeerConnection plays a painted desktop, so the room renders without a neko server.
// SCREENS=desktop,phone limits the run to the named shots; DUMP=<selector> prints the matching
// elements with their computed layout, to debug a shot. The page is served over HTTPS (a throwaway
// self-signed certificate), as HLS needs.
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { execFileSync } from 'node:child_process'
import { chromium } from 'playwright'
import { createServer } from 'vite'

const out = process.argv[2] || 'screens-out'
fs.mkdirSync(out, { recursive: true })
const only = process.env.SCREENS?.split(',')

const certDir = fs.mkdtempSync(path.join(os.tmpdir(), 'neko-screens-'))
execFileSync(
  'openssl',
  [
    'req',
    '-x509',
    '-newkey',
    'rsa:2048',
    '-nodes',
    '-days',
    '1',
    '-subj',
    '/CN=localhost',
    '-keyout',
    `${certDir}/key.pem`,
    '-out',
    `${certDir}/cert.pem`,
  ],
  { stdio: 'ignore' },
)
const https = { key: fs.readFileSync(`${certDir}/key.pem`), cert: fs.readFileSync(`${certDir}/cert.pem`) }
fs.rmSync(certDir, { recursive: true })
const server = await createServer({ logLevel: 'error', server: { port: 4174, strictPort: false, https } })
await server.listen()
const base = server.resolvedUrls.local[0]

// a pre-installed Chromium (cloud sessions) or Playwright's own
const executablePath = ['/opt/pw-browsers/chromium/chrome-linux/chrome', '/opt/pw-browsers/chromium'].find(
  (p) => fs.existsSync(p) && fs.statSync(p).isFile(),
)
const browser = await chromium.launch({ executablePath, args: ['--autoplay-policy=no-user-gesture-required'] })

const profile = (name, over = {}) => ({
  name,
  is_admin: false,
  can_login: true,
  can_connect: true,
  can_watch: true,
  can_host: true,
  can_share_media: false,
  can_access_clipboard: true,
  sends_inactive_cursor: false,
  can_see_inactive_cursors: false,
  ...over,
})
const member = (id, name, over) => ({
  id,
  profile: profile(name, over),
  state: { is_connected: true, is_watching: true },
})

// the room: Matze (admin, this session) with three others
function room({ host = 's3', admin = true, lockedControls = false, viewOnly = false } = {}) {
  return {
    session_id: 's1',
    sessions: {
      s1: viewOnly
        ? member('s1', 'Gast', { is_view_only: true, can_host: false, can_access_clipboard: false })
        : member('s1', 'Matze', { is_admin: admin }),
      s2: member('s2', 'Nele'),
      s3: member('s3', 'Alex'),
      s4: member('s4', 'Bo', { is_admin: true }),
    },
    settings: {
      private_mode: false,
      locked_logins: false,
      locked_controls: lockedControls,
      control_protection: false,
      implicit_hosting: false,
      inactive_cursors: false,
      merciful_reconnect: true,
    },
    screen_size: { width: 1920, height: 1080, rate: 30 },
    control_host: host ? { has_host: true, host_id: host } : { has_host: false },
    touch_events: false,
    webrtc: { videos: ['hq'], pointer_in_video: false },
  }
}
// what a server with HLS enabled advertises and offers (docs/HLS_LL_HLS.md)
const capabilities = {
  version: 1,
  backend: 'hls',
  modes: ['hls', 'll-hls'],
  container: 'fmp4',
  video_codec: 'h264-high-3.1',
  audio_codec: 'aac-lc',
  audio_rate: 48000,
  variants: [
    ['high', 1280, 720, 25],
    ['medium', 854, 480, 20],
    ['low', 640, 360, 15],
  ].map(([id, width, height, frame_rate]) => ({
    id,
    source_id: 'high',
    width,
    height,
    frame_rate,
    video_codec: 'avc1.64001f',
    container: 'video/mp4',
    bandwidth: 4000000,
  })),
  limits: {
    idle_expires_in_ms: 30000,
    maximum_playlist_bytes: 65536,
    max_requests_per_lease: 8,
    max_blocking_per_lease: 4,
  },
}
const chat = [
  ['s2', 'Hat jemand den Link zur Präsentation?'],
  ['s3', 'Ich teile gleich meinen Bildschirm, Moment'],
  ['s1', 'Passt, ich schaue zu 👀'],
]

// before the app: a token for autologin, and a peer connection whose stream is a painted desktop
function fakeBrowser(token) {
  if (token) localStorage.setItem('neko_session', token)
  const paint = () => {
    const c = document.createElement('canvas')
    c.width = 1920
    c.height = 1080
    const g = c.getContext('2d')
    const draw = () => {
      const bg = g.createLinearGradient(0, 0, 1920, 1080)
      bg.addColorStop(0, '#1d2b4a')
      bg.addColorStop(1, '#3a2346')
      g.fillStyle = bg
      g.fillRect(0, 0, 1920, 1080)
      // a browser window
      g.fillStyle = '#f4f5f7'
      g.fillRect(240, 120, 1440, 820)
      g.fillStyle = '#dfe1e6'
      g.fillRect(240, 120, 1440, 56)
      for (const [x, col] of [
        [272, '#ff5f57'],
        [302, '#febc2e'],
        [332, '#28c840'],
      ]) {
        g.fillStyle = col
        g.beginPath()
        g.arc(x, 148, 9, 0, 7)
        g.fill()
      }
      g.fillStyle = '#ffffff'
      g.fillRect(380, 134, 900, 28)
      g.fillStyle = '#5f6368'
      g.font = '18px sans-serif'
      g.fillText('https://example.org/praesentation', 396, 154)
      g.fillStyle = '#1f2937'
      g.font = 'bold 54px sans-serif'
      g.fillText('Quartalsplanung', 320, 290)
      g.font = '28px sans-serif'
      g.fillStyle = '#4b5563'
      const lines = [
        'Ziele für das nächste Quartal',
        'Stand der laufenden Projekte',
        'Offene Fragen und nächste Schritte',
      ]
      lines.forEach((l, i) => g.fillText('•  ' + l, 340, 380 + i * 56))
      g.fillStyle = '#e5e7eb'
      g.fillRect(1100, 250, 500, 320)
      g.fillStyle = '#26e6b4'
      ;[180, 240, 120, 280].forEach((h, i) => g.fillRect(1150 + i * 110, 540 - h, 70, h))
      g.fillStyle = '#9ca3af'
      g.font = '22px sans-serif'
      g.fillText(new Date().toLocaleTimeString('de-DE'), 1530, 910)
    }
    draw()
    setInterval(draw, 500)
    return c.captureStream(15)
  }
  class FakePeer {
    constructor() {
      this.connectionState = 'new'
      this.iceConnectionState = 'new'
      this.signalingState = 'stable'
      this.remoteDescription = null
      this.localDescription = null
    }
    async setRemoteDescription(d) {
      this.remoteDescription = d
    }
    async createAnswer() {
      return { type: 'answer', sdp: 'v=0\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\n' }
    }
    async createOffer() {
      return { type: 'offer', sdp: 'v=0\r\n' }
    }
    async setLocalDescription(d) {
      this.localDescription = d
      setTimeout(() => {
        const stream = paint()
        for (const track of stream.getTracks()) this.ontrack?.({ track, streams: [stream], receiver: {} })
        this.connectionState = this.iceConnectionState = 'connected'
        this.onconnectionstatechange?.()
      }, 50)
    }
    async addIceCandidate() {}
    getSenders() {
      return []
    }
    addTrack() {}
    removeTrack() {}
    close() {}
  }
  window.RTCPeerConnection = FakePeer
}
// a browser without WebRTC (an old TV)
const noWebRTC = () => delete window.RTCPeerConnection

async function shot(
  name,
  {
    width,
    height,
    mobile = false,
    token = 'tok',
    state,
    side,
    tab,
    hash = '',
    viewer,
    after,
    hls,
    webrtc = true,
    stored = {},
  },
) {
  if (only && !only.includes(name)) return
  const ctx = await browser.newContext({
    viewport: { width, height },
    ignoreHTTPSErrors: true,
    deviceScaleFactor: 2,
    isMobile: mobile,
    hasTouch: mobile,
    locale: 'de-DE',
  })
  await ctx.addInitScript(fakeBrowser, token)
  if (!webrtc) await ctx.addInitScript(noWebRTC)
  await ctx.addInitScript(
    ([side, tab, viewer, stored]) => {
      // the stored settings as src/state/storage.ts writes them
      localStorage.setItem('lang', 'de')
      if (side !== undefined) localStorage.setItem('side', side ? '1' : '0')
      if (tab) localStorage.setItem('tab', tab)
      if (viewer) localStorage.setItem('view_name', viewer)
      for (const [k, v] of Object.entries(stored)) localStorage.setItem(k, v)
    },
    [side, tab, viewer, stored],
  )
  const page = await ctx.newPage()
  page.on('pageerror', (e) => console.log(`${name}: pageerror ${e.message}`))
  await page.route('**/api/**', (r) => {
    const path = new URL(r.request().url()).pathname
    if (path.endsWith('/whoami'))
      return token
        ? r.fulfill({ json: member('s1', 'Matze') })
        : r.fulfill({ status: 401, json: { message: 'no session' } })
    if (path.endsWith('/login')) return r.fulfill({ json: { token: 'tok' } })
    // the HLS bootstrap: still starting (no answer yet), or refused
    if (path.endsWith('/media/hls/session')) return hls === 'fail' ? r.fulfill({ status: 503, body: '' }) : undefined
    return r.fulfill({ json: {} })
  })
  await page.routeWebSocket(/\/api\/ws/, (ws) => {
    const send = (event, payload) => ws.send(JSON.stringify({ event, payload }))
    send('system/init', room(state))
    send('chat/init', { enabled: true })
    chat.forEach(([id, text], i) =>
      send('chat/message', { id, created: new Date(Date.now() - (3 - i) * 60000).toISOString(), content: { text } }),
    )
    ws.onMessage((m) => {
      const { event } = JSON.parse(m)
      if (event === 'signal/request')
        send('signal/provide', { sdp: 'v=0\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\n', iceservers: [] })
      if (event === 'media/hls/capabilities/request') send('media/hls/capabilities', capabilities)
      if (event === 'media/hls/create')
        send('media/hls/offer', {
          version: 1,
          backend: 'hls',
          mode: JSON.parse(m).payload.mode,
          path: '/api/media/hls/session',
          ticket: 'T'.repeat(32),
          expires_in_ms: 10000,
        })
    })
  })
  await page.goto(base + hash)
  await page.waitForTimeout(1500)
  if (after) await after(page)
  if (process.env.DUMP)
    console.log(
      await page.evaluate(
        (sel) =>
          [...document.querySelectorAll(sel)]
            .map((e) => {
              const c = getComputedStyle(e)
              const r = e.getBoundingClientRect()
              return [
                e.outerHTML.slice(0, 400),
                c.display,
                c.flexDirection,
                c.fontSize,
                c.lineHeight,
                c.color,
                c.height,
                r.width,
                r.height,
                c.overflow,
              ].join(' | ')
            })
            .join('\n'),
        process.env.DUMP,
      ),
    )
  await page.screenshot({ path: `${out}/${name}.png` })
  console.log(`${out}/${name}.png`)
  await ctx.close()
}

const desktop = { width: 1400, height: 850 }
const phone = { width: 390, height: 844, mobile: true }
await shot('login', { ...desktop, token: '' })
await shot('desktop', { ...desktop, side: false })
await shot('desktop-chat', { ...desktop, side: true, tab: 'chat', state: { lockedControls: true } })
await shot('desktop-settings', { ...desktop, side: true, tab: 'settings', state: { host: 's1' } })
await shot('desktop-narrow', { width: 900, height: 700, side: false })
await shot('phone', { ...phone, side: false })
await shot('phone-chat', { ...phone, side: true, tab: 'chat' })
await shot('phone-landscape', { width: 844, height: 390, mobile: true, side: false })
// a view-only share link (#/<token>): its login screen, then the picture with the playback bar
const share = { hash: '#/ShareToken-12345', token: '', state: { viewOnly: true, host: 's3' } }
await shot('share-login', { ...desktop, ...share })
await shot('share-desktop', { ...desktop, ...share, viewer: 'Gast' })
await shot('share-phone', { ...phone, ...share, viewer: 'Gast' })
// HLS (?media=hls): starting, failed with Retry / WebRTC, the choice in the settings, and on a share link
await shot('no-webrtc', { ...phone, webrtc: false })
await shot('hls-connecting', { ...desktop, hash: '?media=hls', side: false })
await shot('hls-failed', { ...phone, hash: '?media=hls', side: false, hls: 'fail' })
await shot('desktop-media', { ...desktop, side: true, tab: 'settings', state: { host: 's1' } })
// touch input: the trackpad's pointer and the special keys' bar on a phone; the settings with force touch
await shot('phone-trackpad', {
  ...phone,
  side: false,
  state: { host: 's1' },
  stored: { trackpad_mode: '1' },
  after: (page) => page.getByTestId('keyboard-helper-toggle').click(),
})
await shot('desktop-input', {
  ...desktop,
  side: true,
  tab: 'settings',
  state: { host: 's1' },
  stored: { force_touch: '1', trackpad_mode: '1' },
})
await shot('share-hls', { ...phone, ...share, hash: '?media=ll-hls#/ShareToken-12345', viewer: 'Gast', hls: 'fail' })

await browser.close()
await server.close()
