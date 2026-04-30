<template>
  <div class="members">
    <div class="members-container">
      <ul class="members-list">
        <!-- Self -->
        <li v-if="member">
          <div :class="[{ host: member.id === host }, 'self', 'member']">
            <neko-avatar class="avatar" :seed="member.displayname" :size="50" />
            <!-- Status dot: always online for self -->
            <span class="status-dot status-online" aria-label="Online" />
          </div>
        </li>
        <!-- Other members -->
        <template v-for="(member, index) in members">
          <li
            v-if="member.id !== id && member.connected"
            :key="index"
            v-tooltip="{ content: member.displayname, placement: 'bottom', offset: -15, boundariesElement: 'body' }"
          >
            <div
              :class="[{ host: member.id === host, admin: member.admin }, 'member']"
              @contextmenu.stop.prevent="onContext($event, { member })"
            >
              <neko-avatar class="avatar" :seed="member.displayname" :size="50" />
              <!-- Status dot: green = connected -->
              <span class="status-dot status-online" aria-label="Online" />
            </div>
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
    padding-bottom: var(--space-3);
    scrollbar-width: thin;
    scrollbar-color: var(--color-surface-offset) var(--color-bg);
    min-height: 60px;
    display: flex;

    &::-webkit-scrollbar {
      height: 4px;
    }

    &::-webkit-scrollbar-track {
      background-color: var(--color-bg);
    }

    &::-webkit-scrollbar-thumb {
      background-color: var(--color-surface-offset);
      border-radius: var(--radius-full);

      &:hover {
        background-color: var(--color-surface-dynamic);
      }
    }

    .members-container {
      display: block;
      clear: both;
      padding: 0 var(--space-5);
      margin: 0 auto;

      .members-list {
        white-space: nowrap;
        clear: both;

        li {
          display: inline-block;

          // Separator line before the second item (first other member)
          &:nth-child(2) {
            margin-left: var(--space-5);

            &::before {
              position: absolute;
              content: '';
              height: 45px;
              width: 1px;
              background: var(--color-border);
              margin-top: var(--space-3);
              margin-left: calc(var(--space-5) * -1 + var(--space-1));
            }
          }

          .member {
            position: relative;
            display: block;
            width: 50px;
            height: 50px;
            margin: var(--space-2) var(--space-1);
            // Hover lift animation
            transition: transform var(--transition-interactive);
            cursor: default;

            &:hover {
              transform: translateY(-2px) scale(1.06);
            }

            &:active {
              transform: scale(0.95);
            }

            // ── Status dot ──────────────────────────────────────────────
            .status-dot {
              position: absolute;
              bottom: 1px;
              right: 1px;
              width: 12px;
              height: 12px;
              border-radius: var(--radius-full);
              border: 2px solid var(--color-bg);
              transition: background-color var(--transition-interactive);
              pointer-events: none;

              &.status-online  { background-color: var(--color-success); }
              &.status-away    { background-color: var(--color-warning); }
              &.status-busy    { background-color: var(--color-error); }
              &.status-offline { background-color: var(--color-text-faint); }
            }

            // ── Self badge ────────────────────────────────────────────────
            &.self::before {
              font-family: 'Font Awesome 6 Free';
              font-weight: 900;
              content: '\f2bd';
              background: var(--color-surface-offset);
              color: var(--color-primary);
              position: absolute;
              width: 15px;
              height: 15px;
              line-height: 15px;
              font-size: 20px;
              text-align: center;
              top: -2px;
              right: -2px;
              border-radius: var(--radius-full);
            }

            // ── Admin badge ─────────────────────────────────────────────
            &.admin::before {
              display: block;
              font-family: 'Font Awesome 6 Free';
              font-weight: 900;
              content: '\f3ed';
              color: var(--color-primary);
              background: transparent;
              position: absolute;
              width: 14px;
              height: 14px;
              font-size: 14px;
              text-align: center;
              top: -2px;
              right: -2px;
            }

            // ── Host crown badge ───────────────────────────────────────
            &.host::after {
              display: block;
              font-family: 'Font Awesome 6 Free';
              font-weight: 900;
              content: '\f521';
              background: var(--color-primary);
              color: var(--color-bg);
              position: absolute;
              width: 20px;
              height: 20px;
              line-height: 20px;
              font-size: 10px;
              text-align: center;
              bottom: -4px;
              left: -6px;
              border-radius: var(--radius-full);
              // Crown badge sits above status dot
              z-index: 1;
            }

            .avatar {
              border-radius: var(--radius-full);
              overflow: hidden;
              width: 100%;
            }
          }
        }
      }
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
  export default class extends Vue {
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
