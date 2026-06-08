<template>
  <div v-if="isTouchDevice && hosting" class="keyboard-helper-wrapper">
    <!-- Floating Modifier Bar -->
    <div
      v-if="expanded"
      class="modifier-bar"
      :style="barStyle"
      @click.stop=""
      @touchstart.stop=""
      @touchmove.stop=""
      @touchend.stop=""
    >
      <div class="bar-header">
        <span class="bar-title">
          <i class="fas fa-sliders-h icon" />
          {{ $t('clipboard_manager.keyboard_helper_title') }}
        </span>
        <div class="header-actions">
          <button class="reset-btn" @click.stop.prevent="resetAll" v-tooltip="{ content: 'Reset keys', placement: 'top' }">
            <i class="fas fa-undo-alt" />
          </button>
          <button class="close-btn" @click.stop.prevent="toggleExpanded">
            <i class="fas fa-times" />
          </button>
        </div>
      </div>

      <div class="bar-content">
        <!-- FIRST ROW: Modifiers & Commands -->
        <div class="button-row">
          <button
            :class="['kbd-btn modifier', ctrlActive ? 'active' : '']"
            @click.stop.prevent="toggleModifier('ctrl')"
          >
            CTRL
          </button>
          <button
            :class="['kbd-btn modifier', altActive ? 'active' : '']"
            @click.stop.prevent="toggleModifier('alt')"
          >
            ALT
          </button>
          <button
            :class="['kbd-btn modifier', shiftActive ? 'active' : '']"
            @click.stop.prevent="toggleModifier('shift')"
          >
            SHIFT
          </button>
          <button
            :class="['kbd-btn modifier', superActive ? 'active' : '']"
            @click.stop.prevent="toggleModifier('super')"
          >
            WIN
          </button>
          <div class="divider" />
          <button class="kbd-btn action" @click.stop.prevent="pressKey(XK_Escape)">
            ESC
          </button>
          <button class="kbd-btn action" @click.stop.prevent="pressKey(XK_Tab)">
            TAB
          </button>
        </div>

        <!-- SECOND ROW: Macros & Arrows -->
        <div class="button-row">
          <button class="kbd-btn macro" @click.stop.prevent="triggerMacro('all')">
            <i class="fas fa-border-all icon-macro" />ALL
          </button>
          <button class="kbd-btn macro" @click.stop.prevent="triggerMacro('copy')">
            <i class="fas fa-copy icon-macro" />COPY
          </button>
          <button class="kbd-btn macro" @click.stop.prevent="triggerMacro('paste')">
            <i class="fas fa-paste icon-macro" />PASTE
          </button>
          <button class="kbd-btn macro" @click.stop.prevent="triggerMacro('undo')">
            <i class="fas fa-undo icon-macro" />UNDO
          </button>

          <div class="divider" />

          <!-- Arrow pad -->
          <div class="arrow-container">
            <button class="kbd-btn arrow" @click.stop.prevent="pressKey(XK_Left)">
              <i class="fas fa-arrow-left" />
            </button>
            <div class="arrow-vertical">
              <button class="kbd-btn arrow top" @click.stop.prevent="pressKey(XK_Up)">
                <i class="fas fa-arrow-up" />
              </button>
              <button class="kbd-btn arrow bottom" @click.stop.prevent="pressKey(XK_Down)">
                <i class="fas fa-arrow-down" />
              </button>
            </div>
            <button class="kbd-btn arrow" @click.stop.prevent="pressKey(XK_Right)">
              <i class="fas fa-arrow-right" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .keyboard-helper-wrapper {
    position: relative;
    user-select: none;
    -webkit-user-select: none;

    .modifier-bar {
      background-color: rgba(18, 21, 30, 0.9);
      backdrop-filter: blur(20px);
      -webkit-backdrop-filter: blur(20px);
      border-top: 1px solid var(--glass-border);
      padding: 10px 14px;
      box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.6);
      display: flex;
      flex-direction: column;
      gap: 8px;

      .bar-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-bottom: 1px solid rgba(255, 255, 255, 0.05);
        padding-bottom: 6px;

        .bar-title {
          font-weight: 600;
          font-size: 11px;
          color: var(--text-pure);
          text-transform: uppercase;
          letter-spacing: 0.5px;
          display: flex;
          align-items: center;
          gap: 6px;

          .icon {
            color: var(--color-cyber-mint);
          }
        }

        .header-actions {
          display: flex;
          gap: 12px;
          align-items: center;

          button {
            background: none;
            border: none;
            color: var(--text-muted-ok);
            cursor: pointer;
            font-size: 12px;
            transition: var(--transition-fluid);
            padding: 2px 6px;

            &:hover {
              color: var(--text-pure);
            }
          }

          .reset-btn:hover {
            color: var(--color-cyber-mint);
          }
        }
      }

      .bar-content {
        display: flex;
        flex-direction: column;
        gap: 6px;

        .button-row {
          display: flex;
          align-items: center;
          gap: 6px;
          width: 100%;
          overflow-x: auto;
          scrollbar-width: none;
          padding: 2px 0;

          &::-webkit-scrollbar {
            display: none;
          }

          .divider {
            width: 1px;
            height: 20px;
            background: rgba(255, 255, 255, 0.1);
            align-self: center;
            flex-shrink: 0;
          }
        }
      }

      .kbd-btn {
        background: rgba(255, 255, 255, 0.03);
        border: 1px solid var(--glass-border);
        color: var(--text-subtle);
        border-radius: var(--radius-button);
        cursor: pointer;
        font-family: inherit;
        font-size: 10px;
        font-weight: 600;
        height: 32px;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        transition: var(--transition-fluid);
        min-width: 44px;
        padding: 0 8px;

        &:active {
          transform: scale(0.95);
        }

        &.modifier {
          &.active {
            background: var(--color-cyber-mint);
            border-color: var(--color-cyber-mint-active);
            color: #000;
            box-shadow: 0 0 10px var(--color-cyber-mint-glow);
          }
        }

        &.action {
          background: rgba(255, 255, 255, 0.05);
          color: var(--text-pure);

          &:hover {
            border-color: var(--text-pure);
          }
        }

        &.macro {
          background: rgba(38, 230, 180, 0.04);
          border-color: rgba(38, 230, 180, 0.15);
          color: var(--color-cyber-mint);
          display: flex;
          gap: 4px;

          .icon-macro {
            font-size: 8px;
          }

          &:hover {
            background: rgba(38, 230, 180, 0.1);
            border-color: var(--color-cyber-mint);
          }
        }
      }

      .arrow-container {
        display: flex;
        align-items: center;
        gap: 4px;
        margin-left: auto;
        flex-shrink: 0;

        .arrow-vertical {
          display: flex;
          flex-direction: column;
          gap: 2px;

          .arrow {
            height: 16px;
            font-size: 8px;
            width: 32px;
            min-width: 32px;
            border-radius: 4px;

            &.top {
              border-bottom-left-radius: 0;
              border-bottom-right-radius: 0;
            }
            &.bottom {
              border-top-left-radius: 0;
              border-top-right-radius: 0;
            }
          }
        }

        .arrow {
          width: 32px;
          min-width: 32px;
          height: 32px;
          font-size: 10px;
          padding: 0;
          display: flex;
          align-items: center;
          justify-content: center;
        }
      }
    }
  }

  @keyframes kbd-pulse {
    0% {
      transform: scale(0.95);
      opacity: 0.5;
    }
    100% {
      transform: scale(1.4);
      opacity: 0;
    }
  }
