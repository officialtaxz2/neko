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
        <neko-chat     v-if="tab === 'chat'"          key="chat" />
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
      background: color-mix(in srgb, var(--color-bg) 75%, transparent);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      height: $menu-height;
      max-width: 100%;
      display: flex;
      // align-items: stretch lets the ul + li fill the full height,
      // enabling the ::after underline to sit flush at the bottom edge.
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
        // No negative margin — old overlap trick caused the "Halbkreis" artefact.
        margin: 0;

        li {
          // Full-height pill tab — no top-only radius, no translateY overlap.
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

          // Hover: neutral surface tint, icon micro-scale
          &:hover {
            color: var(--color-text);
            background: color-mix(in srgb, var(--color-surface-offset) 60%, transparent);

            i { transform: scale(1.12); }
          }

          // Active press
          &:active {
            transform: scale(0.95);
          }

          // Selected tab: primary color text + underline indicator
          &.active {
            color: var(--color-primary);
            font-weight: 600;

            i { color: var(--color-primary); }

            // 2 px underline flush with the container bottom edge.
            // top-radius prevents a sharp corner where it meets the border.
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

          // Keyboard focus ring
          &:focus-visible {
            outline: 2px solid var(--color-primary);
            outline-offset: -2px;
          }
        }
      }
    }

    // ── Tab content area ────────────────────────────────────────────────────
    // overflow-y: auto allows the bento grid in settings tab to scroll
    // without clipping card shadows or content
    .page-container {
      max-height: 100%;
      flex-grow: 1;
      display: flex;
      flex-direction: column;
      overflow-y: auto;
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
