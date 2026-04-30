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
          microphoneActive ? 'fa-microphone' : 'fa-microphone-slash',
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
  // ------------------------------------------------------------------
  // Shake animation — keyboard-request attention feedback
  // ------------------------------------------------------------------
  .shake {
    animation: shake 1.25s cubic-bezier(0, 0, 0, 1);
  }

  @keyframes shake {
    0%   { transform: scale(1)    translate(0px,   0px)   rotate(0deg);   }
    10%  { transform: scale(1.25) translate(-2px,  -2px)  rotate(-20deg); }
    20%  { transform: scale(1.5)  translate(4px,   -4px)  rotate(20deg);  }
    30%  { transform: scale(1.75) translate(-4px,  -6px)  rotate(-20deg); }
    40%  { transform: scale(2)    translate(6px,   -8px)  rotate(20deg);  }
    50%  { transform: scale(2.25) translate(-6px,  -10px) rotate(-20deg); }
    60%  { transform: scale(2)    translate(6px,   -8px)  rotate(20deg);  }
    70%  { transform: scale(1.75) translate(-4px,  -6px)  rotate(-20deg); }
    80%  { transform: scale(1.5)  translate(4px,   -4px)  rotate(20deg);  }
    90%  { transform: scale(1.25) translate(-2px,  -2px)  rotate(-20deg); }
    100% { transform: scale(1)    translate(0px,   0px)   rotate(0deg);   }
  }

  // ------------------------------------------------------------------
  // Controls list
  // ------------------------------------------------------------------
  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    list-style: none;

    li {
      // Touch target ≥ 44×44 px (WCAG 2.5.5 / mobile UX)
      min-width: 44px;
      min-height: 44px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;

      &.no-pointer {
        cursor: default;
      }

      // ------------------------------------------------------------------
      // Icon micro-animations
      // ------------------------------------------------------------------
      i {
        font-size: 24px;
        color: var(--color-text);
        will-change: transform;
        transition:
          transform var(--transition-fast),
          color    var(--transition-interactive),
          opacity  var(--transition-interactive);

        // Hover: lift + teal accent (skip for non-interactive / disabled)
        &:hover:not(.disabled):not(.shake) {
          transform: scale(1.18);
          color: var(--color-primary);
        }

        // Active: press-down feedback
        &:active:not(.disabled) {
          transform: scale(0.88);
          transition-duration: 80ms;
        }

        &.faded {
          color: var(--color-text);
          opacity: 0.4;
        }

        &.disabled {
          color: var(--color-error);
          opacity: 0.4;
          pointer-events: none;
        }
      }

      // ------------------------------------------------------------------
      // Volume control
      // ------------------------------------------------------------------
      .volume {
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: var(--space-2);
        white-space: nowrap;

        input[type='range'] {
          width: 150px;
          height: 20px;
          background: transparent;
          -webkit-appearance: none;
          appearance: none;
          cursor: pointer;

          // Firefox — track
          &::-moz-range-track {
            width: 100%;
            height: 4px;
            background: var(--color-primary);
            border-radius: var(--radius-full);
            cursor: pointer;
          }

          // Firefox — thumb: box-shadow ensures visibility on light surfaces
          &::-moz-range-thumb {
            height: 12px;
            width: 12px;
            border-radius: var(--radius-full);
            background: #fff;
            cursor: pointer;
            border: none;
            box-shadow: var(--shadow-sm);
            transition: transform var(--transition-fast);
          }

          &:hover::-moz-range-thumb {
            transform: scale(1.35);
          }

          // WebKit — track
          &::-webkit-slider-runnable-track {
            width: 100%;
            height: 4px;
            background: var(--color-primary);
            border-radius: var(--radius-full);
            cursor: pointer;
          }

          // WebKit — thumb: box-shadow ensures visibility on light surfaces
          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 12px;
            width: 12px;
            border-radius: var(--radius-full);
            background: #fff;
            cursor: pointer;
            margin-top: -4px;
            box-shadow: var(--shadow-sm);
            transition: transform var(--transition-fast);
          }

          &:hover::-webkit-slider-thumb {
            transform: scale(1.35);
          }
        }
      }

      // ------------------------------------------------------------------
      // Custom lock toggle switch
      // Track uses var(--color-surface-offset) instead of var(--color-surface)
      // so it's visible in light mode (surface = near-white = invisible track).
      // ------------------------------------------------------------------
      .switch {
        margin: 0 var(--space-1);
        display: block;
        position: relative;
        width: 42px;
        height: 24px;

        input[type='checkbox'] {
          opacity: 0;
          width: 0;
          height: 0;
        }

        span {
          position: absolute;
          cursor: pointer;
          inset: 0;
          background-color: var(--color-surface-offset);
          border-radius: var(--radius-full);
          transition: background-color var(--transition-interactive);

          &::before {
            font-family: 'Font Awesome 6 Free';
            font-weight: 900;
            content: '\f3c1'; // fa-circle-dot (unlocked indicator)
            font-size: 8px;
            line-height: 18px;
            text-align: center;
            color: var(--color-bg);
            position: absolute;
            height: 18px;
            width: 18px;
            left: 3px;
            bottom: 3px;
            background-color: #fff;
            border-radius: 50%;
            box-shadow: var(--shadow-sm);
            transition: transform var(--transition-interactive);
            will-change: transform;
          }
        }
      }

      input[type='checkbox'] {
        &:checked + span {
          background-color: var(--color-primary);

          &::before {
            content: '\f023'; // fa-lock
            transform: translateX(18px);
          }
        }

        &:disabled + span {
          opacity: 0.45;
          cursor: not-allowed;

          &::before {
            content: '';
            background-color: var(--color-text-muted);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class extends Vue {
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
          this.$swal({
            title: this.$t('controls.mic_error') as string,
            text: err.message,
            icon: 'error',
          })
        }
      }
    }
  }
</script>