</style>

<script lang="ts">
  import { Component, Prop, Vue } from 'vue-property-decorator'
  import { GuacamoleKeyboardInterface } from '~/utils/guacamole-keyboard'

  @Component({
    name: 'keyboard-helper',
  })
  export default class KeyboardHelper extends Vue {
    @Prop({ required: true }) readonly keyboard!: GuacamoleKeyboardInterface

    /* Keyboard keysym map constants */
    XK_Control_L = 0xffe3
    XK_Alt_L = 0xffe9
    XK_Shift_L = 0xffe1
    XK_Super_L = 0xffeb
    XK_Escape = 0xff1b
    XK_Tab = 0xff09
    
    XK_Left = 0xff51
    XK_Up = 0xff52
    XK_Right = 0xff53
    XK_Down = 0xff54

    /* Local states */
    private expanded = false
    private ctrlActive = false
    private altActive = false
    private shiftActive = false
    private superActive = false

    private barStyle: Record<string, string> = {
      position: 'fixed',
      left: '0px',
      bottom: '0px',
      width: '100%',
      zIndex: '999999',
      boxSizing: 'border-box',
    }

    get isTouchDevice() {
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

    get hosting() {
      return this.$accessor.remote.hosting
    }

    mounted() {
      this.attachListeners()
      this.updatePosition()
    }

    beforeDestroy() {
      this.detachListeners()
      this.resetAll()
    }

    attachListeners() {
      const viewport = window.visualViewport
      if (viewport) {
        viewport.addEventListener('resize', this.updatePosition)
        viewport.addEventListener('scroll', this.updatePosition)
      }
    }

    detachListeners() {
      const viewport = window.visualViewport
      if (viewport) {
        viewport.removeEventListener('resize', this.updatePosition)
        viewport.removeEventListener('scroll', this.updatePosition)
      }
    }

    toggleExpanded() {
      this.expanded = !this.expanded
      if (this.expanded) {
        this.$nextTick(() => {
          this.updatePosition()
        })
      }
    }

    updatePosition() {
      const viewport = window.visualViewport
      const isIframe = window.self !== window.top
      if (!viewport || isIframe || !this.isTouchDevice) {
        this.barStyle = {
          position: 'fixed',
          left: '0px',
          bottom: '0px',
          width: '100%',
          maxWidth: '100vw',
          zIndex: '999999',
          boxSizing: 'border-box',
        }
        return
      }

      const innerHeight = window.innerHeight
      const offsetTop = viewport.offsetTop
      const height = viewport.height
      const offsetLeft = viewport.offsetLeft
      const width = viewport.width

      // On cellular keyboard open, viewport height diminishes. Keyboard height is innerHeight - bottom of viewport.
      const keyboardHeight = innerHeight - (offsetTop + height)

      this.barStyle = {
        position: 'fixed',
        left: `${offsetLeft}px`,
        bottom: `${Math.max(0, keyboardHeight)}px`,
        width: `${width}px`,
        maxWidth: '100vw',
        zIndex: '999999',
        boxSizing: 'border-box',
      }
    }

    toggleModifier(modifierName: 'ctrl' | 'alt' | 'shift' | 'super') {
      if (!this.keyboard) return

      let keysym = this.XK_Control_L
      let activeState = false

      if (modifierName === 'ctrl') {
        this.ctrlActive = !this.ctrlActive
        keysym = this.XK_Control_L
        activeState = this.ctrlActive
      } else if (modifierName === 'alt') {
        this.altActive = !this.altActive
        keysym = this.XK_Alt_L
        activeState = this.altActive
      } else if (modifierName === 'shift') {
        this.shiftActive = !this.shiftActive
        keysym = this.XK_Shift_L
        activeState = this.shiftActive
      } else if (modifierName === 'super') {
        this.superActive = !this.superActive
        keysym = this.XK_Super_L
        activeState = this.superActive
      }

      if (activeState) {
        this.keyboard.press(keysym)
      } else {
        this.keyboard.release(keysym)
      }
    }

    pressKey(keysym: number) {
      if (!this.keyboard) return
      this.keyboard.press(keysym)
      // Instant release for standard keys
      setTimeout(() => {
        this.keyboard.release(keysym)
      }, 50)
    }

    triggerMacro(type: 'copy' | 'paste' | 'all' | 'undo') {
      if (!this.keyboard) return

      // Simulates Ctrl combinations
      // If Ctrl is not already active, temporarily press and release it
      const tempCtrl = !this.ctrlActive

      if (tempCtrl) {
        this.keyboard.press(this.XK_Control_L)
      }

      let keyChar = 0x0063 // 'c'
      if (type === 'paste') keyChar = 0x0076 // 'v'
      if (type === 'all') keyChar = 0x0061 // 'a'
      if (type === 'undo') keyChar = 0x007a // 'z'

      setTimeout(() => {
        this.keyboard.press(keyChar)
        setTimeout(() => {
          this.keyboard.release(keyChar)
          if (tempCtrl) {
            setTimeout(() => {
              this.keyboard.release(this.XK_Control_L)
            }, 50)
          }
        }, 50)
      }, 50)
    }

    resetAll() {
      if (!this.keyboard) return
      this.keyboard.reset()
      this.ctrlActive = false
      this.altActive = false
      this.shiftActive = false
      this.superActive = false
    }
  }
</script>
