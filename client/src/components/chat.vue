<template>
  <div class="chat">
    <!-- ── Skeleton: connecting ──────────────────────────────────────── -->
    <ul
      v-if="loading"
      class="chat-history chat-history--skeleton"
      aria-hidden="true"
    >
      <li v-for="n in 4" :key="'sk' + n" class="skeleton-msg">
        <div class="skeleton sk-avatar"></div>
        <div class="sk-content">
          <div class="skeleton sk-name"></div>
          <div class="skeleton sk-line"></div>
          <div class="skeleton sk-line sk-line--short"></div>
        </div>
      </li>
    </ul>

    <!-- ── Real messages ─────────────────────────────────────────────── -->
    <ul v-else class="chat-history" ref="history" @click="onClick">
      <template v-for="(message, index) in history">
        <!-- Text message -->
        <li
          :key="index"
          class="message"
          v-if="message.type === 'text'"
          :class="{
            bulk: index > 0 && history[index - 1].id == message.id && history[index - 1].type === 'text',
          }"
        >
          <div class="author" @contextmenu.stop.prevent="onContext($event, { member: member(message.id) })">
            <neko-avatar class="avatar" :seed="member(message.id).displayname" :size="40" />
          </div>
          <div class="content">
            <div class="content-head">
              <!-- Pill-shaped username badge -->
              <span class="username">{{ member(message.id).displayname }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <neko-markdown class="content-body" :source="message.content" />
          </div>
        </li>
        <!-- Event message -->
        <li :key="index" class="event" v-if="message.type === 'event'">
          <div
            class="content"
            v-tooltip="{
              content: timestamp(message.created),
              placement: 'left',
              offset: 3,
              boundariesElement: 'body',
            }"
          >
            <strong v-if="message.id === id && $te('you')">{{ $t('you') }}</strong>
            <strong v-else>{{ member(message.id).displayname }}</strong>
            {{ message.content }}
          </div>
        </li>
      </template>
    </ul>

    <neko-context ref="context" />
    <div v-if="!muted" class="chat-send">
      <div class="divider" />
      <div class="text-container">
        <textarea
          ref="input"
          :placeholder="$t('send_a_message')"
          @keydown="onKeyDown"
          v-model="content"
        />
        <neko-emoji v-if="emoji" @picked="onEmojiPicked" @done="emoji = false" />
        <i class="emoji-menu fas fa-laugh" @click.stop.prevent="onEmoji" aria-label="Pick emoji" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  // ── Shimmer animation ────────────────────────────────────────────────────
  @keyframes shimmer {
    0%   { background-position: -200% 0; }
    100% { background-position:  200% 0; }
  }

  .skeleton {
    background: linear-gradient(
      90deg,
      var(--color-surface-offset)  25%,
      var(--color-surface-dynamic) 50%,
      var(--color-surface-offset)  75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s ease-in-out infinite;
    border-radius: var(--radius-sm);
  }

  // ── Skeleton message row ─────────────────────────────────────────────────
  .skeleton-msg {
    display: flex;
    flex-direction: row;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-3) var(--space-3);
    border-top: 1px solid var(--color-divider);

    .sk-avatar {
      flex-shrink: 0;
      width: 40px;
      height: 40px;
      border-radius: var(--radius-full);
    }

    .sk-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: var(--space-2);
      padding-top: var(--space-1);

      .sk-name {
        height: 12px;
        width: 72px;
        border-radius: var(--radius-full);
      }

      .sk-line {
        height: 11px;
        width: 100%;

        &--short { width: 58%; }
      }
    }
  }

  .chat {
    flex: 1;
    flex-direction: column;
    display: flex;
    max-height: 100%;
    max-width: 100%;
    overflow-x: hidden;

    // ── Message history ────────────────────────────────────────────────────
    .chat-history {
      flex: 1;
      overflow-y: scroll;
      overflow-x: hidden;
      max-width: 100%;
      scrollbar-width: thin;
      scrollbar-color: var(--color-surface-offset) transparent;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: var(--color-surface-offset);
        border-radius: var(--radius-full);

        &:hover {
          background-color: var(--color-surface-dynamic);
        }
      }

      ::v-deep *::selection {
        background: var(--color-primary-highlight);
        color: var(--color-text);
      }

      li {
        flex: 1;
        border-top: 1px solid var(--color-divider);
        padding: var(--space-2) var(--space-2) 0 var(--space-3);
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        overflow: hidden;
        user-select: text;
        word-wrap: break-word;
        transition: background-color var(--transition-fast);

        &:hover {
          background-color: var(--color-surface-offset);
        }

        // ── Text message ─────────────────────────────────────────
        &.message {
          padding-top: var(--space-4);
          font-size: var(--text-base);

          .author {
            flex-grow: 0;
            flex-shrink: 0;
            overflow: hidden;
            width: 40px;
            height: 40px;
            border-radius: var(--radius-full);
            background: var(--color-primary-highlight);
            margin-right: var(--space-3);
            transition: transform var(--transition-interactive);
            cursor: pointer;

            &:hover {
              transform: scale(1.08);
            }

            .avatar {
              width: 100%;
            }
          }

          .content {
            flex: 1;
            display: flex;
            flex-direction: column;
            box-sizing: border-box;
            word-wrap: break-word;
            min-width: 0;

            .content-head {
              cursor: default;
              width: 100%;
              margin-bottom: var(--space-1);
              display: flex;
              align-items: baseline;
              gap: var(--space-2);

              // Pill-shaped username badge
              .username {
                display: inline-flex;
                align-items: center;
                padding: 0 var(--space-2);
                height: 18px;
                border-radius: var(--radius-full);
                background: var(--color-primary-highlight);
                color: var(--color-primary);
                font-size: var(--text-xs);
                font-weight: 600;
                font-family: var(--font-body);
                letter-spacing: 0.01em;
                transition:
                  background-color var(--transition-interactive),
                  color            var(--transition-interactive);

                &:hover {
                  background: var(--color-primary);
                  color: var(--color-bg);
                }
              }

              .timestamp {
                color: var(--color-text-faint);
                font-size: var(--text-xs);
                font-weight: 400;
                line-height: 1;

                &::first-letter {
                  text-transform: uppercase;
                }
              }
            }

            ::v-deep .content-body {
              color: var(--color-text);
              font-size: var(--text-sm);
              line-height: 1.55;
              word-wrap: break-word;
              overflow-wrap: break-word;
              padding-bottom: var(--space-2);

              a {
                color: var(--color-link);
                text-underline-offset: 2px;

                &:hover {
                  color: var(--color-primary);
                }
              }

              strong { font-weight: 700; }
              em     { font-style: italic; }

              // Blockquote with teal accent line
              blockquote {
                border-left: 3px solid var(--color-primary);
                padding-left: var(--space-3);
                margin: var(--space-1) 0;
                color: var(--color-text-muted);
              }

              // Spoiler
              span {
                &.spoiler {
                  background: var(--color-surface-offset);
                  padding: 0 var(--space-1);
                  border-radius: var(--radius-sm);
                  cursor: pointer;
                  transition: background-color var(--transition-interactive);

                  span { opacity: 0; transition: opacity var(--transition-interactive); }
                }

                &.spoiler.active {
                  background: var(--color-surface-dynamic);
                  cursor: default;

                  span { opacity: 1; }
                }
              }

              // Inline code
              code {
                font-family: var(--font-mono);
                background: var(--color-surface-offset);
                border-radius: var(--radius-sm);
                padding: 1px var(--space-1);
                font-size: 0.85em;
                line-height: 1.4;
                white-space: pre-wrap;
              }

              // Code block
              pre {
                background: var(--color-surface-offset);
                border: 1px solid var(--color-border);
                border-radius: var(--radius-md);
                padding: var(--space-3) var(--space-4);
                margin: var(--space-2) 0;
                display: block;
                overflow-x: auto;
                color: var(--color-text);

                code {
                  background: transparent;
                  padding: 0;
                  border-radius: 0;
                  display: block;
                  font-size: var(--text-xs);
                  line-height: 1.6;
                }
              }
            }
          }

          // Bulk: hide repeated avatar + header
          &.bulk {
            padding-top: 0;

            .author {
              visibility: hidden;
              height: 0;
            }

            .content-head {
              display: none;
            }
          }
        }

        // ── Event message ─────────────────────────────────────────
        &.event {
          color: var(--color-text-muted);
          font-size: var(--text-sm);
          cursor: default;

          .content {
            min-width: 0;
            box-sizing: border-box;
            word-wrap: break-word;
            display: inline-block;
            vertical-align: baseline;
            line-height: 1.5;

            strong { font-weight: 600; color: var(--color-text); }
            i      { font-style: italic; font-size: var(--text-xs); }
          }
        }
      }
    }

    // ── Send input area ───────────────────────────────────────────────
    .chat-send {
      flex-shrink: 0;
      padding: 0 var(--space-3) var(--space-3);
      flex-direction: column;
      display: flex;

      .divider {
        width: 100%;
        height: 1px;
        background: var(--color-divider);
        margin: var(--space-2) 0;
      }

      .text-container {
        flex: 1;
        width: 100%;
        min-height: 60px;
        background-color: var(--color-surface-offset);
        border: 1px solid var(--color-border);
        border-radius: var(--radius-lg);
        position: relative;
        display: flex;
        align-items: flex-start;
        transition:
          border-color      var(--transition-interactive),
          background-color  var(--transition-interactive);

        &:focus-within {
          border-color: var(--color-primary);
          background-color: var(--color-surface-2);
        }

        .emoji-menu {
          width: 36px;
          height: 36px;
          font-size: 18px;
          display: flex;
          align-items: center;
          justify-content: center;
          margin: var(--space-1) var(--space-1) 0 0;
          flex-shrink: 0;
          cursor: pointer;
          color: var(--color-text-muted);
          border-radius: var(--radius-md);
          transition:
            color            var(--transition-interactive),
            background-color var(--transition-interactive),
            transform        var(--transition-interactive);

          &:hover {
            color: var(--color-primary);
            background-color: var(--color-primary-highlight);
            transform: scale(1.1);
          }

          &:active {
            transform: scale(0.92);
          }
        }

        textarea {
          flex: 1;
          font-family: var(--font-body);
          font-size: var(--text-sm);
          border: none;
          outline: none;
          caret-color: var(--color-primary);
          color: var(--color-text);
          resize: none;
          padding: var(--space-2) var(--space-3);
          background-color: transparent;
          line-height: 1.5;
          scrollbar-width: thin;
          scrollbar-color: var(--color-surface-offset) transparent;

          &::placeholder {
            color: var(--color-text-faint);
          }

          &::-webkit-scrollbar        { width: 4px; }
          &::-webkit-scrollbar-track  { background-color: transparent; }
          &::-webkit-scrollbar-thumb  {
            background-color: var(--color-surface-offset);
            border-radius: var(--radius-full);
          }

          &::selection {
            background: var(--color-primary-highlight);
            color: var(--color-text);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Watch, Vue } from 'vue-property-decorator'
  import { formatRelative } from 'date-fns'

  import { Member } from '~/neko/types'

  import Markdown from './markdown'
  import Content from './context.vue'
  import Emoji from './emoji.vue'
  import Avatar from './avatar.vue'

  const length = 512 // max length of message

  @Component({
    name: 'neko-chat',
    components: {
      'neko-markdown': Markdown,
      'neko-context': Content,
      'neko-emoji': Emoji,
      'neko-avatar': Avatar,
    },
  })
  export default class extends Vue {
    @Ref('input') readonly _input!: HTMLTextAreaElement
    @Ref('history') readonly _history!: HTMLElement
    @Ref('context') readonly _context!: any

    emoji = false
    content = ''

    get id() {
      return this.$accessor.user.id
    }

    get muted() {
      return this.$accessor.user.muted
    }

    get history() {
      return this.$accessor.chat.history
    }

    // True while the WebSocket connection is being established
    get loading() {
      return this.$accessor.connecting
    }

    @Watch('history')
    onHistroyChange() {
      this.$nextTick(() => {
        if (this._history) this._history.scrollTop = this._history.scrollHeight
      })
    }

    // Once loading finishes, scroll chat to bottom
    @Watch('loading')
    onLoadingChange(val: boolean) {
      if (!val) {
        this.$nextTick(() => {
          if (this._history) this._history.scrollTop = this._history.scrollHeight
        })
      }
    }

    @Watch('muted')
    onMutedChange(muted: boolean) {
      if (muted) {
        this.content = ''
      }
    }

    mounted() {
      this.$nextTick(() => {
        if (this._history) this._history.scrollTop = this._history.scrollHeight
      })
    }

    member(id: string) {
      return this.$accessor.user.members[id] || { id, displayname: this.$t('somebody') }
    }

    timestamp(time: Date) {
      const str = formatRelative(time, new Date())
      return `${str.charAt(0).toUpperCase()}${str.slice(1)}`
    }

    onEmoji() {
      this.emoji = !this.emoji
      this._input.focus()
    }

    onEmojiPicked(emoji: string) {
      const text = `:${emoji}:`
      if (this._input.selectionStart || this._input.selectionStart === 0) {
        var startPos = this._input.selectionStart
        var endPos = this._input.selectionEnd
        this.content = this.content.substring(0, startPos) + text + this.content.substring(endPos, this.content.length)
        this.$nextTick(() => {
          this._input.selectionStart = startPos + text.length
          this._input.selectionEnd = startPos + text.length
        })
      } else {
        this.content += text
      }
      this._input.focus()
      this.emoji = false
    }

    onContext(event: MouseEvent, { member }: { member: Member }) {
      if (member.id === this.id) {
        return
      }
      this._context.open(event, { member })
    }

    onClick(event: { target?: HTMLElement; preventDefault(): void }) {
      const { target } = event
      if (!target) return

      if (target.tagName.toLowerCase() === 'span' && target.classList.contains('spoiler')) {
        target.classList.add('active')
        event.preventDefault()
      }

      if (!target.parentElement) return

      if (
        target.parentElement.tagName.toLowerCase() === 'span' &&
        target.parentElement.classList.contains('spoiler')
      ) {
        target.parentElement.classList.add('active')
        event.preventDefault()
      }
    }

    onKeyDown(event: KeyboardEvent) {
      if (this.muted) return

      if (this.content.length > length) {
        this.content = this.content.substring(0, length)
      }

      if (this.content.length == length) {
        if (
          [8, 16, 17, 18, 20, 33, 34, 35, 36, 37, 38, 39, 40, 45, 46, 91, 93, 144].includes(event.keyCode) ||
          (event.ctrlKey && [67, 65, 88].includes(event.keyCode))
        ) {
          return
        }
        event.preventDefault()
        return
      }

      if (event.keyCode !== 13 || event.shiftKey) return
      if (this.content === '') {
        event.preventDefault()
        return
      }

      this.$accessor.chat.sendMessage(this.content)
      this.content = ''
      event.preventDefault()
    }
  }
</script>
