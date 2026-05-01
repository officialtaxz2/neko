<template>
  <div id="neko">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>
    <template v-else>
      <!-- Login-Page / Room toggle: page-fade out-in prevents any layout bleed -->
      <transition name="page-fade" mode="out-in">
        <!-- Login-Page: only rendered when not connected — zero Room DOM -->
        <neko-login-page v-if="!connected" key="login" />

        <!-- Room: only rendered after successful login -->
        <div
          v-else
          key="room"
          class="neko-room-wrapper"
          :class="[!videoOnly && side ? 'expanded' : '']"
        >
          <!-- ref="mainContainer" used for scroll-listener (sticky header state) -->
          <main class="neko-main" ref="mainContainer">
            <div v-if="!videoOnly" class="header-container" :class="{ 'is-scrolled': headerScrolled }">
              <neko-header :currentTheme="currentTheme" @toggle-theme="toggleTheme" />
            </div>
            <div class="video-container">
              <neko-video
                ref="video"
                :hideControls="hideControls"
                :extraControls="isEmbedMode"
                @control-attempt="controlAttempt"
              />
            </div>
            <div v-if="!videoOnly" class="room-container">
              <div class="room-menu">
                <div class="settings">
                  <neko-menu />
                </div>
                <div class="controls">
                  <neko-controls :shakeKbd="shakeKbd" />
                </div>
                <div class="emotes">
                  <neko-emotes />
                </div>
              </div>
            </div>
          </main>
          <!-- Sidebar: implicit within connected branch — no extra connected check needed -->
          <transition name="side">
            <neko-side v-if="!videoOnly && side" />
          </transition>
          <neko-about v-if="about" />
          <notifications
            v-if="!videoOnly"
            group="neko"
            position="top left"
            style="top: 50px; pointer-events: none"
            :ignoreDuplicates="true"
          />
        </div>
      </transition>
    </template>
  </div>
</template>

