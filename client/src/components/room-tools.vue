<template>
  <div class="room-tools">
    <button
      type="button"
      class="icon-btn tool-btn"
      :aria-label="$t('controls.fullscreen')"
      v-tooltip="tooltip($t('controls.fullscreen'))"
      @click.stop.prevent="$emit('action', 'fullscreen')"
    >
      <i class="fas fa-expand" aria-hidden="true" />
    </button>
    <div ref="anchor" class="more">
      <button
        ref="moreButton"
        type="button"
        :class="['icon-btn', 'tool-btn', { active: open }]"
        :aria-label="$t('controls.tools')"
        aria-haspopup="dialog"
        :aria-expanded="open ? 'true' : 'false'"
        v-tooltip="open ? null : tooltip($t('controls.tools'))"
        @click.stop.prevent="toggle"
      >
        <i class="fas fa-ellipsis-h" aria-hidden="true" />
      </button>
      <div
        v-if="open"
        ref="popover"
        class="popover"
        role="dialog"
        :aria-label="$t('controls.tools')"
        @keydown.esc.stop.prevent="close"
      >
        <neko-more-tools :pipAvailable="pipAvailable" @action="run" @done="close" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .room-tools {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .tool-btn {
    width: 40px;
    height: 40px;
    border-radius: var(--radius-inner, 10px);
    border: 1px solid var(--glass-border);
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-subtle);
    font-size: 15px;
    transition: var(--transition-fluid);

    &:hover {
      color: var(--text-pure);
      background: rgba(255, 255, 255, 0.08);
    }

    &.active {
      color: var(--color-cyber-mint);
      border-color: rgba(38, 230, 180, 0.45);
      background: rgba(38, 230, 180, 0.1);
    }
  }

  .more {
    position: relative;
  }

  .popover {
    position: absolute;
    right: 0;
    bottom: calc(100% + 14px);
    width: 320px;
    max-width: calc(100vw - 24px);
    box-sizing: border-box;
    padding: 14px;
    border-radius: 16px;
    background: var(--bg-obsidian-floating, #12141d);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 18px 50px rgba(0, 0, 0, 0.55);
    z-index: 200;
  }
</style>

<script lang="ts">
  import { Component, Prop, Ref, Vue } from 'vue-property-decorator'
  import MoreTools, { ToolAction } from './more-tools.vue'

  // Desktop/tablet/TV: a direct fullscreen button plus a "More" popover with
  // the same tools as the phone sheet. Actions go up to the parent, which
  // forwards them to the video component.
  @Component({ name: 'neko-room-tools', components: { 'neko-more-tools': MoreTools } })
  export default class NekoRoomTools extends Vue {
    @Prop(Boolean) readonly pipAvailable!: boolean
    @Ref('anchor') readonly anchor!: HTMLElement
    @Ref('popover') readonly popover?: HTMLElement
    @Ref('moreButton') readonly moreButton!: HTMLButtonElement

    open = false

    tooltip(content: string) {
      return { content, placement: 'top', offset: 5, boundariesElement: 'body', delay: { show: 300, hide: 100 } }
    }

    toggle() {
      if (this.open) {
        this.close()
        return
      }
      this.open = true
      document.addEventListener('mousedown', this.onOutside, true)
      this.$nextTick(() => {
        const first = this.popover && this.popover.querySelector('button')
        if (first) (first as HTMLButtonElement).focus()
      })
    }

    close() {
      if (!this.open) return
      this.open = false
      document.removeEventListener('mousedown', this.onOutside, true)
      this.$nextTick(() => this.moreButton && this.moreButton.focus())
    }

    onOutside(event: MouseEvent) {
      if (this.anchor && !this.anchor.contains(event.target as Node)) {
        this.open = false
        document.removeEventListener('mousedown', this.onOutside, true)
      }
    }

    run(action: ToolAction, event?: MouseEvent) {
      this.open = false
      document.removeEventListener('mousedown', this.onOutside, true)
      this.$emit('action', action, event)
    }

    beforeDestroy() {
      document.removeEventListener('mousedown', this.onOutside, true)
    }
  }
</script>
