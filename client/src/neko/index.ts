import Vue from 'vue'
import EventEmitter from 'eventemitter3'
import { BaseClient, BaseEvents } from './base'
import { Member } from './types'
import { EVENT } from './events'
import { accessor } from '~/store'

import {
  SystemMessagePayload,
  MemberListPayload,
  MemberDisconnectPayload,
  MemberPayload,
  ControlPayload,
  ControlTargetPayload,
  ChatPayload,
  EmotePayload,
  ControlClipboardPayload,
  ScreenConfigurationsPayload,
  ScreenResolutionPayload,
  BroadcastStatusPayload,
  AdminTargetPayload,
  AdminLockMessage,
  SystemInitPayload,
  AdminLockResource,
  FileTransferListPayload,
} from './messages'

interface NekoEvents extends BaseEvents {}

export class NekoClient extends BaseClient implements EventEmitter<NekoEvents> {
  private $vue!: Vue
  private $accessor!: typeof accessor
  private url!: string

  public isDemo = false
  private demoInterval?: any
  private simMouseX = 640
  private simMouseY = 360
  private simKeys: string[] = []
  private simRipples: Array<{ x: number; y: number; r: number; max: number }> = []
  private simMembers: any[] = []

  public setDemoMode(isDemo: boolean) {
    this.isDemo = isDemo
  }

  init(vue: Vue) {
    let port: string | undefined = undefined
    try {
      if (typeof process !== 'undefined' && process.env) {
        port = process.env.VUE_APP_SERVER_PORT
      }
    } catch (e) {}

    try {
      const meta = import.meta as any
      if (meta && meta.env) {
        port = port || (meta.env.VITE_APP_SERVER_PORT as string)
      }
    } catch (e) {}

    if (!port || port === 'undefined') {
      port = location.port || '8080'
    }

    let isDev = false
    try {
      if (typeof process !== 'undefined' && process.env) {
        isDev = process.env.NODE_ENV === 'development'
      } else {
        const meta = import.meta as any
        if (meta && meta.env) {
          isDev = meta.env.DEV
        }
      }
    } catch (e) {}

    const isSecure = location.protocol === 'https:'
    const wsScheme = isSecure ? 'wss' : 'ws'
    const hostname = location.hostname

    let basePath = location.pathname
    // strip standard file names if any to avoid invalid ws paths on HTTPS
    basePath = basePath.replace(/\/(index|login)\.html?$/i, '/')
    basePath = basePath.replace(/\/$/, '')

    let url = ''
    if (isDev) {
      if (hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '0.0.0.0') {
        url = `${wsScheme}://${hostname}:${port}/ws`
      } else {
        url = `${wsScheme}://${location.host}/ws`
      }
    } else {
      url = `${wsScheme}://${location.host}${basePath}/ws`
    }

    this.initWithURL(vue, url)
  }

  initWithURL(vue: Vue, url: string) {
    this.$vue = vue
    this.$accessor = vue.$accessor
    this.url = url
    // convert ws url to http url
    this.$vue.$http.defaults.baseURL = url.replace(/^ws/, 'http').replace(/\/ws$/, '')
  }

  private cleanup() {
    this.$accessor.setConnected(false)
    this.$accessor.remote.reset()
    this.$accessor.user.reset()
    this.$accessor.video.reset()
    this.$accessor.chat.reset()
  }

  sendData(event: string, data: any) {
    if (this.isDemo) {
      if (event === 'mousemove') {
        this.simMouseX = data.x
        this.simMouseY = data.y
      } else if (event === 'mousedown') {
        this.simRipples.push({ x: this.simMouseX, y: this.simMouseY, r: 1, max: 25 })
      } else if (event === 'keydown') {
        const keyName = this.getKeyName(data.key)
        this.simKeys.push(keyName)
        if (this.simKeys.length > 5) this.simKeys.shift()
      }
      return
    }
    super.sendData(event as any, data)
  }

  login(password: string, displayname: string) {
    if (this.isDemo || password.toLowerCase() === 'demo') {
      this.startDemo(displayname)
    } else {
      this.connect(this.url, password, displayname)
    }
  }

