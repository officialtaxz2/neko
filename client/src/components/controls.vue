<template>
  <ul>
    <li v-if="showPill" class="pill-item">
      <neko-control-pill :shakeKbd="shakeKbd" />
    </li>
    <li v-if="showPill" class="sep" aria-hidden="true" />
    <li>
      <button
        type="button"
        :class="['icon-btn', 'ctrl', 'play', { disabled: !playable }]"
        :disabled="!playable"
        :aria-label="playing ? $t('controls.pause') : $t('controls.play')"
        v-tooltip="tooltip(playing ? $t('controls.pause') : $t('controls.play'))"
        @click.stop.prevent="toggleMedia"
      >
        <i :class="['fas', playing ? 'fa-pause' : 'fa-play']" aria-hidden="true" />
      </button>
    </li>
    <li>
      <div class="volume">
        <button
          type="button"
          class="icon-btn"
          :aria-label="muted ? $t('controls.unmute') : $t('controls.mute')"
          :aria-pressed="muted ? 'true' : 'false'"
          v-tooltip="tooltip(muted ? $t('controls.unmute') : $t('controls.mute'))"
          @click.stop.prevent="toggleMute"
        >
          <i :class="[muted ? 'fa-volume-mute' : 'fa-volume-up', 'fas']" aria-hidden="true" />
        </button>
        <input type="range" min="0" max="100" v-model="volume" :aria-label="$t('controls.volume')" />
      </div>
    </li>
    <li v-if="micAllowed">
      <button
        type="button"
        :class="['icon-btn', 'ctrl', { 'glowing-mic': microphoneActive, faded: !microphoneActive }]"
        :disabled="!playable"
        :aria-pressed="microphoneActive ? 'true' : 'false'"
        :aria-label="microphoneActive ? $t('controls.mic_off') : $t('controls.mic_on')"
        v-tooltip="tooltip(microphoneActive ? $t('controls.mic_off') : $t('controls.mic_on'))"
        @click.stop.prevent="toggleMicrophone"
      >
        <i :class="['fas', microphoneActive ? 'fa-microphone' : 'fa-microphone-slash']" aria-hidden="true" />
      </button>
    </li>
    <li v-if="viewOnly"><neko-media-selector /></li>
  </ul>
</template>

<style lang="scss" scoped>
  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    gap: 8px;
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      display: inline-flex;
      align-items: center;

      &.sep {
        width: 1px;
        height: 28px;
        margin: 0 4px;
        background: rgba(255, 255, 255, 0.1);
      }

      .ctrl {
        width: 40px;
        height: 40px;
        background: rgba(255, 255, 255, 0.04);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        color: var(--text-subtle);
        font-size: 15px;
        transition: var(--transition-fluid);

        &:hover:not(:disabled) {
          color: var(--color-cyber-mint);
          background: rgba(255, 255, 255, 0.08);
          border-color: rgba(38, 230, 180, 0.25);
        }

        &:active:not(:disabled) {
          transform: scale(0.95);
        }

        &:disabled {
          opacity: 0.45;
        }

        &.faded {
          color: rgba(255, 255, 255, 0.45);
        }

        &.glowing-mic {
          color: var(--color-cyber-mint);
          border-color: var(--color-cyber-mint);
          background: rgba(38, 230, 180, 0.12);
          animation: mic-pulse 1.8s infinite ease-in-out;
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
        align-items: center;
        gap: 6px;
        background: rgba(255, 255, 255, 0.04);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        padding: 0 12px 0 2px;
        height: 40px;
        transition: var(--transition-fluid);

        &:hover {
          border-color: rgba(38, 230, 180, 0.15);
        }

        .icon-btn {
          width: 34px;
          height: 34px;
          border-radius: 9px;
          color: var(--text-subtle);
          font-size: 15px;

          &:hover {
            color: var(--color-cyber-mint);
          }
        }

        input[type='range'] {
          width: 90px;
          height: 4px;
          background: rgba(255, 255, 255, 0.14);
          border-radius: 4px;
          transition: var(--transition-fluid);
          -webkit-appearance: none;
          appearance: none;

          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 14px;
            width: 14px;
            border-radius: 50%;
            background: var(--color-cyber-mint);
            box-shadow: 0 0 8px var(--color-cyber-mint-glow);
            cursor: pointer;
            margin-top: -5px;
          }

          &::-webkit-slider-runnable-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: transparent;
            border-radius: 2px;
          }

          &::-moz-range-thumb {
            height: 14px;
            width: 14px;
            border: none;
            border-radius: 50%;
            background: var(--color-cyber-mint);
            box-shadow: 0 0 8px var(--color-cyber-mint-glow);
            cursor: pointer;
          }

          &::-moz-range-track {
            width: 100%;
            height: 4px;
            cursor: pointer;
            background: transparent;
            border-radius: 2px;
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Mixins, Prop, Watch } from 'vue-property-decorator'
  import MediaControls from './mixins/media-controls'
  import NekoMediaSelector from './media-selector.vue'
  import NekoControlPill from './control-pill.vue'

  @Component({ name: 'neko-controls', components: { NekoMediaSelector, NekoControlPill } })
  export default class NekoControls extends Mixins(MediaControls) {
    @Prop(Boolean) readonly shakeKbd!: boolean

    get showPill() {
      return !this.viewOnly && !this.receiveOnly
    }

    tooltip(content: string) {
      return { content, placement: 'top', offset: 5, boundariesElement: 'body', delay: { show: 300, hide: 100 } }
    }

    // Auto-disable microphone when the user loses control (e.g. another user
    // takes host, or admin releases control). This ensures the mic track is
    // cleaned up and the server-side audio input is freed for the new host.
    @Watch('controlling')
    onControllingChanged(isControlling: boolean) {
      if (!isControlling && this.microphoneActive) {
        this.$client.disableMicrophone()
        this.$accessor.video.setMicrophone(false)
      }
    }
  }
</script>
