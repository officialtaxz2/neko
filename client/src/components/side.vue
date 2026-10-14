<template>
  <aside class="neko-menu">
    <div class="tabs-container">
      <div class="tabs" role="tablist" :aria-label="$t('side.tabs')">
        <button
          v-for="t in tabs"
          :key="t.id"
          type="button"
          role="tab"
          :class="['icon-btn', 'tab', { active: tab === t.id }]"
          :aria-selected="tab === t.id ? 'true' : 'false'"
          @click.stop.prevent="change(t.id)"
        >
          <i :class="['fas', t.icon]" aria-hidden="true" />
          <span>{{ $t(`side.${t.id}`) }}</span>
        </button>
      </div>
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
      border-bottom: 1px solid oklch(from var(--text-subtle) l c h / 0.06);
      padding: 10px;
      flex-shrink: 0;

      .tabs {
        display: flex;
        gap: 4px;

        .tab {
          flex: 1 1 0;
          min-width: 0;
          flex-direction: column;
          gap: 4px;
          padding: 8px 2px 6px;
          border-radius: 10px;
          font-size: 11.5px;
          font-weight: 600;
          color: var(--text-muted-ok);
          transition: var(--transition-fluid);

          i {
            font-size: 15px;
          }

          span {
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          &:hover {
            color: var(--text-pure);
            background: rgba(255, 255, 255, 0.04);
          }

          &.active {
            color: var(--color-cyber-mint);
            background: rgba(38, 230, 180, 0.12);
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

    openContext(event: MouseEvent | KeyboardEvent, data: any) {
      // Keyboard activation has no pointer position: anchor below the element.
      let x = (event as MouseEvent).clientX
      let y = (event as MouseEvent).clientY
      if (!x && !y && event.currentTarget instanceof HTMLElement) {
        const anchor = event.currentTarget.getBoundingClientRect()
        x = anchor.left
        y = anchor.bottom + 4
      }
      // The context menu is positioned inside the side panel.
      const menuEl = this.$el as HTMLElement
      const rect = menuEl ? menuEl.getBoundingClientRect() : { left: 0, top: 0 }
      this._context.open({ clientX: x - rect.left, clientY: y - rect.top, target: event.target }, data)
    }

    get tabs() {
      const tabs = [
        { id: 'chat', icon: 'fa-comment-alt' },
        { id: 'members', icon: 'fa-users' },
      ]
      if (this.filetransferAllowed) {
        tabs.push({ id: 'files', icon: 'fa-folder' })
      }
      tabs.push({ id: 'settings', icon: 'fa-sliders-h' })
      return tabs
    }

    get canDownload() {
      return this.$accessor.user.admin || this.$accessor.files.userDownload
    }

    get canUpload() {
      return this.$accessor.user.admin || this.$accessor.files.userUpload
    }

    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer &&
        (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer')) &&
        (this.canDownload || this.canUpload)
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
