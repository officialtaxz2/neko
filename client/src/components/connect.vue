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
        
        <div class="demo-checkbox-container" v-if="!autoPassword && showDemo">
          <input type="checkbox" id="demo-mode" v-model="demoMode" />
          <label for="demo-mode">Demo- / Testmodus (Lokale Sim)</label>
        </div>

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
    background: radial-gradient(circle at 50% 50%, rgba(12, 13, 18, 0.85) 0%, rgba(7, 8, 12, 0.95) 100%);
    backdrop-filter: blur(12px) saturate(140%);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;

    .window {
      width: 340px;
      background: rgba(26, 30, 42, 0.45);
      border-radius: var(--radius-outer);
      border: 1px solid var(--glass-border);
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.05);
      padding: 24px;
      backdrop-filter: blur(8px);
      transition: var(--transition-fluid);

      &:hover {
        border-color: rgba(38, 230, 180, 0.25);
        box-shadow: 0 24px 60px rgba(0, 0, 0, 0.75), 0 0 15px rgba(38, 230, 180, 0.08);
      }

      .logo {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
        cursor: pointer;
        margin-bottom: 20px;
        user-select: none;

        img {
          height: 64px;
          margin-right: 12px;
          filter: drop-shadow(0 0 8px rgba(38, 230, 180, 0.3));
          transition: transform 0.5s cubic-bezier(0.16, 1, 0.3, 1);
        }

        &:hover img {
          transform: rotate(10deg) scale(1.05);
        }

        span {
          font-size: 28px;
          font-weight: 300;
          letter-spacing: -0.5px;
          color: var(--text-pure);

          b {
            font-weight: 900;
            color: var(--color-cyber-mint);
            text-shadow: 0 0 12px var(--color-cyber-mint-glow);
          }
        }
      }

      .message {
        display: flex;
        flex-direction: column;

        span {
          display: block;
          text-align: center;
          text-transform: uppercase;
          font-size: 11px;
          font-weight: 700;
          letter-spacing: 1.5px;
          color: var(--color-cyber-mint);
          margin-bottom: 12px;
          text-shadow: 0 0 8px rgba(38, 230, 180, 0.15);
        }

        input {
          border: 1px solid rgba(255, 255, 255, 0.08);
          padding: 10px 14px;
          font-size: 14px;
          border-radius: var(--radius-inner);
          margin: 6px 0;
          background: rgba(12, 13, 18, 0.5);
          color: var(--text-pure);
          transition: var(--transition-fluid);
          outline: none;

          &::placeholder {
            color: var(--text-muted-ok);
            opacity: 0.8;
          }

          &:focus, &:focus-visible {
            border-color: var(--color-cyber-mint);
            background: rgba(12, 13, 18, 0.75);
            box-shadow: 0 0 0 3px rgba(38, 230, 180, 0.15);
          }

          &::selection {
            background: var(--color-cyber-mint);
            color: rgba(0, 0, 0, 0.9);
          }
        }

        button {
          cursor: pointer;
          border-radius: var(--radius-inner);
          padding: 10px;
          background: linear-gradient(135deg, var(--color-cyber-mint) 0%, var(--color-cyber-mint-active) 100%);
          color: #050508;
          text-align: center;
          text-transform: uppercase;
          font-size: 13px;
          font-weight: 800;
          letter-spacing: 1px;
          margin-top: 14px;
          border: none;
          transition: var(--transition-fluid);
          box-shadow: 0 4px 12px var(--color-cyber-mint-glow);

          &:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(38, 230, 180, 0.45);
            filter: brightness(1.05);
          }

          &:active {
            transform: translateY(0);
            box-shadow: 0 2px 8px var(--color-cyber-mint-glow);
          }

          &:focus-visible {
            outline: 2px solid var(--text-pure);
            outline-offset: 2px;
          }
        }

        .demo-checkbox-container {
          display: flex;
          align-items: center;
          justify-content: flex-start;
          margin: 10px 4px 4px 4px;
          font-size: 13px;
          color: var(--text-muted-ok);
          cursor: pointer;

          input[type="checkbox"] {
            cursor: pointer;
            width: auto;
            margin: 0 8px 0 0;
            background: none;
            height: auto;
            accent-color: var(--color-cyber-mint);
          }

          label {
            cursor: pointer;
            user-select: none;
            text-transform: none !important;
            line-height: normal !important;
            font-size: 12.5px !important;
            color: var(--text-subtle);
            transition: var(--transition-fluid);
          }

          &:hover label {
            color: var(--color-cyber-mint);
          }
        }
      }

      .loader {
        width: 60px;
        height: 60px;
        position: relative;
        margin: 20px auto 10px auto;

        .bounce1,
        .bounce2 {
          width: 100%;
          height: 100%;
          border-radius: 50%;
          background-color: var(--color-cyber-mint);
          opacity: 0.6;
          position: absolute;
          top: 0;
          left: 0;
          box-shadow: 0 0 15px var(--color-cyber-mint-glow);

          -webkit-animation: bounce 2s infinite ease-in-out;
          animation: bounce 2s infinite ease-in-out;
        }

        .bounce2 {
          -webkit-animation-delay: -1s;
          animation-delay: -1s;
        }
      }
    }

    @keyframes bounce {
      0%,
      100% {
        transform: scale(0);
        opacity: 0.2;
      }
      50% {
        transform: scale(1);
        opacity: 0.8;
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-connect' })
  export default class NekoConnect extends Vue {
    private autoPassword: string | null = new URLSearchParams(location.search).get('pwd')

    private displayname: string = ''
    private password: string = ''
    private demoMode: boolean = false
    private showDemo: boolean = false

    private get isDev(): boolean {
      try {
        const meta = import.meta as any
        if (meta && meta.env) {
          return !!meta.env.DEV
        }
      } catch (e) {}
      return false
    }

    created() {
      this.showDemo = this.isDev
      this.demoMode = this.isDev
    }

    mounted() {
      const params = new URLSearchParams(location.search)
      // auto-password fill
      let password = this.$accessor.password
      if (this.autoPassword !== null) {
        this.removeUrlParam('pwd')
        password = this.autoPassword
      }

      // auto-user fill
      let displayname = this.$accessor.displayname
      const usr = params.get('usr')
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

      const client = this.$client as any
      if (this.demoMode) {
        if (client && typeof client.setDemoMode === 'function') {
          client.setDemoMode(true)
        }
        password = 'demo'
      } else {
        if (client && typeof client.setDemoMode === 'function') {
          client.setDemoMode(false)
        }
      }

      this.$accessor.login({ displayname: this.displayname, password })
      this.autoPassword = null
    }

    about() {
      this.$accessor.client.toggleAbout()
    }
  }
</script>
