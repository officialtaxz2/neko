<template>
  <div class="settings">
    <!-- Stream -->
    <section aria-labelledby="settings-stream">
      <h3 id="settings-stream">{{ $t('setting.group_stream') }}</h3>
      <div class="media-backend" role="radiogroup" :aria-label="$t('setting.media_backend')">
        <label
          v-for="option in media_options"
          :key="option.value"
          :class="['media-option', { selected: media_backend === option.value, disabled: option.disabled }]"
        >
          <input
            type="radio"
            name="media-backend"
            :value="option.value"
            :checked="media_backend === option.value"
            :disabled="option.disabled"
            @change="media_backend = option.value"
          />
          <span class="radio" aria-hidden="true" />
          <span class="text">
            <span class="name">
              {{ option.label }}
              <em v-if="option.value === 'webrtc'">{{ $t('setting.media_default') }}</em>
            </span>
            <small>{{ option.disabled ? $t('media.unavailable') : option.description }}</small>
          </span>
        </label>
      </div>
      <div class="media-backend-status" role="status">
        <i :class="['fas', media_backend_status_icon]" aria-hidden="true" />
        <small>{{ media_backend_status }}</small>
      </div>
      <div v-if="media_backend_override" class="media-backend-override">
        <small class="note">
          {{ $t(media_backend_invalid_override ? 'setting.media_backend_invalid_override' : 'setting.media_backend_override') }}
        </small>
        <button type="button" class="btn secondary small" @click.stop.prevent="$client.clearMediaBackendOverride()">
          {{ $t('setting.media_backend_clear_override') }}
        </button>
      </div>
      <small v-if="$accessor.hls.selected" class="note">{{ $t('media.hls_limitations') }}</small>
      <ul>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.autoplay') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="autoplay" />
              <span />
            </span>
          </label>
        </li>
      </ul>
    </section>

    <!-- Input -->
    <section aria-labelledby="settings-input">
      <h3 id="settings-input">{{ $t('setting.group_input') }}</h3>
      <ul>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.scroll') }}</span>
            <span class="slider">
              <input type="range" min="1" max="100" v-model="scroll" />
            </span>
          </label>
        </li>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.scroll_invert') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="scroll_invert" />
              <span />
            </span>
          </label>
        </li>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.keyboard_layout') }}</span>
            <span class="select">
              <select v-model="keyboard_layout">
                <option v-for="(name, code) in keyboard_layouts_list" :key="code" :value="code">{{ name }}</option>
              </select>
              <span />
            </span>
          </label>
        </li>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.force_touch') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="force_touch" />
              <span />
            </span>
          </label>
        </li>
        <template v-if="is_touch_device">
          <li>
            <label class="row">
              <span class="label">{{ $t('setting.trackpad_mode') }}</span>
              <span class="switch">
                <input type="checkbox" v-model="trackpad_mode" />
                <span />
              </span>
            </label>
          </li>
          <li v-if="trackpad_mode">
            <label class="row">
              <span class="label">{{ $t('setting.trackpad_cursor_hidden') }}</span>
              <span class="switch">
                <input type="checkbox" v-model="trackpad_cursor_hidden" />
                <span />
              </span>
            </label>
          </li>
        </template>
      </ul>
    </section>

    <!-- Chat -->
    <section aria-labelledby="settings-chat">
      <h3 id="settings-chat">{{ $t('setting.group_chat') }}</h3>
      <ul>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.chat_sound') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="chat_sound" />
              <span />
            </span>
          </label>
        </li>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.ignore_emotes') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="ignore_emotes" />
              <span />
            </span>
          </label>
        </li>
        <li v-if="$accessor.openinapp.enabled">
          <label class="row">
            <span class="label">{{ $t('setting.open_links_in_app') }}</span>
            <span class="switch">
              <input type="checkbox" v-model="links_in_app" />
              <span />
            </span>
          </label>
        </li>
      </ul>
    </section>

    <!-- Room (admin only) -->
    <section v-if="admin" aria-labelledby="settings-room">
      <h3 id="settings-room">
        {{ $t('setting.group_room') }}
        <span class="badge">{{ $t('setting.admin_badge') }}</span>
      </h3>
      <ul>
        <li v-for="resource in lock_resources" :key="resource">
          <label class="row">
            <span class="label">
              {{ $t(`setting.lock_${resource}`) }}
              <small>{{ $t(`setting.lock_${resource}_hint`) }}</small>
            </span>
            <span class="switch">
              <input type="checkbox" :checked="isLocked(resource)" @change="toggleLock(resource)" />
              <span />
            </span>
          </label>
        </li>
        <li class="broadcast">
          <div class="row">
            <span class="label">
              {{ $t('setting.broadcast_title') }}
              <small>{{ $t('setting.broadcast_hint') }}</small>
            </span>
            <button
              v-if="!broadcast_is_active"
              type="button"
              class="btn primary small"
              @click.stop.prevent="$accessor.settings.broadcastCreate(broadcast_url)"
            >
              <i class="fas fa-play" aria-hidden="true" />
              {{ $t('setting.broadcast_start') }}
            </button>
            <button v-else type="button" class="btn danger small" @click.stop.prevent="$accessor.settings.broadcastDestroy()">
              <i class="fas fa-stop" aria-hidden="true" />
              {{ $t('setting.broadcast_stop') }}
            </button>
          </div>
          <input
            v-model="broadcast_url"
            :disabled="broadcast_is_active"
            class="input"
            :aria-label="$t('setting.broadcast_title')"
            placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
          />
        </li>
      </ul>
    </section>

    <!-- Account -->
    <section aria-labelledby="settings-account">
      <h3 id="settings-account">{{ $t('setting.group_account') }}</h3>
      <ul>
        <li>
          <label class="row">
            <span class="label">{{ $t('setting.language') }}</span>
            <span class="select">
              <select v-model="$i18n.locale">
                <option v-for="lang in langs" :key="lang" :value="lang">{{ languageName(lang) }}</option>
              </select>
              <span />
            </span>
          </label>
        </li>
        <li v-if="connected" class="logout">
          <button type="button" class="btn danger-outline" @click.stop.prevent="logout">
            <i class="fas fa-sign-out-alt" aria-hidden="true" />
            {{ $t('logout') }}
          </button>
        </li>
      </ul>
    </section>
  </div>
