<template>
  <aside class="neko-menu">
    <!-- Tab navigation -->
    <div class="tabs-container">
      <ul role="tablist">
        <li
          role="tab"
          :aria-selected="tab === 'chat'"
          :class="{ active: tab === 'chat' }"
          @click.stop.prevent="change('chat')"
        >
          <i class="fas fa-comment-alt" aria-hidden="true" />
          <span>{{ $t('side.chat') }}</span>
        </li>
        <li
          v-if="filetransferAllowed"
          role="tab"
          :aria-selected="tab === 'files'"
          :class="{ active: tab === 'files' }"
          @click.stop.prevent="change('files')"
        >
          <i class="fas fa-file" aria-hidden="true" />
          <span>{{ $t('side.files') }}</span>
        </li>
        <li
          role="tab"
          :aria-selected="tab === 'settings'"
          :class="{ active: tab === 'settings' }"
          @click.stop.prevent="change('settings')"
        >
          <i class="fas fa-sliders-h" aria-hidden="true" />
          <span>{{ $t('side.settings') }}</span>
        </li>
      </ul>
    </div>

    <!-- Tab content with smooth fade+slide transition -->
    <div class="page-container" role="tabpanel">
      <transition name="tab-fade" mode="out-in">
        <neko-chat     v-if="tab === 'chat'"     key="chat" />
        <neko-files    v-else-if="tab === 'files'"    key="files" />
        <neko-settings v-else-if="tab === 'settings'" key="settings" />
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
      // Glassmorphism: slightly more transparent than the panel body
      background: color-mix(in srgb, var(--color-bg) 75%, transparent);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      height: $menu-height;
      max-width: 100%;
      display: flex;
      align-items: flex-end;
      flex-shrink: 0;
      border-bottom: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
      padding: 0 var(--space-3);
      transition: background-color var(--transition-slow);

      ul {
        display: flex;
        align-items: center;
        gap: var(--space-1);
        list-style: none;
        padding: 0;
        margin: 0 0 calc(var(--space-2) * -1) 0; // overlap the border-bottom

        li {
          // Pill shape
          display: inline-flex;
          align-items: center;
          gap: var(--space-2);
          padding: var(--space-2) var(--space-4);
          min-height: 44px; // touch target
          border-radius: var(--radius-full) var(--radius-full) 0 0;
          font-size: var(--text-sm);
          font-weight: 500;
          font-family: var(--font-body);
          color: var(--color-text-muted);
          background: transparent;
          cursor: pointer;
          user-select: none;
          // Smooth transitions for all interactive states
          transition:
            color            var(--transition-interactive),
            background-color var(--transition-interactive),
            transform        var(--transition-interactive);

          i {
            font-size: var(--text-xs);
            transition: transform var(--transition-interactive);
          }

          // Hover state
          &:hover {
            color: var(--color-text);
            background: color-mix(in srgb, var(--color-surface-offset) 70%, transparent);
            transform: translateY(-1px);

            i {
              transform: scale(1.15);
            }
          }

          // Active press state
          &:active {
            transform: scale(0.96) translateY(0);
          }

          // Selected tab
          &.active {
            color: var(--color-primary);
            background: color-mix(in srgb, var(--color-surface) 80%, transparent);
            font-weight: 600;

            i {
              color: var(--color-primary);
            }

            &:hover {
              transform: none;
              background: color-mix(in srgb, var(--color-surface) 80%, transparent);
            }
          }

          // Keyboard focus
          &:focus-visible {
            outline: 2px solid var(--color-primary);
            outline-offset: -2px;
          }
        }
      }
    }

    // ── Tab content area ────────────────────────────────────────────────────
    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
  }

  // ── Tab fade+slide transition ──────────────────────────────────────────────
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
    get filetransferAllowed() {
      return (
        this.$accessor.remote.fileTransfer &&
        (this.$accessor.user.admin || !this.$accessor.isLocked('file_transfer'))
      )
    }

    get tab() {
      return this.$accessor.client.tab
    }

    @Watch('tab', { immediate: true })
    @Watch('filetransferAllowed', { immediate: true })
    onTabChange() {
      // Do not show the files tab if file transfer is disabled
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
