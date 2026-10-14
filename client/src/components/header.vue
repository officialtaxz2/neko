<template>
  <div class="header">
    <button type="button" class="icon-btn neko" :aria-label="$t('controls.about')" @click="about">
      <img src="@/assets/images/logo.svg" alt="" />
      <span><b>n</b>.eko</span>
    </button>
    <div class="status" role="status">
      <span v-for="resource in activeLocks" :key="resource" class="chip warn" :title="$t(`locks.chip_${resource}`)">
        <i class="fas fa-lock" aria-hidden="true" />
        <span class="chip-text">{{ $t(`locks.chip_${resource}`) }}</span>
      </span>
      <span v-if="admin" class="chip admin" :title="$t('admin_loggedin')">
        <i class="fas fa-shield-alt" aria-hidden="true" />
        <span class="chip-text">{{ $t('header.admin') }}</span>
      </span>
      <span class="chip ok" :title="$t('header.in_room', { count: connectedCount })">
        <span class="dot" aria-hidden="true" />
        <span class="chip-text">{{ $t('header.in_room', { count: connectedCount }) }}</span>
        <span class="chip-count" aria-hidden="true">{{ connectedCount }}</span>
      </span>
    </div>
    <button
      type="button"
      :class="['icon-btn', 'panel-toggle', { active: side }]"
      :aria-expanded="side ? 'true' : 'false'"
      :aria-label="side ? $t('header.close_panel') : $t('header.open_panel')"
      @click="toggleMenu"
    >
      <i class="fas fa-comment-alt" aria-hidden="true" />
      <span class="label">{{ $t('header.chat') }}</span>
      <span v-if="showBadge" class="badge" :title="$t('header.new_messages')" />
    </button>
  </div>
</template>

<style lang="scss" scoped>
  .header {
    flex: 1;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 10px;
    padding: 0 12px 0 16px;
    min-width: 0;

    .neko {
      flex: none;
      gap: 9px;
      color: var(--text-pure);
      border-radius: 8px;
      padding: 2px 4px;

      img {
        display: block;
        height: 28px;
      }

      span {
        font-size: 24px;
        line-height: 28px;
        font-weight: 500;

        b {
          font-weight: 900;
          color: var(--color-cyber-mint);
        }
      }
    }

    .status {
      flex: 1;
      min-width: 0;
      display: flex;
      justify-content: flex-end;
      align-items: center;
      gap: 8px;
      overflow: hidden;
    }

    .chip {
      flex: none;
      display: inline-flex;
      align-items: center;
      gap: 6px;
      height: 28px;
      padding: 0 10px;
      border-radius: 999px;
      border: 1px solid rgba(255, 255, 255, 0.12);
      background: rgba(255, 255, 255, 0.03);
      font-size: 12.5px;
      font-weight: 500;
      color: var(--text-subtle);
      white-space: nowrap;

      i {
        font-size: 11px;
      }

      &.warn {
        border-color: rgba(245, 194, 107, 0.4);
        color: #f5c26b;
        background: rgba(245, 194, 107, 0.07);
      }

      &.admin i {
        color: #ffb86c;
      }

      .dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        background: var(--color-cyber-mint);
        box-shadow: 0 0 8px var(--color-cyber-mint);
      }

      .chip-count {
        display: none;
      }
    }

    .panel-toggle {
      position: relative;
      flex: none;
      gap: 8px;
      height: 38px;
      padding: 0 14px;
      border-radius: 10px;
      border: 1px solid var(--glass-border);
      background: rgba(255, 255, 255, 0.04);
      color: var(--text-subtle);
      font-size: 13.5px;
      font-weight: 600;
      transition: var(--transition-fluid);

      &:hover {
        color: var(--text-pure);
        background: rgba(255, 255, 255, 0.08);
      }

      &.active {
        color: var(--color-cyber-mint);
        border-color: rgba(38, 230, 180, 0.35);
        background: rgba(38, 230, 180, 0.1);
      }

      .badge {
        position: absolute;
        top: 5px;
        right: 5px;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #ff5a66;
        box-shadow: 0 0 0 2px var(--bg-obsidian-deep);
      }
    }
  }

  // Narrow screens: keep every status visible, but icon-only.
  @media (max-width: 600px) {
    .header {
      gap: 6px;
      padding: 0 8px 0 10px;

      .neko span {
        font-size: 20px;
      }

      .chip {
        padding: 0 8px;

        .chip-text {
          display: none;
        }

        .chip-count {
          display: inline;
        }
      }

      .panel-toggle {
        width: 40px;
        padding: 0;

        .label {
          display: none;
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'
  import { AdminLockResource } from '~/neko/messages'

  @Component({ name: 'neko-header' })
  export default class NekoHeader extends Vue {
    get admin() {
      return this.$accessor.user.admin
    }

    get side() {
      return this.$accessor.client.side
    }

    get texts() {
      return this.$accessor.chat.texts
    }

    get showBadge() {
      return !this.side && this.readTexts != this.texts
    }

    get fileTransfer() {
      return this.$accessor.remote.fileTransfer
    }

    get activeLocks(): AdminLockResource[] {
      const resources: AdminLockResource[] = ['control', 'login']
      if (this.fileTransfer) resources.push('file_transfer')
      return resources.filter((resource) => this.$accessor.isLocked(resource))
    }

    get connectedCount() {
      return Object.values(this.$accessor.user.members).filter((member) => member.connected).length
    }

    about() {
      this.$accessor.client.toggleAbout()
    }

    readTexts: number = 0
    toggleMenu() {
      this.$accessor.client.toggleSide()
      this.readTexts = this.texts
    }
  }
</script>
