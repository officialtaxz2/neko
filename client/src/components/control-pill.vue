<template>
  <div v-if="visible" :class="['control-pill', state, { compact, shake: shakeKbd }]" role="group" :aria-label="title">
    <!-- Nobody controls: one primary action -->
    <button v-if="state === 'free'" type="button" class="icon-btn take" :disabled="!playable" @click.stop.prevent="toggleControl">
      <i class="fas fa-mouse-pointer" aria-hidden="true" />
      <span>{{ $t('controls.take') }}</span>
    </button>

    <template v-else>
      <span class="lead" aria-hidden="true">
        <neko-avatar v-if="state === 'other' && hostMember" :seed="hostMember.displayname" :avatar="hostMember.avatar" :size="compact ? 28 : 30" />
        <i v-else-if="state === 'locked'" class="fas fa-lock" />
        <i v-else :class="['fas', compact && inputLocked ? 'fa-pause' : 'fa-mouse-pointer']" />
      </span>
      <span class="label">
        <span class="title">{{ title }}</span>
        <small v-if="subtitle && !compact">{{ subtitle }}</small>
      </span>

      <template v-if="state === 'mine' || state === 'implicit'">
        <!-- On phones the input pause lives in the dock's "More" sheet. -->
        <button
          v-if="!compact"
          type="button"
          :class="['icon-btn', 'act', 'act-icon', { pressed: inputLocked }]"
          :aria-pressed="inputLocked ? 'true' : 'false'"
          :aria-label="inputLocked ? $t('controls.unlock_input') : $t('controls.lock_input')"
          v-tooltip="tooltip(inputLocked ? $t('controls.unlock_input') : $t('controls.lock_input'))"
          @click.stop.prevent="inputLocked = !inputLocked"
        >
          <i :class="['fas', inputLocked ? 'fa-lock' : 'fa-lock-open']" aria-hidden="true" />
        </button>
        <button v-if="!implicitHosting" type="button" class="icon-btn act" @click.stop.prevent="toggleControl">
          {{ $t('controls.release_short') }}
        </button>
      </template>

      <template v-else-if="state === 'other'">
        <button v-if="admin" type="button" class="icon-btn act primary" @click.stop.prevent="toggleControl">
          {{ $t('controls.force_take') }}
        </button>
        <button v-else-if="requested" type="button" class="icon-btn act" disabled>
          <i class="fas fa-check" aria-hidden="true" />
          {{ $t('controls.requested') }}
        </button>
        <button v-else type="button" class="icon-btn act primary" :disabled="!playable" @click.stop.prevent="request">
          <i class="fas fa-hand-paper" aria-hidden="true" />
          {{ $t('controls.request_short') }}
        </button>
      </template>
    </template>
  </div>
</template>

<style lang="scss" scoped>
  .control-pill {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    height: 44px;
    padding: 0 6px 0 7px;
    min-width: 0;
    border-radius: 999px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-pure);
    font-weight: 500;
    white-space: nowrap;
    transition: var(--transition-fluid);

    .lead {
      flex: none;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-width: 30px;
      font-size: 15px;

      ::v-deep .avatar {
        font-size: 12px !important;
        font-weight: 700;
      }
    }

    .label {
      display: flex;
      flex-direction: column;
      min-width: 0;
      line-height: 1.15;

      .title {
        overflow: hidden;
        text-overflow: ellipsis;
      }

      small {
        font-size: 11px;
        color: var(--text-muted-ok);
        font-weight: 500;
      }
    }

    .act {
      flex: none;
      gap: 6px;
      height: 32px;
      padding: 0 14px;
      border-radius: 999px;
      background: rgba(255, 255, 255, 0.08);
      font-size: 13px;
      font-weight: 600;
      color: inherit;
      transition: var(--transition-fluid);

      &:hover:not(:disabled) {
        background: rgba(255, 255, 255, 0.14);
      }

      &:disabled {
        opacity: 0.6;
      }

      &.primary {
        background: var(--color-cyber-mint);
        color: #06281f;

        &:hover:not(:disabled) {
          background: var(--color-cyber-mint-active);
        }
      }

      &.act-icon {
        width: 32px;
        padding: 0;
      }
    }

    .take {
      gap: 8px;
      height: 100%;
      padding: 0 12px 0 9px;
      color: var(--color-cyber-mint);
      font-weight: 600;
      font-size: 13.5px;

      &:disabled {
        opacity: 0.55;
      }
    }

    &.free {
      border-color: rgba(38, 230, 180, 0.5);
      background: rgba(38, 230, 180, 0.12);
      padding: 0 4px;

      &:hover {
        background: rgba(38, 230, 180, 0.18);
      }
    }

    &.mine {
      background: var(--color-cyber-mint);
      border-color: transparent;
      color: #06281f;
      box-shadow: 0 0 16px var(--color-cyber-mint-glow);

      .lead {
        margin-left: 4px;
      }

      .label small {
        color: rgba(6, 40, 31, 0.72);
      }

      .act {
        background: rgba(6, 40, 31, 0.14);

        &:hover:not(:disabled) {
          background: rgba(6, 40, 31, 0.24);
        }

        &.pressed {
          background: #06281f;
          color: var(--color-cyber-mint);
        }
      }
    }

    &.locked .lead {
      color: #f5c26b;
    }

    &.implicit {
      .lead {
        color: var(--color-cyber-mint);
      }
    }

    &.compact {
      gap: 6px;
      height: 44px;

      .act {
        padding: 0 11px;
      }
    }

    &.shake {
      animation: pill-nudge 1.1s cubic-bezier(0.36, 0.07, 0.19, 0.97);
      border-color: var(--color-cyber-mint);
    }
  }

  @keyframes pill-nudge {
    0%,
    100% {
      transform: translateX(0);
    }
    15%,
    55% {
      transform: translateX(-4px);
    }
    35%,
    75% {
      transform: translateX(4px);
    }
  }
