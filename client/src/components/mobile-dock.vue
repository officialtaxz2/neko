<template>
  <div class="mobile-dock">
    <div class="dock">
      <neko-control-pill v-if="showPill" class="dock-pill" compact :shakeKbd="shakeKbd" />
      <div class="dock-buttons">
        <button
          type="button"
          class="icon-btn dock-btn"
          :disabled="!playable"
          :aria-label="playing ? $t('controls.pause') : $t('controls.play')"
          @click.stop.prevent="toggleMedia"
        >
          <i :class="['fas', playing ? 'fa-pause' : 'fa-play']" aria-hidden="true" />
        </button>
        <neko-emotes v-if="!viewOnly" compact />
        <button
          ref="moreButton"
          type="button"
          :class="['icon-btn', 'dock-btn', { active: open }]"
          :aria-label="$t('controls.more')"
          aria-haspopup="dialog"
          :aria-expanded="open ? 'true' : 'false'"
          @click.stop.prevent="open = !open"
        >
          <i class="fas fa-ellipsis-h" aria-hidden="true" />
          <span v-if="muted" class="dot" aria-hidden="true" />
        </button>
      </div>
    </div>

    <transition name="sheet">
      <div v-if="open" class="sheet-backdrop" @click.self="close">
        <div
          ref="sheet"
          class="sheet"
          role="dialog"
          aria-modal="true"
          :aria-label="$t('controls.more')"
          @keydown.esc.stop.prevent="close"
        >
          <div class="handle" aria-hidden="true" />

          <div class="sheet-row sound">
            <button
              ref="first"
              type="button"
              :class="['icon-btn', 'round', { pressed: muted }]"
              :aria-label="muted ? $t('controls.unmute') : $t('controls.mute')"
              :aria-pressed="muted ? 'true' : 'false'"
              @click.stop.prevent="toggleMute"
            >
              <i :class="['fas', muted ? 'fa-volume-mute' : 'fa-volume-up']" aria-hidden="true" />
            </button>
            <input
              v-if="volumeAdjustable"
              v-model="volume"
              type="range"
              min="0"
              max="100"
              :aria-label="$t('controls.volume')"
            />
            <span v-else class="sound-label">{{ muted ? $t('controls.unmute') : $t('controls.mute') }}</span>
          </div>

          <div class="grid">
            <button
              v-if="showPill && hosting"
              type="button"
              :class="['icon-btn', 'tile', { on: inputLocked }]"
              :aria-pressed="inputLocked ? 'true' : 'false'"
              @click.stop.prevent="inputLocked = !inputLocked"
            >
              <i :class="['fas', inputLocked ? 'fa-lock' : 'fa-lock-open']" aria-hidden="true" />
              <span>{{ inputLocked ? $t('controls.resume_input') : $t('controls.pause_input') }}</span>
            </button>
            <button v-if="micAllowed" type="button" :class="['icon-btn', 'tile', { on: microphoneActive }]" :aria-pressed="microphoneActive ? 'true' : 'false'" @click.stop.prevent="toggleMicrophone">
              <i :class="['fas', microphoneActive ? 'fa-microphone' : 'fa-microphone-slash']" aria-hidden="true" />
              <span>{{ $t('controls.microphone') }}</span>
            </button>
            <button type="button" class="icon-btn tile" @click.stop.prevent="run('fullscreen')">
              <i class="fas fa-expand" aria-hidden="true" />
              <span>{{ $t('controls.fullscreen') }}</span>
            </button>
            <template v-if="hosting && touchDevice">
              <button type="button" class="icon-btn tile" @click.stop.prevent="run('keyboard')">
                <i class="fas fa-keyboard" aria-hidden="true" />
                <span>{{ $t('controls.keyboard') }}</span>
              </button>
              <button type="button" class="icon-btn tile" @click.stop.prevent="run('keyboard-helper')">
                <i class="fas fa-sliders-h" aria-hidden="true" />
                <span>{{ $t('controls.keyboard_helper') }}</span>
              </button>
            </template>
            <button v-if="hosting" type="button" class="icon-btn tile" @click.stop.prevent="run('clipboard')">
              <i class="fas fa-clipboard" aria-hidden="true" />
              <span>{{ $t('controls.clipboard') }}</span>
            </button>
            <button v-if="pipAvailable" type="button" class="icon-btn tile" @click.stop.prevent="run('pip')">
              <i class="fas fa-external-link-alt" aria-hidden="true" />
              <span>{{ $t('controls.pip') }}</span>
            </button>
            <button v-if="admin && !receiveOnly" type="button" class="icon-btn tile" @click.stop.prevent="run('resolution', $event)">
              <i class="fas fa-desktop" aria-hidden="true" />
              <span>{{ $t('controls.resolution') }}</span>
            </button>
            <button type="button" class="icon-btn tile" @click.stop.prevent="openStreamSettings">
              <i class="fas fa-broadcast-tower" aria-hidden="true" />
              <span>{{ $t('controls.stream', { name: streamName }) }}</span>
            </button>
          </div>

          <button type="button" class="icon-btn close" @click.stop.prevent="close">{{ $t('controls.close') }}</button>
        </div>
      </div>
    </transition>
  </div>
</template>