  logout() {
    this.stopDemo()
    this.disconnect()
    this.cleanup()
    this.$vue.$swal({
      title: this.$vue.$t('connection.logged_out'),
      icon: 'info',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    }).then(() => {
      window.location.reload()
    })
  }

  private startDemo(displayname: string) {
    this.isDemo = true
    this._displayname = displayname
    this[EVENT.CONNECTING]()

    setTimeout(() => {
      if (!this.isDemo) return

      this._id = 'demo-user-id'
      this._state = 'connected'

      this._ws = {
        readyState: WebSocket.OPEN,
        send: (msg: string) => {
          this.handleDemoWSMessage(msg)
        },
        close: () => {
          this.stopDemo()
        }
      } as any

      this._peer = {
        close: () => {
          this.stopDemo()
        }
      } as any

      this[EVENT.CONNECTED]()

      // Init system locks & files
      this[EVENT.SYSTEM.INIT]({
        implicit_hosting: false,
        locks: {},
        file_transfer: true,
        heartbeat_interval: 10,
      })

      this.simMembers = [
        { id: 'demo-user-id', displayname, admin: true, muted: false },
        { id: 'nekobot', displayname: 'NekoBot 🐱', admin: false, muted: false },
        { id: 'alice', displayname: 'Alice 🌸', admin: false, muted: false },
        { id: 'bob', displayname: 'Bob 🚀', admin: true, muted: false },
      ]

      // Set initial player list
      this[EVENT.MEMBER.LIST]({
        members: this.simMembers
      })

      // Set resolution mapping
      this[EVENT.SCREEN.RESOLUTION]({
        id: '',
        width: 1280,
        height: 720,
        rate: 30
      })

      // Send initial configurations list
      this[EVENT.SCREEN.CONFIGURATIONS]({
        configurations: {
          '0': { width: 1920, height: 1080, rates: { '0': 30, '1': 60 } },
          '1': { width: 1280, height: 720, rates: { '0': 30, '2': 60 } },
          '2': { width: 1024, height: 768, rates: { '0': 30, '3': 60 } },
          '3': { width: 800, height: 600, rates: { '0': 30 } }
        }
      })

      // Nice welcoming system chat lines
      setTimeout(() => {
        this[EVENT.CHAT.MESSAGE]({
          id: 'nekobot',
          content: 'Nyan! Hello ' + displayname + '! Willkommen im n.eko Sandbox-Modus! 🐱',
        })
      }, 700)

      setTimeout(() => {
        this[EVENT.CHAT.MESSAGE]({
          id: 'alice',
          content: 'Hi! Du kannst die Steuerung oben rechts (Maus-Symbol) aktivieren und den Cursor im Fenster bewegen!',
        })
      }, 1600)

      // Initialize the canvas and streamer
      this.initDemoStream()
    }, 600)
  }

  private stopDemo() {
    this.isDemo = false
    if (this.demoInterval) {
      clearInterval(this.demoInterval)
      this.demoInterval = undefined
    }
    this._ws = undefined
    this._peer = undefined
  }