</template>

<style lang="scss" scoped>
  .settings {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 6px 20px 20px;
    overflow-y: auto;
    overflow-x: hidden;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.1) transparent;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-track {
      background-color: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background-color: rgba(255, 255, 255, 0.1);
      border-radius: 3px;
    }
  }

  section {
    padding: 12px 0 4px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);

    &:last-child {
      border-bottom: none;
    }
  }

  h3 {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 8px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted-ok);

    .badge {
      padding: 2px 7px;
      border-radius: 999px;
      font-size: 10px;
      letter-spacing: 0.04em;
      color: #ffb86c;
      background: rgba(255, 184, 108, 0.12);
      border: 1px solid rgba(255, 184, 108, 0.3);
    }
  }

  ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li + li {
    border-top: 1px solid rgba(255, 255, 255, 0.04);
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 48px;
    padding: 6px 0;
    cursor: pointer;
  }

  .label {
    display: flex;
    flex-direction: column;
    min-width: 0;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-subtle);
    line-height: 1.3;

    small {
      margin-top: 2px;
      font-size: 12px;
      font-weight: 400;
      color: var(--text-muted-ok);
    }
  }

  .note {
    display: block;
    margin: 4px 0 6px;
    font-size: 12px;
    line-height: 1.4;
    color: var(--text-muted-ok);
  }

  // Stream choice
  .media-backend {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .media-option {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: var(--radius-inner, 10px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.02);
    cursor: pointer;
    transition: var(--transition-fluid);

    &:hover {
      background: rgba(255, 255, 255, 0.05);
    }

    input {
      position: absolute;
      opacity: 0;
      width: 0;
      height: 0;
    }

    .radio {
      flex: none;
      width: 18px;
      height: 18px;
      border-radius: 50%;
      border: 2px solid rgba(255, 255, 255, 0.3);
      transition: var(--transition-fluid);
    }

    .text {
      display: flex;
      flex-direction: column;
      min-width: 0;
      line-height: 1.3;

      .name {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-pure);

        em {
          margin-left: 6px;
          font-style: normal;
          font-size: 11px;
          font-weight: 500;
          color: var(--text-muted-ok);
        }
      }

      small {
        font-size: 12px;
        color: var(--text-muted-ok);
      }
    }

    &.selected {
      border-color: rgba(38, 230, 180, 0.45);
      background: rgba(38, 230, 180, 0.08);

      .radio {
        border-color: var(--color-cyber-mint);
        box-shadow: inset 0 0 0 3px var(--bg-obsidian-deep, #0a0b10);
        background: var(--color-cyber-mint);
      }
    }

    &.disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    &:focus-within {
      outline: 2px solid var(--color-cyber-mint);
      outline-offset: 2px;
    }
  }

  .media-backend-status {
    display: flex;
    align-items: center;
    gap: 7px;
    margin: 10px 2px 4px;
    color: var(--text-subtle);

    i {
      flex: none;
      color: var(--color-cyber-mint);
      font-size: 8px;
    }

    small {
      line-height: 1.35;
    }
  }

  .media-backend-override {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  // Buttons
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    height: 40px;
    padding: 0 16px;
    border-radius: var(--radius-button, 8px);
    border: 1px solid transparent;
    font-family: inherit;
    font-size: 13.5px;
    font-weight: 600;
    cursor: pointer;
    transition: var(--transition-fluid);

    &.small {
      flex: none;
      height: 34px;
      padding: 0 14px;
      font-size: 13px;
    }

    &.primary {
      background: var(--color-cyber-mint);
      color: #06281f;

      &:hover {
        background: var(--color-cyber-mint-active);
      }
    }

    &.secondary {
      background: rgba(255, 255, 255, 0.06);
      border-color: var(--glass-border);
      color: var(--text-pure);

      &:hover {
        background: rgba(255, 255, 255, 0.1);
      }
    }

    &.danger {
      background: #a62626;
      color: #fff;

      &:hover {
        background: #bf2d2d;
      }
    }

    &.danger-outline {
      width: 100%;
      background: transparent;
      border-color: rgba(255, 90, 102, 0.45);
      color: #ff7b85;

      &:hover {
        background: rgba(255, 90, 102, 0.1);
      }
    }
  }

  .logout {
    padding: 12px 0 4px;
  }

  // Switch
  .switch {
    flex: none;
    position: relative;
    width: 44px;
    height: 24px;

    input {
      position: absolute;
      opacity: 0;
      width: 0;
      height: 0;
    }

    span {
      position: absolute;
      inset: 0;
      cursor: pointer;
      background-color: rgba(255, 255, 255, 0.05);
      border: 1px solid var(--glass-border);
      border-radius: 34px;
      transition: var(--transition-fluid);

      &:before {
        position: absolute;
        content: '';
        height: 16px;
        width: 16px;
        left: 3px;
        bottom: 3px;
        background-color: var(--text-pure);
        border-radius: 50%;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        transition: var(--transition-fluid);
      }
    }

    input:checked + span {
      background-color: var(--color-cyber-mint);
      border-color: transparent;
      box-shadow: 0 0 10px var(--color-cyber-mint-glow);

      &:before {
        transform: translateX(20px);
        background-color: #050508;
      }
    }

    input:focus-visible + span {
      outline: 2px solid var(--color-cyber-mint);
      outline-offset: 2px;
    }
  }

  // Range slider
  .slider {
    flex: none;
    width: 140px;
    display: flex;
    align-items: center;

    input[type='range'] {
      appearance: none;
      width: 100%;
      height: 4px;
      border-radius: 4px;
      background: rgba(255, 255, 255, 0.12);
      cursor: pointer;

      &::-moz-range-thumb {
        height: 14px;
        width: 14px;
        border: none;
        border-radius: 50%;
        background: var(--color-cyber-mint);
        box-shadow: 0 0 8px var(--color-cyber-mint-glow);
      }

      &::-webkit-slider-thumb {
        appearance: none;
        height: 14px;
        width: 14px;
        border-radius: 50%;
        background: var(--color-cyber-mint);
        box-shadow: 0 0 8px var(--color-cyber-mint-glow);
      }
    }
  }

  // Select
  .select {
    flex: none;
    position: relative;
    width: 160px;

    select {
      appearance: none;
      display: block;
      width: 100%;
      height: 36px;
      padding: 0 28px 0 12px;
      font-family: inherit;
      font-weight: 500;
      font-size: 13px;
      text-overflow: ellipsis;
      border: 1px solid var(--glass-border);
      border-radius: 8px;
      color: var(--text-pure);
      background-color: var(--bg-obsidian-floating);
      cursor: pointer;
      transition: var(--transition-fluid);

      &:hover,
      &:focus {
        border-color: var(--color-cyber-mint);
        outline: none;
      }

      option {
        background-color: var(--bg-obsidian-floating);
        color: var(--text-pure);
      }
    }

    span {
      position: absolute;
      right: 12px;
      top: 50%;
      width: 7px;
      height: 7px;
      border-right: 1.5px solid var(--text-subtle);
      border-bottom: 1.5px solid var(--text-subtle);
      transform: translateY(-65%) rotate(45deg);
      pointer-events: none;
    }
  }

  // Broadcast
  .broadcast {
    padding-bottom: 10px;

    .row {
      cursor: default;
    }

    .input {
      display: block;
      width: 100%;
      box-sizing: border-box;
      height: 36px;
      padding: 0 12px;
      border: 1px solid var(--glass-border);
      border-radius: 8px;
      color: var(--text-pure);
      background-color: var(--bg-obsidian-floating);
      font-family: inherit;
      font-size: 13px;
      user-select: auto;

      &:focus {
        border-color: var(--color-cyber-mint);
        outline: none;
      }

      &[disabled] {
        background: rgba(255, 255, 255, 0.02);
        border-color: transparent;
        color: var(--text-muted-ok);
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Watch, Vue } from 'vue-property-decorator'
  import { messages } from '~/locale'
  import { AdminLockResource } from '~/neko/messages'

  // Native language names, so people find their own language in any UI locale.
  const LANGUAGE_NAMES: Record<string, string> = {
    en: 'English',
    es: 'Español',
    sk: 'Slovenčina',
    sv: 'Svenska',
    nb: 'Norsk bokmål',
    fr: 'Français',
    de: 'Deutsch',
    ko: '한국어',
    fi: 'Suomi',
    ru: 'Русский',
    cn: '简体中文',
    tw: '繁體中文',
    ja: '日本語',
    id: 'Bahasa Indonesia',
    pl: 'Polski',
  }

  interface MediaOption {
    value: string
    label: string
    description: string
    disabled: boolean
  }

  @Component({ name: 'neko-settings' })
  export default class NekoSettings extends Vue {
    private broadcast_url: string = ''

    get admin() {
      return this.$accessor.user.admin
    }

    get connected() {
      return this.$accessor.connected
    }

    get langs() {
      return Object.keys(messages)
    }

    languageName(lang: string) {
      return LANGUAGE_NAMES[lang] || lang
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

    get media_backend() {
      return this.$accessor.settings.media_backend
    }

    set media_backend(value: string) {
      if (!this.$client.canSelectMediaBackend(value)) return
      this.$accessor.settings.setMediaBackend(value)
      this.$client.changeMediaBackend(value)
    }

    get hls_modes() {
      return this.$accessor.hls.availableModes
    }

    // WebRTC first, then the advertised HLS modes (LL-HLS before HLS). A stored
    // HLS preference the server no longer advertises stays visible, disabled.
    get media_options(): MediaOption[] {
      const options: MediaOption[] = [
        {
          value: 'webrtc',
          label: this.$t('setting.media_backend_webrtc') as string,
          description: this.$t('setting.media_desc_webrtc') as string,
          disabled: false,
        },
      ]
      for (const mode of ['ll-hls', 'hls']) {
        const available = (this.hls_modes as string[]).includes(mode)
        if (!available && this.media_backend !== mode) continue
        options.push({
          value: mode,
          label: mode === 'll-hls' ? 'LL-HLS' : 'HLS',
          description: this.$t(mode === 'll-hls' ? 'setting.media_desc_ll_hls' : 'setting.media_desc_hls') as string,
          disabled: !available,
        })
      }
      return options
    }

    get media_backend_override() {
      return this.$client.mediaBackendOverridden
    }

    get media_backend_invalid_override() {
      return this.$client.mediaBackendOverrideInvalid
    }

    get media_backend_status_icon() {
      if (this.$accessor.hls.selected) {
        const status = this.$accessor.hls.status
        return status === 'streaming' || status === 'paused'
          ? 'fa-circle'
          : status === 'terminal'
            ? 'fa-exclamation-circle'
            : 'fa-circle-notch fa-spin'
      }
      return 'fa-circle'
    }

    get media_backend_status() {
      if (this.$accessor.hls.selected) {
        const hls = this.$accessor.hls
        return `${hls.selected === 'll-hls' ? 'LL-HLS' : 'HLS'} — ${this.$t(`media.status.${hls.status}`)}${hls.detail ? `: ${hls.detail}` : ''}`
      }
      return `${this.$t('setting.media_backend_current')}: ${this.$t('setting.media_backend_webrtc')}`
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

    get force_touch() {
      return this.$accessor.settings.force_touch
    }

    set force_touch(value: boolean) {
      this.$accessor.settings.setForceTouch(value)
    }

    get is_touch_device() {
      if (typeof window === 'undefined') return false
      if (this.$accessor.settings.force_touch) return true
      return (
        'ontouchstart' in window ||
        navigator.maxTouchPoints > 0 ||
        /Mobi|Android|iPhone|iPad|iPod|Windows Phone/i.test(navigator.userAgent) ||
        (navigator.maxTouchPoints > 1 && /Macintosh|MacIntel/i.test(navigator.userAgent || navigator.platform)) ||
        (window.matchMedia && window.matchMedia('(pointer: coarse)').matches) ||
        (window.matchMedia && window.matchMedia('(any-pointer: coarse)').matches) ||
        window.innerWidth <= 1024
      )
    }

    get trackpad_mode() {
      return this.$accessor.settings.trackpad_mode
    }

    set trackpad_mode(value: boolean) {
      this.$accessor.settings.setTrackpadMode(value)
    }

    get trackpad_cursor_hidden() {
      return this.$accessor.settings.trackpad_cursor_hidden
    }

    set trackpad_cursor_hidden(value: boolean) {
      this.$accessor.settings.setTrackpadCursorHidden(value)
    }

    get links_in_app() {
      return this.$accessor.settings.links_in_app
    }

    set links_in_app(value: boolean) {
      this.$accessor.settings.setLinksInApp(value)
    }

    get keyboard_layouts_list() {
      return this.$accessor.settings.keyboard_layouts_list
    }

    get keyboard_layout() {
      return this.$accessor.settings.keyboard_layout
    }

    set keyboard_layout(value: string) {
      this.$accessor.settings.setKeyboardLayout(value)
      this.$accessor.remote.changeKeyboard()
    }

    get lock_resources(): AdminLockResource[] {
      const resources: AdminLockResource[] = ['control', 'login']
      if (this.$accessor.remote.fileTransfer) resources.push('file_transfer')
      return resources
    }

    isLocked(resource: AdminLockResource): boolean {
      return this.$accessor.isLocked(resource)
    }

    toggleLock(resource: AdminLockResource) {
      this.$accessor.toggleLock(resource)
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

    async logout() {
      const result = await this.$swal({
        title: this.$t('setting.logout_confirm_title') as string,
        text: this.$t('setting.logout_confirm_text') as string,
        icon: 'question',
        showCancelButton: true,
        confirmButtonText: this.$t('logout') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })
      if (result.isConfirmed) {
        this.$accessor.logout()
      }
    }
  }
</script>
