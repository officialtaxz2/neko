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
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0;
    margin: 0;
    list-style: none;

    li {
      display: inline-flex;
      align-items: center;

      i {
        font-size: 20px;
        cursor: pointer;
        color: var(--text-subtle);
        transition: var(--transition-fluid);

        &:hover {
          color: var(--color-cyber-mint);
          filter: drop-shadow(0 0 6px var(--color-cyber-mint-glow));
          transform: translateY(-1px);
        }

        &:active {
          transform: translateY(0);
        }
      }
    }
  }

  select {
    appearance: none;
    background-color: rgba(12, 13, 18, 0.45);
    border: 1px solid var(--glass-border);
    color: var(--text-pure);
    cursor: pointer;
    border-radius: var(--radius-button);
    padding: 2px 10px;
    height: 30px;
    line-height: 26px;
    font-size: 13px;
    font-weight: 600;
    font-family: inherit;
    vertical-align: middle;
    display: inline-block;
    transition: var(--transition-fluid);
    outline: none;

    option {
      font-weight: normal;
      color: var(--text-pure);
      background-color: var(--bg-obsidian-floating);
    }

    &:hover {
      border-color: rgba(38, 230, 180, 0.25);
      background-color: rgba(12, 13, 18, 0.6);
    }

    &:focus {
      border-color: var(--color-cyber-mint);
      box-shadow: 0 0 0 3px rgba(38, 230, 180, 0.15);
    }
  }
</style>

<script lang="ts">
  import { Component, Vue, Watch } from 'vue-property-decorator'
  import { messages } from '~/locale'
  import { set } from '~/utils/localstorage'

  @Component({ name: 'neko-menu' })
  export default class NekoMenu extends Vue {
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
