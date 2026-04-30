<template>
  <aside class="neko-menu">
    <!-- Tab navigation -->
    <div class="tabs-container">
      <ul role="tablist">
        <li
          role="tab"
          :aria-selected="activeChat"
          :class="{ active: activeChat }"
          @click.stop.prevent="toggleChat"
        >
          <i class="fas fa-comment-alt" aria-hidden="true" />
          <span>{{ $t('side.chat') }}</span>
        </li>
        <li
          role="tab"
          :aria-selected="activeUsers"
          :class="{ active: activeUsers }"
          @click.stop.prevent="toggleUsers"
        >
          <i class="fas fa-users" aria-hidden="true" />
          <span>{{ $t('side.users') }}</span>
        </li>
        <li
          v-if="filetransferAllowed"
          role="tab"
          :aria-selected="isFilesActive"
          :class="{ active: isFilesActive }"
          @click.stop.prevent="activateFiles"
        >
          <i class="fas fa-file" aria-hidden="true" />
          <span>{{ $t('side.files') }}</span>
        </li>
        <li
          role="tab"
          :aria-selected="activeSettings"
          :class="{ active: activeSettings }"
          @click.stop.prevent="toggleSettings"
        >
          <i class="fas fa-sliders-h" aria-hidden="true" />
          <span>{{ $t('side.settings') }}</span>
        </li>
      </ul>
    </div>

    <!-- Tab content -->
    <div class="page-container" role="tabpanel">
      <!-- Files panel: exclusive, store-managed -->
      <transition name="tab-fade" mode="out-in">
        <neko-files v-if="isFilesActive" key="files" />
      </transition>

      <!-- Settings panel: exclusive -->
      <transition name="tab-fade" mode="out-in">
        <neko-settings v-if="activeSettings && !isFilesActive" key="settings" />
      </transition>

      <!-- Chat panel: toggleable -->
      <transition name="panel-grow">
        <neko-chat v-if="activeChat && !isFilesActive" key="chat" />
      </transition>

      <!-- Split divider: visible only when both panels are active -->
      <div
        v-if="activeChat && activeUsers && !isFilesActive"
        class="panel-divider"
        aria-hidden="true"
      />

      <!-- Users panel: toggleable — R2 replaces placeholder with <neko-userlist /> -->
      <transition name="panel-grow">
        <div v-if="activeUsers && !isFilesActive" key="users" class="users-panel">
          <div class="users-placeholder">
            <i class="fas fa-users" aria-hidden="true" />
            <span>{{ $t('side.users') }}</span>
          </div>
        </div>
      </transition>
    </div>
  </aside>
</template>