<style lang="scss">
  // Login-Page ↔ Room fade
  .page-fade-enter-active,
  .page-fade-leave-active {
    transition: opacity var(--transition-slow);
  }
  .page-fade-enter,
  .page-fade-leave-to {
    opacity: 0;
  }

  // prefers-reduced-motion: skip page-fade entirely
  @media (prefers-reduced-motion: reduce) {
    .page-fade-enter-active,
    .page-fade-leave-active {
      transition: none;
    }
  }

  #neko {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    max-width: 100vw;
    max-height: 100vh;
    display: flex;
    background: var(--color-bg);
    overflow: hidden;
    transition: background-color var(--transition-slow);
  }

  // Room wrapper: carries .expanded when sidebar is open
  .neko-room-wrapper {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: row;
    overflow: hidden;

    // When sidebar is open: right-edge glow on main as visual split separator
    &.expanded .neko-main {
      box-shadow: 4px 0 20px color-mix(in srgb, var(--color-primary) 10%, transparent);
    }

    .neko-main {
      min-width: 360px;
      max-width: 100%;
      flex-grow: 1;
      flex-direction: column;
      display: flex;
      overflow: auto;
      scroll-behavior: smooth;
      transition: box-shadow var(--transition-slow);

      .header-container {
        background: transparent;
        height: $menu-height;
        flex-shrink: 0;
        display: flex;
        position: sticky;
        top: 0;
        z-index: 10;
        transition:
          background-color var(--transition-slow),
          backdrop-filter  var(--transition-slow),
          box-shadow       var(--transition-slow);

        &.is-scrolled {
          background: color-mix(in srgb, var(--color-surface) 92%, transparent);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          box-shadow:
            0 1px 0 color-mix(in srgb, var(--color-border) 70%, transparent),
            var(--shadow-sm);
        }
      }

      .video-container {
        background: var(--color-bg);
        max-width: 100%;
        flex-grow: 1;
        display: flex;
        transition: background-color var(--transition-slow);
      }

      .room-container {
        background: linear-gradient(
          to top,
          var(--color-bg) 0%,
          color-mix(in srgb, var(--color-surface) 85%, transparent) 100%
        );
        // backdrop-filter intentionally NOT here — see original comment re: vue-context
        // (backdrop-filter creates a new containing block for position:fixed descendants)
        height: $controls-height;
        max-width: 100%;
        flex-shrink: 0;
        flex-direction: column;
        display: flex;
        border-top: 1px solid color-mix(in srgb, var(--color-primary) 14%, var(--color-border));
        transition: background-color var(--transition-slow);
        position: sticky;
        bottom: 0;
        z-index: 10;

        // Glass blur via pseudo-element — avoids containing-block side effect on fixed children
        &::before {
          content: '';
          position: absolute;
          inset: 0;
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          pointer-events: none;
          z-index: -1;
        }

        .room-menu {
          max-width: 100%;
          flex: 1;
          display: flex;

          .settings {
            margin-left: var(--space-3);
            flex: 1;
            justify-content: flex-start;
            align-items: center;
            display: flex;
          }

          .controls {
            flex: 1;
            justify-content: center;
            align-items: center;
            display: flex;
          }

          .emotes {
            margin-right: var(--space-3);
            flex: 1;
            justify-content: flex-end;
            align-items: center;
            display: flex;
          }
        }
      }
    }
  }

  // ------------------------------------------------------------------
  // Sidebar open/close transition — slide from right + fade
  // Vue 2 hook names: -enter / -leave-to
  // ------------------------------------------------------------------
  .side-enter-active,
  .side-leave-active {
    transition:
      transform var(--transition-slow),
      opacity   var(--transition-slow);
    overflow: hidden;
  }

  .side-enter,
  .side-leave-to {
    transform: translateX(100%);
    opacity: 0;
  }

  @media only screen and (max-width: 1024px) {
    html,
    body {
      // overflow-x:hidden prevents any sub-pixel horizontal bleed that
      // causes mobile browsers to reserve a scrollbar gutter on the right.
      overflow-x: hidden !important;
      overflow-y: auto !important;
      // width/height auto — let the browser calculate naturally
      width: auto !important;
      height: auto !important;
    }

    body > p {
      display: none;
    }

    #neko {
      position: relative;
      // Use width/max-width 100% instead of 100vw.
      // On mobile, 100vw can include the scrollbar gutter width,
      // producing a dead strip on the right side of the screen.
      width: 100% !important;
      max-width: 100% !important;
      max-height: initial !important;
    }

    .neko-room-wrapper {
      position: relative;
      flex-direction: column;
      // Ensure the wrapper never exceeds the viewport width
      width: 100%;
      max-width: 100%;

      .neko-main {
        height: 100vh;
      }

      .neko-menu {
        height: 100vh;
        width: 100% !important;
      }

      // No split glow on mobile
      &.expanded .neko-main {
        box-shadow: none;
      }
    }

    // On mobile: sidebar slides up from bottom
    .side-enter,
    .side-leave-to {
      transform: translateY(100%);
      opacity: 0;
    }
  }

  @media only screen and (max-width: 1024px) and (orientation: portrait) {
    .neko-room-wrapper {
      &.expanded .neko-main {
        height: 40vh;
      }

      &.expanded .neko-menu {
        height: 60vh;
        width: 100% !important;
      }
    }
  }

  @media only screen and (max-width: 768px) {
    .neko-room-wrapper .neko-main .room-container {
      display: none;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  import LoginPage from '~/components/login.vue'
  import Video from '~/components/video.vue'
  import Menu from '~/components/menu.vue'
  import Side from '~/components/side.vue'
  import Controls from '~/components/controls.vue'
  import Emotes from '~/components/emotes.vue'
  import About from '~/components/about.vue'
  import Header from '~/components/header.vue'
  import Unsupported from '~/components/unsupported.vue'

  type Theme = 'dark' | 'light'

  @Component({
    name: 'neko',
    components: {
      'neko-login-page': LoginPage,
      'neko-video': Video,
      'neko-menu': Menu,
      'neko-side': Side,
      'neko-controls': Controls,
      'neko-emotes': Emotes,
      'neko-about': About,
      'neko-header': Header,
      'neko-unsupported': Unsupported,
    },
  })
  export default class extends Vue {
    @Ref('video') video!: Video
    @Ref('mainContainer') mainContainer!: HTMLElement

    shakeKbd = false
    currentTheme: Theme = 'dark'

    headerScrolled = false
    private mainScrollHandler: (() => void) | null = null

    mounted() {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      this.currentTheme = prefersDark ? 'dark' : 'light'
      this.applyTheme(this.currentTheme)

      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (!this._userChoseTheme) {
          this.currentTheme = e.matches ? 'dark' : 'light'
          this.applyTheme(this.currentTheme)
        }
      })

      this.$nextTick(() => {
        if (this.mainContainer) {
          this.mainScrollHandler = () => {
            this.headerScrolled = this.mainContainer.scrollTop > 4
          }
          this.mainContainer.addEventListener('scroll', this.mainScrollHandler, { passive: true })
        }
      })
    }

    beforeDestroy() {
      if (this.mainContainer && this.mainScrollHandler) {
        this.mainContainer.removeEventListener('scroll', this.mainScrollHandler)
      }
    }

    private _userChoseTheme = false

    toggleTheme() {
      this._userChoseTheme = true
      this.currentTheme = this.currentTheme === 'dark' ? 'light' : 'dark'
      this.applyTheme(this.currentTheme)
    }

    private applyTheme(theme: Theme) {
      document.documentElement.setAttribute('data-theme', theme)
    }

    get volume() {
      const numberParam = parseFloat(new URL(location.href).searchParams.get('volume') || '1.0')
      return Math.max(0.0, Math.min(!isNaN(numberParam) ? numberParam * 100 : 100, 100))
    }

    get isCastMode() {
      return !!new URL(location.href).searchParams.get('cast')
    }

    get isEmbedMode() {
      return !!new URL(location.href).searchParams.get('embed')
    }

    get hideControls() {
      return this.isCastMode
    }

    get videoOnly() {
      return this.isCastMode || this.isEmbedMode
    }

    @Watch('volume', { immediate: true })
    onVolume(volume: number) {
      if (new URL(location.href).searchParams.has('volume')) {
        this.$accessor.video.setVolume(volume)
      }
    }

    @Watch('hideControls', { immediate: true })
    onHideControls(enabled: boolean) {
      if (enabled) {
        this.$accessor.video.setMuted(false)
        this.$accessor.settings.setSound(false)
      }
    }

    @Watch('side')
    onSide(side: boolean) {
      if (side) {
        this.$nextTick(() => {
          const sideEl = document.querySelector('aside')
          if (sideEl) {
            sideEl.scrollIntoView({ behavior: 'smooth', block: 'start' })
          }
        })
      }
    }

    controlAttempt() {
      if (this.shakeKbd || this.$accessor.remote.hosted) return
      this.shakeKbd = true
      window.setTimeout(() => (this.shakeKbd = false), 5000)
    }

    get about() {
      return this.$accessor.client.about
    }

    get side() {
      return this.$accessor.client.side
    }

    get connected() {
      return this.$accessor.connected
    }
  }
</script>
