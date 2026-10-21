<template>
  <div class="emotes" @mouseleave="stopSendingEmotes" @mouseup="stopSendingEmotes">
    <ul v-if="!muted">
      <template v-if="!compact">
        <li v-for="emote in recent" :key="emote">
          <button
            type="button"
            class="icon-btn recent"
            :aria-label="emote"
            @mousedown.stop.prevent="startSendingEmotes(emote)"
            @keydown.enter.space.prevent="sendEmote(emote)"
          >
            <span :class="['emote', emote]" aria-hidden="true" />
          </button>
        </li>
      </template>
      <li>
        <button
          type="button"
          :class="['icon-btn', 'open', { boxed: compact }]"
          :aria-label="$t('controls.emotes')"
          aria-haspopup="true"
          v-tooltip="{ content: $t('controls.emotes'), placement: 'top', offset: 5, boundariesElement: 'body' }"
          @click.stop.prevent="open"
        >
          <i class="far fa-grin-beam" aria-hidden="true"></i>
        </button>
      </li>
    </ul>
    <vue-context class="context" ref="context">
      <li v-for="emote in emotes" :key="emote">
        <button type="button" class="icon-btn" :aria-label="emote" @click="sendEmote(emote)">
          <span :class="['emote', emote]" aria-hidden="true" />
        </button>
      </li>
    </vue-context>
  </div>
</template>

<style lang="scss" scoped>
  .emotes {
    ul {
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;

      li {
        font-size: 24px;
        margin: 0 3px;

        .recent {
          width: 34px;
          height: 34px;
          border-radius: 10px;

          &:hover {
            background: rgba(255, 255, 255, 0.06);
          }

          .emote {
            display: block;
            width: 24px;
            height: 24px;
          }
        }

        .open {
          width: 40px;
          height: 40px;
          border-radius: var(--radius-inner);
          border: 1px solid var(--glass-border);
          background: rgba(255, 255, 255, 0.04);
          color: var(--text-subtle);
          font-size: 18px;
          transition: var(--transition-fluid);

          &:hover {
            color: var(--color-cyber-mint);
            background: rgba(255, 255, 255, 0.08);
          }

          &.boxed {
            width: 44px;
            height: 44px;
          }
        }
      }
    }

    .context {
      background-color: $background-floating;
      background-clip: padding-box;
      border-radius: 0.25rem;
      display: flex;
      margin: 0;
      padding: 5px;
      width: 220px;
      z-index: 1500;
      position: fixed;
      list-style: none;
      box-sizing: border-box;
      max-height: calc(100% - 50px);
      color: $interactive-normal;
      flex-wrap: wrap;
      user-select: none;
      box-shadow: $elevation-high;

      > li {
        margin: 0;
        position: relative;
        align-content: center;
        padding: 5px;
        border-radius: 3px;

        .icon-btn {
          display: block;
        }

        .emote {
          display: block;
          width: 24px;
          height: 24px;
        }

        &:hover,
        &:focus {
          text-decoration: none;
          background-color: $background-modifier-hover;
          color: $interactive-hover;
        }

        &:focus {
          outline: 0;
        }
      }

      &:focus {
        outline: 0;
      }
    }
  }
</style>

<script lang="ts">
  import { Vue, Ref, Component, Prop } from 'vue-property-decorator'
  import { getList, set } from '../utils/localstorage'

  // @ts-ignore
  import { VueContext } from 'vue-context'

  @Component({
    name: 'neko-emotes',
    components: {
      'vue-context': VueContext,
    },
  })
  export default class NekoEmotes extends Vue {
    @Ref('context') readonly context!: any
    @Prop(Boolean) readonly compact!: boolean
    recent: string[] = getList('emote_recent')

    get emotes() {
      return [
        'anger',
        'bomb',
        'sleep',
        'explode',
        'sweat',
        'poo',
        'hundred',
        'alert',
        'punch',
        'wave',
        'okay',
        'thumbs-up',
        'clap',
        'prey',
        'celebrate',
        'flame',
        'goof',
        'love',
        'cool',
        'smerk',
        'worry',
        'ouch',
        'cry',
        'surprised',
        'quiet',
        'rage',
        'annoy',
        'steamed',
        'scared',
        'terrified',
        'sleepy',
        'dead',
        'happy',
        'roll-eyes',
        'thinking',
        'clown',
        'sick',
        'rofl',
        'drule',
        'sniff',
        'sus',
        'party',
        'odd',
        'hot',
        'cold',
        'blush',
        'sad',
      ].filter((v) => !this.recent.includes(v))
    }

    get muted() {
      return this.$accessor.user.muted
    }

    open(event: MouseEvent) {
      this.context.open(event)
    }

    sendEmote(emote: string) {
      if (!this.recent.includes(emote)) {
        if (this.recent.length > 4) {
          this.recent.shift()
        }
        this.recent.push(emote)
        set('emote_recent', JSON.stringify(this.recent))
      }
      this.$accessor.chat.sendEmote(emote)
    }

    private interval!: number

    startSendingEmotes(emote: string) {
      this.$accessor.chat.sendEmote(emote)
      this.stopSendingEmotes()

      this.interval = window.setInterval(() => {
        this.$accessor.chat.sendEmote(emote)
      }, 350)
    }

    stopSendingEmotes() {
      if (this.interval) {
        clearInterval(this.interval)
      }
    }
  }
</script>