  private handleDemoWSMessage(msgJson: string) {
    try {
      const data = JSON.parse(msgJson)
      const event = data.event
      const payload = data

      if (event === EVENT.CHAT.MESSAGE) {
        const text = payload.content
        setTimeout(() => {
          this[EVENT.CHAT.MESSAGE]({
            id: 'demo-user-id',
            content: text
          })

          setTimeout(() => {
            let reply = ''
            const lower = text.toLowerCase()
            if (lower.includes('hallo') || lower.includes('hi') || lower.includes('hello')) {
              reply = 'Hallo! Wie gefällt dir das n.eko Interface? Es läuft alles direkt in-browser!'
            } else if (lower.includes('maus') || lower.includes('steuerung') || lower.includes('mouse') || lower.includes('control')) {
              reply = 'Wenn du die Steuerung übernimmst, siehst du den Mauszeiger und Klick-Wellen live auf dem Desktop!'
            } else if (lower.includes('cat') || lower.includes('neko') || lower.includes('katze')) {
              reply = 'Miau! 🐱 Katzen sind hier herzlich willkommen!'
            } else {
              reply = 'Super! Ich habe deine Nachricht erhalten: "' + text + '". Der Chat funktioniert einwandfrei!'
            }

            const responders = ['nekobot', 'alice', 'bob']
            const responder = responders[Math.floor(Math.random() * responders.length)]
            this[EVENT.CHAT.MESSAGE]({
              id: responder,
              content: reply
            })
          }, 850)
        }, 50)
      } else if (event === EVENT.CONTROL.REQUEST) {
        setTimeout(() => {
          this[EVENT.CONTROL.LOCKED]({
            id: 'demo-user-id'
          })
          this[EVENT.CHAT.MESSAGE]({
            id: 'nekobot',
            content: 'Du steuerst jetzt den Desktop! Bewege/klicke deine Maus über dem Bildschirm, um Interaktionen zu testen.'
          })
        }, 100)
      } else if (event === EVENT.CONTROL.RELEASE) {
        setTimeout(() => {
          this[EVENT.CONTROL.RELEASE]({
            id: 'demo-user-id'
          })
        }, 100)
      } else if (event === EVENT.CHAT.EMOTE) {
        setTimeout(() => {
          this[EVENT.CHAT.EMOTE]({
            id: 'demo-user-id',
            emote: payload.emote
          })
        }, 50)
      } else if (event === EVENT.FILETRANSFER.REFRESH) {
        setTimeout(() => {
          this[EVENT.FILETRANSFER.LIST]({
            cwd: '/home/neko/Downloads',
            files: [
              { name: 'lustiges_katzenbild.jpg', type: 'file', size: 1048576 },
              { name: 'neko_simulations_anleitung.txt', type: 'file', size: 2450 },
              { name: 'Web_Projekte_Ordner', type: 'dir', size: 0 },
              { name: 'mein_privater_schluessel.pem', type: 'file', size: 1024 }
            ]
          })
        }, 200)
      } else if (event === EVENT.SCREEN.SET) {
        setTimeout(() => {
          this[EVENT.SCREEN.RESOLUTION]({
            id: 'demo-user-id',
            width: payload.width,
            height: payload.height,
            rate: payload.rate
          })
        }, 200)
      } else if (event === EVENT.SCREEN.CONFIGURATIONS) {
        setTimeout(() => {
          this[EVENT.SCREEN.CONFIGURATIONS]({
            configurations: {
              '0': { width: 1920, height: 1080, rates: { '0': 30, '1': 60 } },
              '1': { width: 1280, height: 720, rates: { '0': 30, '2': 60 } },
              '2': { width: 1024, height: 768, rates: { '0': 30, '3': 60 } },
              '3': { width: 800, height: 600, rates: { '0': 30 } }
            }
          })
        }, 100)
      } else if (event === EVENT.ADMIN.LOCK) {
        setTimeout(() => {
          this[EVENT.ADMIN.LOCK]({
            event: EVENT.ADMIN.LOCK,
            id: 'demo-user-id',
            resource: payload.resource
          })
        }, 100)
      } else if (event === EVENT.ADMIN.UNLOCK) {
        setTimeout(() => {
          this[EVENT.ADMIN.UNLOCK]({
            event: EVENT.ADMIN.UNLOCK,
            id: 'demo-user-id',
            resource: payload.resource
          })
        }, 100)
      } else if (event === EVENT.ADMIN.MUTE) {
        setTimeout(() => {
          this.simMembers = this.simMembers.map(m => m.id === payload.id ? { ...m, muted: true } : m)
          this[EVENT.ADMIN.MUTE]({
            id: 'demo-user-id',
            target: payload.id
          })
        }, 100)
      } else if (event === EVENT.ADMIN.UNMUTE) {
        setTimeout(() => {
          this.simMembers = this.simMembers.map(m => m.id === payload.id ? { ...m, muted: false } : m)
          this[EVENT.ADMIN.UNMUTE]({
            id: 'demo-user-id',
            target: payload.id
          })
        }, 100)
      } else if (event === EVENT.ADMIN.BAN) {
        setTimeout(() => {
          const targetId = payload.id
          this[EVENT.ADMIN.BAN]({
            id: 'demo-user-id',
            target: targetId
          })
          this.simMembers = this.simMembers.filter(m => m.id !== targetId)
          this[EVENT.MEMBER.DISCONNECTED]({
            id: targetId
          })
        }, 100)
      } else if (event === EVENT.ADMIN.KICK) {
        setTimeout(() => {
          const targetId = payload.id
          this[EVENT.ADMIN.KICK]({
            id: 'demo-user-id',
            target: targetId
          })
          this.simMembers = this.simMembers.filter(m => m.id !== targetId)
          this[EVENT.MEMBER.DISCONNECTED]({
            id: targetId
          })
        }, 100)
      } else if (event === EVENT.CONTROL.GIVE || event === EVENT.ADMIN.GIVE) {
        setTimeout(() => {
          const handler = event === EVENT.CONTROL.GIVE ? this[EVENT.CONTROL.GIVE] : this[EVENT.ADMIN.GIVE]
          handler.call(this, {
            id: 'demo-user-id',
            target: payload.id
          })
        }, 100)
      } else if (event === EVENT.ADMIN.CONTROL) {
        setTimeout(() => {
          this[EVENT.ADMIN.CONTROL]({
            id: 'demo-user-id',
            target: this.$accessor.remote.id
          })
        }, 100)
      } else if (event === EVENT.ADMIN.RELEASE) {
        setTimeout(() => {
          this[EVENT.ADMIN.RELEASE]({
            id: 'demo-user-id',
            target: this.$accessor.remote.id
          })
        }, 100)
      }
    } catch (e) {
      console.error('Demo WebSocket parse/exec error:', e)
    }
  }

