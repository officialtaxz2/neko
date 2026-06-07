<template>
  <aside class="neko-menu">
    <div class="tabs-container">
      <ul>
        <div class="pill-slider" :style="pillStyle" />
        <li :class="{ active: tab === 'chat' }" @click.stop.prevent="change('chat')">
          <i class="fas fa-comment-alt" />
          <span>{{ $t('side.chat') }}</span>
        </li>
        <li v-if="filetransferAllowed" :class="{ active: tab === 'files' }" @click.stop.prevent="change('files')">
          <i class="fas fa-file" />
          <span>{{ $t('side.files') }}</span>
        </li>
        <li :class="{ active: tab === 'settings' }" @click.stop.prevent="change('settings')">
          <i class="fas fa-sliders-h" />
          <span>{{ $t('side.settings') }}</span>
        </li>
      </ul>
    </div>
    <div class="page-container">
      <neko-chat v-if="tab === 'chat'" />
      <neko-files v-if="tab === 'files'" />
      <neko-settings v-if="tab === 'settings'" />
    </div>
  </aside>
</template>

<style lang="scss">
  .neko-menu {
    width: $side-width;
    background-color: rgba(18, 21, 30, 0.45);
    backdrop-filter: blur(16px);
    border-left: 1px solid oklch(from var(--text-subtle) l c h / 0.08);
    flex-shrink: 0;
    max-height: 100%;
    max-width: 100%;
    display: flex;
    flex-direction: column;
    box-shadow: -10px 0 30px rgba(0, 0, 0, 0.25);

    .tabs-container {
      background: transparent;
      border-bottom: 1px solid oklch(from var(--text-subtle) l c h / 0.04);
      height: 56px;
      max-height: 100%;
      max-width: 100%;
      display: flex;
      align-items: center;
      padding: 0 16px;
      flex-shrink: 0;

      ul {
        background: rgba(255, 255, 255, 0.02);
        border: 1px solid oklch(from var(--text-subtle) l c h / 0.04);
        border-radius: 12px; // outer radius
        padding: 3px; // spacing padding
        position: relative;
        display: flex;
        flex-direction: row;
        margin: 0;
        width: 100%;
        list-style: none;

        .pill-slider {
          position: absolute;
          top: 3px;
          bottom: 3px;
          border-radius: 9px; // inner-radius = outer (12px) - gap (3px) = 9px! PERFECT Nested Radius!
          background: linear-gradient(135deg, var(--color-cyber-mint) 0%, var(--color-cyber-mint-active) 100%);
          box-shadow: 0 4px 12px var(--color-cyber-mint-glow);
          transition: transform 0.35s cubic-bezier(0.16, 1, 0.3, 1), width 0.35s cubic-bezier(0.16, 1, 0.3, 1), left 0.35s cubic-bezier(0.16, 1, 0.3, 1);
          z-index: 1;
        }

        li {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 8px 0;
          font-size: 13px;
          font-weight: 600;
          color: var(--text-subtle);
          cursor: pointer;
          transition: var(--transition-fluid);
          user-select: none;
          z-index: 2; // above pill-slider!
          border: none;
          background: transparent;

          i {
            font-size: 11px;
            color: inherit;
          }

          &:hover {
            color: var(--text-pure);
          }

          &.active {
            color: #050508;
            font-weight: 700;
          }
        }
      }
    }

    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      overflow: hidden;
      padding-top: 5px;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Watch } from 'vue-property-decorator'

  import Settings from '~/components/settings.vue'
  import Chat from '~/components/chat.vue'
  import Files from '~/components/files.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-settings': Settings,
      'neko-chat': Chat,
      'neko-files': Files,
    },
  })
  export default class NekoSide extends Vue {
    get currentTabIndex() {
      if (this.tab === 'chat') return 0
      if (this.tab === 'files') return 1
      if (this.tab === 'settings') return this.filetransferAllowed ? 2 : 1
      return 0
    }

    get totalTabs() {
      return this.filetransferAllowed ? 3 : 2
    }

    get pillStyle() {
      const idx = this.currentTabIndex
      const total = this.totalTabs
      const widthPercent = 100 / total
      const leftPercent = widthPercent * idx
      return {
        width: `calc(${widthPercent}% - 6px)`,
        left: `calc(${leftPercent}% + 3px)`,
      }
    }

    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer && (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer'))
      )
    }

    get tab() {
      return this.$accessor.client.tab
    }

    @Watch('tab', { immediate: true })
    @Watch('filetransferAllowed', { immediate: true })
    onTabChange() {
      // do not show the files tab if file transfer is disabled
      if (this.tab === 'files' && !this.filetransferAllowed) {
        this.change('chat')
      }
    }

    @Watch('filetransferAllowed')
    onFileTransferAllowedChange() {
      if (this.filetransferAllowed) {
        this.$accessor.files.refresh()
      }
    }

    change(tab: string) {
      this.$accessor.client.setTab(tab)
    }
  }
</script>
