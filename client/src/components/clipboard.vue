<template>
  <div class="clipboard-panel" v-if="opened" @click="$event.stopPropagation()">
    <div class="panel-header">
      <span class="panel-title">
        <i class="fas fa-history icon" />
        {{ $t('clipboard_manager.history_title') }}
      </span>
      <button class="close-btn" @click.stop.prevent="close">
        <i class="fas fa-times" />
      </button>
    </div>

    <!-- Active Manual TextArea -->
    <div class="manual-input-section">
      <textarea
        ref="textarea"
        v-model="clipboard"
        :placeholder="$t('clipboard_manager.manual_placeholder')"
        @focus="$event.target.select()"
      />
    </div>

    <!-- History list -->
    <div class="history-section">
      <div class="history-header">
        <span class="history-count" v-if="history.length > 0">
          {{ history.length }} {{ history.length === 1 ? 'Item' : 'Items' }}
        </span>
        <button v-if="history.length > 0" class="clear-btn" @click.stop.prevent="clearHistory">
          <i class="fas fa-trash-alt icon" />
          {{ $t('clipboard_manager.clear') }}
        </button>
      </div>

      <div class="history-list-wrapper">
        <p v-if="history.length === 0" class="empty-state">
          {{ $t('clipboard_manager.empty') }}
        </p>
        <ul v-else class="history-list">
          <li v-for="(item, index) in history" :key="index" class="history-item">
            <span class="item-text" :title="item">{{ truncate(item) }}</span>
            <div class="item-actions">
              <!-- Copy & Sync to Local + VM -->
              <button
                class="action-btn"
                v-tooltip="{ content: $t('clipboard_manager.copy_to_host'), placement: 'top', offset: 5 }"
                @click.stop.prevent="copyToHost(item)"
              >
                <i :class="['fas', copiedText === item ? 'fa-check success-check' : 'fa-copy']" />
                <span class="toast-indicator" v-if="copiedText === item">Copied</span>
              </button>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <!-- Hint footer info -->
    <div class="clipboard-hint">
      <i class="fas fa-info-circle icon" />
      <span>{{ $t('clipboard_manager.sync_hint') }}</span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .clipboard-panel {
    background-color: var(--bg-obsidian-slate);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-outer);
    display: flex;
    flex-direction: column;
    padding: 12px;
    position: absolute;
    bottom: 20px;
    right: 20px;
    width: 340px;
    max-height: 400px;
    height: auto;
    box-shadow: $elevation-high;
    z-index: 1000;
    transition: var(--transition-fluid);
    backdrop-filter: blur(12px);

    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 10px;
      padding-bottom: 8px;
      border-bottom: 1px solid var(--glass-border);

      .panel-title {
        font-weight: 600;
        font-size: 13px;
        color: var(--text-pure);
        display: flex;
        align-items: center;
        gap: 6px;

        .icon {
          color: var(--color-cyber-mint);
          font-size: 12px;
        }
      }

      .close-btn {
        background: none;
        border: none;
        color: var(--text-muted-ok);
        cursor: pointer;
        font-size: 14px;
        transition: var(--transition-fluid);
        padding: 4px;

        &:hover {
          color: var(--text-pure);
        }
      }
    }

    .manual-input-section {
      width: 100%;
      margin-bottom: 12px;

      textarea {
        width: 100%;
        height: 80px;
        min-height: 80px;
        max-height: 120px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.2);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        color: var(--text-pure);
        font-size: 12px;
        font-family: inherit;
        resize: vertical;
        outline: none;
        transition: var(--transition-fluid);

        &:focus {
          border-color: var(--color-cyber-mint);
          box-shadow: 0 0 8px var(--color-cyber-mint-glow);
        }

        &::selection {
          background: var(--color-cyber-mint);
          color: #000;
        }
      }
    }

    .history-section {
      display: flex;
      flex-direction: column;
      flex-grow: 1;
      min-height: 0;

      .history-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .history-count {
          font-size: 11px;
          color: var(--text-muted-ok);
          font-weight: 500;
        }

        .clear-btn {
          background: none;
          border: none;
          color: var(--text-muted-ok);
          font-size: 11px;
          cursor: pointer;
          display: flex;
          align-items: center;
          gap: 4px;
          padding: 2px 6px;
          border-radius: var(--radius-button);
          transition: var(--transition-fluid);

          &:hover {
            color: $style-error;
            background: rgba(255, 74, 90, 0.1);
          }

          .icon {
            font-size: 10px;
          }
        }
      }

      .history-list-wrapper {
        overflow-y: auto;
        max-height: 180px;
        min-height: 60px;
        border-radius: var(--radius-inner);
        background: rgba(0, 0, 0, 0.1);
        padding: 4px;

        &::-webkit-scrollbar {
          width: 4px;
        }
        &::-webkit-scrollbar-track {
          background: transparent;
        }
        &::-webkit-scrollbar-thumb {
          background: var(--glass-border);
          border-radius: 2px;
        }

        .empty-state {
          font-size: 11px;
          color: var(--text-muted-ok);
          text-align: center;
          padding: 20px 0;
          font-style: italic;
        }

        .history-list {
          list-style: none;
          padding: 0;
          margin: 0;
          display: flex;
          flex-direction: column;
          gap: 4px;

          .history-item {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 6px 8px;
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid transparent;
            border-radius: var(--radius-button);
            transition: var(--transition-fluid);

            &:hover {
              background: rgba(255, 255, 255, 0.04);
              border-color: var(--glass-border);
            }

            .item-text {
              font-size: 11px;
              color: var(--text-subtle);
              white-space: nowrap;
              overflow: hidden;
              text-overflow: ellipsis;
              flex-grow: 1;
              margin-right: 8px;
              user-select: text;
            }

            .item-actions {
              display: flex;
              gap: 4px;
              align-items: center;

              .action-btn {
                background: rgba(255, 255, 255, 0.04);
                border: 1px solid var(--glass-border);
                color: var(--text-subtle);
                border-radius: var(--radius-button);
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
                transition: var(--transition-fluid);
                font-size: 10px;
                position: relative;

                &:hover {
                  color: var(--color-cyber-mint);
                  border-color: var(--color-cyber-mint);
                  background: var(--color-cyber-mint-glow);
                }

                &.send-btn:hover {
                  color: var(--color-cyber-mint);
                }

                .success-check {
                  color: var(--color-cyber-mint);
                }

                .toast-indicator {
                  position: absolute;
                  bottom: calc(100% + 6px);
                  right: 50%;
                  transform: translateX(50%);
                  background: var(--bg-obsidian-floating);
                  border: 1px solid var(--color-cyber-mint);
                  color: var(--color-cyber-mint);
                  padding: 2px 6px;
                  font-size: 9px;
                  border-radius: 4px;
                  white-space: nowrap;
                  pointer-events: none;
                }
              }
            }
          }
        }
      }
    }

    .clipboard-hint {
      display: flex;
      align-items: flex-start;
      gap: 6px;
      margin-top: 10px;
      padding-top: 8px;
      border-top: 1px dashed rgba(255, 255, 255, 0.08);
      font-size: 10px;
      color: var(--text-muted-ok);
      line-height: 1.4;

      .icon {
        color: var(--color-cyber-mint);
        margin-top: 2px;
        font-size: 10px;
        flex-shrink: 0;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'

  @Component({
    name: 'neko-clipboard',
  })
  export default class NekoClipboard extends Vue {
    @Ref('textarea') readonly _textarea!: HTMLTextAreaElement

    private opened: boolean = false
    private typing: number | null = null
    private copiedText: string = ''

    get clipboard() {
      return this.$accessor.remote.clipboard
    }

    set clipboard(data: string) {
      this.$accessor.remote.setClipboard(data)

      if (this.typing) {
        clearTimeout(this.typing)
        this.typing = null
      }

      this.typing = window.setTimeout(() => this.$accessor.remote.sendClipboard(this.clipboard), 500)
    }

    get history() {
      return this.$accessor.remote.clipboardHistory
    }

    open() {
      this.opened = true
      document.body.addEventListener('click', this.close)
      window.setTimeout(() => {
        if (this._textarea) {
          this._textarea.focus()
        }
      }, 0)
    }

    close() {
      this.opened = false
      document.body.removeEventListener('click', this.close)
    }

    async copyToHost(text: string) {
      try {
        // Automatically sync to VM's clipboard store first
        this.$accessor.remote.setClipboard(text)
        this.$accessor.remote.sendClipboard(text)

        // Then copy to native client's system clipboard
        await navigator.clipboard.writeText(text)
        this.copiedText = text
        setTimeout(() => {
          if (this.copiedText === text) {
            this.copiedText = ''
          }
        }, 1500)
      } catch (err: any) {
        this.$log.error('Could not copy to local clipboard: ', err)
      }
    }

    clearHistory() {
      this.$accessor.remote.clearClipboardHistory()
    }

    truncate(text: string, length = 45) {
      if (text.length <= length) return text
      return text.substring(0, length) + '...'
    }
  }
</script>
