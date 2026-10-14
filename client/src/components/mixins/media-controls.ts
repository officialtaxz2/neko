import { Vue, Component } from 'vue-property-decorator'

// Shared state and actions for the room's media and control buttons.
// Used by the desktop control bar, the control pill and the mobile dock so
// that every surface reads the same store state and triggers the same actions.
@Component
export default class MediaControls extends Vue {
  get viewOnly() {
    return this.$accessor.user.viewOnly
  }

  get receiveOnly() {
    return this.$client.receiveOnlySelected
  }

  get admin() {
    return this.$accessor.user.admin
  }

  get controlLocked() {
    return 'control' in this.$accessor.locked && this.$accessor.locked['control'] && !this.$accessor.user.admin
  }

  get hosted() {
    return this.$accessor.remote.hosted
  }

  get hosting() {
    return !this.receiveOnly && this.$accessor.remote.hosting
  }

  get controlling() {
    return !this.receiveOnly && this.$accessor.remote.controlling
  }

  get implicitHosting() {
    return !this.receiveOnly && this.$accessor.remote.implicitHosting
  }

  // Microphone is allowed when the user is actively controlling (has host).
  // With implicit hosting, the controlling getter is true only when the user
  // has actually been assigned as host (clicked inside the video), not for
  // everyone by default. This prevents multiple users from sharing their
  // microphone simultaneously — only the person in control can.
  get micAllowed() {
    return !this.viewOnly && !this.receiveOnly && this.controlling
  }

  get microphoneActive() {
    return this.$accessor.video.microphone
  }

  get volume() {
    return this.$accessor.video.volume
  }

  set volume(volume: number) {
    this.$accessor.video.setVolume(volume)
  }

  get muted() {
    return this.$accessor.video.muted || this.volume === 0
  }

  get playing() {
    return this.$accessor.video.playing
  }

  get playable() {
    return this.$accessor.video.playable
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

  toggleMedia() {
    if (!this.playable) {
      return
    }
    if (this.$client.hlsSelected) {
      if (this.playing) this.$client.pauseHLS()
      else void this.$client.playHLS()
      return
    }
    this.$accessor.video.togglePlay()
  }

  toggleMute() {
    this.$accessor.video.toggleMute()
  }

  async toggleMicrophone() {
    if (!this.playable || !this.micAllowed) {
      return
    }

    if (this.microphoneActive) {
      this.$client.disableMicrophone()
      this.$accessor.video.setMicrophone(false)
      return
    }

    try {
      await this.$client.enableMicrophone()
      this.$accessor.video.setMicrophone(true)
    } catch (err: any) {
      let extraText = ''
      if (window.self !== window.top) {
        extraText =
          '\n\nHinweis: Da der Client in einem iframe läuft, öffnen Sie die App am besten in einem neuen Tab (über den Button oben rechts), damit Ihr Browser die Erlaubnis abfragen und erteilen kann.'
      }
      this.$swal({
        title: this.$t('controls.mic_error') as string,
        text: (err.message || err.name || String(err)) + extraText,
        icon: 'error',
      })
    }
  }
}