  private initDemoStream() {
    const canvas = document.createElement('canvas')
    canvas.width = 1280
    canvas.height = 720
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let bounceX = 200
    let bounceY = 200
    let bounceDX = 4
    let bounceDY = 3
    let gradAngle = 0

    const drawFrame = () => {
      if (!this.isDemo) return

      ctx.fillStyle = '#000000'
      ctx.fillRect(0, 0, canvas.width, canvas.height)

      gradAngle += 0.005
      const grad = ctx.createLinearGradient(
        640 + Math.cos(gradAngle) * 300,
        360 + Math.sin(gradAngle) * 200,
        640 - Math.cos(gradAngle) * 300,
        360 - Math.sin(gradAngle) * 200
      )
      grad.addColorStop(0, '#101216')
      grad.addColorStop(0.5, '#171a21')
      grad.addColorStop(1, '#0c0e12')
      ctx.fillStyle = grad
      ctx.fillRect(0, 0, canvas.width, canvas.height)

      ctx.strokeStyle = 'rgba(114, 137, 218, 0.04)'
      ctx.lineWidth = 1
      for (let x = 0; x < canvas.width; x += 40) {
        ctx.beginPath()
        ctx.moveTo(x, 0)
        ctx.lineTo(x, canvas.height)
        ctx.stroke()
      }
      for (let y = 0; y < canvas.height; y += 40) {
        ctx.beginPath()
        ctx.moveTo(0, y)
        ctx.lineTo(canvas.width, y)
        ctx.stroke()
      }

      ctx.fillStyle = 'rgba(23, 26, 33, 0.85)'
      ctx.strokeStyle = 'rgba(114, 137, 218, 0.4)'
      ctx.lineWidth = 2
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(80, 80, 1120, 560, 10)
      } else {
        ctx.rect(80, 80, 1120, 560)
      }
      ctx.fill()
      ctx.stroke()

      ctx.fillStyle = 'rgba(12, 14, 18, 0.95)'
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(80, 80, 1120, 50, [10, 10, 0, 0])
      } else {
        ctx.rect(80, 80, 1120, 50)
      }
      ctx.fill()
      ctx.stroke()

      const dotColors = ['#ff5f56', '#ffbd2e', '#27c93f']
      dotColors.forEach((color, i) => {
        ctx.fillStyle = color
        ctx.beginPath()
        ctx.arc(110 + i * 20, 105, 6, 0, Math.PI * 2)
        ctx.fill()
      })

      ctx.fillStyle = '#8e9297'
      ctx.font = 'bold 15px sans-serif'
      ctx.fillText('n.eko Workspace - Simulated Environment (AISTUDIO_CONTAINER)', 180, 111)

