<template>
  <div class="settings">
    <div class="bento-grid" ref="grid">

      <!-- ── GROUP: Video & Audio ───────────────────────────────────── -->
      <div class="settings-group bento-full">
        <button class="settings-group-header" @click="toggle('video')" :aria-expanded="isOpen('video')">
          <i class="fas fa-film" aria-hidden="true" />
          <span class="group-title">Video &amp; Audio</span>
          <i class="fas fa-chevron-down group-chevron" :class="{ 'group-chevron--open': isOpen('video') }" aria-hidden="true" />
        </button>
        <div class="settings-group-content" :class="{ collapsed: !isOpen('video') }">
          <div class="settings-group-inner">
            <div class="bento-grid bento-grid--inner">

              <!-- Scrolling card -->
              <div class="bento-card bento-full">
                <div class="card-header">
                  <i class="fas fa-mouse" aria-hidden="true" />
                  <span class="card-title">{{ $t('setting.scroll') }}</span>
                </div>
                <div class="card-body">
                  <div class="row">
                    <span>{{ $t('setting.scroll') }}</span>
                    <label class="slider">
                      <div class="slider-wrap">
                        <input
                          type="range" min="1" max="20" v-model="scroll"
                          :style="{ '--fill': ((+scroll - 1) / 19 * 100) + '%' }"
                          @mousedown="showTooltip"
                          @touchstart="showTooltip"
                          @mouseup="hideTooltip"
                          @touchend="hideTooltip"
                        />
                        <span
                          class="slider-tooltip"
                          :class="{ 'slider-tooltip--visible': tooltipVisible }"
                          :style="{ left: ((+scroll - 1) / 19 * 100) + '%' }"
                          aria-hidden="true"
                        >{{ scroll }}</span>
                      </div>
                    </label>
                  </div>
                  <div class="row">
                    <span>{{ $t('setting.scroll_invert') }}</span>
                    <label class="switch">
                      <input type="checkbox" v-model="scroll_invert" />
                      <span />
                    </label>
                  </div>
                  <div class="row" :class="{ 'row--dimmed': !is_touch_device }">
                    <div class="row-label">
                      <span>{{ $t('setting.trackpad_mode') }}</span>
                      <span class="badge" :class="is_touch_device ? 'badge--mobile' : 'badge--desktop'">
                        {{ is_touch_device ? $t('setting.mobile') : $t('setting.desktop_only_inactive') }}
                      </span>
                    </div>
                    <label class="switch">
                      <input type="checkbox" v-model="trackpad_mode" :disabled="!is_touch_device" />
                      <span />
                    </label>
                  </div>
                </div>
              </div>

              <!-- Video card -->
              <div class="bento-card">
                <div class="card-header">
                  <i class="fas fa-film" aria-hidden="true" />
                  <span class="card-title">Video</span>
                </div>
                <div class="card-body">
                  <div class="row">
                    <span>Show FPS / Bitrate stats</span>
                    <label class="switch">
                      <input type="checkbox" v-model="show_stats" />
                      <span />
                    </label>
                  </div>
                </div>
              </div>

              <!-- Chat card -->
              <div class="bento-card">
                <div class="card-header">
                  <i class="fas fa-comment-alt" aria-hidden="true" />
                  <span class="card-title">Chat</span>
                </div>
                <div class="card-body">
                  <div class="row">
                    <span>{{ $t('setting.autoplay') }}</span>
                    <label class="switch"><input type="checkbox" v-model="autoplay" /><span /></label>
                  </div>
                  <div class="row">
                    <span>{{ $t('setting.ignore_emotes') }}</span>
                    <label class="switch"><input type="checkbox" v-model="ignore_emotes" /><span /></label>
                  </div>
                  <div class="row">
                    <span>{{ $t('setting.chat_sound') }}</span>
                    <label class="switch"><input type="checkbox" v-model="chat_sound" /><span /></label>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>
      </div>

      <!-- ── GROUP: Input & Controls ───────────────────────────────── -->
      <div class="settings-group bento-full">
        <button class="settings-group-header" @click="toggle('input')" :aria-expanded="isOpen('input')">
          <i class="fas fa-keyboard" aria-hidden="true" />
          <span class="group-title">Input &amp; Controls</span>
          <i class="fas fa-chevron-down group-chevron" :class="{ 'group-chevron--open': isOpen('input') }" aria-hidden="true" />
        </button>
        <div class="settings-group-content" :class="{ collapsed: !isOpen('input') }">
          <div class="settings-group-inner">
            <div class="bento-grid bento-grid--inner">

              <!-- Input / Keyboard card -->
              <div class="bento-card bento-full">
                <div class="card-header">
                  <i class="fas fa-keyboard" aria-hidden="true" />
                  <span class="card-title">Input</span>
                </div>
                <div class="card-body">
                  <div class="row row--select">
                    <span>{{ $t('setting.keyboard_layout') }}</span>
                    <label class="select">
                      <select v-model="keyboard_layout">
                        <option v-for="(name, code) in keyboard_layouts_list" :key="code" :value="code">{{ name }}</option>
                      </select>
                      <span />
                    </label>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>
      </div>

      <!-- ── GROUP: Admin (admin-only) ─────────────────────────────── -->
      <div class="settings-group bento-full" v-if="admin">
        <button class="settings-group-header" @click="toggle('admin')" :aria-expanded="isOpen('admin')">
          <i class="fas fa-shield-alt" aria-hidden="true" />
          <span class="group-title">Admin</span>
          <i class="fas fa-chevron-down group-chevron" :class="{ 'group-chevron--open': isOpen('admin') }" aria-hidden="true" />
        </button>
        <div class="settings-group-content" :class="{ collapsed: !isOpen('admin') }">
          <div class="settings-group-inner">
            <div class="bento-grid bento-grid--inner">

              <!-- Display card -->
              <div class="bento-card">
                <div class="card-header">
                  <i class="fas fa-desktop" aria-hidden="true" />
                  <span class="card-title">Display</span>
                </div>
                <div class="card-body">
                  <div class="row row--select">
                    <span>Resolution</span>
                    <label class="select">
                      <select v-model="resValue">
                        <option
                          v-for="(conf, i) in resConfigurations"
                          :key="i"
                          :value="`${conf.width}x${conf.height}@${conf.rate}`"
                        >
                          {{ conf.width }} × {{ conf.height }} @ {{ conf.rate }} fps
                        </option>
                      </select>
                      <span />
                    </label>
                  </div>
                </div>
              </div>

              <!-- Broadcast card (full width) -->
              <div class="bento-card bento-full">
                <div class="card-header">
                  <i class="fas fa-broadcast-tower" aria-hidden="true" />
                  <span class="card-title">{{ $t('setting.broadcast_title') }}</span>
                  <button
                    v-if="!broadcast_is_active"
                    @click.stop.prevent="$accessor.settings.broadcastCreate(broadcast_url)"
                    class="btn-icon btn-icon--ghost"
                    :aria-label="$t('setting.broadcast_title')"
                  >
                    <i class="fas fa-play" aria-hidden="true" />
                  </button>
                  <button
                    v-else
                    @click.stop.prevent="$accessor.settings.broadcastDestroy()"
                    class="btn-icon btn-red"
                    :aria-label="$t('setting.broadcast_title')"
                  >
                    <i class="fas fa-stop" aria-hidden="true" />
                  </button>
                </div>
                <div class="card-body">
                  <div class="row row-full">
                    <input
                      v-model="broadcast_url"
                      :disabled="broadcast_is_active"
                      class="input"
                      placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
                    />
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>
      </div>

      <!-- ── Session card (standalone, always visible) ─────────────── -->
      <div class="bento-card bento-full" v-if="connected">
        <div class="card-body">
          <div class="row">
            <button class="btn-logout" @click.stop.prevent="logout">{{ $t('logout') }}</button>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<style lang="scss" scoped>
  .settings {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    padding: var(--space-4) var(--space-3);

    .bento-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--space-3);
      align-items: start;
    }

    // Inner grids inside group content reset gap slightly
    .bento-grid--inner {
      gap: var(--space-3);
    }

    .bento-card {
      position: relative;
      overflow: hidden;
      background: color-mix(in srgb, var(--color-surface-2) 90%, transparent);
      border: 1px solid var(--color-border);
      border-radius: var(--radius-lg);
      transition: box-shadow var(--transition-interactive);

      &:hover { box-shadow: var(--shadow-md); }
      &.bento-full { grid-column: 1 / -1; }

      // Cursor-shine overlay
      &::before {
        content: '';
        position: absolute;
        inset: 0;
        border-radius: inherit;
        pointer-events: none;
        background: radial-gradient(
          320px circle at var(--cursor-x, 50%) var(--cursor-y, 50%),
          color-mix(in srgb, var(--color-primary) 10%, transparent) 0%,
          transparent 70%
        );
        opacity: 0;
        transition: opacity 220ms ease;
        z-index: 0;
      }

      &:hover::before { opacity: 1; }

      > * { position: relative; z-index: 1; }
    }

    // ── Collapsible group ─────────────────────────────────────────────
    .settings-group {
      grid-column: 1 / -1;
      border: 1px solid var(--color-border);
      border-radius: var(--radius-lg);
      overflow: hidden;
      background: color-mix(in srgb, var(--color-surface) 85%, transparent);
    }

    .settings-group-header {
      width: 100%;
      display: flex;
      align-items: center;
      gap: var(--space-2);
      padding: var(--space-3) var(--space-4);
      background: var(--color-surface);
      border: none;
      border-bottom: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
      cursor: pointer;
      text-align: left;
      transition: background var(--transition-interactive);

      &:hover { background: var(--color-surface-2); }

      &:focus-visible {
        outline: 2px solid var(--color-primary);
        outline-offset: -2px;
      }

      i:first-child {
        color: var(--color-text-muted);
        font-size: var(--text-xs);
        flex-shrink: 0;
      }

      .group-title {
        flex: 1;
        font-size: var(--text-xs);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--color-text-muted);
      }

      .group-chevron {
        color: var(--color-text-muted);
        font-size: var(--text-xs);
        flex-shrink: 0;
        // Collapsed state: chevron points right (default); open: points down
        transform: rotate(-90deg);
        transition: transform 280ms cubic-bezier(0.16, 1, 0.3, 1);

        &.group-chevron--open { transform: rotate(0deg); }
      }
    }

    .settings-group-content {
      display: grid;
      grid-template-rows: 1fr;
      overflow: hidden;
      transition: grid-template-rows 280ms cubic-bezier(0.16, 1, 0.3, 1);

      &.collapsed { grid-template-rows: 0fr; }
    }

    .settings-group-inner {
      // Required for grid-template-rows collapse trick
      min-height: 0;
      padding: var(--space-3);
    }

    .card-header {
      display: flex;
      align-items: center;
      gap: var(--space-2);
      padding: var(--space-2) var(--space-4);
      border-bottom: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
      background: var(--color-surface);

      i {
        color: var(--color-text-muted);
        font-size: var(--text-xs);
        flex-shrink: 0;
      }

      .card-title {
        flex: 1;
        font-size: var(--text-xs);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--color-text-muted);
      }

      .btn-icon {
        min-width: 28px;
        min-height: 28px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
        cursor: pointer;
        font-size: var(--text-xs);
        transition:
          background var(--transition-interactive),
          color       var(--transition-interactive),
          border-color var(--transition-interactive),
          transform   var(--transition-fast);

        border: 1px solid var(--color-primary);
        background: var(--color-primary);
        color: var(--color-text-inverse);

        &:hover  { background: var(--color-primary-hover); border-color: var(--color-primary-hover); }
        &:active { transform: scale(0.92); }

        &.btn-icon--ghost {
          background: transparent;
          color: var(--color-primary);
          border: 1px solid color-mix(in srgb, var(--color-primary) 55%, transparent);

          &:hover {
            background: var(--color-primary);
            color: var(--color-text-inverse);
            border-color: var(--color-primary);
          }
        }

        &.btn-red {
          background: var(--color-error);
          border-color: var(--color-error);
          color: #fff;
          &:hover { background: var(--color-error-hover); border-color: var(--color-error-hover); }
        }
      }
    }

    .card-body {
      padding: var(--space-1) 0;

      .row {
        min-height: 44px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 var(--space-4);
        gap: var(--space-3);
        border-bottom: 1px solid color-mix(in srgb, var(--color-border) 40%, transparent);

        &:last-child { border-bottom: none; }

        > span {
          flex: 1;
          color: var(--color-text);
          font-size: var(--text-sm);
        }

        &.row--dimmed {
          opacity: 0.75;
          .switch { cursor: not-allowed; }
        }

        &.row-full { padding: var(--space-3) var(--space-4); }

        &.row--select {
          flex-direction: column;
          align-items: stretch;
          justify-content: center;
          padding-top: var(--space-2);
          padding-bottom: var(--space-2);
          gap: var(--space-2);

          > span { flex: none; }

          .select {
            width: 100%;
            min-height: unset;

            select { width: 100%; }
          }
        }

        .row-label {
          flex: 1;
          display: flex;
          align-items: center;
          gap: var(--space-2);
        }

        .badge {
          display: inline-flex;
          align-items: center;
          padding: 2px var(--space-2);
          border-radius: var(--radius-full);
          font-size: var(--text-xs);
          font-weight: 500;
          line-height: 1;
          white-space: nowrap;

          &.badge--mobile {
            background: color-mix(in srgb, var(--color-primary) 16%, transparent);
            color: var(--color-primary);
            border: 1px solid color-mix(in srgb, var(--color-primary) 30%, transparent);
          }

          &.badge--desktop {
            background: color-mix(in srgb, var(--color-text-faint) 16%, transparent);
            color: var(--color-text-muted);
            border: 1px solid color-mix(in srgb, var(--color-text-faint) 30%, transparent);
          }
        }
      }
    }

    // ── Scroll sensitivity slider ─────────────────────────────────────
    .slider {
      display: flex;
      align-items: center;

      .slider-wrap {
        position: relative;
        display: flex;
        align-items: center;
        width: 120px;
      }

      .slider-tooltip {
        position: absolute;
        top: -28px;
        transform: translateX(-50%);
        background: var(--color-surface-offset);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-sm);
        padding: 1px var(--space-2);
        font-size: var(--text-xs);
        color: var(--color-text);
        white-space: nowrap;
        pointer-events: none;
        opacity: 0;
        transition: opacity 150ms ease;

        &::after {
          content: '';
          position: absolute;
          top: 100%;
          left: 50%;
          transform: translateX(-50%);
          border: 4px solid transparent;
          border-top-color: var(--color-border);
        }

        &.slider-tooltip--visible { opacity: 1; }
      }

      input[type='range'] {
        width: 120px;
        max-width: 120px;
        -webkit-appearance: none;
        appearance: none;
        height: 4px;
        border-radius: var(--radius-full);
        outline: none;
        cursor: pointer;

        &::-moz-range-track {
          height: 4px;
          border-radius: var(--radius-full);
          background: linear-gradient(
            to right,
            var(--color-primary)         0% var(--fill, 21%),
            var(--color-surface-dynamic) var(--fill, 21%) 100%
          );
        }

        &::-moz-range-thumb {
          width: 14px; height: 14px;
          border-radius: var(--radius-full);
          background: var(--color-primary);
          border: none;
          cursor: grab;
        }
        &:hover::-moz-range-thumb { transform: scale(1.35); }

        &::-webkit-slider-runnable-track {
          height: 4px;
          border-radius: var(--radius-full);
          background: linear-gradient(
            to right,
            var(--color-primary)         0% var(--fill, 21%),
            var(--color-surface-dynamic) var(--fill, 21%) 100%
          );
        }

        &::-webkit-slider-thumb {
          -webkit-appearance: none;
          width: 14px; height: 14px;
          margin-top: -5px;
          border-radius: var(--radius-full);
          background: var(--color-primary);
          border: none;
          cursor: grab;
          transition: transform var(--transition-fast);
        }
        &:hover::-webkit-slider-thumb { transform: scale(1.35); }
      }
    }

    .switch {
      position: relative;
      display: inline-block;
      width: 42px;
      height: 24px;
      flex-shrink: 0;

      input {
        opacity: 0;
        width: 0;
        height: 0;
        position: absolute;

        &:checked + span { background: var(--color-primary); }
        &:checked + span::before { transform: translateX(18px); }
        &:disabled + span { opacity: 0.45; cursor: not-allowed; }
        &:focus-visible + span { outline: 2px solid var(--color-primary); outline-offset: 2px; }
      }

      span {
        position: absolute;
        inset: 0;
        background: var(--color-surface-dynamic);
        border-radius: var(--radius-full);
        cursor: pointer;
        transition: background var(--transition-interactive);

        &::before {
          content: '';
          position: absolute;
          width: 18px; height: 18px;
          left: 3px; top: 3px;
          background: #fff;
          border-radius: var(--radius-full);
          transition: transform var(--transition-interactive);
        }
      }
    }

    .select {
      position: relative;
      display: inline-flex;
      align-items: center;
      min-height: 52px;

      select {
        -webkit-appearance: none;
        appearance: none;
        background: var(--color-surface-offset);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-md);
        padding: var(--space-2) var(--space-8) var(--space-2) var(--space-3);
        color: var(--color-text);
        font-size: var(--text-sm);
        cursor: pointer;
        max-width: 100%;
        transition:
          border-color var(--transition-interactive),
          background   var(--transition-interactive);

        &:hover   { border-color: var(--color-primary); }
        &:focus   { outline: none; border-color: var(--color-primary); }
        &:disabled { opacity: 0.5; cursor: not-allowed; }
      }

      span {
        position: absolute;
        right: var(--space-3);
        pointer-events: none;
        color: var(--color-text-muted);

        &::after { content: '▾'; font-size: var(--text-xs); }
      }
    }

    .input {
      flex: 1;
      background: var(--color-surface-offset);
      border: 1px solid var(--color-border);
      border-radius: var(--radius-md);
      padding: var(--space-2) var(--space-3);
      color: var(--color-text);
      font-size: var(--text-sm);
      max-width: 100%;
      transition: border-color var(--transition-interactive);

      &::placeholder { color: var(--color-text-faint); }
      &:hover        { border-color: var(--color-primary); }
      &:focus        { outline: none; border-color: var(--color-primary); }
      &:disabled     { opacity: 0.5; cursor: not-allowed; }
    }

    .btn-logout {
      width: 100%;
      min-height: 44px;
      background: color-mix(in srgb, var(--color-error) 10%, transparent);
      border: 1px solid color-mix(in srgb, var(--color-error) 35%, transparent);
      border-radius: var(--radius-md);
      color: var(--color-error);
      font-size: var(--text-sm);
      font-weight: 500;
      cursor: pointer;
      transition:
        background    var(--transition-interactive),
        border-color  var(--transition-interactive),
        color         var(--transition-interactive);

      &:hover {
        background: var(--color-error);
        border-color: var(--color-error);
        color: #fff;
      }

      &:active { transform: scale(0.98); }
    }

    // ── prefers-reduced-motion ────────────────────────────────────────
    @media (prefers-reduced-motion: reduce) {
      .bento-card::before { display: none; }

      .settings-group-content,
      .group-chevron { transition: none; }
    }
  }
