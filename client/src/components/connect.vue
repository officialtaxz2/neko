<template>
  <div class="connect">
    <div class="window">
      <div class="logo" title="About n.eko" @click.stop.prevent="about">
        <img src="@/assets/images/logo.svg" alt="n.eko" />
        <span><b>n</b>.eko</span>
      </div>
      <form class="message" v-if="!connecting" @submit.stop.prevent="connect">
        <span v-if="!autoPassword">{{ $t('connect.login_title') }}</span>
        <span v-else>{{ $t('connect.invitation_title') }}</span>
        <input type="text" :placeholder="$t('connect.displayname')" v-model="displayname" />
        <input type="password" :placeholder="$t('connect.password')" v-model="password" v-if="!autoPassword" />
        <button type="submit" @click.stop.prevent="login">
          {{ $t('connect.connect') }}
        </button>
      </form>
      <div class="loader" v-if="connecting">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  .connect {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    // Overlay: semi-transparent bg via color-mix (no $scss-var needed)
    background: color-mix(in srgb, var(--color-bg) 85%, transparent);
    display: flex;
    justify-content: center;
    align-items: center;

    .window {
      width: 300px;
      background: var(--color-surface);
      border-radius: var(--radius-lg);
      border: 1px solid var(--color-border);
      padding: var(--space-3);

      .logo {
        width: 100%;
        // Touch target: logo button is ~90px tall — already exceeds 44px
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
        cursor: pointer;
        transition: opacity var(--transition-interactive);

        &:hover {
          opacity: 0.8;
        }

        &:active {
          opacity: 0.6;
        }

        img {
          height: 90px;
          margin-right: var(--space-2);
        }

        span {
          font-size: 30px;
          line-height: 56px;
          color: var(--color-text);

          b {
            font-weight: 900;
          }
        }
      }

      .message {
        display: flex;
        flex-direction: column;
        gap: var(--space-1);

        span {
          display: block;
          text-align: center;
          text-transform: uppercase;
          line-height: 30px;
          color: var(--color-text-muted);
          font-size: var(--text-sm);
        }

        input {
          // Touch target ≥ 44px
          min-height: 44px;
          border: 1px solid transparent;
          padding: var(--space-2) var(--space-3);
          border-radius: var(--radius-md);
          background: var(--color-bg);
          color: var(--color-text);
          font-size: var(--text-sm);
          transition:
            border-color     var(--transition-interactive),
            background-color var(--transition-interactive);

          &:focus {
            border-color: var(--color-primary);
            outline: none;
          }

          &::placeholder {
            color: var(--color-text-faint);
          }

          &::selection {
            background: var(--color-primary-highlight);
          }
        }

        button {
          // Touch target ≥ 44px
          min-height: 44px;
          cursor: pointer;
          border-radius: var(--radius-md);
          padding: var(--space-2) var(--space-4);
          background: var(--color-primary);
          color: var(--color-text-inverse);
          text-align: center;
          text-transform: uppercase;
          font-weight: 700;
          font-size: var(--text-sm);
          border: none;
          transition:
            background-color var(--transition-interactive),
            transform        var(--transition-fast);

          &:hover {
            background: var(--color-primary-hover);
          }

          &:active {
            background: var(--color-primary-active);
            transform: scale(0.97);
          }
        }
      }

      .loader {
        width: 90px;
        height: 90px;
        position: relative;
        margin: var(--space-3) auto;

        .bounce1,
        .bounce2 {
          width: 100%;
          height: 100%;
          border-radius: 50%;
          background-color: var(--color-primary);
          opacity: 0.6;
          position: absolute;
          top: 0;
          left: 0;
          animation: bounce 2s infinite ease-in-out;
          -webkit-animation: bounce 2s infinite ease-in-out;
        }

        .bounce2 {
          animation-delay: -1s;
          -webkit-animation-delay: -1s;
        }
      }
    }

    @keyframes bounce {
      0%,
      100% {
        transform: scale(0);
        -webkit-transform: scale(0);
      }
      50% {
        transform: scale(1);
        -webkit-transform: scale(1);
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-connect' })
  export default class extends Vue {
    private autoPassword: string | null = new URL(location.href).searchParams.get('pwd')

    private displayname: string = ''
    private password: string = ''

    mounted() {
      // auto-password fill
      let password = this.$accessor.password
      if (this.autoPassword !== null) {
        this.removeUrlParam('pwd')
        password = this.autoPassword
      }

      // auto-user fill
      let displayname = this.$accessor.displayname
      const usr = new URL(location.href).searchParams.get('usr')
      if (usr) {
        this.removeUrlParam('usr')
        displayname = this.$accessor.displayname || usr
      }

      if (displayname !== '' && password !== '') {
        this.$accessor.login({ displayname, password })
        this.autoPassword = null
      }
    }

    get connecting() {
      return this.$accessor.connecting
    }

    removeUrlParam(param: string) {
      let url = document.location.href
      let urlparts = url.split('?')

      if (urlparts.length >= 2) {
        let urlBase = urlparts.shift()
        let queryString = urlparts.join('?')

        let prefix = encodeURIComponent(param) + '='
        let pars = queryString.split(/[&;]/g)
        for (let i = pars.length; i-- > 0; ) {
          if (pars[i].lastIndexOf(prefix, 0) !== -1) {
            pars.splice(i, 1)
          }
        }

        url = urlBase + (pars.length > 0 ? '?' + pars.join('&') : '')
        window.history.pushState('', document.title, url)
      }
    }

    login() {
      let password = this.password
      if (this.autoPassword !== null) {
        password = this.autoPassword
      }

      if (this.displayname == '') {
        this.$swal({
          title: this.$t('connect.error') as string,
          text: this.$t('connect.empty_displayname') as string,
          icon: 'error',
        })
        return
      }

      this.$accessor.login({ displayname: this.displayname, password })
      this.autoPassword = null
    }

    about() {
      this.$accessor.client.toggleAbout()
    }
  }
</script>
