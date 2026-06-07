<template>
  <div class="side-members">
    <div class="members-header">
      <span>{{ $t('side.members') }} ({{ totalConnected }})</span>
    </div>
    <div class="members-list-wrapper">
      <ul class="members-list">
        <!-- Self / Current User -->
        <li
          v-if="myMember"
          class="member-item"
          :class="{ host: myMember.id === host, admin: myMember.admin }"
          @contextmenu.stop.prevent="onContext($event, { member: myMember })"
          @click.stop.prevent="onContext($event, { member: myMember })"
        >
          <div class="avatar-container">
            <neko-avatar class="avatar" :seed="myMember.displayname" :size="38" />
          </div>
          <div class="member-info">
            <span class="displayname">
              {{ myMember.displayname }} <span class="self-tag">({{ $t('you') }})</span>
            </span>
          </div>
          <div class="member-badges">
            <span v-if="myMember.id === host" class="badge host-badge" v-tooltip="{ content: 'Host', placement: 'top' }">
              <i class="fas fa-crown" />
            </span>
            <span v-if="myMember.admin" class="badge admin-badge" v-tooltip="{ content: 'Admin', placement: 'top' }">
              <i class="fas fa-shield-alt" />
            </span>
            <span v-if="myMember.muted" class="badge mute-badge" v-tooltip="{ content: 'Muted', placement: 'top' }">
              <i class="fas fa-microphone-slash" />
            </span>
          </div>
        </li>

        <!-- Other Users -->
        <template v-for="(memberItem, index) in connectedMembers">
          <li
            :key="index"
            class="member-item"
            :class="{ host: memberItem.id === host, admin: memberItem.admin }"
            @contextmenu.stop.prevent="onContext($event, { member: memberItem })"
            @click.stop.prevent="onContext($event, { member: memberItem })"
          >
            <div class="avatar-container">
              <neko-avatar class="avatar" :seed="memberItem.displayname" :size="38" />
            </div>
            <div class="member-info">
              <span class="displayname">{{ memberItem.displayname }}</span>
            </div>
            <div class="member-badges">
              <span v-if="memberItem.id === host" class="badge host-badge" v-tooltip="{ content: 'Host', placement: 'top' }">
                <i class="fas fa-crown" />
              </span>
              <span v-if="memberItem.admin" class="badge admin-badge" v-tooltip="{ content: 'Admin', placement: 'top' }">
                <i class="fas fa-shield-alt" />
              </span>
              <span v-if="memberItem.muted" class="badge mute-badge" v-tooltip="{ content: 'Muted', placement: 'top' }">
                <i class="fas fa-microphone-slash" />
              </span>
              <span v-if="memberItem.ignored" class="badge ignore-badge" v-tooltip="{ content: 'Ignored', placement: 'top' }">
                <i class="fas fa-eye-slash" />
              </span>
            </div>
          </li>
        </template>
      </ul>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .side-members {
    display: flex;
    flex-direction: column;
    flex: 1;
    height: 100%;
    max-height: 100%;
    overflow: hidden;

    .members-header {
      padding: 12px 16px;
      font-size: 12px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-muted-ok);
      border-bottom: 1px solid rgba(255, 255, 255, 0.03);
    }

    .members-list-wrapper {
      flex: 1;
      overflow-y: auto;
      overflow-x: hidden;
      padding: 8px 12px;
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
    }

    .members-list {
      display: flex;
      flex-direction: column;
      gap: 6px;
      list-style: none;
      padding: 0;
      margin: 0;
    }

    .member-item {
      display: flex;
      flex-direction: row;
      align-items: center;
      padding: 8px 10px;
      border-radius: var(--radius-inner);
      cursor: pointer;
      user-select: none;
      border: 1px solid transparent;
      background: rgba(255, 255, 255, 0.01);
      transition: var(--transition-fluid);

      &:hover {
        background: rgba(255, 255, 255, 0.03);
        border-color: rgba(255, 255, 255, 0.05);
      }

      &.host {
        background: rgba(38, 230, 180, 0.02);
        border-color: rgba(38, 230, 180, 0.1);

        &:hover {
          background: rgba(38, 230, 180, 0.04);
          border-color: rgba(38, 230, 180, 0.2);
        }
      }

      .avatar-container {
        flex-shrink: 0;
        width: 38px;
        height: 38px;
        border-radius: 50%;
        margin-right: 12px;
        border: 1px solid var(--glass-border);
        overflow: hidden;
        background: rgba(255, 255, 255, 0.05);
        transition: var(--transition-fluid);
      }

      &.host .avatar-container {
        border-color: var(--color-cyber-mint);
        box-shadow: 0 0 10px var(--color-cyber-mint-glow);
      }

      .member-info {
        flex-grow: 1;
        min-width: 0;

        .displayname {
          font-size: 14.5px;
          font-weight: 600;
          color: var(--text-pure);
          text-overflow: ellipsis;
          overflow: hidden;
          white-space: nowrap;
          display: block;

          .self-tag {
            font-size: 11px;
            color: var(--color-cyber-mint);
            font-weight: 500;
            margin-left: 4px;
          }
        }
      }

      .member-badges {
        display: flex;
        flex-direction: row;
        align-items: center;
        gap: 8px;
        flex-shrink: 0;

        .badge {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 20px;
          height: 20px;
          border-radius: 50%;
          font-size: 10px;

          &.host-badge {
            color: var(--color-cyber-mint);
            background: rgba(38, 230, 180, 0.1);
          }

          &.admin-badge {
            color: #ffb86c;
            background: rgba(255, 184, 108, 0.1);
          }

          &.mute-badge {
            color: #ff4a5a;
            background: rgba(255, 74, 90, 0.1);
          }

          &.ignore-badge {
            color: var(--text-muted-ok);
            background: rgba(255, 255, 255, 0.05);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'
  import { Member } from '~/neko/types'

  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-side-members',
    components: {
      'neko-avatar': Avatar,
    },
  })
  export default class NekoSideMembers extends Vue {
    get id() {
      return this.$accessor.user.id
    }

    get host() {
      return this.$accessor.remote.id
    }

    get myMember() {
      return this.$accessor.user.member
    }

    get members() {
      return this.$accessor.user.members
    }

    get connectedMembers() {
      return Object.values(this.members).filter((m) => m.id !== this.id && m.connected)
    }

    get totalConnected() {
      let count = 0
      if (this.myMember) count++
      count += this.connectedMembers.length
      return count
    }

    onContext(event: MouseEvent, data: { member: Member }) {
      const parent = this.$parent as any
      if (parent && typeof parent.openContext === 'function') {
        parent.openContext(event, data)
      }
    }
  }
</script>
