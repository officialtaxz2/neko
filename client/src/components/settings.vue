<template>
  <div class="settings">
    <div class="bento-grid">

      <!-- Scrolling card (full width) -->
      <div class="bento-card bento-full">
        <div class="card-header">
          <i class="fas fa-mouse" aria-hidden="true" />
          <span class="card-title">{{ $t('setting.scroll') }}</span>
        </div>
        <div class="card-body">
          <div class="row">
            <span>{{ $t('setting.scroll') }}</span>
            <label class="slider">
              <input type="range" min="1" max="100" v-model="scroll" />
            </label>
          </div>
          <div class="row">
            <span>{{ $t('setting.scroll_invert') }}</span>
            <label class="switch">
              <input type="checkbox" v-model="scroll_invert" />
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
            <label class="switch">
              <input type="checkbox" v-model="autoplay" />
              <span />
            </label>
          </div>
          <div class="row">
            <span>{{ $t('setting.ignore_emotes') }}</span>
            <label class="switch">
              <input type="checkbox" v-model="ignore_emotes" />
              <span />
            </label>
          </div>
          <div class="row">
            <span>{{ $t('setting.chat_sound') }}</span>
            <label class="switch">
              <input type="checkbox" v-model="chat_sound" />
              <span />
            </label>
          </div>
        </div>
      </div>

      <!-- Input / Keyboard card -->
      <div class="bento-card">
        <div class="card-header">
          <i class="fas fa-keyboard" aria-hidden="true" />
          <span class="card-title">Input</span>
        </div>
        <div class="card-body">
          <div class="row">
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

      <!-- Broadcast card (admin only, full width) -->
      <div class="bento-card bento-full" v-if="admin">
        <div class="card-header">
          <i class="fas fa-broadcast-tower" aria-hidden="true" />
          <span class="card-title">{{ $t('setting.broadcast_title') }}</span>
          <button
            v-if="!broadcast_is_active"
            @click.stop.prevent="$accessor.settings.broadcastCreate(broadcast_url)"
            class="btn-icon"
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

      <!-- Session card (full width, only when connected) -->
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

    // ── Bento Grid ───────────────────────────────────────────────────────────
    .bento-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--space-3);
      align-items: start;
    }

    // ── Bento Card ───────────────────────────────────────────────────────────
    .bento-card {
      background: color-mix(in srgb, var(--color-surface-2) 90%, transparent);
      // Full --color-border opacity: clearly visible in both light and dark mode.
      // Previously 70% alpha caused near-invisible borders on light backgrounds.
      border: 1px solid var(--color-border);
      border-radius: var(--radius-lg);
      overflow: hidden;
      transition: box-shadow var(--transition-interactive);

      &:hover {
        box-shadow: var(--shadow-md);
      }

      // Full-width card spans both columns
      &.bento-full {
        grid-column: 1 / -1;
      }
    }

    // ── Card Header ──────────────────────────────────────────────────────────
    .card-header {
      display: flex;
      align-items: center;
      gap: var(--space-2);
      padding: var(--space-2) var(--space-4);
      // Neutral surface — removed teal 5% tint to reduce color saturation.
      // Only the active/checked toggle and the broadcast status button use primary color.
      border-bottom: 1px solid color-mix(in srgb, var(--color-border) 55%, transparent);
      background: var(--color-surface);

      i {
        // --color-text-muted instead of --color-primary:
        // icons provide structural context, not emphasis — neutral is correct here.
        // Teal is reserved for interactive/active states (toggles, active tabs).
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

      // Compact icon button in card header (broadcast start/stop)
      .btn-icon {
        min-width: 28px;
        min-height: 28px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
        border: none;
        background: var(--color-primary);
        color: var(--color-text-inverse);
        cursor: pointer;
        font-size: var(--text-xs);
        transition:
          background var(--transition-interactive),
          transform  var(--transition-fast);

        &:hover  { background: var(--color-primary-hover); }
        &:active { transform: scale(0.92); }

        &.btn-red {
          background: var(--color-error);
          &:hover { background: var(--color-error-hover); }
        }
      }
    }

    // ── Card Body rows ────────────────────────────────────────────────────────
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

        // Full-width row variant (broadcast input)
        &.row-full {
          padding: var(--space-3) var(--space-4);
        }
      }
    }

    // ── Logout button ─────────────────────────────────────────────────────────
    .btn-logout {
      width: 100%;
      min-height: 44px;
      border-radius: var(--radius-md);
      border: none;
      background: var(--color-primary);
      color: var(--color-text-inverse);
      font-size: var(--text-sm);
      font-weight: 700;
      text-transform: uppercase;
      cursor: pointer;
      transition:
        background var(--transition-interactive),
        transform  var(--transition-fast);

      &:hover  { background: var(--color-primary-hover); }
      &:active { background: var(--color-primary-active); transform: scale(0.97); }
    }

    // ── Toggle Switch — 44×44px touch target, 42×24px visible pill ────────────
    //
    // OFF-state track: --color-text-faint instead of --color-border.
    // In light mode --color-border is hsl(220,10%,84%) on a near-white bg
    // = ~1.2:1 contrast (WCAG fail). --color-text-faint is hsl(220,10%,64%)
    // = ~2.6:1 — clearly visible as a grey pill in both modes.
    .switch {
      flex-shrink: 0;
      display: inline-block;
      position: relative;
      width: 44px;
      height: 44px;
      padding: 10px 1px;
      cursor: pointer;

      input {
        position: absolute;
        opacity: 0;
        width: 0;
        height: 0;
      }

      span {
        position: absolute;
        cursor: pointer;
        top: 10px;
        left: 1px;
        right: 1px;
        bottom: 10px;
        background-color: var(--color-text-faint);
        border-radius: var(--radius-full);
        transition: background-color var(--transition-interactive);

        &::before {
          content: '';
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

      &:hover span { background-color: var(--color-text-muted); }
    }

    input[type='checkbox'] {
      &:checked + span {
        background-color: var(--color-primary);
        &::before { transform: translateX(18px); }
      }
      &:checked + span:hover { background-color: var(--color-primary-hover); }
    }

    // ── Scroll speed slider ───────────────────────────────────────────────────
    .slider {
      flex-shrink: 0;
      max-width: 120px;

      input[type='range'] {
        display: inline-block;
        background: transparent;
        appearance: none;
        -webkit-appearance: none;
        height: 24px;
        max-width: 120px;
        cursor: pointer;

        &::-moz-range-track {
          width: 100%; height: 4px; cursor: pointer;
          background: var(--color-primary); border-radius: var(--radius-full);
        }
        &::-moz-range-thumb {
          height: 12px; width: 12px; border-radius: var(--radius-full);
          background: #fff; cursor: pointer; border: none;
          box-shadow: var(--shadow-sm); transition: transform var(--transition-fast);
        }
        &:hover::-moz-range-thumb { transform: scale(1.35); }

        &::-webkit-slider-runnable-track {
          width: 100%; height: 4px; cursor: pointer;
          background: var(--color-primary); border-radius: var(--radius-full);
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

    // ── Keyboard layout select ────────────────────────────────────────────────
    .select {
      flex-shrink: 0;
      max-width: 120px;
      text-align: right;

      select {
        -webkit-appearance: none;
        -moz-appearance: none;
        appearance: none;
        display: block;
        width: 100%;
        max-width: 100%;
        height: 30px;
        text-align: right;
        padding: 0 var(--space-1) 0 var(--space-2);
        margin: 0;
        line-height: 30px;
        font-weight: 400;
        font-size: var(--text-xs);
        text-overflow: ellipsis;
        border: 1px solid transparent;
        border-radius: var(--radius-md);
        color: var(--color-text);
        background-color: var(--color-bg);
        cursor: pointer;
        transition:
          border-color     var(--transition-interactive),
          background-color var(--transition-interactive);

        &:hover {
          border-color: var(--color-border);
          background-color: var(--color-surface-offset);
        }

        option {
          font-weight: normal;
          color: var(--color-text);
          background-color: var(--color-bg);
        }
      }
    }

    // ── Broadcast URL input ───────────────────────────────────────────────────
    .input {
      display: block;
      width: 100%;
      height: 36px;
      padding: 0 var(--space-3);
      line-height: 36px;
      font-size: var(--text-sm);
      font-weight: 300;
      border: 1px solid transparent;
      border-radius: var(--radius-md);
      color: var(--color-text);
      background-color: color-mix(in srgb, var(--color-bg) 80%, transparent);
      user-select: auto;
      transition:
        border-color     var(--transition-interactive),
        background-color var(--transition-interactive);

      &:focus {
        border-color: var(--color-primary);
        outline: none;
      }

      &::selection { background: var(--color-primary-highlight); }

      &[disabled] {
        opacity: 0.5;
        cursor: not-allowed;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-settings' })
  export default class extends Vue {
    private broadcast_url: string = ''

    get admin() {
      return this.$accessor.user.admin
    }

    get connected() {
      return this.$accessor.connected
    }

    get scroll() {
      return this.$accessor.settings.scroll.toString()
    }

    set scroll(value: string) {
      this.$accessor.settings.setScroll(parseInt(value))
    }

    get scroll_invert() {
      return this.$accessor.settings.scroll_invert
    }

    set scroll_invert(value: boolean) {
      this.$accessor.settings.setInvert(value)
    }

    get autoplay() {
      return this.$accessor.settings.autoplay
    }

    set autoplay(value: boolean) {
      this.$accessor.settings.setAutoplay(value)
    }

    get ignore_emotes() {
      return this.$accessor.settings.ignore_emotes
    }

    set ignore_emotes(value: boolean) {
      this.$accessor.settings.setIgnore(value)
    }

    get chat_sound() {
      return this.$accessor.settings.chat_sound
    }

    set chat_sound(value: boolean) {
      this.$accessor.settings.setSound(value)
    }

    get keyboard_layouts_list() {
      return this.$accessor.settings.keyboard_layouts_list
    }

    get keyboard_layout() {
      return this.$accessor.settings.keyboard_layout
    }

    get broadcast_is_active() {
      return this.$accessor.settings.broadcast_is_active
    }

    get broadcast_url_remote() {
      return this.$accessor.settings.broadcast_url
    }

    @Watch('broadcast_url_remote', { immediate: true })
    onBroadcastUrlChange() {
      this.broadcast_url = this.broadcast_url_remote
    }

    set keyboard_layout(value: string) {
      this.$accessor.settings.setKeyboardLayout(value)
      this.$accessor.remote.changeKeyboard()
    }

    logout() {
      this.$accessor.logout()
    }
  }
</script>
