<template>
  <div id="neko" :class="[!videoOnly && side ? 'expanded' : '']">
    <template v-if="!$client.supported">
      <neko-unsupported />
    </template>
    <template v-else>
      <main class="neko-main">
        <div v-if="!videoOnly" class="header-container">
          <neko-header />
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
          <neko-members />
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
      <neko-side v-if="!videoOnly && side" />
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
    background-color: var(--bg-obsidian-deep);

    .neko-main {
      min-width: 360px;
      max-width: 100%;
      flex-grow: 1;
      flex-direction: column;
      display: flex;
      overflow: hidden;
      position: relative;

      .header-container {
        background: rgba(12, 13, 18, 0.65);
        backdrop-filter: blur(12px);
        height: $menu-height;
        flex-shrink: 0;
        display: flex;
        border-bottom: 1px solid var(--glass-border);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        z-index: 100;
      }

      .video-container {
        background: radial-gradient(circle at 50% 50%, rgba(12, 13, 18, 0.2) 0%, rgba(7, 8, 12, 0.7) 100%);
        max-width: 100%;
        flex-grow: 1;
        display: flex;
        position: relative;
        overflow: hidden;
      }

      .room-container {
        background: rgba(18, 21, 30, 0.98);
        height: $controls-height;
        max-width: 100%;
        flex-shrink: 0;
        flex-direction: column;
        display: flex;
        border-top: 1px solid var(--glass-border);
        box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.25);
        z-index: 10;
        padding: 4px 0;

        .room-menu {
          max-width: 100%;
          flex: 1;
          display: flex;

          .settings {
            margin-left: 15px;
            flex: 1;
            justify-content: flex-start;
            align-items: center;
            display: flex;
          }

          .controls {
            flex: 1.5;
            justify-content: center;
            align-items: center;
            display: flex;
          }

          .emotes {
            margin-right: 15px;
            flex: 1;
            justify-content: flex-end;
            align-items: center;
            display: flex;
          }
        }
      }
    }
  }

  @media only screen and (max-width: 1024px) {
    html,
    body {
      overflow-y: auto !important;
      width: auto !important;
      height: auto !important;

      /* Elegant, modern scrollbar integration */
      scrollbar-width: thin;
      scrollbar-color: rgba(255, 255, 255, 0.15) transparent;

      &::-webkit-scrollbar {
        width: 6px;
        height: 6px;
      }
      &::-webkit-scrollbar-track {
        background: transparent;
      }
      &::-webkit-scrollbar-thumb {
        background: rgba(255, 255, 255, 0.15);
        border-radius: 3px;
        transition: background 0.3s ease;
        
        &:hover {
          background: rgba(255, 255, 255, 0.3);
        }
      }
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

  /* Premium, modern glassmorphic notification styling */
  .vue-notification-group {
    left: 20px !important;
    top: 60px !important;
    max-width: 340px !important;
    z-index: 99999 !important;
    padding: 0 !important;
    
    @media only screen and (max-width: 480px) {
      left: 10px !important;
      right: 10px !important;
      top: 15px !important;
      max-width: calc(100% - 20px) !important;
    }
  }

  .vue-notification {
    margin: 0 0 10px 0 !important;
    padding: 14px 16px !important;
    font-size: 13px !important;
    font-family: inherit !important;
    color: #e5e7eb !important; /* Soft white/gray body */
    border-radius: 12px !important;
    border: 1px solid rgba(255, 255, 255, 0.08) !important;
    background: rgba(18, 20, 29, 0.85) !important;
    backdrop-filter: blur(12px) !important;
    -webkit-backdrop-filter: blur(12px) !important;
    box-shadow: 
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 10px 30px -5px rgba(0, 0, 0, 0.3),
      inset 0 1px 0 0 rgba(255, 255, 255, 0.05) !important;
    transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1) !important;
    border-left: 4px solid #6366f1 !important; /* Indigo primary glow */

    .notification-title {
      font-weight: 600 !important;
      font-size: 13px !important;
      letter-spacing: -0.0125em;
      line-height: 1.4 !important;
      color: #ffffff !important;
      margin-bottom: 2px !important;
    }

    .notification-content {
      font-size: 12px !important;
      opacity: 0.8 !important;
      line-height: 1.5 !important;
      color: #cbd5e1 !important;
    }

    /* Type overrides for feedback states matching DesignSystem.md */
    &.success {
      background: rgba(16, 24, 21, 0.88) !important;
      border-color: rgba(34, 197, 94, 0.15) !important;
      border-left: 4px solid #22c55e !important;
      box-shadow: 
        0 10px 30px -5px rgba(16, 24, 21, 0.5),
        inset 0 1px 0 0 rgba(34, 197, 94, 0.1) !important;
      
      .notification-title {
        color: #86efac !important;
      }
    }

    &.warn, &.warning {
      background: rgba(28, 23, 15, 0.88) !important;
      border-color: rgba(234, 179, 8, 0.15) !important;
      border-left: 4px solid #eab308 !important;
      box-shadow: 
        0 10px 30px -5px rgba(28, 23, 15, 0.5),
        inset 0 1px 0 0 rgba(234, 179, 8, 0.1) !important;

      .notification-title {
        color: #fde047 !important;
      }
    }

    &.error {
      background: rgba(28, 15, 15, 0.88) !important;
      border-color: rgba(239, 68, 68, 0.15) !important;
      border-left: 4px solid #ef4444 !important;
      box-shadow: 
        0 10px 30px -5px rgba(28, 15, 15, 0.5),
        inset 0 1px 0 0 rgba(239, 68, 68, 0.1) !important;

      .notification-title {
        color: #fca5a5 !important;
      }
    }

    &.info {
      background: rgba(15, 23, 42, 0.88) !important;
      border-color: rgba(59, 130, 246, 0.15) !important;
      border-left: 4px solid #3b82f6 !important;
      box-shadow: 
        0 10px 30px -5px rgba(15, 23, 42, 0.4),
        inset 0 1px 0 0 rgba(59, 130, 246, 0.1) !important;

      .notification-title {
        color: #93c5fd !important;
      }
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
  import Members from '~/components/members.vue'
  import Emotes from '~/components/emotes.vue'
  import About from '~/components/about.vue'
  import Header from '~/components/header.vue'
  import Unsupported from '~/components/unsupported.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-connect': Connect,
      'neko-video': Video,
      'neko-menu': Menu,
      'neko-side': Side,
      'neko-controls': Controls,
      'neko-members': Members,
      'neko-emotes': Emotes,
      'neko-about': About,
      'neko-header': Header,
      'neko-unsupported': Unsupported,
    },
  })
  export default class Neko extends Vue {
    @Ref('video') video!: Video

    shakeKbd = false

    get volume() {
      const numberParam = parseFloat(new URLSearchParams(location.search).get('volume') || '1.0')
      return Math.max(0.0, Math.min(!isNaN(numberParam) ? numberParam * 100 : 100, 100))
    }

    get isCastMode() {
      return !!new URLSearchParams(location.search).get('cast')
    }

    get isEmbedMode() {
      return !!new URLSearchParams(location.search).get('embed')
    }

    get hideControls() {
      return this.isCastMode
    }

    get videoOnly() {
      return this.isCastMode || this.isEmbedMode
    }

    @Watch('volume', { immediate: true })
    onVolume(volume: number) {
      if (new URLSearchParams(location.search).has('volume')) {
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
        console.log('side enabled')
        // scroll to the side
        this.$nextTick(() => {
          const side = document.querySelector('aside')
          if (side) {
            side.scrollIntoView({ behavior: 'smooth', block: 'start' })
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
