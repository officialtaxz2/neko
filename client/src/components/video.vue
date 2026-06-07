<template>
  <div ref="component" class="video">
    <div ref="player" :class="['player', hosted ? 'has-control' : 'no-control']">
      <div ref="container" class="player-container">
        <video ref="video" playsinline />
        <div class="emotes">
          <template v-for="(emote, index) in emotes">
            <neko-emote :id="index" :key="index" />
          </template>
        </div>
        <textarea
          ref="overlay"
          class="overlay"
          spellcheck="false"
          tabindex="0"
          data-gramm="false"
          :style="{ pointerEvents: hosting ? 'auto' : 'none' }"
          @click.stop.prevent
          @contextmenu.stop.prevent
          @wheel.stop.prevent="onWheel"
          @mousemove.stop.prevent="onMouseMove"
          @mousedown.stop.prevent="onMouseDown"
          @mouseup.stop.prevent="onMouseUp"
          @mouseenter.stop.prevent="onMouseEnter"
          @mouseleave.stop.prevent="onMouseLeave"
          @touchmove.stop.prevent="onTouchHandler"
          @touchstart.stop.prevent="onTouchHandler"
          @touchend.stop.prevent="onTouchHandler"
          @compositionstart="onCompositionStartHandler"
          @compositionend="onCompositionEndHandler"
        />
        <div v-if="!playing && playable" class="player-overlay" @click.stop.prevent="playAndUnmute">
          <i class="fas fa-play-circle" />
        </div>
        <div v-else-if="mutedOverlay && muted" class="player-overlay" @click.stop.prevent="unmute">
          <i class="fas fa-volume-up" />
        </div>
        <div
          v-if="trackpadActive && hosting && !locked"
          class="trackpad-cursor"
          :style="{ left: cursorX + 'px', top: cursorY + 'px' }"
        />
        <div ref="aspect" class="player-aspect" />
      </div>
      <ul v-if="!fullscreen && !hideControls" class="video-menu top">
        <li><i @click.stop.prevent="requestFullscreen" class="fas fa-expand"></i></li>
        <li v-if="admin"><i @click.stop.prevent="openResolution" class="fas fa-desktop"></i></li>
        <li v-if="!controlLocked && !implicitHosting" :class="extraControls || 'extra-control'">
          <i
            :class="[
              hosted && !hosting ? 'disabled' : '',
              !hosted && !hosting ? 'faded' : '',
              'fas',
              'fa-computer-mouse',
            ]"
            @click.stop.prevent="toggleControl"
          />
        </li>
      </ul>
      <ul v-if="!fullscreen && !hideControls" class="video-menu bottom">
        <li v-if="hosting">
          <i
            @click.stop.prevent="openClipboard"
            v-tooltip="{ content: $t('clipboard_manager.history_title'), placement: 'left', offset: 5 }"
            class="fas fa-clipboard"
          ></i>
        </li>
        <li>
          <i
            v-if="pip_available"
            @click.stop.prevent="requestPictureInPicture"
            v-tooltip="{ content: 'Picture-in-Picture', placement: 'left', offset: 5, boundariesElement: 'body' }"
            class="fas fa-external-link-alt"
          />
        </li>
        <li
          v-if="hosting && is_touch_device"
          :class="extraControls || 'extra-control'"
          @click.stop.prevent="toggleKeyboardHelper"
        >
          <i
            class="fas fa-sliders-h"
            v-tooltip="{ content: $t('clipboard_manager.keyboard_helper_title'), placement: 'left', offset: 5 }"
          />
        </li>
        <li
          v-if="hosting && is_touch_device"
          :class="extraControls || 'extra-control'"
          @click.stop.prevent="openMobileKeyboard"
        >
          <i class="fas fa-keyboard" />
        </li>
      </ul>
      <neko-resolution ref="resolution" v-if="admin" />
      <neko-clipboard ref="clipboard" v-if="hosting" />
      <neko-keyboard-helper ref="keyboardHelper" :keyboard="keyboard" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .video {
    width: 100%;
    height: 100%;
    padding: 0;

    .player {
      position: absolute;
      top: 0;
      left: 0;
      display: flex;
      justify-content: center;
      align-items: center;
      background: transparent;
      overflow: visible; // Allows the beautiful glow to expand elegantly beyond the video player borders!

      &.no-control .player-container {
        border-color: rgba(124, 58, 237, 0.35);
        animation: breathGlowNeutral 6s ease-in-out infinite;
      }

      &.has-control .player-container {
        border-color: rgba(38, 230, 180, 0.45);
        animation: breathGlowCyber 4s ease-in-out infinite;
      }

      @keyframes breathGlowNeutral {
        0%, 100% {
          box-shadow: 
            0 0 15px rgba(124, 58, 237, 0.25),
            0 0 40px rgba(99, 102, 241, 0.3),
            0 0 80px rgba(99, 102, 241, 0.15);
        }
        50% {
          box-shadow: 
            0 0 25px rgba(124, 58, 237, 0.4),
            0 0 60px rgba(99, 102, 241, 0.45),
            0 0 110px rgba(99, 102, 241, 0.22);
        }
      }

      @keyframes breathGlowCyber {
        0%, 100% {
          box-shadow: 
            0 0 15px rgba(38, 230, 180, 0.3),
            0 0 40px rgba(38, 230, 180, 0.3),
            0 0 80px rgba(38, 230, 180, 0.15);
        }
        50% {
          box-shadow: 
            0 0 30px rgba(38, 230, 180, 0.5),
            0 0 65px rgba(38, 230, 180, 0.5),
            0 0 110px rgba(38, 230, 180, 0.25);
        }
      }

      .video-menu {
        position: absolute;
        right: 18px;
        background: rgba(12, 13, 18, 0.55);
        backdrop-filter: blur(12px);
        border: 1px solid var(--glass-border);
        border-radius: 30px;
        padding: 5px 6px;
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 6px;
        list-style: none;
        z-index: 10;
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);

        &.top {
          top: 18px;
        }

        &.bottom {
          bottom: 18px;
        }

        li {
          margin: 0;
          display: flex;

          i {
            width: 34px;
            height: 34px;
            background: transparent;
            border-radius: 50%;
            line-height: 34px;
            font-size: 15px;
            text-align: center;
            color: var(--text-subtle);
            cursor: pointer;
            transition: var(--transition-fluid);

            &:hover {
              background: rgba(255, 255, 255, 0.08);
              color: var(--color-cyber-mint);
              transform: scale(1.1);
              text-shadow: 0 0 8px rgba(38, 230, 180, 0.4);
            }

            &:active {
              transform: scale(0.95);
            }

            &.faded {
              color: rgba(255, 255, 255, 0.25);
            }

            &.disabled {
              color: var(--style-error);
              opacity: 0.6;
            }
          }

          &.extra-control {
            display: none;
          }
          @media (max-width: 1024px) {
            &.extra-control {
              display: block;
            }
          }
        }
      }

      .player-container {
        position: relative;
        width: 100%;
        max-width: calc(16 / 9 * 100vh);
        border-radius: var(--radius-outer);
        overflow: hidden;
        border: 1.5px solid rgba(255, 255, 255, 0.08);
        transition: border-color var(--transition-fluid), box-shadow var(--transition-fluid);
        isolation: isolate;

        video {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
          display: flex;
          background: #000;

          &::-webkit-media-controls {
            display: none !important;
          }
        }

        .player-overlay,
        .emotes {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
          overflow: hidden;
        }

        .player-overlay {
          background: radial-gradient(circle at 50% 50%, rgba(12, 13, 18, 0.7) 0%, rgba(5, 5, 8, 0.95) 100%);
          backdrop-filter: blur(14px) saturate(120%);
          display: flex;
          justify-content: center;
          align-items: center;
          cursor: pointer;
          z-index: 5;
          animation: overlayFadeIn 0.5s ease;

          i::before {
            font-size: 96px;
            text-align: center;
            color: var(--color-cyber-mint);
            text-shadow: 0 0 25px var(--color-cyber-mint-glow);
            transition: var(--transition-fluid);
          }

          &:hover i::before {
            color: var(--text-pure);
            transform: scale(1.15) rotate(5deg);
            filter: drop-shadow(0 0 35px var(--color-cyber-mint));
          }

          &.hidden {
            display: none;
          }
        }

        @keyframes overlayFadeIn {
          from {
            opacity: 0;
            backdrop-filter: blur(0px);
          }
          to {
            opacity: 1;
            backdrop-filter: blur(14px);
          }
        }

        .overlay {
          position: absolute;
          top: 0;
          bottom: 0;
          width: 100%;
          height: 100%;
          cursor: default;
          outline: none !important;
          border: 0;
          color: transparent;
          background: transparent;
          resize: none;
          box-shadow: none !important;

          &:focus, &:focus-visible {
            outline: none !important;
            box-shadow: none !important;
          }
        }

        .player-aspect {
          display: block;
          padding-bottom: 56.25%;
        }

        .trackpad-cursor {
          position: absolute;
          width: 16px;
          height: 16px;
          border: 1.5px solid var(--color-cyber-mint);
          background: rgba(12, 13, 18, 0.45);
          border-radius: 50%;
          pointer-events: none;
          transform: translate(-50%, -50%);
          z-index: 20;
          box-shadow: 0 0 10px var(--color-cyber-mint-glow);
          transition: transform 0.1s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s;
          
          &::after {
            content: '';
            position: absolute;
            top: 50%;
            left: 50%;
            width: 6px;
            height: 6px;
            background: var(--color-cyber-mint);
            border-radius: 50%;
            transform: translate(-50%, -50%);
            box-shadow: 0 0 6px var(--color-cyber-mint);
          }
          
          &:active, &.active {
            transform: translate(-50%, -50%) scale(0.85);
            box-shadow: 0 0 15px var(--color-cyber-mint);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue, Prop } from 'vue-property-decorator'
  import ResizeObserver from 'resize-observer-polyfill'
  import { elementRequestFullscreen, onFullscreenChange, isFullscreen, lockKeyboard, unlockKeyboard } from '~/utils'

  import Emote from './emote.vue'
  import Resolution from './resolution.vue'
  import Clipboard from './clipboard.vue'
  import KeyboardHelper from './keyboard_helper.vue'

  // @ts-ignore
  import GuacamoleKeyboard from '~/utils/guacamole-keyboard.ts'

  const WHEEL_LINE_HEIGHT = 19

  @Component({
    name: 'neko-video',
    components: {
      'neko-emote': Emote,
      'neko-resolution': Resolution,
      'neko-clipboard': Clipboard,
      'neko-keyboard-helper': KeyboardHelper,
    },
  })
  export default class NekoVideo extends Vue {
    @Ref('component') readonly _component!: HTMLElement
    @Ref('container') readonly _container!: HTMLElement
    @Ref('overlay') readonly _overlay!: HTMLTextAreaElement
    @Ref('aspect') readonly _aspect!: HTMLElement
    @Ref('player') readonly _player!: HTMLElement
    @Ref('video') readonly _video!: HTMLVideoElement
    @Ref('resolution') readonly _resolution!: Resolution
    @Ref('clipboard') readonly _clipboard!: Clipboard
    @Ref('keyboardHelper') readonly _keyboardHelper!: any

    // all controls are hidden (e.g. for cast mode)
    @Prop(Boolean) readonly hideControls!: boolean
    // extra controls are shown (e.g. for embed mode)
    @Prop(Boolean) readonly extraControls!: boolean

    private keyboard = GuacamoleKeyboard()
    private observer = new ResizeObserver(this.onResize.bind(this))
    private focused = false
    private fullscreen = false
    private mutedOverlay = true
    private lastTextAreaValue = ''

    private cursorX = 0
    private cursorY = 0
    private touchLastX = 0
    private touchLastY = 0
    private longPressTimer: any = null
    private hasMoved = false
    private didLongPress = false

    private onVideoCanPlayThrough = () => {
      if (!this._video) return
      this.$accessor.video.setPlayable(true)
      if (this.autoplay) {
        this.$nextTick(() => {
          if (this._video) {
            this.$accessor.video.play()
          }
        })
      }
    }

    private onVideoEnded = () => {
      this.$accessor.video.setPlayable(false)
    }

    private onVideoError = (event: ErrorEvent) => {
      this.$log.error(event.error)
      this.$accessor.video.setPlayable(false)
    }

    private onVideoVolumeChange = () => {
      if (!this._video) return
      this.$accessor.video.setMuted(this._video.muted)
      this.$accessor.video.setVolume(this._video.volume * 100)
    }

    private onVideoPlaying = () => {
      this.$accessor.video.play()
    }

    private onVideoPause = () => {
      this.$accessor.video.pause()
    }

    private onFullscreenChangeHandler = () => {
      this.fullscreen = isFullscreen()
      this.fullscreen ? lockKeyboard() : unlockKeyboard()
      this.onResize()
    }

    get admin() {
      return this.$accessor.user.admin
    }

    get connected() {
      return this.$accessor.connected
    }

    get connecting() {
      return this.$accessor.connecting
    }

    get controlling() {
      return this.$accessor.remote.controlling
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
    }

    get hosted() {
      return this.$accessor.remote.hosted
    }

    get volume() {
      return this.$accessor.video.volume
    }

    get muted() {
      return this.$accessor.video.muted
    }

    get stream() {
      return this.$accessor.video.stream
    }

    get playing() {
      return this.$accessor.video.playing
    }

    get playable() {
      return this.$accessor.video.playable
    }

    get emotes() {
      return this.$accessor.chat.emotes
    }

    get autoplay() {
      return this.$accessor.settings.autoplay
    }

    // server-side lock
    get controlLocked() {
      return 'control' in this.$accessor.locked && this.$accessor.locked['control'] && !this.$accessor.user.admin
    }

    get locked() {
      return this.$accessor.remote.locked || (this.controlLocked && (!this.hosting || this.implicitHosting))
    }

    get scroll() {
      return this.$accessor.settings.scroll
    }

    get scroll_invert() {
      return this.$accessor.settings.scroll_invert
    }

    get pip_available() {
      //@ts-ignore
      return typeof document.createElement('video').requestPictureInPicture === 'function'
    }

    get clipboard_read_available() {
      return (
        'clipboard' in navigator &&
        typeof navigator.clipboard.readText === 'function' &&
        // Firefox 122+ incorrectly reports that it can read the clipboard but it can't
        // instead it hangs when reading clipboard, until user clicks on the page
        // and the click itself is not handled by the page at all, also the clipboard
        // reads always fail with "Clipboard read operation is not allowed."
        navigator.userAgent.indexOf('Firefox') == -1
      )
    }

    get clipboard_write_available() {
      return 'clipboard' in navigator && typeof navigator.clipboard.writeText === 'function'
    }

    get clipboard() {
      return this.$accessor.remote.clipboard
    }

    get width() {
      return this.$accessor.video.width
    }

    get height() {
      return this.$accessor.video.height
    }

    get rate() {
      return this.$accessor.video.rate
    }

    get vertical() {
      return this.$accessor.video.vertical
    }

    get horizontal() {
      return this.$accessor.video.horizontal
    }

    get is_touch_device() {
      if (typeof window === 'undefined') return false
      return (
        'ontouchstart' in window ||
        navigator.maxTouchPoints > 0 ||
        /Mobi|Android|iPhone|iPad|iPod|Windows Phone/i.test(navigator.userAgent) ||
        window.innerWidth <= 1024
      )
    }

    get trackpad_mode() {
      return this.$accessor.settings.trackpad_mode
    }

    get trackpadActive() {
      return this.trackpad_mode && this.is_touch_device
    }

    @Watch('width')
    onWidthChanged() {
      this.onResize()
    }

    @Watch('height')
    onHeightChanged() {
      this.onResize()
    }

    @Watch('volume')
    onVolumeChanged(volume: number) {
      volume /= 100

      if (this._video && this._video.volume != volume) {
        this._video.volume = volume
      }
    }

    @Watch('muted')
    onMutedChanged(muted: boolean) {
      if (this._video && this._video.muted != muted) {
        this._video.muted = muted

        if (!muted) {
          this.mutedOverlay = false
        }
      }
    }

    @Watch('stream')
    onStreamChanged(stream?: MediaStream) {
      if (!this._video || !stream) {
        return
      }

      if ('srcObject' in this._video) {
        this._video.srcObject = stream
      } else {
        // @ts-ignore
        this._video.src = window.URL.createObjectURL(this.stream) // for older browsers
      }
    }

    @Watch('playing')
    async onPlayingChanged(playing: boolean) {
      if (this._video && this._video.paused && playing) {
        // if autoplay is disabled, play() will throw an error
        // and we need to properly save the state otherwise we
        // would be thinking we're playing when we're not
        try {
          await this._video.play()
        } catch (err: any) {
          if (!this._video.muted) {
            // video.play() can fail if audio is set due restrictive
            // browsers autoplay policy -> retry with muted audio
            try {
              this.$accessor.video.setMuted(true)
              this._video.muted = true
              await this._video.play()
            } catch (err: any) {
              // if it still fails, we're not playing anything
              this.$accessor.video.pause()
            }
          } else {
            this.$accessor.video.pause()
          }
        }
      }

      if (this._video && !this._video.paused && !playing) {
        this.pause()
      }
    }

    @Watch('clipboard')
    async onClipboardChanged(clipboard: string) {
      if (this.clipboard_write_available) {
        try {
          await navigator.clipboard.writeText(clipboard)
          this.$accessor.remote.setClipboard(clipboard)
        } catch (err: any) {
          if (err && (err.name === 'NotAllowedError' || err.name === 'SecurityError' || String(err.message).includes('permissions policy'))) {
            this.$log.debug('Clipboard write blocked by permissions policy:', err.message)
          } else {
            this.$log.error(err)
          }
        }
      }
    }

    mounted() {
      this._container.addEventListener('resize', this.onResize)
      this.onVolumeChanged(this.volume)
      this.onMutedChanged(this.muted)
      this.onStreamChanged(this.stream)
      this.onResize()

      this.observer.observe(this._component)

      onFullscreenChange(this._player, this.onFullscreenChangeHandler)

      this._video.addEventListener('canplaythrough', this.onVideoCanPlayThrough)
      this._video.addEventListener('ended', this.onVideoEnded)
      this._video.addEventListener('error', this.onVideoError)
      this._video.addEventListener('volumechange', this.onVideoVolumeChange)
      this._video.addEventListener('playing', this.onVideoPlaying)
      this._video.addEventListener('pause', this.onVideoPause)

      /* Initialize Guacamole Keyboard */
      this.keyboard.onkeydown = (key: number) => {
        if (!this.hosting || this.locked) {
          return true
        }

        this.$client.sendData('keydown', { key: this.keyMap(key) })
        return false
      }
      this.keyboard.onkeyup = (key: number) => {
        if (!this.hosting || this.locked) {
          return
        }

        this.$client.sendData('keyup', { key: this.keyMap(key) })
      }
      this.keyboard.listenTo(this._overlay)
    }

    beforeDestroy() {
      this.observer.disconnect()
      this.$accessor.video.setPlayable(false)

      if (this._container) {
        this._container.removeEventListener('resize', this.onResize)
      }

      if (this._video) {
        this._video.removeEventListener('canplaythrough', this.onVideoCanPlayThrough)
        this._video.removeEventListener('ended', this.onVideoEnded)
        this._video.removeEventListener('error', this.onVideoError)
        this._video.removeEventListener('volumechange', this.onVideoVolumeChange)
        this._video.removeEventListener('playing', this.onVideoPlaying)
        this._video.removeEventListener('pause', this.onVideoPause)
      }

      if (this._player) {
        this._player.onfullscreenchange = null
        // @ts-ignore
        this._player.onwebkitfullscreenchange = null
        // @ts-ignore
        this._player.onmozfullscreenchange = null
        // @ts-ignore
        this._player.onmsfullscreenchange = null
      }

      if (this.longPressTimer !== null) {
        window.clearTimeout(this.longPressTimer)
        this.longPressTimer = null
      }
      /* Guacamole Keyboard does not provide destroy functions */
    }

    get hasMacOSKbd() {
      return /(Mac|iPhone|iPod|iPad)/i.test(navigator.platform)
    }

    KeyTable = {
      XK_ISO_Level3_Shift: 0xfe03, // AltGr
      XK_Mode_switch: 0xff7e, // Character set switch
      XK_Control_L: 0xffe3, // Left control
      XK_Control_R: 0xffe4, // Right control
      XK_Meta_L: 0xffe7, // Left meta
      XK_Meta_R: 0xffe8, // Right meta
      XK_Alt_L: 0xffe9, // Left alt
      XK_Alt_R: 0xffea, // Right alt
      XK_Super_L: 0xffeb, // Left super
      XK_Super_R: 0xffec, // Right super
    }

    keyMap(key: number): number {
      // Alt behaves more like AltGraph on macOS, so shuffle the
      // keys around a bit to make things more sane for the remote
      // server. This method is used by noVNC, RealVNC and TigerVNC
      // (and possibly others).
      if (this.hasMacOSKbd) {
        switch (key) {
          case this.KeyTable.XK_Meta_L:
            key = this.KeyTable.XK_Control_L
            break
          case this.KeyTable.XK_Super_L:
            key = this.KeyTable.XK_Alt_L
            break
          case this.KeyTable.XK_Super_R:
            key = this.KeyTable.XK_Super_L
            break
          case this.KeyTable.XK_Alt_L:
            key = this.KeyTable.XK_Mode_switch
            break
          case this.KeyTable.XK_Alt_R:
            key = this.KeyTable.XK_ISO_Level3_Shift
            break
        }
      }

      return key
    }

    async play() {
      if (!this._video.paused || !this.playable) {
        return
      }

      try {
        await this._video.play()
        this.onResize()
      } catch (err: any) {
        if (err && (err.name === 'NotAllowedError' || String(err.message).includes('user didn\'t interact'))) {
          this.$log.debug('Autoplay prevented by browser security policy. Waiting for user interaction.', err.message)
        } else {
          this.$log.error(err)
        }
      }
    }

    pause() {
      if (this._video.paused || !this.playable) {
        return
      }

      this._video.pause()
    }

    toggle() {
      if (!this.playable) {
        return
      }

      if (!this.playing) {
        this.$accessor.video.play()
      } else {
        this.$accessor.video.pause()
      }
    }

    playAndUnmute() {
      this.$accessor.video.play()
      this.$accessor.video.setMuted(false)
    }

    unmute() {
      this.$accessor.video.setMuted(false)
    }

    toggleControl() {
      if (!this.playable) {
        return
      }

      if (this.admin && this.hosted && !this.hosting) {
        this.$accessor.remote.adminControl()
      } else {
        this.$accessor.remote.toggle()
      }
    }

    requestControl() {
      this.$accessor.remote.request()
    }

    requestFullscreen() {
      // try to fullscreen player element
      if (elementRequestFullscreen(this._player)) {
        this.onResize()
        return
      }

      // fallback to fullscreen video itself (on mobile devices)
      if (elementRequestFullscreen(this._video)) {
        this.onResize()
        return
      }
    }

    requestPictureInPicture() {
      //@ts-ignore
      this._video.requestPictureInPicture()
      this.onResize()
    }

    openResolution(event: MouseEvent) {
      this._resolution.open(event)
    }

    openClipboard() {
      this._clipboard.open()
    }

    toggleKeyboardHelper() {
      if (this._keyboardHelper) {
        this._keyboardHelper.toggleExpanded()
      }
    }

    async syncClipboard() {
      if (this.clipboard_read_available && window.document.hasFocus()) {
        try {
          const text = await navigator.clipboard.readText()
          if (this.clipboard !== text) {
            this.$accessor.remote.setClipboard(text)
            this.$accessor.remote.sendClipboard(text)
          }
        } catch (err: any) {
          if (err && (err.name === 'NotAllowedError' || err.name === 'SecurityError' || String(err.message).includes('permissions policy'))) {
            this.$log.debug('Clipboard read blocked by permissions policy:', err.message)
          } else {
            this.$log.error(err)
          }
        }
      }
    }

    sendMousePos(e: MouseEvent) {
      const { w, h } = this.$accessor.video.resolution
      const rect = this._overlay.getBoundingClientRect()

      this.$client.sendData('mousemove', {
        x: Math.round((w / rect.width) * (e.clientX - rect.left)),
        y: Math.round((h / rect.height) * (e.clientY - rect.top)),
      })
    }

    wheelThrottle = false
    onWheel(e: WheelEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      let x = e.deltaX
      let y = e.deltaY

      // Pixel units unless it's non-zero.
      // Note that if deltamode is line or page won't matter since we aren't
      // sending the mouse wheel delta to the server anyway.
      // The difference between pixel and line can be important however since
      // we have a threshold that can be smaller than the line height.
      if (e.deltaMode !== 0) {
        x *= WHEEL_LINE_HEIGHT
        y *= WHEEL_LINE_HEIGHT
      }

      if (this.scroll_invert) {
        x = x * -1
        y = y * -1
      }

      x = Math.min(Math.max(x, -this.scroll), this.scroll)
      y = Math.min(Math.max(y, -this.scroll), this.scroll)

      this.sendMousePos(e)

      if (!this.wheelThrottle) {
        this.wheelThrottle = true
        this.$client.sendData('wheel', { x, y })

        window.setTimeout(() => {
          this.wheelThrottle = false
        }, 100)
      }
    }

    onTouchHandler(e: TouchEvent) {
      if (this.trackpadActive) {
        this.onTrackpadTouch(e)
        return
      }

      let first = e.changedTouches[0]
      let type = ''
      switch (e.type) {
        case 'touchstart':
          type = 'mousedown'
          break
        case 'touchmove':
          type = 'mousemove'
          break
        case 'touchend':
          type = 'mouseup'
          break
        default:
          return
      }

      const simulatedEvent = new MouseEvent(type, {
        bubbles: true,
        cancelable: true,
        view: window,
        screenX: first.screenX,
        screenY: first.screenY,
        clientX: first.clientX,
        clientY: first.clientY,
      })
      first.target.dispatchEvent(simulatedEvent)
    }

    onTrackpadTouch(e: TouchEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      const rect = this._overlay.getBoundingClientRect()
      const touch = e.changedTouches[0]

      if (e.type === 'touchstart') {
        this.touchLastX = touch.clientX
        this.touchLastY = touch.clientY
        this.hasMoved = false
        this.didLongPress = false

        if (this.longPressTimer !== null) {
          window.clearTimeout(this.longPressTimer)
        }
        this.longPressTimer = window.setTimeout(() => {
          if (!this.hasMoved && !this.didLongPress) {
            this.didLongPress = true
            this.triggerTrackpadClick(3) // Right-click (key 3)
            if (typeof navigator !== 'undefined' && 'vibrate' in navigator) {
              navigator.vibrate(40)
            }
          }
        }, 600)

      } else if (e.type === 'touchmove') {
        const dx = touch.clientX - this.touchLastX
        const dy = touch.clientY - this.touchLastY

        this.touchLastX = touch.clientX
        this.touchLastY = touch.clientY

        this.cursorX = Math.max(0, Math.min(rect.width, this.cursorX + dx))
        this.cursorY = Math.max(0, Math.min(rect.height, this.cursorY + dy))

        if (Math.abs(dx) > 2 || Math.abs(dy) > 2) {
          this.hasMoved = true
          if (this.longPressTimer !== null) {
            window.clearTimeout(this.longPressTimer)
            this.longPressTimer = null
          }
        }

        this.sendTrackpadMousePos()

      } else if (e.type === 'touchend') {
        if (this.longPressTimer !== null) {
          window.clearTimeout(this.longPressTimer)
          this.longPressTimer = null
        }

        if (!this.hasMoved && !this.didLongPress) {
          this.triggerTrackpadClick(1) // Left-click (key 1)
        }
      }
    }

    sendTrackpadMousePos() {
      const { w, h } = this.$accessor.video.resolution
      if (!this._overlay) return
      const rect = this._overlay.getBoundingClientRect()

      this.$client.sendData('mousemove', {
        x: Math.round((w / rect.width) * this.cursorX),
        y: Math.round((h / rect.height) * this.cursorY),
      })
    }

    triggerTrackpadClick(button: number) {
      if (!this.controlling) {
        if (this.implicitHosting) {
          this.$accessor.remote.request()
        }
        return
      }

      this.sendTrackpadMousePos()
      this.$client.sendData('mousedown', { key: button })
      window.setTimeout(() => {
        this.$client.sendData('mouseup', { key: button })
      }, 50)
    }

    onCompositionStartHandler() {
      this.lastTextAreaValue = this._overlay.value
    }

    onCompositionEndHandler() {
      this._overlay.value = this.lastTextAreaValue
    }

    isMouseDown = false

    onMouseDown(e: MouseEvent) {
      this.isMouseDown = true

      if (this.locked) {
        return
      }

      if (!this.controlling) {
        this.implicitHostingRequest(e)
        return
      }

      this.sendMousePos(e)
      this.$client.sendData('mousedown', { key: e.button + 1 })
    }

    onMouseUp(e: MouseEvent) {
      // only if we are the one who started the mouse down
      if (!this.isMouseDown) return
      this.isMouseDown = false

      if (this.locked) {
        return
      }

      if (!this.controlling) {
        this.implicitHostingRequest(e)
        return
      }

      this.sendMousePos(e)
      this.$client.sendData('mouseup', { key: e.button + 1 })
    }

    private reqMouseDown: MouseEvent | null = null
    private reqMouseUp: MouseEvent | null = null

    @Watch('controlling')
    onControlChange(controlling: boolean) {
      if (controlling && this.reqMouseDown) {
        this.onMouseDown(this.reqMouseDown)
      }

      if (controlling && this.reqMouseUp) {
        this.onMouseUp(this.reqMouseUp)
      }

      this.reqMouseDown = null
      this.reqMouseUp = null
    }

    implicitHostingRequest(e: MouseEvent) {
      if (this.implicitHosting) {
        if (e.type === 'mousedown') {
          this.reqMouseDown = e
          this.reqMouseUp = null
          this.$accessor.remote.request()
        } else if (e.type === 'mouseup') {
          this.reqMouseUp = e
        }
        return
      }

      if (e.type === 'mousedown') {
        this.$emit('control-attempt', e)
      }
    }

    onMouseMove(e: MouseEvent) {
      if (!this.hosting || this.locked) {
        return
      }

      this.sendMousePos(e)
    }

    onMouseEnter(e: MouseEvent) {
      if (this.hosting) {
        this.$accessor.remote.syncKeyboardModifierState({
          capsLock: e.getModifierState('CapsLock'),
          numLock: e.getModifierState('NumLock'),
          scrollLock: e.getModifierState('ScrollLock'),
        })

        this.syncClipboard()
      }

      this.focused = true
    }

    onMouseLeave(e: MouseEvent) {
      if (this.hosting) {
        this.$accessor.remote.setKeyboardModifierState({
          capsLock: e.getModifierState('CapsLock'),
          numLock: e.getModifierState('NumLock'),
          scrollLock: e.getModifierState('ScrollLock'),
        })
      }

      this.keyboard.reset()
      this.focused = false
    }

    onResize() {
      const { offsetWidth, offsetHeight } = !this.fullscreen ? this._component : document.body
      this._player.style.width = `${offsetWidth}px`
      this._player.style.height = `${offsetHeight}px`
      this._container.style.maxWidth = `${(this.horizontal / this.vertical) * offsetHeight}px`
      this._aspect.style.paddingBottom = `${(this.vertical / this.horizontal) * 100}%`

      this.$nextTick(() => {
        if (this._overlay) {
          const rect = this._overlay.getBoundingClientRect()
          if (this.cursorX === 0 && this.cursorY === 0) {
            this.cursorX = rect.width / 2
            this.cursorY = rect.height / 2
          } else {
            this.cursorX = Math.max(0, Math.min(rect.width, this.cursorX))
            this.cursorY = Math.max(0, Math.min(rect.height, this.cursorY))
          }
        }
      })
    }

    @Watch('focused')
    @Watch('hosting')
    @Watch('locked')
    onFocus() {
      // focus opens the keyboard on mobile
      if (this.is_touch_device) {
        return
      }

      // in order to capture key events, overlay must be focused
      if (this.focused && this.hosting && !this.locked) {
        this._overlay.focus()
      }
    }

    openMobileKeyboard() {
      // focus opens the keyboard on mobile
      this._overlay.focus()
    }
  }
</script>