<style lang="scss">
  .neko-menu {
    width: $side-width;
    // Glassmorphism: semi-transparent surface + blur
    background-color: color-mix(in srgb, var(--color-surface) 80%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    flex-shrink: 0;
    max-height: 100%;
    max-width: 100%;
    display: flex;
    flex-direction: column;
    border-left: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
    transition: background-color var(--transition-slow);

    // ── Tab navigation bar ──────────────────────────────────────────────────
    .tabs-container {
      background: color-mix(in srgb, var(--color-bg) 75%, transparent);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      height: $menu-height;
      max-width: 100%;
      display: flex;
      align-items: stretch;
      flex-shrink: 0;
      border-bottom: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
      padding: 0 var(--space-3);
      transition: background-color var(--transition-slow);

      ul {
        display: flex;
        align-items: stretch;
        gap: var(--space-1);
        list-style: none;
        padding: 0;
        margin: 0;

        li {
          display: inline-flex;
          align-items: center;
          gap: var(--space-2);
          padding: 0 var(--space-4);
          border-radius: var(--radius-md);
          position: relative;
          font-size: var(--text-sm);
          font-weight: 500;
          font-family: var(--font-body);
          color: var(--color-text-muted);
          background: transparent;
          cursor: pointer;
          user-select: none;
          transition:
            color            var(--transition-interactive),
            background-color var(--transition-interactive),
            transform        var(--transition-interactive);

          i {
            font-size: var(--text-xs);
            transition: transform var(--transition-interactive);
          }

          &:hover {
            color: var(--color-text);
            background: color-mix(in srgb, var(--color-surface-offset) 60%, transparent);

            i { transform: scale(1.12); }
          }

          &:active {
            transform: scale(0.95);
          }

          &.active {
            color: var(--color-primary);
            font-weight: 600;

            i { color: var(--color-primary); }

            &::after {
              content: '';
              position: absolute;
              bottom: 0;
              left: var(--space-3);
              right: var(--space-3);
              height: 2px;
              background: var(--color-primary);
              border-radius: var(--radius-full) var(--radius-full) 0 0;
            }

            &:hover {
              background: color-mix(in srgb, var(--color-primary-highlight) 55%, transparent);
            }
          }

          &:focus-visible {
            outline: 2px solid var(--color-primary);
            outline-offset: -2px;
          }
        }
      }
    }

    // ── Tab content area ────────────────────────────────────────────────────
    // overflow hidden: each panel handles its own scroll
    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    // ── Split divider between Chat and Users ────────────────────────────────
    .panel-divider {
      height: 1px;
      flex-shrink: 0;
      background: color-mix(in srgb, var(--color-border) 60%, transparent);
    }

    // ── Users placeholder panel (R2: replace with neko-userlist) ────────────
    .users-panel {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      overflow-y: auto;
      min-height: 0;

      .users-placeholder {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-3);
        color: var(--color-text-faint);
        font-size: var(--text-sm);
        font-family: var(--font-body);
        padding: var(--space-8);

        i { font-size: var(--text-xl); }
      }
    }
  }

  // ── Tab fade+slide transition (files / settings exclusive panels) ──────────
  .tab-fade-enter-active,
  .tab-fade-leave-active {
    transition:
      opacity   var(--transition-interactive),
      transform var(--transition-interactive);
  }

  .tab-fade-enter {
    opacity: 0;
    transform: translateY(4px);
  }

  .tab-fade-leave-to {
    opacity: 0;
    transform: translateY(-4px);
  }

  // ── Panel-grow transition (chat / users toggleable panels) ────────────────
  // max-height trick: smooth height expand/collapse without JS measurements
  .panel-grow-enter-active,
  .panel-grow-leave-active {
    transition:
      max-height var(--transition-slow),
      opacity    var(--transition-interactive);
    overflow: hidden;
  }

  .panel-grow-enter,
  .panel-grow-leave-to {
    max-height: 0;
    opacity: 0;
  }

  .panel-grow-enter-to,
  .panel-grow-leave {
    // large enough to cover any realistic panel height
    max-height: 100vh;
    opacity: 1;
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
  export default class extends Vue {
    activeChat = true
    activeUsers = false
    activeSettings = false

    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer &&
        (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer'))
      )
    }

    // Store tab is used exclusively to track the Files panel state
    get tab() {
      return this.$accessor.client.tab
    }

    get isFilesActive() {
      return this.tab === 'files'
    }

    @Watch('tab', { immediate: true })
    @Watch('filetransferAllowed', { immediate: true })
    onTabChange() {
      // Fall back out of files if file transfer gets disabled
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

    // Chat: toggle; exit exclusive modes; prevent deactivating last active panel
    toggleChat() {
      if (this.isFilesActive) this.change('chat')
      if (this.activeSettings) {
        this.activeSettings = false
        this.activeChat = true
        return
      }
      if (this.activeChat && !this.activeUsers) return
      this.activeChat = !this.activeChat
    }

    // Users: toggle; exit exclusive modes; prevent deactivating last active panel
    toggleUsers() {
      if (this.isFilesActive) this.change('chat')
      if (this.activeSettings) {
        this.activeSettings = false
        this.activeUsers = true
        return
      }
      if (this.activeUsers && !this.activeChat) return
      this.activeUsers = !this.activeUsers
    }

    // Settings: exclusive — deactivates Chat + Users; restores Chat on exit
    toggleSettings() {
      if (this.activeSettings) {
        this.activeSettings = false
        this.activeChat = true
      } else {
        if (this.isFilesActive) this.change('chat')
        this.activeChat = false
        this.activeUsers = false
        this.activeSettings = true
      }
    }

    // Files: store-managed; clears settings state on activation
    activateFiles() {
      this.activeSettings = false
      this.change('files')
    }

    change(tab: string) {
      this.$accessor.client.setTab(tab)
    }
  }
</script>
