<template>
  <ul>
    <li v-if="!implicitHosting && (!controlLocked || hosting)">
      <i
        :class="[
          !disabeld && shakeKbd ? 'shake' : '',
          disabeld && !hosting ? 'disabled' : '',
          !disabeld && !hosting ? 'faded' : '',
          'fas',
          'fa-keyboard',
          'request',
        ]"
        v-tooltip="{
          content: !disabeld || hosting ? (hosting ? $t('controls.release') : $t('controls.request')) : '',
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
        @click.stop.prevent="toggleControl"
      />
    </li>
    <li class="no-pointer" v-if="implicitHosting">
      <i
        :class="[controlLocked ? 'disabled' : '', 'fas', 'fa-mouse-pointer']"
        v-tooltip="{
          content: controlLocked ? $t('controls.hasnot') : $t('controls.has'),
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
      />
    </li>
    <li v-if="implicitHosting || (!implicitHosting && (!controlLocked || hosting))">
      <label
        class="switch"
        v-tooltip="{
          content: hosting ? (locked ? $t('controls.unlock') : $t('controls.lock')) : '',
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
      >
        <input type="checkbox" v-model="locked" :disabled="!hosting || (implicitHosting && controlLocked)" />
        <span />
      </label>
    </li>
    <li>
      <i
        :class="[{ disabled: !playable }, playing ? 'fa-pause-circle' : 'fa-play-circle', 'fas', 'play']"
        @click.stop.prevent="toggleMedia"
      />
    </li>
    <li v-if="micAllowed">
      <i
        :class="[
          { disabled: !playable },
          microphoneActive ? 'fa-microphone glowing-mic' : 'fa-microphone-slash',
          microphoneActive ? '' : 'faded',
          'fas',
        ]"
        v-tooltip="{
          content: microphoneActive ? $t('controls.mic_off') : $t('controls.mic_on'),
          placement: 'top',
          offset: 5,
          boundariesElement: 'body',
          delay: { show: 300, hide: 100 },
        }"
        @click.stop.prevent="toggleMicrophone"
      />
    </li>
    <li>
      <div class="volume">
        <i
          :class="[volume === 0 || muted ? 'fa-volume-mute' : 'fa-volume-up', 'fas']"
          @click.stop.prevent="toggleMute"
        />
        <input type="range" min="0" max="100" v-model="volume" />
      </div>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  .shake {
    animation: shake 1.25s cubic-bezier(0, 0, 0, 1);
  }

  @keyframes shake {
    0% {
      transform: scale(1) translate(0px, 0) rotate(0);
    }
    10% {
      transform: scale(1.25) translate(-2px, -2px) rotate(-20deg);
    }
    20% {
      transform: scale(1.5) translate(4px, -4px) rotate(20deg);
    }
    30% {
      transform: scale(1.75) translate(-4px, -6px) rotate(-20deg);
    }
    40% {
      transform: scale(2) translate(6px, -8px) rotate(20deg);
    }
    50% {
      transform: scale(2.25) translate(-6px, -10px) rotate(-20deg);
    }
    60% {
      transform: scale(2) translate(6px, -8px) rotate(20deg);
    }
    70% {
      transform: scale(1.75) translate(-4px, -6px) rotate(-20deg);
    }
    80% {
      transform: scale(1.5) translate(4px, -4px) rotate(20deg);
    }
    90% {
      transform: scale(1.25) translate(-2px, -2px) rotate(-20deg);
    }
    100% {
      transform: scale(1) translate(0px, 0) rotate(0);
    }
  }

  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    gap: 12px;
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      display: inline-flex;
      align-items: center;
      cursor: pointer;

      &.no-pointer {
        cursor: default;
      }

      & > i {
        width: 38px;
        height: 38px;
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-subtle);
        font-size: 15px;
        transition: var(--transition-fluid);

        &:hover:not(.disabled) {
          color: var(--color-cyber-mint);
          background: rgba(255, 255, 255, 0.08);
          border-color: rgba(38, 230, 180, 0.25);
          filter: drop-shadow(0 0 6px var(--color-cyber-mint-glow));
          transform: scale(1.08);
        }

        &:active:not(.disabled) {
          transform: scale(0.95);
        }

        &.faded {
          color: rgba(255, 255, 255, 0.25);
          border-color: rgba(255, 255, 255, 0.04);
        }

        &.glowing-mic {
          color: var(--color-cyber-mint) !important;
          border-color: var(--color-cyber-mint) !important;
          background: rgba(38, 230, 180, 0.12) !important;
          animation: mic-pulse 1.8s infinite ease-in-out;
        }

        &.disabled {
          color: var(--style-error);
          opacity: 0.6;
          border-color: rgba(255, 74, 90, 0.15);
          cursor: not-allowed;
        }
      }

      @keyframes mic-pulse {
        0% {
          box-shadow: 0 0 0 0 rgba(38, 230, 180, 0.4);
        }
        70% {
          box-shadow: 0 0 0 8px rgba(38, 230, 180, 0);
        }
        100% {
          box-shadow: 0 0 0 0 rgba(38, 230, 180, 0);
        }
      }

      .volume {
        white-space: nowrap;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
        gap: 10px;
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        padding: 0 12px;
        height: 38px;
        list-style: none;
        transition: var(--transition-fluid);

        &:hover {
          border-color: rgba(38, 230, 180, 0.15);
          background: rgba(255, 255, 255, 0.05);
        }

        i {
          width: 18px;
          height: auto;
          background: none;
          border: none;
          padding: 0;
          color: var(--text-subtle);
          font-size: 14px;

          &:hover {
            color: var(--color-cyber-mint);
            background: none;
            border: none;
            transform: none;
            filter: none;
          }
        }

        input[type='range'] {
          width: 90px;
          height: 4px;
          background: rgba(255, 255, 255, 0.12);
          border-radius: 4px;
          outline: none;
          transition: var(--transition-fluid);
          -webkit-appearance: none;

          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 12px;
            width: 12px;
            border-radius: 50%;
            background: var(--color-cyber-mint);
            box-shadow: 0 0 8px var(--color-cyber-mint-glow);
            cursor: pointer;
            transition: var(--transition-fluid);
            margin-top: -4px;

            &:hover {
              transform: scale(1.35);
              background: #fff;
              box-shadow: 0 0 12px var(--color-cyber-mint), 0 0 0 4px var(--color-cyber-mint-glow);
            }
          }

          &::-webkit-slider-runnable-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: transparent;
            border-radius: 2px;
          }

          &::-moz-range-thumb {
            height: 12px;
            width: 12px;
            border: none;
            border-radius: 50%;
            background: var(--color-cyber-mint);
            box-shadow: 0 0 8px var(--color-cyber-mint-glow);
            cursor: pointer;
            transition: var(--transition-fluid);

            &:hover {
              transform: scale(1.35);
              background: #fff;
              box-shadow: 0 0 12px var(--color-cyber-mint), 0 0 0 4px var(--color-cyber-mint-glow);
            }
          }

          &::-moz-range-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: transparent;
            border-radius: 2px;
          }

          &:hover {
            background: rgba(255, 255, 255, 0.22);
            height: 5px;
          }
        }
      }

      .switch {
        margin: 0;
        display: block;
        position: relative;
        width: 48px;
        height: 26px;

        input[type='checkbox'] {
          opacity: 0;
          width: 0;
          height: 0;
        }

        span {
          position: absolute;
          cursor: pointer;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background-color: rgba(255, 255, 255, 0.05);
          border: 1px solid var(--glass-border);
          transition: var(--transition-fluid);
          border-radius: 34px;

          &:before {
            color: #050508;
            font-weight: 900;
            font-family: 'Font Awesome 6 Free';
            content: '\f3c1';
            font-size: 8px;
            line-height: 20px;
            text-align: center;
            position: absolute;
            height: 20px;
            width: 20px;
            left: 2px;
            bottom: 2px;
            background-color: var(--text-pure);
            transition: var(--transition-fluid);
            border-radius: 50%;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
          }
        }
      }

      input[type='checkbox'] {
        &:checked + span {
          background-color: var(--color-cyber-mint);
          border-color: transparent;
          box-shadow: 0 0 10px var(--color-cyber-mint-glow);

          &:before {
            content: '\f023';
            transform: translateX(22px) rotate(360deg);
            background-color: #050508;
            color: var(--color-cyber-mint);
          }
        }

        &:disabled + span {
          opacity: 0.5;
          &:before {
            content: '';
            background-color: rgba(255, 255, 255, 0.2);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class NekoControls extends Vue {
    @Prop(Boolean) readonly shakeKbd!: boolean

    get controlLocked() {
      return 'control' in this.$accessor.locked && this.$accessor.locked['control'] && !this.$accessor.user.admin
    }

    get disabeld() {
      return this.$accessor.remote.hosted
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get controlling() {
      return this.$accessor.remote.controlling
    }

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
    }

    // Microphone is allowed when the user is actively controlling (has host).
    // With implicit hosting, the controlling getter is true only when the user
    // has actually been assigned as host (clicked inside the video), not for
    // everyone by default. This prevents multiple users from sharing their
    // microphone simultaneously — only the person in control can.
    get micAllowed() {
      return this.controlling
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

    get locked() {
      return this.$accessor.remote.locked && this.$accessor.remote.hosting
    }

    set locked(locked: boolean) {
      this.$accessor.remote.setLocked(locked)
    }

    toggleControl() {
      if (!this.playable) {
        return
      }
      this.$accessor.remote.toggle()
    }

    toggleMedia() {
      if (!this.playable) {
        return
      }
      this.$accessor.video.togglePlay()
    }

    toggleMute() {
      this.$accessor.video.toggleMute()
    }

    microphoneActive = false

    // Auto-disable microphone when the user loses control (e.g. another user
    // takes host, or admin releases control). This ensures the mic track is
    // cleaned up and the server-side audio input is freed for the new host.
    @Watch('controlling')
    onControllingChanged(isControlling: boolean) {
      if (!isControlling && this.microphoneActive) {
        this.$client.disableMicrophone()
        this.microphoneActive = false
      }
    }

    async toggleMicrophone() {
      if (!this.playable || !this.micAllowed) {
        return
      }

      if (this.microphoneActive) {
        this.$client.disableMicrophone()
        this.microphoneActive = false
      } else {
        try {
          await this.$client.enableMicrophone()
          this.microphoneActive = true
        } catch (err: any) {
          let extraText = ''
          if (window.self !== window.top) {
            extraText = '\n\nHinweis: Da der Client in einem iframe läuft, öffnen Sie die App am besten in einem neuen Tab (über den Button oben rechts), damit Ihr Browser die Erlaubnis abfragen und erteilen kann.'
          }
          this.$swal({
            title: this.$t('controls.mic_error') as string,
            text: (err.message || err.name || String(err)) + extraText,
            icon: 'error',
          })
        }
      }
    }
  }
</script>
