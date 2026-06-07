<template>
  <div class="chat">
    <ul class="chat-history" ref="history" @click="onClick">
      <template v-for="(message, index) in history">
        <li
          :key="index"
          class="message"
          v-if="message.type === 'text'"
          :class="{
            bulk: index > 0 && history[index - 1].id == message.id && history[index - 1].type === 'text',
          }"
        >
          <div :class="['author', { pilot: activeHostId === message.id }]" @contextmenu.stop.prevent="onContext($event, { member: member(message.id) })">
            <neko-avatar class="avatar" :seed="member(message.id).displayname" :size="40" />
          </div>
          <div class="content">
            <div class="content-head">
              <span>{{ member(message.id).displayname }}</span>
              <span class="timestamp">{{ timestamp(message.created) }}</span>
            </div>
            <neko-markdown class="content-body" :source="message.content" />
          </div>
        </li>
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
      <div class="accent" />
      <div class="text-container">
        <textarea ref="input" :placeholder="$t('send_a_message')" @keydown="onKeyDown" v-model="content" />
        <neko-emoji v-if="emoji" @picked="onEmojiPicked" @done="emoji = false" />
        <i class="emoji-menu fas fa-laugh" @click.stop.prevent="onEmoji"></i>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .chat {
    flex: 1;
    flex-direction: column;
    display: flex;
    max-height: 100%;
    max-width: 100%;
    overflow-x: hidden;
    background: transparent;

    .chat-history {
      flex: 1;
      overflow-y: auto;
      overflow-x: hidden;
      max-width: 100%;
      padding: 10px 14px;
      scrollbar-width: thin;
      scrollbar-color: rgba(255, 255, 255, 0.1) transparent;

      &::-webkit-scrollbar {
        width: 6px;
      }

      &::-webkit-scrollbar-track {
        background-color: transparent;
      }

      &::-webkit-scrollbar-thumb {
        background-color: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
        transition: var(--transition-fluid);

        &:hover {
          background-color: var(--color-cyber-mint);
        }
      }

      ::v-deep *::selection {
        background: var(--color-cyber-mint);
        color: #050508;
      }

      li {
        flex: 1;
        border-top: 1px solid rgba(255, 255, 255, 0.03);
        padding: 12px 8px;
        display: flex;
        flex-direction: row;
        flex-wrap: nowrap;
        overflow: hidden;
        user-select: text;
        word-wrap: break-word;
        animation: messageFadeIn 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;

        &.message {
          font-size: 15px;
          border-radius: var(--radius-inner);
          transition: var(--transition-fluid);

          &:hover {
            background: rgba(255, 255, 255, 0.02);
            border-color: rgba(255, 255, 255, 0.05);
          }

          .author {
            flex-grow: 0;
            flex-shrink: 0;
            overflow: visible; // Prevent clipping the beautiful cyber-mint glow
            width: 38px;
            height: 38px;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.05);
            margin-right: 12px;
            border: 1px solid var(--glass-border);
            transition: var(--transition-fluid);

            &.pilot {
              border-color: var(--color-cyber-mint) !important;
              box-shadow: 0 0 10px var(--color-cyber-mint-glow);
              animation: avatarPulse 2s infinite ease-in-out;
            }

            .avatar {
              width: 100%;
              height: 100%;
              border-radius: 50%; // Keep avatar rounded since parent is visible
            }
          }

          @keyframes avatarPulse {
            0%, 100% { box-shadow: 0 0 4px var(--color-cyber-mint-glow); }
            50% { box-shadow: 0 0 12px var(--color-cyber-mint); }
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
              margin-bottom: 4px;
              display: flex;
              align-items: baseline;
              gap: 6px;

              span {
                display: inline-block;
                color: var(--text-pure);
                font-size: 14px;
                font-weight: 600;
              }

              .timestamp {
                color: var(--text-muted-ok);
                font-size: 11px;
                font-weight: 500;
                line-height: 12px;
                opacity: 0.75;

                &::first-letter {
                  text-transform: uppercase;
                }
              }
            }

            ::v-deep .content-body {
              color: var(--text-subtle);
              line-height: 22px;
              word-wrap: break-word;
              overflow-wrap: break-word;
              font-size: 14.5px;

              a {
                color: var(--color-cyber-mint);
                text-decoration: none;
                transition: var(--transition-fluid);

                &:hover {
                  text-shadow: 0 0 6px var(--color-cyber-mint-glow);
                  text-decoration: underline;
                }
              }

              strong {
                font-weight: 800;
                color: var(--text-pure);
              }

              em {
                font-style: italic;
              }

              blockquote {
                border-left: 3px var(--color-cyber-mint) solid;
                padding-left: 8px;
                background: rgba(255, 255, 255, 0.02);
                margin: 4px 0;
                border-radius: 0 var(--radius-button) var(--radius-button) 0;
              }

              span {
                &.spoiler {
                  background: rgba(255, 255, 255, 0.08);
                  padding: 2px 6px;
                  border-radius: 4px;
                  cursor: pointer;
                  filter: blur(5px);
                  transition: filter 0.4s cubic-bezier(0.16, 1, 0.3, 1), background 0.3s;
                  user-select: none;

                  span {
                    opacity: 0.1;
                    transition: opacity 0.4s ease;
                  }
                }

                &.spoiler.active {
                  background: rgba(255, 255, 255, 0.02);
                  filter: blur(0px);
                  cursor: default;
                  user-select: text;

                  span {
                    opacity: 1;
                  }
                }
              }

              code {
                font-family: inherit;
                background: rgba(12, 13, 18, 0.45);
                border: 1px solid rgba(255, 255, 255, 0.05);
                border-radius: 5px;
                padding: 2px 6px;
                font-size: 13px;
                color: var(--color-cyber-mint);
              }

              pre {
                flex: 1;
                color: var(--text-subtle);
                border: 1px solid rgba(255, 255, 255, 0.05);
                background: rgba(12, 13, 18, 0.45);
                padding: 10px 12px;
                margin: 8px 0;
                border-radius: var(--radius-inner);
                display: block;

                code {
                  display: block;
                  background: transparent;
                  border: none;
                  padding: 0;
                  color: inherit;
                }
              }
            }
          }

          &.bulk {
            padding-top: 2px;
            padding-bottom: 2px;

            .author {
              visibility: hidden;
              height: 0;
              margin-top: 0;
              margin-bottom: 0;
            }

            .content-head {
              display: none;
            }
          }
        }

        &.event {
          color: var(--text-muted-ok);
          font-size: 13px;
          opacity: 0.85;
          padding: 8px 12px;
          background: rgba(255, 255, 255, 0.012);
          border-radius: var(--radius-inner);
          margin: 4px 0;

          .content {
            min-width: 0;
            box-sizing: border-box;
            word-wrap: break-word;
            display: inline-block;
            vertical-align: baseline;
            line-height: 20px;

            strong {
              font-weight: 700;
              color: var(--text-pure);
            }

            i {
              font-style: italic;
              font-size: 11px;
            }
          }
        }
      }
    }

    @keyframes messageFadeIn {
      from {
        opacity: 0;
        transform: translateY(12px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    .chat-send {
      flex-shrink: 0;
      height: auto;
      min-height: 72px;
      padding: 10px 14px 14px 14px;
      flex-direction: column;
      display: flex;
      border-top: 1px solid var(--glass-border);
      background: rgba(12, 13, 18, 0.35);

      .accent {
        display: none;
      }

      .text-container {
        flex: 1;
        width: 100%;
        background-color: rgba(12, 13, 18, 0.45);
        border: 1px solid var(--glass-border);
        border-radius: var(--radius-inner);
        position: relative;
        display: flex;
        align-items: center;
        padding: 4px 8px;
        transition: var(--transition-fluid);

        &:focus-within {
          border-color: var(--color-cyber-mint);
          box-shadow: 0 0 0 3px rgba(38, 230, 180, 0.15);
          background-color: rgba(12, 13, 18, 0.6);
        }

        .emoji-menu {
          width: 32px;
          height: 32px;
          font-size: 18px;
          color: var(--text-muted-ok);
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 50%;
          transition: var(--transition-fluid);

          &:hover {
            color: var(--color-cyber-mint);
            background: rgba(255, 255, 255, 0.05);
            transform: scale(1.1);
          }
        }

        textarea {
          flex: 1;
          font-family: inherit;
          font-size: 14.5px;
          border: none;
          outline: none !important;
          box-shadow: none !important;
          caret-color: var(--color-cyber-mint);
          color: var(--text-pure);
          resize: none;
          margin: 6px 4px;
          background-color: transparent;
          height: 34px;
          line-height: 20px;
          scrollbar-width: none;

          &:focus, &:focus-visible {
            outline: none !important;
            box-shadow: none !important;
          }

          &::placeholder {
            color: var(--text-muted-ok);
            opacity: 0.75;
          }

          &::-webkit-scrollbar {
            display: none;
          }

          &::selection {
            background: var(--color-cyber-mint);
            color: #050508;
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
  export default class NekoChat extends Vue {
    @Ref('input') readonly _input!: HTMLTextAreaElement
    @Ref('history') readonly _history!: HTMLElement
    @Ref('context') readonly _context!: any

    emoji = false
    content = ''

    get activeHostId() {
      return this.$accessor.remote.id
    }

    get id() {
      return this.$accessor.user.id
    }

    get muted() {
      return this.$accessor.user.muted
    }

    get history() {
      return this.$accessor.chat.history
    }

    @Watch('history')
    onHistroyChange() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
      })
    }

    @Watch('muted')
    onMutedChange(muted: boolean) {
      if (muted) {
        this.content = ''
      }
    }

    mounted() {
      this.$nextTick(() => {
        this._history.scrollTop = this._history.scrollHeight
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
      if (!target) {
        return
      }

      if (target.tagName.toLowerCase() === 'span' && target.classList.contains('spoiler')) {
        target.classList.add('active')
        event.preventDefault()
      }

      if (!target.parentElement) {
        return
      }

      if (target.parentElement.tagName.toLowerCase() === 'span' && target.parentElement.classList.contains('spoiler')) {
        target.parentElement.classList.add('active')
        event.preventDefault()
      }
    }

    onKeyDown(event: KeyboardEvent) {
      if (this.muted) {
        return
      }

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

      if (event.keyCode !== 13 || event.shiftKey) {
        return
      }

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
