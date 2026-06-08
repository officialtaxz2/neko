<template>
  <div class="settings">
    <ul>
      <li>
        <span>{{ $t('setting.scroll') }}</span>
        <label class="slider">
          <input type="range" min="1" max="100" v-model="scroll" />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.scroll_invert') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="scroll_invert" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.autoplay') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="autoplay" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.ignore_emotes') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="ignore_emotes" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.chat_sound') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="chat_sound" />
          <span />
        </label>
      </li>
      <li>
        <span>{{ $t('setting.force_touch') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="force_touch" />
          <span />
        </label>
      </li>
      <li :class="{ 'disabled-setting': !is_touch_device }">
        <span>{{ $t('setting.trackpad_mode') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="trackpad_mode" :disabled="!is_touch_device" />
          <span />
        </label>
      </li>
      <li v-if="trackpad_mode" :class="{ 'disabled-setting': !is_touch_device }">
        <span>{{ $t('setting.trackpad_cursor_hidden') }}</span>
        <label class="switch">
          <input type="checkbox" v-model="trackpad_cursor_hidden" :disabled="!is_touch_device" />
          <span />
        </label>
      </li>

      <li>
        <span>{{ $t('setting.keyboard_layout') }}</span>
        <label class="select">
          <select v-model="keyboard_layout">
            <option v-for="(name, code) in keyboard_layouts_list" :key="code" :value="code">{{ name }}</option>
          </select>
          <span />
        </label>
      </li>
      <li class="broadcast" v-if="admin">
        <div>
          <span>{{ $t('setting.broadcast_title') }}</span>
          <button v-if="!broadcast_is_active" @click.stop.prevent="$accessor.settings.broadcastCreate(broadcast_url)">
            <i class="fas fa-play"></i>
          </button>
          <button v-else @click.stop.prevent="$accessor.settings.broadcastDestroy()" class="btn-red">
            <i class="fas fa-stop"></i>
          </button>
        </div>
        <input
          v-model="broadcast_url"
          :disabled="broadcast_is_active"
          class="input"
          placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
        />
      </li>
      <li v-if="connected">
        <button @click.stop.prevent="logout">{{ $t('logout') }}</button>
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .settings {
    flex: 1;
    display: flex;
    flex-direction: column;
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
      transition: var(--transition-fluid);

      &:hover {
        background-color: rgba(255, 255, 255, 0.2);
      }
    }

    ul {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: 10px 24px;

      li {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: space-between;
        border-bottom: 1px solid oklch(from var(--text-subtle) l c h / 0.05);
        padding: 14px 0;
        white-space: nowrap;

        &:last-child {
          border-bottom: none;
        }

        span {
          margin-right: auto;
          font-size: 14px;
          font-weight: 500;
          color: var(--text-subtle);
        }

        button {
          cursor: pointer;
          border-radius: 8px;
          padding: 8px 16px;
          background: linear-gradient(135deg, var(--color-cyber-mint) 0%, var(--color-cyber-mint-active) 100%);
          color: #050508;
          text-align: center;
          text-transform: uppercase;
          font-weight: 800;
          font-size: 13px;
          letter-spacing: 0.5px;
          transition: var(--transition-fluid);
          border: none;
          display: block;
          width: 100%;
          box-shadow: 0 4px 12px rgba(38, 230, 180, 0.15);

          &:hover {
            transform: translateY(-1.5px);
            box-shadow: 0 6px 18px var(--color-cyber-mint-glow);
            filter: brightness(1.05);
          }

          &:focus-visible {
            box-shadow: 0 0 0 3px var(--bg-obsidian-deep), 0 0 0 6px var(--color-cyber-mint);
          }

          &:active {
            transform: translateY(0);
          }
        }

        .switch {
          justify-self: flex-end;
          position: relative;
          width: 44px;
          height: 24px;

          input {
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
              position: absolute;
              content: '';
              height: 16px;
              width: 16px;
              left: 3px;
              bottom: 3px;
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
          }

          &:checked + span:before {
            transform: translateX(20px);
            background-color: #050508;
          }
        }

        .slider {
          white-space: nowrap;
          max-width: 140px;
          width: 100%;
          display: flex;
          align-items: center;

          input[type='range'] {
            display: inline-block;
            background: rgba(255, 255, 255, 0.12);
            appearance: none;
            height: 4px;
            width: 100%;
            border-radius: 4px;
            outline: none;
            transition: var(--transition-fluid);

            &::-moz-range-thumb {
              height: 12px;
              width: 12px;
              border-radius: 50%;
              background: var(--color-cyber-mint);
              cursor: pointer;
              transition: var(--transition-fluid);
              box-shadow: 0 0 8px var(--color-cyber-mint-glow);
              border: none;

              &:hover {
                transform: scale(1.3);
                background: #fff;
                box-shadow: 0 0 12px var(--color-cyber-mint);
              }
            }

            &::-moz-range-track {
              width: 100%;
              height: 4px;
              background: transparent;
              border-radius: 2px;
            }

            &::-webkit-slider-thumb {
              appearance: none;
              height: 12px;
              width: 12px;
              border-radius: 50%;
              background: var(--color-cyber-mint);
              cursor: pointer;
              margin-top: -4px;
              transition: var(--transition-fluid);
              box-shadow: 0 0 8px var(--color-cyber-mint-glow);

              &:hover {
                transform: scale(1.3);
                background: #fff;
                box-shadow: 0 0 12px var(--color-cyber-mint);
              }
            }

            &::-webkit-slider-runnable-track {
              width: 100%;
              height: 4px;
              background: transparent;
              border-radius: 2px;
            }

            &:hover {
              background: rgba(255, 255, 255, 0.22);
              height: 5px;
            }
          }
        }

        .select {
          width: 100%;
          max-width: 154px;
          position: relative;

          select {
            -webkit-appearance: none;
            -moz-appearance: none;
            appearance: none;
            display: block;
            width: 100%;
            height: 34px;
            text-align: right;
            padding: 0 28px 0 12px;
            line-height: 34px;
            font-weight: 500;
            font-size: 13px;
            text-overflow: ellipsis;
            border: 1px solid var(--glass-border);
            border-radius: 8px;
            color: var(--text-pure);
            background-color: var(--bg-obsidian-floating);
            cursor: pointer;
            transition: var(--transition-fluid);
            direction: rtl;

            &:hover, &:focus {
              border-color: var(--color-cyber-mint);
              box-shadow: 0 0 8px var(--color-cyber-mint-glow);
              outline: none;
            }

            option {
              background-color: var(--bg-obsidian-floating);
              color: var(--text-pure);
              text-align: left;
              direction: ltr;
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
            transition: var(--transition-fluid);
          }

          &:focus-within span {
            border-color: var(--color-cyber-mint);
          }
        }

        .input {
          display: block;
          height: 34px;
          text-align: right;
          padding: 0 12px;
          margin-left: 10px;
          line-height: 34px;
          text-overflow: ellipsis;
          border: 1px solid var(--glass-border);
          border-radius: 8px;
          color: var(--text-pure);
          background-color: var(--bg-obsidian-floating);
          font-weight: 400;
          font-size: 13px;
          transition: var(--transition-fluid);
          user-select: auto;

          &:focus {
            border-color: var(--color-cyber-mint);
            box-shadow: 0 0 8px var(--color-cyber-mint-glow);
            outline: none;
          }

          &::selection {
            background: rgba(38, 230, 180, 0.35);
          }

          &[disabled] {
            background: rgba(255, 255, 255, 0.02);
            border-color: transparent;
            color: var(--text-muted-ok);
          }
        }

        &.broadcast {
          display: flex;
          flex-direction: column;
          align-items: stretch;

          div {
            margin-bottom: 12px;
            display: flex;
            align-items: center;
            justify-content: space-between;

            button {
              flex-shrink: 1;
              width: auto !important;
              margin: 0;
              padding: 0 16px;
              height: 34px;
              line-height: 34px;

              &.btn-red {
                background: linear-gradient(135deg, #a62626 0%, #801d1d 100%);
                color: #fff;
                box-shadow: 0 4px 12px rgba(166, 38, 38, 0.2);

                &:hover {
                  box-shadow: 0 6px 18px rgba(166, 38, 38, 0.35);
                }
              }
            }
          }

          .input {
            text-align: left;
            width: auto !important;
            margin: 0;
          }
        }

        &.disabled-setting {
          opacity: 0.4;

          .switch span {
            cursor: not-allowed;
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Watch, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-settings' })
  export default class NekoSettings extends Vue {
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
