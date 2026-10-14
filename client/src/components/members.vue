<template>
  <div class="members">
    <div class="members-container">
      <ul class="members-list" :aria-label="$t('side.members')">
        <li v-if="member">
          <div :class="[{ host: member.id === host }, 'self', 'member']" :title="member.displayname">
            <neko-avatar class="avatar" :seed="member.displayname" :avatar="member.avatar" :size="34" />
          </div>
        </li>
        <template v-for="(member, index) in members">
          <li
            v-if="member.id !== id && member.connected"
            :key="index"
            v-tooltip="{ content: member.displayname, html: false, placement: 'top', offset: 5, boundariesElement: 'body' }"
          >
            <button
              type="button"
              :class="['icon-btn', { host: member.id === host, admin: member.admin }, 'member']"
              :aria-label="member.displayname"
              aria-haspopup="menu"
              @contextmenu.stop.prevent="onContext($event, { member })"
              @click.stop.prevent="onContext($event, { member })"
            >
              <neko-avatar class="avatar" :seed="member.displayname" :avatar="member.avatar" :size="34" />
            </button>
          </li>
        </template>
      </ul>
    </div>
    <neko-context ref="context" />
  </div>
</template>

<style lang="scss" scoped>
  .members {
    flex: 1;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
    min-height: 48px;
    display: flex;
    align-items: center;

    &::-webkit-scrollbar {
      height: 4px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background-color: rgba(255, 255, 255, 0.15);
      border-radius: 4px;
      transition: var(--transition-fluid);

      &:hover {
        background-color: var(--color-cyber-mint);
      }
    }

    .members-container {
      display: block;
      clear: both;
      padding: 4px 4px;
      margin: 0;

      .members-list {
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 6px;
        white-space: nowrap;
        list-style: none;
        padding: 0;

        li {
          display: inline-block;
          position: relative;

          .member {
            position: relative;
            display: flex;
            align-items: center;
            justify-content: center;
            width: 38px;
            height: 38px;
            border-radius: 50%;
            padding: 2px;
            background: rgba(255, 255, 255, 0.05);
            border: 1.5px solid rgba(255, 255, 255, 0.1);
            transition: var(--transition-fluid);

            &:hover {
              transform: translateY(-2px);
              border-color: rgba(255, 255, 255, 0.3);
              box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
            }

            &.self {
              &::before {
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f2bd';
                background: var(--bg-obsidian-slate);
                color: var(--color-cyber-mint);
                position: absolute;
                width: 16px;
                height: 16px;
                line-height: 15px;
                font-size: 11px;
                text-align: center;
                top: -4px;
                right: -4px;
                border-radius: 50%;
                border: 1.5px solid var(--glass-border);
                box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
                z-index: 2;
              }
            }

            &.admin {
              &::before {
                display: block;
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f3ed';
                color: #ffb86c;
                background: var(--bg-obsidian-slate);
                position: absolute;
                width: 16px;
                height: 16px;
                line-height: 15px;
                font-size: 10px;
                text-align: center;
                top: -4px;
                right: -4px;
                border: 1.5px solid var(--glass-border);
                border-radius: 50%;
                box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
                z-index: 2;
              }
            }

            &.host {
              border-color: var(--color-cyber-mint);
              box-shadow: 0 0 12px var(--color-cyber-mint-glow);
              animation: auroraPulse 2.5s infinite alternate ease-in-out;

              &::after {
                display: block;
                font-family: 'Font Awesome 6 Free';
                font-weight: 900;
                content: '\f521';
                background: var(--color-cyber-mint);
                color: #050508;
                position: absolute;
                width: 16px;
                height: 16px;
                line-height: 15px;
                font-size: 9px;
                text-align: center;
                bottom: -4px;
                right: -4px;
                border-radius: 50%;
                border: 1.5px solid rgba(255, 255, 255, 0.2);
                box-shadow: 0 2px 6px rgba(38, 230, 180, 0.3);
                z-index: 3;
              }
            }

            .avatar {
              border-radius: 50%;
              overflow: hidden;
              width: 100%;
              height: 100%;
              background: rgba(0, 0, 0, 0.2);
            }
          }

          &:nth-child(2) {
            margin-left: 8px;

            &::before {
              position: absolute;
              content: ' ';
              height: 24px;
              width: 1px;
              background: rgba(255, 255, 255, 0.1);
              top: 7px;
              left: -7px;
            }
          }
        }
      }
    }
  }

  @keyframes auroraPulse {
    0% {
      box-shadow: 0 0 6px rgba(38, 230, 180, 0.2), inset 0 0 4px rgba(38, 230, 180, 0.1);
      border-color: rgba(38, 230, 180, 0.5);
    }
    100% {
      box-shadow: 0 0 16px rgba(38, 230, 180, 0.5), inset 0 0 8px rgba(38, 230, 180, 0.25);
      border-color: var(--color-cyber-mint);
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'

  import Content from './context.vue'
  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-members',
    components: {
      'neko-context': Content,
      'neko-avatar': Avatar,
    },
  })
  export default class NekoMembers extends Vue {
    @Ref('context') readonly _context!: any

    get id() {
      return this.$accessor.user.id
    }

    get host() {
      return this.$accessor.remote.id
    }

    get member() {
      return this.$accessor.user.member
    }

    get members() {
      return this.$accessor.user.members
    }

    onContext(event: MouseEvent, data: any) {
      this._context.open(event, data)
    }
  }
</script>
