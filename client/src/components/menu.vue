<template>
  <ul>
    <li>
      <button
        type="button"
        class="icon-btn"
        :aria-label="$t('controls.about')"
        v-tooltip="{ content: $t('controls.about'), placement: 'top', offset: 5, boundariesElement: 'body' }"
        @click.stop.prevent="about"
      >
        <i class="fas fa-question-circle" aria-hidden="true" />
      </button>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  ul {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0;
    margin: 0;
    list-style: none;

    li {
      display: inline-flex;
      align-items: center;

      .icon-btn {
        width: 40px;
        height: 40px;
        border-radius: var(--radius-inner);
        border: 1px solid var(--glass-border);
        background: rgba(255, 255, 255, 0.04);
        font-size: 16px;
        color: var(--text-subtle);
        transition: var(--transition-fluid);

        &:hover {
          color: var(--color-cyber-mint);
          background: rgba(255, 255, 255, 0.08);
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue, Watch } from 'vue-property-decorator'
  import { messages } from '~/locale'
  import { set } from '~/utils/localstorage'

  @Component({ name: 'neko-menu' })
  export default class NekoMenu extends Vue {
    get langs() {
      return Object.keys(messages)
    }

    about() {
      this.$accessor.client.toggleAbout()
    }

    @Watch('$i18n.locale')
    onLanguageChange(newLang: string) {
      set('lang', newLang)
    }

    mounted() {
      const params = new URLSearchParams(location.search)
      const default_lang = params.get('lang')
      if (default_lang && this.langs.includes(default_lang)) {
        this.$i18n.locale = default_lang
      }
      const show_side = params.get('show_side')
      if (show_side !== null) {
        this.$accessor.client.setSide(show_side === '1')
      }
      const mute_chat = params.get('mute_chat')
      if (mute_chat !== null) {
        this.$accessor.settings.setSound(mute_chat !== '1')
      }
    }
  }
</script>
