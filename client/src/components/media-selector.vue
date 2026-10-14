<template>
  <select v-model="backend" :aria-label="$t('setting.media_backend')" class="media-selector">
    <option value="webrtc">WebRTC</option>
    <option value="webcodecs-ws">WebCodecs</option>
    <option v-for="mode in modes" :key="mode" :value="mode">{{ mode === 'll-hls' ? 'LL-HLS' : 'HLS' }}</option>
    <option v-if="unavailable" :value="backend" disabled>{{ backend === 'll-hls' ? 'LL-HLS' : 'HLS' }} — {{ $t('media.unavailable') }}</option>
  </select>
</template>

<script lang="ts">
  import { Vue, Component } from 'vue-property-decorator'
  import { isHLSBackend } from '~/neko/media-selection.js'
  import type { HLSMode } from '~/neko/messages'

  @Component({ name: 'neko-media-selector' })
  export default class NekoMediaSelector extends Vue {
    get modes() { return this.$accessor.hls.availableModes }
    get backend() { return this.$client.effectiveMediaBackend }
    set backend(value: string) {
      if (!this.$client.canSelectMediaBackend(value)) return
      this.$accessor.settings.setMediaBackend(value)
      this.$client.changeMediaBackend(value)
    }
    get unavailable() { return isHLSBackend(this.backend) && !this.modes.includes(this.backend as HLSMode) }
  }
</script>

<style lang="scss" scoped>
  .media-selector {
    max-width: 130px;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: var(--color-dark, #12121a);
    color: var(--text-pure, white);
    padding: 6px;
    font: inherit;
  }
</style>