</style>

<script lang="ts">
  import { Component, Mixins, Prop, Watch } from 'vue-property-decorator'
  import MediaControls from './mixins/media-controls'
  import Avatar from './avatar.vue'

  type PillState = 'free' | 'mine' | 'other' | 'locked' | 'implicit'

  @Component({ name: 'neko-control-pill', components: { 'neko-avatar': Avatar } })
  export default class NekoControlPill extends Mixins(MediaControls) {
    @Prop(Boolean) readonly shakeKbd!: boolean
    @Prop(Boolean) readonly compact!: boolean

    requested = false
    private requestTimer?: number

    get visible() {
      return !this.viewOnly && !this.receiveOnly
    }

    get hostMember() {
      const id = this.$accessor.remote.id
      return id ? this.$accessor.user.members[id] : null
    }

    get state(): PillState {
      if (this.implicitHosting) {
        if (this.controlLocked) return 'locked'
        return this.controlling ? 'mine' : 'implicit'
      }
      if (this.hosting) return 'mine'
      if (this.controlLocked) return 'locked'
      if (this.hosted && this.hostMember) return 'other'
      return 'free'
    }

    get title() {
      switch (this.state) {
        case 'mine':
          return this.$t('controls.you_control') as string
        case 'other':
          return this.$t('controls.other_controls', { name: this.hostMember!.displayname }) as string
        case 'locked':
          return this.$t('controls.locked_by_admin') as string
        case 'implicit':
          return this.$t('controls.click_to_control') as string
        default:
          return this.$t('controls.take') as string
      }
    }

    get subtitle() {
      switch (this.state) {
        case 'mine':
          return (this.inputLocked ? this.$t('controls.input_locked') : this.$t('controls.input_active')) as string
        case 'other':
          return (
            this.requested
              ? this.$t('controls.waiting_for', { name: this.hostMember!.displayname })
              : this.$t('controls.watching')
          ) as string
        case 'locked':
          return this.$t('controls.admin_only') as string
        default:
          return ''
      }
    }

    // Local input lock: the host keeps control but stops sending mouse and
    // keyboard input. Same store flag as the former lock switch.
    get inputLocked() {
      return this.$accessor.remote.locked && this.hosting
    }

    set inputLocked(locked: boolean) {
      this.$accessor.remote.setLocked(locked)
    }

    tooltip(content: string) {
      return { content, placement: 'top', offset: 5, boundariesElement: 'body', delay: { show: 300, hide: 100 } }
    }

    request() {
      if (!this.playable) return
      this.$accessor.remote.request()
      this.requested = true
      window.clearTimeout(this.requestTimer)
      this.requestTimer = window.setTimeout(() => (this.requested = false), 60000)
    }

    get hostId() {
      return this.$accessor.remote.id
    }

    @Watch('hostId')
    onHostChange() {
      this.requested = false
      window.clearTimeout(this.requestTimer)
    }

    beforeDestroy() {
      window.clearTimeout(this.requestTimer)
    }
  }
</script>
