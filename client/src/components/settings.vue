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

    ul {
      flex: 1;
      display: flex;
      flex-direction: column;
      padding: var(--space-1) var(--space-5);

      li {
        min-height: 44px;
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: space-between;
        border-bottom: 1px solid var(--color-border);
        gap: var(--space-3);
        white-space: nowrap;

        &:last-child {
          border-bottom: none;
        }

        > span {
          margin-right: auto;
          color: var(--color-text);
          font-size: var(--text-sm);
        }

        button {
          min-height: 44px;
          cursor: pointer;
          border-radius: var(--radius-md);
          padding: var(--space-1) var(--space-3);
          background: var(--color-primary);
          color: var(--color-text-inverse);
          text-align: center;
          text-transform: uppercase;
          font-weight: 700;
          font-size: var(--text-sm);
          border: none;
          display: block;
          width: 100%;
          transition:
            background-color var(--transition-interactive),
            transform        var(--transition-fast);

          &:hover {
            background: var(--color-primary-hover);
          }

          &:active {
            background: var(--color-primary-active);
            transform: scale(0.97);
          }

          &.btn-red {
            background: var(--color-error);

            &:hover {
              background: var(--color-error-hover);
            }
          }
        }

        // ------------------------------------------------------------------
        // Custom Toggle Switch — 44×44px touch target, 42×24px visible pill
        //
        // Off-state track uses --color-border (84% L in light / 23% L in dark)
        // so the pill is clearly visible on --color-bg/--color-surface in
        // both modes. Previously used --color-surface-offset (94% L in light)
        // which was near-invisible on the white settings panel background.
        // ------------------------------------------------------------------
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
            // 84% L in light (clearly visible gray), 23% L in dark (visible)
            background-color: var(--color-border);
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

          // Hover darkens toward text-faint (64% L light / 34% L dark)
          // so direction is always: hover = darker than base
          &:hover span {
            background-color: var(--color-text-faint);
          }
        }

        input[type='checkbox'] {
          &:checked + span {
            background-color: var(--color-primary);

            &::before {
              transform: translateX(18px);
            }
          }

          &:checked + span:hover {
            background-color: var(--color-primary-hover);
          }
        }

        // ------------------------------------------------------------------
        // Scroll speed slider
        // ------------------------------------------------------------------
        .slider {
          flex-shrink: 0;
          white-space: nowrap;
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
              width: 100%;
              height: 4px;
              cursor: pointer;
              background: var(--color-primary);
              border-radius: var(--radius-full);
            }

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

            &::-webkit-slider-runnable-track {
              width: 100%;
              height: 4px;
              cursor: pointer;
              background: var(--color-primary);
              border-radius: var(--radius-full);
            }

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
        // Keyboard layout select
        // ------------------------------------------------------------------
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

        // ------------------------------------------------------------------
        // Broadcast URL input
        // ------------------------------------------------------------------
        .input {
          display: block;
          height: 30px;
          text-align: right;
          padding: 0 var(--space-2);
          margin-left: var(--space-2);
          line-height: 30px;
          text-overflow: ellipsis;
          border: 1px solid transparent;
          border-radius: var(--radius-md);
          color: var(--color-text);
          background-color: var(--color-bg);
          font-weight: 300;
          user-select: auto;
          transition:
            border-color     var(--transition-interactive),
            background-color var(--transition-interactive);

          &:focus {
            border-color: var(--color-primary);
            outline: none;
          }

          &::selection {
            background: var(--color-primary-highlight);
          }

          &[disabled] {
            background: none;
            opacity: 0.5;
            cursor: not-allowed;
          }
        }

        // ------------------------------------------------------------------
        // Broadcast section layout
        // ------------------------------------------------------------------
        &.broadcast {
          flex-direction: column;
          align-items: stretch;
          min-height: unset;

          div {
            margin-bottom: var(--space-2);
            display: flex;
            justify-content: space-between;

            button {
              flex-shrink: 1;
              width: auto !important;
              margin: 0;
              padding: 0 var(--space-2);
            }
          }

          .input {
            text-align: left;
            width: auto !important;
            margin: 0;
          }
        }
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
