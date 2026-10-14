<template>
  <div class="more-tools">
    <template v-if="inputTools.length">
      <h4 class="group">{{ $t('setting.group_input') }}</h4>
      <div class="grid">
        <button
          v-for="tool in inputTools"
          :key="tool.id"
          type="button"
          :class="['icon-btn', 'tile', { on: tool.on }]"
          :aria-pressed="tool.pressed"
          @click.stop.prevent="tool.run($event)"
        >
          <i :class="['fas', tool.icon]" aria-hidden="true" />
          <span>{{ tool.label }}</span>
        </button>
      </div>
    </template>
    <template v-if="viewTools.length">
      <h4 class="group">{{ $t('controls.view_group') }}</h4>
      <div class="grid">
        <button
          v-for="tool in viewTools"
          :key="tool.id"
          type="button"
          :class="['icon-btn', 'tile', { on: tool.on }]"
          :aria-pressed="tool.pressed"
          @click.stop.prevent="tool.run($event)"
        >
          <i :class="['fas', tool.icon]" aria-hidden="true" />
          <span>{{ tool.label }}</span>
        </button>
      </div>
    </template>
  </div>
</template>

<style lang="scss" scoped>
  .group {
    margin: 4px 2px 8px;
    font-size: 10.5px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted-ok);
  }

  .grid + .group {
    margin-top: 14px;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
  }

  .tile {
    flex-direction: column;
    gap: 7px;
    min-height: 72px;
    padding: 10px 6px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.07);
    color: var(--text-subtle);
    transition: var(--transition-fluid);

    i {
      font-size: 18px;
    }

    span {
      font-size: 12px;
      line-height: 1.25;
      text-align: center;
      overflow-wrap: anywhere;
      hyphens: auto;
    }

    &:hover {
      background: rgba(255, 255, 255, 0.08);
      color: var(--text-pure);
    }

    &.on {
      color: var(--color-cyber-mint);
      border-color: rgba(38, 230, 180, 0.4);
    }
  }
</style>

<script lang="ts">
  import { Component, Mixins, Prop } from 'vue-property-decorator'
  import MediaControls from './mixins/media-controls'
  import { isTouchDevice } from '~/utils/touch'

  export type ToolAction = 'fullscreen' | 'keyboard' | 'keyboard-helper' | 'clipboard' | 'pip' | 'resolution'

  interface Tool {
    id: string
    icon: string
    label: string
    on?: boolean
    pressed?: 'true' | 'false'
    run: (event: MouseEvent) => void
  }

  // The tools that used to float on the video, in one place. The phone sheet
  // shows everything; the desktop popover leaves out what the bar already has
  // (fullscreen, microphone, input pause).
  @Component({ name: 'neko-more-tools' })
  export default class NekoMoreTools extends Mixins(MediaControls) {
    @Prop(Boolean) readonly sheet!: boolean
    @Prop(Boolean) readonly pipAvailable!: boolean

    get touchDevice() {
      return isTouchDevice(this.$accessor.settings.force_touch)
    }

    get canControl() {
      return !this.viewOnly && !this.receiveOnly
    }

    // Local input pause for the current host (same flag as the desktop pill).
    get inputLocked() {
      return this.$accessor.remote.locked && this.hosting
    }

    get streamName() {
      const selected = this.$accessor.hls.selected
      if (selected === 'll-hls') return 'LL-HLS'
      if (selected === 'hls') return 'HLS'
      return 'WebRTC'
    }

    get inputTools(): Tool[] {
      const tools: Tool[] = []
      if (!this.canControl || !this.hosting) return tools
      if (this.touchDevice) {
        tools.push(
          {
            id: 'keyboard',
            icon: 'fa-keyboard',
            label: this.$t('controls.keyboard') as string,
            run: () => this.action('keyboard'),
          },
          {
            id: 'keyboard-helper',
            icon: 'fa-sliders-h',
            label: this.$t('controls.keyboard_helper') as string,
            run: () => this.action('keyboard-helper'),
          },
        )
      }
      if (this.sheet) {
        tools.push({
          id: 'input-pause',
          icon: this.inputLocked ? 'fa-lock' : 'fa-lock-open',
          label: (this.inputLocked ? this.$t('controls.resume_input') : this.$t('controls.pause_input')) as string,
          on: this.inputLocked,
          pressed: this.inputLocked ? 'true' : 'false',
          run: () => this.$accessor.remote.setLocked(!this.inputLocked),
        })
      }
      if (this.sheet && this.micAllowed) {
        tools.push({
          id: 'microphone',
          icon: this.microphoneActive ? 'fa-microphone' : 'fa-microphone-slash',
          label: this.$t('controls.microphone') as string,
          on: this.microphoneActive,
          pressed: this.microphoneActive ? 'true' : 'false',
          run: () => void this.toggleMicrophone(),
        })
      }
      tools.push({
        id: 'clipboard',
        icon: 'fa-clipboard',
        label: this.$t('controls.clipboard') as string,
        run: () => this.action('clipboard'),
      })
      return tools
    }

    get viewTools(): Tool[] {
      const tools: Tool[] = []
      if (this.sheet) {
        tools.push({
          id: 'fullscreen',
          icon: 'fa-expand',
          label: this.$t('controls.fullscreen') as string,
          run: () => this.action('fullscreen'),
        })
      }
      if (this.pipAvailable) {
        tools.push({
          id: 'pip',
          icon: 'fa-external-link-alt',
          label: this.$t('controls.pip') as string,
          run: () => this.action('pip'),
        })
      }
      if (this.admin && !this.receiveOnly) {
        tools.push({
          id: 'resolution',
          icon: 'fa-desktop',
          label: this.$t('controls.resolution') as string,
          run: (event) => this.action('resolution', event),
        })
      }
      if (!this.viewOnly) {
        tools.push({
          id: 'stream',
          icon: 'fa-broadcast-tower',
          label: this.$t('controls.stream', { name: this.streamName }) as string,
          run: () => this.openStreamSettings(),
        })
      }
      return tools
    }

    action(action: ToolAction, event?: MouseEvent) {
      this.$emit('action', action, event)
    }

    openStreamSettings() {
      this.$emit('done')
      this.$accessor.client.setTab('settings')
      this.$accessor.client.setSide(true)
    }
  }
</script>
