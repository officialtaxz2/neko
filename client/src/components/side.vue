<template>
  <aside class="neko-menu">
    <div class="tabs-container">
      <ul>
        <div class="pill-slider" :style="pillStyle" />
        <li :class="{ active: tab === 'chat' }" @click.stop.prevent="change('chat')">
          <i class="fas fa-comment-alt" />
          <span>{{ $t('side.chat') }}</span>
        </li>
        <li :class="{ active: tab === 'members' }" @click.stop.prevent="change('members')">
          <i class="fas fa-users" />
          <span>{{ $t('side.members') }}</span>
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
      <neko-side-members v-if="tab === 'members'" />
      <neko-files v-if="tab === 'files'" />
      <neko-settings v-if="tab === 'settings'" />
    </div>
    <neko-context ref="context" />
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
  import { Vue, Component, Watch, Ref } from 'vue-property-decorator'

  import Settings from '~/components/settings.vue'
  import Chat from '~/components/chat.vue'
  import Files from '~/components/files.vue'
  import SideMembers from '~/components/side_members.vue'
  import Context from '~/components/context.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-settings': Settings,
      'neko-chat': Chat,
      'neko-files': Files,
      'neko-side-members': SideMembers,
      'neko-context': Context,
    },
  })
  export default class NekoSide extends Vue {
    @Ref('context') readonly _context!: any

    openContext(event: MouseEvent, data: any) {
      const menuEl = this.$el as HTMLElement
      if (menuEl) {
        const rect = menuEl.getBoundingClientRect()
        const customEvent = new Proxy(event, {
          get(target, prop) {
            if (prop === 'clientX') {
              return event.clientX - rect.left
            }
            if (prop === 'clientY') {
              return event.clientY - rect.top
            }
            const value = (target as any)[prop]
            if (typeof value === 'function') {
              return value.bind(target)
            }
            return value
          }
        })
        this._context.open(customEvent as any, data)
      } else {
        this._context.open(event, data)
      }
    }

    get currentTabIndex() {
      const order = ['chat', 'members']
      if (this.filetransferAllowed) {
        order.push('files')
      }
      order.push('settings')
      const idx = order.indexOf(this.tab)
      return idx >= 0 ? idx : 0
    }

    get totalTabs() {
      return this.filetransferAllowed ? 4 : 3
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
