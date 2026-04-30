<template>
  <div id="neko" :class="[!videoOnly && side ? 'expanded' : '']">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>
    <template v-else>
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
      <!-- Sidebar: wrapped in <transition> for slide-in/out animation -->
      <transition name="side">
        <neko-side v-if="!videoOnly && side" />
      </transition>
      <neko-connect v-if="!connected" />
      <neko-about v-if="about" />
      <notifications
        v-if="!videoOnly"
        group="neko"
        position="top left"
        style="top: 50px; pointer-events: none"
        :ignoreDuplicates="true"
      />
    </template>
  </div>
</template>

<style lang="scss">
  #neko {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    max-width: 100vw;
    max-height: 100vh;
    flex-direction: row;
    display: flex;
    background: var(--color-bg);
    // Clip sidebar during slide animation
    overflow: hidden;
    transition: background-color var(--transition-slow);

    .neko-main {
      min-width: 360px;
      max-width: 100%;
      flex-grow: 1;
      flex-direction: column;
      display: flex;
      overflow: auto;
      // Native smooth scrolling for all scroll events within this container
      scroll-behavior: smooth;
      // Transition for the split-separator glow
      transition: box-shadow var(--transition-slow);

      .header-container {
        // Transparent baseline: lets header.vue gradient show through
        background: transparent;
        height: $menu-height;
        flex-shrink: 0;
        display: flex;
        // Sticky: stays at top when neko-main is scrolled (mobile)
        position: sticky;
        top: 0;
        z-index: 10;
        transition:
          background-color var(--transition-slow),
          backdrop-filter  var(--transition-slow),
          box-shadow       var(--transition-slow);

        // Elevated state: active when user has scrolled down > 4px
        // Adds glassmorphism blur + shadow to visually lift the header
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
        // Token-based bg (replaces hardcoded rgba(0,0,0,0.4))
        background: var(--color-bg);
        max-width: 100%;
        flex-grow: 1;
        display: flex;
        transition: background-color var(--transition-slow);
      }

      .room-container {
        // Gradient background — semi-transparent at top to let blur show through
        background: linear-gradient(
          to top,
          var(--color-bg) 0%,
          color-mix(in srgb, var(--color-surface) 85%, transparent) 100%
        );
        // NOTE: backdrop-filter intentionally NOT on this element.
        // backdrop-filter creates a new containing block for position:fixed
        // descendants (CSS spec). vue-context (emoji picker) uses position:fixed
        // and is a child of this element — if backdrop-filter were here, the
        // picker's viewport coordinates would be mis-interpreted as local coords,
        // causing it to overflow .neko-main and trigger scrollbars.
        // Glass blur is applied via ::before pseudo-element instead, which has
        // no children and therefore does not affect any fixed-position descendants.
        height: $controls-height;
        max-width: 100%;
        flex-shrink: 0;
        flex-direction: column;
        display: flex;
        // Teal-tinted top border as split accent between video and controls
        border-top: 1px solid color-mix(in srgb, var(--color-primary) 14%, var(--color-border));
        transition: background-color var(--transition-slow);
        // Sticky: stays at bottom when neko-main is scrolled (mobile)
        position: sticky;
        bottom: 0;
        z-index: 10;

        // Glass blur via pseudo-element — same visual result as backdrop-filter
        // on the element, but without the containing-block side effect for
        // position:fixed children (vue-context emoji picker).
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

    // When sidebar is open: right-edge glow on main as visual split separator
    &.expanded .neko-main {
      box-shadow: 4px 0 20px color-mix(in srgb, var(--color-primary) 10%, transparent);
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
    // Prevent layout overflow during animation
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
      overflow-y: auto !important;
      width: auto !important;
      height: auto !important;
    }

    body > p {
      display: none;
    }

    #neko {
      position: relative;
      flex-direction: column;
      max-height: initial !important;

      .neko-main {
        height: 100vh;
      }

      .neko-menu {
        height: 100vh;
        width: 100% !important;
      }

      // No split glow on mobile — sidebar slides up, not beside
      &.expanded .neko-main {
        box-shadow: none;
      }
    }

    // On mobile: sidebar slides up from the bottom instead
    .side-enter,
    .side-leave-to {
      transform: translateY(100%);
      opacity: 0;
    }
  }

  @media only screen and (max-width: 1024px) and (orientation: portrait) {
    #neko {
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
    #neko .neko-main .room-container {
      display: none;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Ref, Watch } from 'vue-property-decorator'

  import Connect from '~/components/connect.vue'
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
      'neko-connect': Connect,
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

    // Tracks whether the main container has been scrolled — drives .is-scrolled on header
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

      // Attach passive scroll listener to detect header elevation state
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
