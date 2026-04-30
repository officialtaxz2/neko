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
              <!--
                :style sets --fill so the CSS track gradient knows the
                filled portion. Formula: (value - min) / (max - min) * 100.
              -->
              <input
                type="range" min="1" max="100" v-model="scroll"
                :style="{ '--fill': ((+scroll - 1) / 99 * 100) + '%' }"
              />
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
          <!--
            Start button: outlined ghost style when inactive (no solid teal fill).
            Hover reveals the solid teal background. Stop button stays solid red.
          -->
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

    .bento-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: var(--space-3);
      align-items: start;
    }

    .bento-card {
      background: color-mix(in srgb, var(--color-surface-2) 90%, transparent);
      border: 1px solid var(--color-border);
      border-radius: var(--radius-lg);
      overflow: hidden;
      transition: box-shadow var(--transition-interactive);

      &:hover { box-shadow: var(--shadow-md); }
      &.bento-full { grid-column: 1 / -1; }
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

      // ── Broadcast control buttons ─────────────────────────────────
      // btn-icon--ghost: outlined, no background fill at rest.
      // Teal only appears on hover so the card header stays neutral
      // when no broadcast is running.
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

        // Default (solid teal) — kept as fallback
        border: 1px solid var(--color-primary);
        background: var(--color-primary);
        color: var(--color-text-inverse);

        &:hover  { background: var(--color-primary-hover); border-color: var(--color-primary-hover); }
        &:active { transform: scale(0.92); }

        // Ghost variant: transparent at rest, teal on hover
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

        &.row-full { padding: var(--space-3) var(--space-4); }
      }
    }

    // LOG OUT: ghost button at rest — destructive action should not use
    // the primary CTA color. Neutral at rest, error-red on hover to signal
    // the destructive nature of the action without being visually aggressive.
    .btn-logout {
      width: 100%;
      min-height: 44px;
      border-radius: var(--radius-md);
      border: 1px solid var(--color-border);
      background: transparent;
      color: var(--color-text-muted);
      font-size: var(--text-sm);
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      cursor: pointer;
      transition:
        background   var(--transition-interactive),
        border-color var(--transition-interactive),
        color        var(--transition-interactive),
        transform    var(--transition-fast);

      &:hover {
        background:   color-mix(in srgb, var(--color-error) 10%, transparent);
        border-color: var(--color-error);
        color:        var(--color-error);
      }
      &:active { transform: scale(0.97); }
    }

    // ── Toggle Switch ─────────────────────────────────────────────────────
    .switch {
      flex-shrink: 0;
      display: inline-block;
      position: relative;
      width: 44px;
      height: 44px;
      padding: 10px 1px;
      cursor: pointer;

      input { position: absolute; opacity: 0; width: 0; height: 0; }

      span {
        position: absolute;
        cursor: pointer;
        top: 10px; left: 1px; right: 1px; bottom: 10px;
        background-color: var(--color-text-muted);
        border-radius: var(--radius-full);
        transition: background-color var(--transition-interactive);

        &::before {
          content: '';
          position: absolute;
          height: 18px; width: 18px;
          left: 3px; bottom: 3px;
          background-color: #fff;
          border-radius: 50%;
          box-shadow: var(--shadow-sm);
          transition: transform var(--transition-interactive);
          will-change: transform;
        }
      }

      &:hover span {
        background-color: color-mix(in srgb, var(--color-text-muted) 70%, var(--color-text));
      }
    }

    input[type='checkbox'] {
      &:checked + span {
        background-color: var(--color-primary);
        &::before { transform: translateX(18px); }
      }
      &:checked + span:hover { background-color: var(--color-primary-hover); }
    }

    // ── Scroll sensitivity slider ─────────────────────────────────────────
    // --fill is set via :style on the <input> from the Vue scroll computed.
    // The track uses a split gradient: filled = teal, unfilled = surface-dynamic.
    // This avoids a solid-teal bar which reads as "too much green" in light mode.
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
          background: linear-gradient(
            to right,
            var(--color-primary)         0% var(--fill, 50%),
            var(--color-surface-dynamic) var(--fill, 50%) 100%
          );
          border-radius: var(--radius-full);
        }
        &::-moz-range-thumb {
          height: 12px; width: 12px; border-radius: var(--radius-full);
          background: #fff; cursor: pointer; border: none;
          box-shadow: var(--shadow-sm); transition: transform var(--transition-fast);
        }
        &:hover::-moz-range-thumb { transform: scale(1.35); }

        &::-webkit-slider-runnable-track {
          width: 100%; height: 4px; cursor: pointer;
          background: linear-gradient(
            to right,
            var(--color-primary)         0% var(--fill, 50%),
            var(--color-surface-dynamic) var(--fill, 50%) 100%
          );
          border-radius: var(--radius-full);
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

      &:focus { border-color: var(--color-primary); outline: none; }
      &::selection { background: var(--color-primary-highlight); }
      &[disabled] { opacity: 0.5; cursor: not-allowed; }
    }
  }
</style>

<script lang="ts">
  import { Component, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-settings' })
  export default class extends Vue {
    private broadcast_url: string = ''

    get admin()     { return this.$accessor.user.admin }
    get connected() { return this.$accessor.connected }

    get scroll()    { return this.$accessor.settings.scroll.toString() }
    set scroll(v: string) { this.$accessor.settings.setScroll(parseInt(v)) }

    get scroll_invert()    { return this.$accessor.settings.scroll_invert }
    set scroll_invert(v: boolean) { this.$accessor.settings.setInvert(v) }

    get autoplay()    { return this.$accessor.settings.autoplay }
    set autoplay(v: boolean) { this.$accessor.settings.setAutoplay(v) }

    get ignore_emotes()    { return this.$accessor.settings.ignore_emotes }
    set ignore_emotes(v: boolean) { this.$accessor.settings.setIgnore(v) }

    get chat_sound()    { return this.$accessor.settings.chat_sound }
    set chat_sound(v: boolean) { this.$accessor.settings.setSound(v) }

    get keyboard_layouts_list() { return this.$accessor.settings.keyboard_layouts_list }
    get keyboard_layout()       { return this.$accessor.settings.keyboard_layout }
    set keyboard_layout(v: string) {
      this.$accessor.settings.setKeyboardLayout(v)
      this.$accessor.remote.changeKeyboard()
    }

    get broadcast_is_active()   { return this.$accessor.settings.broadcast_is_active }
    get broadcast_url_remote()  { return this.$accessor.settings.broadcast_url }

    @Watch('broadcast_url_remote', { immediate: true })
    onBroadcastUrlChange() { this.broadcast_url = this.broadcast_url_remote }

    logout() { this.$accessor.logout() }
  }
</script>