</style>

<script lang="ts">
  import { Vue, Component } from 'vue-property-decorator'
  import Resolution from '~/components/resolution.vue'

  const STATS_STORAGE_KEY = 'neko_show_stats'

  // Group IDs mapped to localStorage keys
  const GROUP_KEYS: Record<string, string> = {
    video:  'neko_settings_group_video',
    input:  'neko_settings_group_input',
    admin:  'neko_settings_group_admin',
  }

  @Component({
    name: 'neko-settings',
    components: {
      'neko-resolution': Resolution,
    },
  })
  export default class extends Vue {
    // ── Collapsible group state ───────────────────────────────────────
    // Reactive map; all groups default open (true)
    private groupOpen: Record<string, boolean> = {
      video: true,
      input: true,
      admin: true,
    }

    created() {
      // Restore per-group state from localStorage (default: open)
      for (const [id, key] of Object.entries(GROUP_KEYS)) {
        const stored = localStorage.getItem(key)
        if (stored !== null) {
          this.$set(this.groupOpen, id, stored !== 'false')
        }
      }
    }

    isOpen(id: string): boolean {
      return this.groupOpen[id] !== false
    }

    toggle(id: string): void {
      const next = !this.isOpen(id)
      this.$set(this.groupOpen, id, next)
      const key = GROUP_KEYS[id]
      if (key) localStorage.setItem(key, String(next))
    }

    // ── Tooltip state ─────────────────────────────────────────────────
    tooltipVisible = false
    private _tooltipTimer: ReturnType<typeof setTimeout> | null = null

    // ── Cursor-shine: track pointer per card ──────────────────────────
    private _onMouseMove: ((e: MouseEvent) => void) | null = null

    mounted() {
      if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

      const grid = this.$refs.grid as HTMLElement | undefined
      if (!grid) return

      this._onMouseMove = (e: MouseEvent) => {
        const cards = grid.querySelectorAll<HTMLElement>('.bento-card')
        cards.forEach((card) => {
          const rect = card.getBoundingClientRect()
          const x = ((e.clientX - rect.left) / rect.width  * 100).toFixed(2) + '%'
          const y = ((e.clientY - rect.top)  / rect.height * 100).toFixed(2) + '%'
          card.style.setProperty('--cursor-x', x)
          card.style.setProperty('--cursor-y', y)
        })
      }

      grid.addEventListener('mousemove', this._onMouseMove, { passive: true })
    }

    beforeDestroy() {
      const grid = this.$refs.grid as HTMLElement | undefined
      if (grid && this._onMouseMove) {
        grid.removeEventListener('mousemove', this._onMouseMove)
      }
      this._onMouseMove = null
    }

    showTooltip() {
      if (this._tooltipTimer !== null) {
        clearTimeout(this._tooltipTimer)
        this._tooltipTimer = null
      }
      this.tooltipVisible = true
    }

    hideTooltip() {
      this._tooltipTimer = setTimeout(() => {
        this.tooltipVisible = false
        this._tooltipTimer = null
      }, 900)
    }

    // ── Stats toggle (localStorage, default: false) ───────────────────
    get show_stats(): boolean {
      return localStorage.getItem(STATS_STORAGE_KEY) === 'true'
    }
    set show_stats(v: boolean) {
      localStorage.setItem(STATS_STORAGE_KEY, String(v))
      window.dispatchEvent(new CustomEvent('neko:stats-toggle', { detail: v }))
    }

    // ── Computed: settings store bindings ─────────────────────────────
    get admin()     { return this.$accessor.user.admin }

    get is_touch_device(): boolean {
      return (
        ('ontouchstart' in window || navigator.maxTouchPoints > 0) &&
        ('ontouchend' in window)
      )
    }

    get scroll()    { return this.$accessor.settings.scroll.toString() }
    set scroll(v: string) { this.$accessor.settings.setScroll(parseInt(v)) }

    get scroll_invert()    { return this.$accessor.settings.scroll_invert }
    set scroll_invert(v: boolean) { this.$accessor.settings.setInvert(v) }

    get trackpad_mode()    { return this.$accessor.settings.trackpad_mode }
    set trackpad_mode(v: boolean) { this.$accessor.settings.setTrackpadMode(v) }

    get autoplay()    { return this.$accessor.settings.autoplay }
    set autoplay(v: boolean) { this.$accessor.settings.setAutoplay(v) }

    get ignore_emotes()    { return this.$accessor.settings.ignore_emotes }
    set ignore_emotes(v: boolean) { this.$accessor.settings.setIgnore(v) }

    get chat_sound()    { return this.$accessor.settings.chat_sound }
    set chat_sound(v: boolean) { this.$accessor.settings.setSound(v) }

    get keyboard_layout()    { return this.$accessor.settings.keyboard_layout }
    set keyboard_layout(v: string) { this.$accessor.settings.setKeyboardLayout(v) }

    get keyboard_layouts_list() { return this.$accessor.settings.keyboard_layouts_list }

    get broadcast_is_active() { return this.$accessor.settings.broadcast_is_active }
    get broadcast_url()       { return this.$accessor.settings.broadcast_url }
    set broadcast_url(v: string) { this.$accessor.settings.setBroadcastStatus({ url: v, isActive: this.broadcast_is_active }) }

    get connected()   { return this.$accessor.connected }

    // ── Resolution helpers ────────────────────────────────────────────
    get resConfigurations() {
      return this.$accessor.video.configurations
    }
    get resValue() {
      const v = this.$accessor.video
      return `${v.width}x${v.height}@${v.rate}`
    }
    set resValue(v: string) {
      const [res, rate] = v.split('@')
      const [width, height] = res.split('x').map(Number)
      this.$accessor.video.screenSet({ width, height, rate: Number(rate) })
    }

    logout() {
      this.$accessor.logout()
    }
  }
</script>
