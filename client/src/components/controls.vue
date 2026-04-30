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

    <li class="stats-badge no-pointer" v-if="playing && statsVisible">
      <div class="stats-inner">
        <span class="stat">
          <span class="stat-val" :key="'f' + statsKey">{{ displayFps }}</span>
          <span class="stat-unit">fps</span>
        </span>
        <span class="stat-sep">&middot;</span>
        <span class="stat">
          <span class="stat-val" :key="'b' + statsKey">{{ displayBitrateLabel }}</span>
          <span class="stat-unit">{{ bitrateUnit }}</span>
        </span>
      </div>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
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

  @keyframes stat-flip {
    0%   { opacity: 0.3; transform: translateY(-4px) scale(0.88); }
    100% { opacity: 1;   transform: translateY(0)    scale(1); }
  }

  ul {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    list-style: none;
    font-family: var(--font-controls);

    li {
      min-width: 44px;
      min-height: 44px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;

      &.no-pointer { cursor: default; }

      i {
        font-size: 24px;
        color: var(--color-text);
        will-change: transform;
        transition:
          transform var(--transition-fast),
          color    var(--transition-interactive),
          opacity  var(--transition-interactive);

        &:hover:not(.disabled):not(.shake) {
          transform: scale(1.18);
          color: var(--color-primary);
        }

        &:active:not(.disabled) {
          transform: scale(0.88);
          transition-duration: 80ms;
        }

        &.faded   { color: var(--color-text-muted); }
        &.disabled { color: var(--color-error); opacity: 0.4; pointer-events: none; }
      }

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

          &::-moz-range-track {
            width: 100%; height: 4px;
            background: var(--color-primary); border-radius: var(--radius-full); cursor: pointer;
          }
          &::-moz-range-thumb {
            height: 12px; width: 12px; border-radius: var(--radius-full);
            background: #fff; cursor: pointer; border: none;
            box-shadow: var(--shadow-sm); transition: transform var(--transition-fast);
          }
          &:hover::-moz-range-thumb { transform: scale(1.35); }

          &::-webkit-slider-runnable-track {
            width: 100%; height: 4px;
            background: var(--color-primary); border-radius: var(--radius-full); cursor: pointer;
          }
          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            height: 12px; width: 12px; border-radius: var(--radius-full);
            background: #fff; cursor: pointer; margin-top: -4px;
            box-shadow: var(--shadow-sm); transition: transform var(--transition-fast);
          }
          &:hover::-webkit-slider-thumb { transform: scale(1.35); }
        }
      }

      // ------------------------------------------------------------------
      // Lock toggle switch — OFF track: --color-text-muted
      // Contrast: light ~5:1 / dark ~3.7:1 (was text-faint: ~2.3:1 both).
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
          background-color: var(--color-text-muted);
          border-radius: var(--radius-full);
          transition: background-color var(--transition-interactive);

          &::before {
            font-family: 'Font Awesome 6 Free';
            font-weight: 900;
            content: '\f3c1';
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
          &::before { content: '\f023'; transform: translateX(18px); }
        }
        &:disabled + span {
          opacity: 0.45;
          cursor: not-allowed;
          &::before { content: ''; background-color: var(--color-text-muted); }
        }
      }

      &.stats-badge {
        min-width: unset;
        padding: 0 var(--space-1);

        .stats-inner {
          display: flex;
          align-items: center;
          gap: var(--space-1);
          padding: var(--space-1) var(--space-2);
          background: color-mix(in srgb, var(--color-surface-2) 85%, transparent);
          border: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
          border-radius: var(--radius-md);
          font-family: var(--font-controls);
          font-size: var(--text-xs);
          color: var(--color-text-muted);
          white-space: nowrap;
          user-select: none;
          transition:
            background-color var(--transition-slow),
            border-color     var(--transition-slow);
        }

        .stat { display: inline-flex; align-items: baseline; gap: 2px; }
        .stat-val {
          color: var(--color-primary);
          font-weight: 500;
          animation: stat-flip 220ms cubic-bezier(0.16, 1, 0.3, 1) both;
          display: inline-block;
          min-width: 2ch;
          text-align: right;
          font-variant-numeric: tabular-nums;
        }
        .stat-unit { color: var(--color-text-faint); }
        .stat-sep  { color: var(--color-text-faint); margin: 0 var(--space-1); line-height: 1; }
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Prop, Watch } from 'vue-property-decorator'

  @Component({ name: 'neko-controls' })
  export default class extends Vue {
    @Prop(Boolean) readonly shakeKbd!: boolean

    private statsInterval: number | null = null
    private animRafId: number = 0
    private lastBytesReceived = 0
    private lastStatsTime = 0
    private targetFps = 0
    private targetBitrateKbps = 0

    statsKey = 0
    displayFps = 0
    displayBitrateKbps = 0

    get statsVisible(): boolean {
      return this.displayFps > 0 || this.displayBitrateKbps > 0
    }

    get displayBitrateLabel(): string {
      return this.targetBitrateKbps >= 1000
        ? (this.displayBitrateKbps / 1000).toFixed(1)
        : String(Math.round(this.displayBitrateKbps))
    }

    get bitrateUnit(): string {
      return this.targetBitrateKbps >= 1000 ? 'Mbps' : 'Kbps'
    }

    mounted() {
      this.statsInterval = window.setInterval(() => void this.pollStats(), 2000)
    }

    beforeDestroy() {
      if (this.statsInterval !== null) clearInterval(this.statsInterval)
      cancelAnimationFrame(this.animRafId)
    }

    @Watch('playing')
    onPlayingChanged(isPlaying: boolean) {
      if (!isPlaying) {
        this.targetFps = 0
        this.targetBitrateKbps = 0
        this.lastBytesReceived = 0
        this.lastStatsTime = 0
        this.scheduleAnimate()
      }
    }

    private async pollStats(): Promise<void> {
      const pc = (this.$client as any).peerConnection as RTCPeerConnection | undefined
      if (!pc) return
      try {
        const now = performance.now()
        const stats = await pc.getStats()
        stats.forEach((r: any) => {
          if (r.type !== 'inbound-rtp' || r.kind !== 'video') return
          if (typeof r.framesPerSecond === 'number') {
            this.targetFps = Math.round(r.framesPerSecond)
          }
          if (typeof r.bytesReceived === 'number' && this.lastStatsTime > 0) {
            const dt = (now - this.lastStatsTime) / 1000
            if (dt > 0) {
              this.targetBitrateKbps = Math.round(
                ((r.bytesReceived - this.lastBytesReceived) * 8) / dt / 1000,
              )
            }
          }
          this.lastBytesReceived = r.bytesReceived ?? this.lastBytesReceived
          this.lastStatsTime = now
        })
        this.statsKey++
        this.scheduleAnimate()
      } catch { /* PC not yet connected */ }
    }

    private scheduleAnimate(): void {
      cancelAnimationFrame(this.animRafId)
      this.animRafId = requestAnimationFrame(this.animateTick)
    }

    private readonly animateTick = (): void => {
      const lerp = (a: number, b: number): number => {
        const d = b - a
        return Math.abs(d) < 1 ? b : Math.round(a + d * 0.18)
      }
      this.displayFps = lerp(this.displayFps, this.targetFps)
      this.displayBitrateKbps = lerp(this.displayBitrateKbps, this.targetBitrateKbps)
      if (this.displayFps !== this.targetFps || this.displayBitrateKbps !== this.targetBitrateKbps) {
        this.animRafId = requestAnimationFrame(this.animateTick)
      }
    }

    get controlLocked() {
      return 'control' in this.$accessor.locked && this.$accessor.locked['control'] && !this.$accessor.user.admin
    }
    get disabeld()        { return this.$accessor.remote.hosted }
    get hosting()         { return this.$accessor.remote.hosting }
    get controlling()     { return this.$accessor.remote.controlling }
    get implicitHosting() { return this.$accessor.remote.implicitHosting }
    get micAllowed()      { return this.controlling }
    get volume()          { return this.$accessor.video.volume }
    set volume(volume: number) { this.$accessor.video.setVolume(volume) }
    get muted()           { return this.$accessor.video.muted || this.volume === 0 }
    get playing()         { return this.$accessor.video.playing }
    get playable()        { return this.$accessor.video.playable }
    get locked()          { return this.$accessor.remote.locked && this.$accessor.remote.hosting }
    set locked(locked: boolean) { this.$accessor.remote.setLocked(locked) }

    toggleControl() { if (!this.playable) return; this.$accessor.remote.toggle() }
    toggleMedia()   { if (!this.playable) return; this.$accessor.video.togglePlay() }
    toggleMute()    { this.$accessor.video.toggleMute() }

    microphoneActive = false

    @Watch('controlling')
    onControllingChanged(isControlling: boolean) {
      if (!isControlling && this.microphoneActive) {
        this.$client.disableMicrophone()
        this.microphoneActive = false
      }
    }

    async toggleMicrophone() {
      if (!this.playable || !this.micAllowed) return
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
