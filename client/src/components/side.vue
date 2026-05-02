<template>
  <aside class="neko-menu">
    <!-- Tab navigation -->
    <div class="tabs-container">
      <ul role="tablist">
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
          role="tab"
          :aria-selected="activeChat"
          :class="{ active: activeChat }"
          @click.stop.prevent="toggleChat"
        >
          <i class="fas fa-comment-alt" aria-hidden="true" />
          <span>{{ $t('side.chat') }}</span>
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
    <div class="page-container" :class="splitClass" role="tabpanel">
      <!-- Files panel: exclusive, store-managed -->
      <transition name="tab-fade" mode="out-in">
        <neko-files v-if="isFilesActive" key="files" />
      </transition>

      <!-- Settings panel: exclusive -->
      <transition name="tab-fade" mode="out-in">
        <neko-settings v-if="activeSettings && !isFilesActive" key="settings" />
      </transition>

      <!-- Users slot (row 1) -->
      <div class="panel-row panel-row--users" v-if="!isFilesActive && !activeSettings">
        <transition name="panel-from-top">
          <neko-userlist v-if="activeUsers" key="users" class="panel-slot" />
        </transition>
      </div>

      <!-- Split divider: only when both panels are active -->
      <div
        v-if="activeChat && activeUsers && !isFilesActive && !activeSettings"
        class="panel-divider"
        aria-hidden="true"
      />

      <!-- Chat slot (row 2) -->
      <div class="panel-row panel-row--chat" v-if="!isFilesActive && !activeSettings">
        <transition name="panel-from-bottom">
          <neko-chat v-if="activeChat" key="chat" class="panel-slot" />
        </transition>
      </div>
    </div>
  </aside>
</template>

<style lang="scss">
  .neko-menu {
    width: $side-width;
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

    // ── Tab navigation bar ──────────────────────────────────────────
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

          &:active { transform: scale(0.95); }

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

    // ── Tab content area ───────────────────────────────────────────
    .page-container {
      max-height: 100%;
      flex-grow: 1;
      // Grid layout drives the animated height split
      display: grid;
      grid-template-rows: 1fr auto 1fr; // users | divider | chat
      overflow: hidden;
      transition: grid-template-rows 280ms cubic-bezier(0.16, 1, 0.3, 1);

      // Users-only: chat row collapses to 0
      &.users-only  { grid-template-rows: 1fr auto 0fr; }
      // Chat-only: users row collapses to 0
      &.chat-only   { grid-template-rows: 0fr auto 1fr; }
      // Both open: equal 50/50
      &.split        { grid-template-rows: 1fr auto 1fr; }
      // Exclusive panels (files/settings): rows irrelevant, display:flex fallback
      &.exclusive    {
        display: flex;
        flex-direction: column;
        transition: none;
      }
    }

    // Panel row wrappers — required for grid-template-rows collapse trick
    .panel-row {
      min-height: 0;
      overflow: hidden;
      display: flex;
      flex-direction: column;
    }

    // Each toggleable panel fills its row
    .panel-slot {
      flex: 1;
      min-height: 0;
      overflow: hidden;
    }

    // ── Split divider between Users and Chat ──────────────────────────
    .panel-divider {
      height: 1px;
      flex-shrink: 0;
      background: color-mix(in srgb, var(--color-border) 60%, transparent);
    }
  }

  // ── Tab fade+slide (files / settings exclusive panels) ─────────────────
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

  // ── Users panel: slides in/out from the TOP ─────────────────────────
  .panel-from-top-enter-active,
  .panel-from-top-leave-active {
    transition:
      opacity   250ms cubic-bezier(0.16, 1, 0.3, 1),
      transform 280ms cubic-bezier(0.16, 1, 0.3, 1);
    will-change: transform, opacity;
  }

  .panel-from-top-enter {
    opacity: 0;
    transform: translateY(-10px);
  }

  .panel-from-top-leave-to {
    opacity: 0;
    transform: translateY(-10px);
  }

  // ── Chat panel: slides in/out from the BOTTOM ──────────────────────
  .panel-from-bottom-enter-active,
  .panel-from-bottom-leave-active {
    transition:
      opacity   250ms cubic-bezier(0.16, 1, 0.3, 1),
      transform 280ms cubic-bezier(0.16, 1, 0.3, 1);
    will-change: transform, opacity;
  }

  .panel-from-bottom-enter {
    opacity: 0;
    transform: translateY(10px);
  }

  .panel-from-bottom-leave-to {
    opacity: 0;
    transform: translateY(10px);
  }

  // ── prefers-reduced-motion ────────────────────────────────────────
  @media (prefers-reduced-motion: reduce) {
    .page-container { transition: none; }

    .panel-from-top-enter-active,
    .panel-from-top-leave-active,
    .panel-from-bottom-enter-active,
    .panel-from-bottom-leave-active,
    .tab-fade-enter-active,
    .tab-fade-leave-active {
      transition-duration: 0.01ms !important;
    }
  }
</style>

<script lang="ts">
  import { Vue, Component, Watch } from 'vue-property-decorator'

  import Settings from '~/components/settings.vue'
  import Chat from '~/components/chat.vue'
  import Files from '~/components/files.vue'
  import Userlist from '~/components/userlist.vue'

  @Component({
    name: 'neko',
    components: {
      'neko-settings': Settings,
      'neko-chat': Chat,
      'neko-files': Files,
      'neko-userlist': Userlist,
    },
  })
  export default class extends Vue {
    activeChat     = true
    activeUsers    = true
    activeSettings = false

    /** CSS class driving the grid-template-rows animation. */
    get splitClass(): string {
      if (this.isFilesActive || this.activeSettings) return 'exclusive'
      if (this.activeUsers && this.activeChat)       return 'split'
      if (this.activeUsers)                          return 'users-only'
      return 'chat-only'
    }

    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer &&
        (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer'))
      )
    }

    get tab() {
      return this.$accessor.client.tab
    }

    get isFilesActive() {
      return this.tab === 'files'
    }

    @Watch('tab', { immediate: true })
    @Watch('filetransferAllowed', { immediate: true })
    onTabChange() {
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

    toggleSettings() {
      if (this.activeSettings) {
        this.activeSettings = false
        this.activeChat = true
      } else {
        if (this.isFilesActive) this.change('chat')
        this.activeChat     = false
        this.activeUsers    = false
        this.activeSettings = true
      }
    }

    activateFiles() {
      this.activeSettings = false
      this.change('files')
    }

    change(tab: string) {
      this.$accessor.client.setTab(tab)
    }
  }
</script>