<style lang="scss" scoped>
  .mobile-dock {
    display: none;
  }

  // The dock replaces the desktop control bar on phones.
  @media only screen and (max-width: 768px) {
    .mobile-dock {
      display: block;
      flex: none;
      padding: 8px 10px calc(8px + env(safe-area-inset-bottom));
      background: rgba(18, 21, 30, 0.98);
      border-top: 1px solid var(--glass-border);
    }
  }

  .dock {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .dock-pill {
    flex: 1 1 auto;
    min-width: 0;
  }

  .dock-buttons {
    flex: none;
    display: flex;
    align-items: center;
    gap: 6px;
    margin-left: auto;
  }

  .dock-btn {
    position: relative;
    width: 44px;
    height: 44px;
    border-radius: var(--radius-inner, 10px);
    border: 1px solid var(--glass-border);
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-subtle);
    font-size: 16px;

    &.active {
      color: var(--color-cyber-mint);
      border-color: rgba(38, 230, 180, 0.4);
    }

    &:disabled {
      opacity: 0.5;
    }

    .dot {
      position: absolute;
      top: 7px;
      right: 7px;
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: #f5c26b;
    }
  }

  .sheet-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1000;
    display: flex;
    align-items: flex-end;
    background: rgba(5, 5, 8, 0.55);
  }

  .sheet {
    width: 100%;
    max-height: 80vh;
    overflow-y: auto;
    box-sizing: border-box;
    padding: 8px 16px calc(16px + env(safe-area-inset-bottom));
    border-radius: 18px 18px 0 0;
    border-top: 1px solid var(--glass-border);
    background: var(--bg-obsidian-floating, #12141d);
    box-shadow: 0 -12px 40px rgba(0, 0, 0, 0.45);

    .handle {
      width: 40px;
      height: 4px;
      margin: 0 auto 14px;
      border-radius: 2px;
      background: rgba(255, 255, 255, 0.18);
    }
  }

  .sheet-row.sound {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 6px 2px 14px;

    .round {
      flex: none;
      width: 44px;
      height: 44px;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.06);
      color: var(--text-pure);
      font-size: 16px;

      &.pressed {
        color: #f5c26b;
      }
    }

    .sound-label {
      color: var(--text-subtle);
      font-size: 14px;
    }

    input[type='range'] {
      flex: 1;
      appearance: none;
      height: 4px;
      border-radius: 4px;
      background: rgba(255, 255, 255, 0.15);

      &::-webkit-slider-thumb {
        appearance: none;
        width: 22px;
        height: 22px;
        border-radius: 50%;
        background: var(--color-cyber-mint);
      }

      &::-moz-range-thumb {
        width: 22px;
        height: 22px;
        border: none;
        border-radius: 50%;
        background: var(--color-cyber-mint);
      }
    }
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  .tile {
    flex-direction: column;
    gap: 7px;
    min-height: 76px;
    padding: 10px 6px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.07);
    color: var(--text-subtle);

    i {
      font-size: 18px;
    }

    span {
      font-size: 12px;
      line-height: 1.25;
      text-align: center;
    }

    &.on {
      color: var(--color-cyber-mint);
      border-color: rgba(38, 230, 180, 0.4);
    }
  }

  .close {
    width: 100%;
    height: 46px;
    margin-top: 14px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--text-pure);
    font-weight: 600;
  }

  .sheet-enter-active,
  .sheet-leave-active {
    transition: opacity 0.2s ease;

    .sheet {
      transition: transform 0.22s ease;
    }
  }

  .sheet-enter,
  .sheet-leave-to {
    opacity: 0;

    .sheet {
      transform: translateY(100%);
    }
  }
</style>

<script lang="ts">
  import { Component, Mixins, Prop, Ref, Watch } from 'vue-property-decorator'
  import MediaControls from './mixins/media-controls'
  import ControlPill from './control-pill.vue'
  import Emotes from './emotes.vue'

  export type DockAction = 'fullscreen' | 'keyboard' | 'keyboard-helper' | 'clipboard' | 'pip' | 'resolution'

  // Phone layout: control pill, pause, emotes and a "More" bottom sheet for
  // everything else (sound, microphone, fullscreen, keyboard, clipboard, ...).
  // Player actions are emitted as `action` events so the parent can forward
  // them to the video component that owns them.
  @Component({
    name: 'neko-mobile-dock',
    components: { 'neko-control-pill': ControlPill, 'neko-emotes': Emotes },
  })
  export default class NekoMobileDock extends Mixins(MediaControls) {
    @Prop(Boolean) readonly shakeKbd!: boolean
    @Prop(Boolean) readonly pipAvailable!: boolean
    @Ref('moreButton') readonly moreButton!: HTMLButtonElement
    @Ref('first') readonly first!: HTMLButtonElement

    open = false

    get showPill() {
      return !this.viewOnly && !this.receiveOnly
    }

    // Local input pause for the current host (same flag as the desktop pill).
    get inputLocked() {
      return this.$accessor.remote.locked && this.hosting
    }

    set inputLocked(locked: boolean) {
      this.$accessor.remote.setLocked(locked)
    }

    get touchDevice() {
      if (this.$accessor.settings.force_touch) return true
      return 'ontouchstart' in window || navigator.maxTouchPoints > 0
    }

    // iOS ignores scripted volume changes; only mute works there.
    get volumeAdjustable() {
      const ua = navigator.userAgent || ''
      const iOS = /iPad|iPhone|iPod/.test(ua) || (navigator.maxTouchPoints > 1 && /Macintosh/.test(ua))
      return !iOS
    }

    get streamName() {
      const selected = this.$accessor.hls.selected
      if (selected === 'll-hls') return 'LL-HLS'
      if (selected === 'hls') return 'HLS'
      return 'WebRTC'
    }

    @Watch('open')
    onOpen(open: boolean) {
      if (open) {
        this.$nextTick(() => this.first && this.first.focus())
      }
    }

    close() {
      this.open = false
      this.$nextTick(() => this.moreButton && this.moreButton.focus())
    }

    run(action: DockAction, event?: MouseEvent) {
      this.open = false
      this.$emit('action', action, event)
    }

    openStreamSettings() {
      this.open = false
      this.$accessor.client.setTab('settings')
      this.$accessor.client.setSide(true)
    }
  }
</script>