      ctx.fillStyle = 'rgba(8, 10, 13, 0.6)'
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(110, 160, 400, 240, 8)
      } else {
        ctx.rect(110, 160, 400, 240)
      }
      ctx.fill()

      ctx.fillStyle = '#ffffff'
      ctx.font = '16px monospace'
      ctx.fillText('⚡ CPU: ' + (15 + Math.sin(Date.now() / 1500) * 8).toFixed(1) + '%', 140, 200)
      ctx.fillText('💾 RAM: 4.1 GB / 8.0 GB', 140, 230)
      ctx.fillText('📡 Ping: ' + (5 + Math.floor(Math.random() * 3)) + ' ms', 140, 260)
      ctx.fillText('📊 FPS: 30.0 (Smooth)', 140, 290)
      ctx.fillText('📁 Mount: /home/neko/Downloads', 140, 320)
      ctx.fillText('⚙️ OS: Alpine Linux x86_64', 140, 350)

      ctx.fillStyle = 'rgba(8, 10, 13, 0.6)'
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(870, 160, 290, 80, 8)
      } else {
        ctx.rect(870, 160, 290, 80)
      }
      ctx.fill()

      ctx.fillStyle = '#7289da'
      ctx.font = 'bold 14px monospace'
      ctx.fillText('Live Server-Zeit', 890, 192)
      ctx.fillStyle = '#ffffff'
      ctx.font = 'bold 22px monospace'
      ctx.fillText(new Date().toLocaleTimeString(), 890, 222)

      ctx.fillStyle = 'rgba(114, 137, 218, 0.15)'
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(110, 420, 1050, 180, 8)
      } else {
        ctx.rect(110, 420, 1050, 180)
      }
      ctx.fill()
      ctx.strokeStyle = 'rgba(114, 137, 218, 0.3)'
      ctx.stroke()

      ctx.fillStyle = '#ffffff'
      ctx.font = 'bold 20px sans-serif'
      ctx.fillText('🚀 n.eko Client Simulations-Modus Aktiv!', 140, 465)
      ctx.font = '15px sans-serif'
      ctx.fillStyle = '#b9bbbe'
      ctx.fillText('Dieses simulierte WebRTC-Videobild läuft komplett lokal im Browser, um KVM/X11-Containerbeschränkungen zu umgehen.', 140, 495)
      ctx.fillText('• Du kannst im Chat-Menü rechts Nachrichten schreiben. Unsere Bots antworten dir sofort!', 140, 520)
      ctx.fillText('• Aktiviere die Steuerung, um Mauszeiger und Klick-Schockwellen live auf dem Bildschirm zu sehen!', 140, 545)
      ctx.fillText('• Durchsuche Dateien, schalte den Ton stumm, passe das Layout an und verwalte die Benutzerliste.', 140, 570)

      bounceX += bounceDX
      bounceY += bounceDY
      if (bounceX < 550 || bounceX > 820) bounceDX = -bounceDX
      if (bounceY < 180 || bounceY > 370) bounceDY = -bounceDY

      ctx.fillStyle = 'rgba(114, 137, 218, 0.9)'
      ctx.beginPath()
      if ((ctx as any).roundRect) {
        ;(ctx as any).roundRect(bounceX, bounceY, 120, 80, 6)
      } else {
        ctx.rect(bounceX, bounceY, 120, 80)
      }
      ctx.fill()

      ctx.fillStyle = '#ffffff'
      ctx.font = 'bold 15px sans-serif'
      ctx.fillText('🐱 n.eko', bounceX + 15, bounceY + 35)
      ctx.font = '12px monospace'
      ctx.fillText('bounce.exe', bounceX + 15, bounceY + 58)

      this.simRipples.forEach((ripple) => {
        ripple.r += 1.5
        ctx.beginPath()
        ctx.arc(ripple.x, ripple.y, ripple.r, 0, Math.PI * 2)
        ctx.strokeStyle = 'rgba(114, 137, 218, ' + (1 - ripple.r / ripple.max) + ')'
        ctx.lineWidth = 3
        ctx.stroke()
      })
      this.simRipples = this.simRipples.filter((ripple) => ripple.r < ripple.max)

      ctx.fillStyle = '#ffffff'
      ctx.strokeStyle = '#000000'
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.moveTo(this.simMouseX, this.simMouseY)
      ctx.lineTo(this.simMouseX + 15, this.simMouseY + 15)
      ctx.lineTo(this.simMouseX + 6, this.simMouseY + 17)
      ctx.closePath()
      ctx.fill()
      ctx.stroke()

      if (this.simKeys.length > 0) {
        ctx.fillStyle = 'rgba(0,0,0,0.7)'
        ctx.beginPath()
        if ((ctx as any).roundRect) {
          ;(ctx as any).roundRect(550, 160, 300, 45, 6)
        } else {
          ctx.rect(550, 160, 300, 45)
        }
        ctx.fill()

        ctx.fillStyle = '#27c93f'
        ctx.font = 'bold 14px monospace'
        ctx.fillText('⌨️ Tasten-Input: ' + this.simKeys.join(' '), 565, 187)
      }
    }

    this.demoInterval = window.setInterval(drawFrame, 33)

    let stream: MediaStream | null = null
    const c = canvas as any
    if (c.captureStream) {
      stream = c.captureStream(15)
    } else if (c.mozCaptureStream) {
      stream = c.mozCaptureStream(15)
    }

    if (stream) {
      const track = stream.getVideoTracks()[0]
      this.$accessor.video.addTrack([track, stream])
      this.$accessor.video.setStream(0)
      this.$accessor.video.setPlayable(true)
      this.$accessor.video.play()
    }
  }

  private getKeyName(code: number): string {
    if (code >= 0x41 && code <= 0x5a) return String.fromCharCode(code)
    if (code >= 0x61 && code <= 0x7a) return String.fromCharCode(code - 32)
    if (code === 0xff0d) return 'Enter'
    if (code === 0xff08) return 'Back'
    if (code === 0xff09) return 'Tab'
    if (code === 0xff1b) return 'Esc'
    if (code === 0x0020) return 'Leertaste'
    return 'Key' + code.toString(16)
  }

  /////////////////////////////
  // Internal Events
  /////////////////////////////
  protected [EVENT.RECONNECTING]() {
    this.$vue.$notify({
      group: 'neko',
      type: 'warning',
      title: this.$vue.$t('connection.reconnecting') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONNECTING]() {
    this.$accessor.setConnnecting()
  }

  protected [EVENT.CONNECTED]() {
    this.$accessor.user.setMember(this.id)
    this.$accessor.setConnected(true)

    this.$vue.$notify({
      group: 'neko',
      clean: true,
    })

    this.$vue.$notify({
      group: 'neko',
      type: 'success',
      title: this.$vue.$t('connection.connected') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.DISCONNECTED](reason?: Error) {
    this.cleanup()

    this.$vue.$notify({
      group: 'neko',
      type: 'error',
      title: this.$vue.$t('connection.disconnected') as string,
      text: reason ? reason.message : undefined,
      duration: 5000,
      speed: 1000,
    })

    try {
      if (typeof window !== 'undefined' && window.location) {
        const hostname = window.location.hostname
        if (
          hostname.includes('ais-dev-') ||
          hostname.includes('ais-pre-') ||
          hostname.includes('localhost') ||
          hostname.includes('127.0.0.1') ||
          hostname.includes('0.0.0.0')
        ) {
          const isRealAttempt = this._displayname && !this.isDemo
          if (isRealAttempt) {
            (this.$vue as any).$swal({
              title: 'Verbindung fehlgeschlagen',
              text: 'In der AI Studio Vorschau läuft standardmäßig kein Neko-Backend auf diesem Server. Möchtest du im simulierten Desktop-Demo-/Testmodus fortfahren?',
              icon: 'question',
              showCancelButton: true,
              confirmButtonText: 'Ja, Demo/Test starten',
              cancelButtonText: 'Nein, abbrechen',
            }).then((result: any) => {
              if (result && result.value) {
                this.startDemo(this._displayname || 'neko')
              }
            })
          }
        }
      }
    } catch (err) {}
  }

  protected [EVENT.TRACK](event: RTCTrackEvent) {
    const { track, streams } = event
    const stream = streams[0] || new MediaStream([track])

    // Log track arrival for debugging stream stability issues
    console.log(`[Neko] Received ${track.kind} track: id=${track.id} readyState=${track.readyState}`)

    if (track.kind === 'audio') {
      // Audio tracks are stored but do not change the active stream index
      this.$accessor.video.addTrack([track, stream])
      return
    }

    // Video track: addTrack will replace any existing video track (not accumulate)
    this.$accessor.video.addTrack([track, stream])
    this.$accessor.video.setStream(0)
    this.$accessor.video.setPlayable(true)
  }

  protected [EVENT.DATA](data: any) {
    if (!(data instanceof ArrayBuffer)) return
    if (data.byteLength < 3) return

    const view = new DataView(data)
    const event = view.getUint8(0)

    if (event === 0x01) { // OP_CURSOR_POSITION
      if (data.byteLength < 7) return
      const x = view.getUint16(3, false) // big endian
      const y = view.getUint16(5, false) // big endian

      this.emit('cursor-position', { x, y })
    }
  }

  /////////////////////////////
  // System Events
  /////////////////////////////
  protected [EVENT.SYSTEM.INIT]({ implicit_hosting, locks, file_transfer, heartbeat_interval, screen_size }: SystemInitPayload) {
    this.$accessor.remote.setImplicitHosting(implicit_hosting)
    this.$accessor.remote.setFileTransfer(file_transfer)

    if (screen_size) {
      this.$accessor.video.setResolution(screen_size)
    }

    for (const resource in locks) {
      this[EVENT.ADMIN.LOCK]({
        event: EVENT.ADMIN.LOCK,
        resource: resource as AdminLockResource,
        id: locks[resource],
      })
    }

    if (heartbeat_interval > 0) {
      if (this._ws_heartbeat) clearInterval(this._ws_heartbeat)
      this._ws_heartbeat = window.setInterval(() => this.sendMessage(EVENT.CLIENT.HEARTBEAT), heartbeat_interval * 1000)
    }
  }

  protected [EVENT.SYSTEM.DISCONNECT]({ message }: SystemMessagePayload) {
    if (message == 'kicked') {
      this.$accessor.logout()
      message = this.$vue.$t('connection.kicked') as string
    }

    this.onDisconnected(new Error(message))

    this.$vue.$swal({
      title: this.$vue.$t('connection.disconnected'),
      text: message,
      icon: 'error',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  protected [EVENT.SYSTEM.ERROR]({ title, message }: SystemMessagePayload) {
    this.$vue.$swal({
      title,
      text: message,
      icon: 'error',
      confirmButtonText: this.$vue.$t('connection.button_confirm') as string,
    })
  }

  /////////////////////////////
  // Member Events
  /////////////////////////////
  protected [EVENT.MEMBER.LIST]({ members }: MemberListPayload) {
    this.$accessor.user.setMembers(members)
    this.$accessor.chat.newMessage({
      id: this.id,
      content: this.$vue.$t('notifications.connected', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.MEMBER.CONNECTED](member: MemberPayload) {
    this.$accessor.user.addMember(member)

    if (member.id !== this.id) {
      this.$accessor.chat.newMessage({
        id: member.id,
        content: this.$vue.$t('notifications.connected', { name: '' }) as string,
        type: 'event',
        created: new Date(),
      })
    }
  }

  protected [EVENT.MEMBER.DISCONNECTED]({ id }: MemberDisconnectPayload) {
    const member = this.member(id)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.disconnected', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })

    this.$accessor.user.delMember(id)
  }

  /////////////////////////////
  // Control Events
  /////////////////////////////
  protected [EVENT.CONTROL.LOCKED]({ id }: ControlPayload) {
    this.$accessor.remote.setHost(id)
    this.$accessor.remote.changeKeyboard()

    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: this.$vue.$t('notifications.controls_taken', {
          name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
        }) as string,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.controls_taken', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.RELEASE]({ id }: ControlPayload) {
    this.$accessor.remote.reset()
    const member = this.member(id)
    if (!member) {
      return
    }

    if (this.id === id) {
      this.$vue.$notify({
        group: 'neko',
        type: 'info',
        title: this.$vue.$t('notifications.controls_released', {
          name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
        }) as string,
        duration: 5000,
        speed: 1000,
      })
    }

    this.$accessor.chat.newMessage({
      id: member.id,
      content: this.$vue.$t('notifications.controls_released', { name: '' }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.REQUEST]({ id }: ControlPayload) {
    const member = this.member(id)
    if (!member) {
      return
    }

    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: this.$vue.$t('notifications.controls_has', { name: member.displayname }) as string,
      text: this.$vue.$t('notifications.controls_has_alt') as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONTROL.REQUESTING]({ id }: ControlPayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$vue.$notify({
      group: 'neko',
      type: 'info',
      title: this.$vue.$t('notifications.controls_requesting', { name: member.displayname }) as string,
      duration: 5000,
      speed: 1000,
    })
  }

  protected [EVENT.CONTROL.GIVE]({ id, target }: ControlTargetPayload) {
    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.remote.setHost(member)
    this.$accessor.remote.changeKeyboard()

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_given', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.CONTROL.CLIPBOARD]({ text }: ControlClipboardPayload) {
    this.$accessor.remote.setClipboard(text)
  }

  /////////////////////////////
  // Chat Events
  /////////////////////////////
  protected [EVENT.CHAT.MESSAGE]({ id, content }: ChatPayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content,
      type: 'text',
      created: new Date(),
    })
  }

  protected [EVENT.CHAT.EMOTE]({ id, emote }: EmotePayload) {
    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newEmote({ type: emote })
  }

  /////////////////////////////
  // File Transfer Events
  /////////////////////////////
  protected [EVENT.FILETRANSFER.LIST]({ cwd, files }: FileTransferListPayload) {
    this.$accessor.files.setCwd(cwd)
    this.$accessor.files.setFileList(files)
  }

  /////////////////////////////
  // Screen Events
  /////////////////////////////
  protected [EVENT.SCREEN.CONFIGURATIONS]({ configurations }: ScreenConfigurationsPayload) {
    this.$accessor.video.setConfigurations(configurations)
  }

  protected ['screen/updated'](payload: ScreenResolutionPayload) {
    this[EVENT.SCREEN.RESOLUTION](payload)
  }

  protected [EVENT.SCREEN.RESOLUTION]({ id, width, height, rate }: ScreenResolutionPayload) {
    this.$accessor.video.setResolution({ width, height, rate })

    if (!id) {
      return
    }

    const member = this.member(id)
    if (!member || member.ignored) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.resolution', {
        width: width,
        height: height,
        rate: rate,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  /////////////////////////////
  // Broadcast Events
  /////////////////////////////
  protected [EVENT.BROADCAST.STATUS](payload: BroadcastStatusPayload) {
    this.$accessor.settings.broadcastStatus(payload)
  }

  /////////////////////////////
  // Admin Events
  /////////////////////////////
  protected [EVENT.ADMIN.BAN]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.banned', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.KICK]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.kicked', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.MUTE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    this.$accessor.user.setMuted({ id: target, muted: true })

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.muted', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.UNMUTE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    this.$accessor.user.setMuted({ id: target, muted: false })

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.unmuted', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.LOCK]({ id, resource }: AdminLockMessage) {
    this.$accessor.setLocked(resource)

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t(`locks.${resource}.notif_locked`) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.UNLOCK]({ id, resource }: AdminLockMessage) {
    this.$accessor.setUnlocked(resource)

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t(`locks.${resource}.notif_unlocked`) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.CONTROL]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.setHost(id)
    this.$accessor.remote.changeKeyboard()

    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: this.$vue.$t('notifications.controls_taken_force') as string,
        type: 'event',
        created: new Date(),
      })
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_taken_steal', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.RELEASE]({ id, target }: AdminTargetPayload) {
    this.$accessor.remote.reset()
    if (!target) {
      this.$accessor.chat.newMessage({
        id,
        content: this.$vue.$t('notifications.controls_released_force') as string,
        type: 'event',
        created: new Date(),
      })
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_released_steal', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  protected [EVENT.ADMIN.GIVE]({ id, target }: AdminTargetPayload) {
    if (!target) {
      return
    }

    const member = this.member(target)
    if (!member) {
      return
    }

    this.$accessor.remote.setHost(member)
    this.$accessor.remote.changeKeyboard()

    this.$accessor.chat.newMessage({
      id,
      content: this.$vue.$t('notifications.controls_given', {
        name: member.id == this.id && this.$vue.$te('you') ? this.$vue.$t('you') : member.displayname,
      }) as string,
      type: 'event',
      created: new Date(),
    })
  }

  // Utilities
  protected member(id: string): Member | undefined {
    return this.$accessor.user.members[id]
  }
}
