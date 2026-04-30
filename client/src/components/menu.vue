<template>
  <ul>
    <li><i @click.stop.prevent="about" class="fas fa-question-circle" /></li>
    <li>
      <i
        class="fas fa-shield-alt"
        v-tooltip="{
          content: $t('admin_loggedin'),
          placement: 'right',
          offset: 5,
          boundariesElement: 'body',
        }"
        v-if="admin"
      />
    </li>
    <li>
      <select v-model="$i18n.locale">
        <option v-for="(lang, i) in langs" :key="`Lang${i}`" :value="lang">
          {{ lang }}
        </option>
      </select>
    </li>
  </ul>
</template>

<style lang="scss" scoped>
  ul {
    li {
      display: inline-block;
      margin-right: 10px;

      i {
        font-size: 24px;
        color: var(--color-text);
        cursor: pointer;
        transition: color var(--transition-interactive), transform var(--transition-fast);

        &:hover {
          color: var(--color-primary);
          transform: scale(1.1);
        }
      }
    }
  }

  select {
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
    // Adapts to both modes via CSS custom properties.
    // Previously used hardcoded dark SCSS vars ($background-tertiary,
    // color: white) which made text invisible in light mode.
    background-color: var(--color-surface-offset);
    border: 1px solid var(--color-border);
    color: var(--color-text);
    cursor: pointer;
    border-radius: var(--radius-md);
    height: 24px;
    padding: 0 var(--space-2);
    vertical-align: text-bottom;
    display: inline-block;
    font-size: var(--text-xs);
    transition:
      background-color var(--transition-interactive),
      border-color     var(--transition-interactive);

    option {
      font-weight: normal;
      color: var(--color-text);
      background-color: var(--color-bg);
    }

    &:hover {
      border-color: var(--color-primary);
      background-color: var(--color-surface-dynamic);
    }
  }
</style>

<script lang="ts">
  import { Component, Vue, Watch } from 'vue-property-decorator'
  import { messages } from '~/locale'
  import { set } from '~/utils/localstorage'

  @Component({ name: 'neko-menu' })
  export default class extends Vue {
    get admin() {
      return this.$accessor.user.admin
    }

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
      const default_lang = new URL(location.href).searchParams.get('lang')
      if (default_lang && this.langs.includes(default_lang)) {
        this.$i18n.locale = default_lang
      }
      const show_side = new URL(location.href).searchParams.get('show_side')
      if (show_side !== null) {
        this.$accessor.client.setSide(show_side === '1')
      }
      const mute_chat = new URL(location.href).searchParams.get('mute_chat')
      if (mute_chat !== null) {
        this.$accessor.settings.setSound(mute_chat !== '1')
      }
    }
  }
</script>
