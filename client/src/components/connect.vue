<template>
  <div class="connect">
    <!-- System Dialog Overlay: shown above login card when a system event fires -->
    <transition name="dialog-fade">
      <div class="system-dialog" v-if="systemDialog" role="dialog" aria-modal="true" :aria-label="systemDialog.title">
        <div class="system-dialog__card">
          <!-- Icon: info (teal) or error (red) -->
          <div class="system-dialog__icon" :class="`system-dialog__icon--${systemDialog.type}`">
            <!-- info icon -->
            <svg v-if="systemDialog.type === 'info'" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 8v1M12 11v5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <!-- error / warning icon -->
            <svg v-else viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 8v5M12 16v1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </div>
          <h2 class="system-dialog__title">{{ systemDialog.title }}</h2>
          <p class="system-dialog__text" v-if="systemDialog.text">{{ systemDialog.text }}</p>
          <button class="system-dialog__btn" @click="dismissDialog" autofocus>
            {{ $t('connection.button_confirm') }}
          </button>
        </div>
      </div>
    </transition>

    <!-- Login card: hidden while system dialog is visible -->
    <transition name="card-fade">
      <div class="window" v-if="!systemDialog">
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
    </transition>
  </div>
</template>

<style lang="scss" scoped>
  .connect {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    // Base color — sits below all background-image layers
    background-color: var(--color-bg);
    // Layer 1 (top): radial Spotlight-Gradient — teal-tinted light cone centred on the login card
    // Layer 2 (below): dot-grid pattern — subtly reveals the tech/terminal aesthetic
    // Dots are transparent-based so the spotlight glow bleeds through naturally in the centre.
    background-image:
      radial-gradient(
        ellipse 80% 70% at 50% 50%,
        color-mix(in srgb, var(--color-primary) 10%, transparent) 0%,
        transparent 70%
      ),
      radial-gradient(
        circle,
        color-mix(in srgb, var(--color-primary) 20%, transparent) 1.5px,
        transparent 1.5px
      );
    background-size: auto, 28px 28px;
    // Blur on overlay for depth behind the card
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    display: flex;
    justify-content: center;
    align-items: center;
  }

  // -----------------------------------------------
  // System Dialog Overlay
  // -----------------------------------------------
  .system-dialog {
    position: absolute;
    inset: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    // Darkens the dot-grid bg behind the dialog for focus
    background: color-mix(in srgb, var(--color-bg) 60%, transparent);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    z-index: 10;

    &__card {
      width: 340px;
      padding: var(--space-8) var(--space-6);
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: var(--space-4);
      // Glassmorphism card — same language as login card
      background: color-mix(in srgb, var(--color-surface) 80%, transparent);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border-radius: var(--radius-xl);
      border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
      border-top-color: color-mix(in srgb, var(--color-text) 15%, transparent);
      box-shadow: var(--shadow-lg);
    }

    &__icon {
      width: 52px;
      height: 52px;
      border-radius: var(--radius-full);
      display: flex;
      align-items: center;
      justify-content: center;
      // Tinted ring behind icon
      border: 2px solid currentColor;
      opacity: 0.9;

      svg {
        width: 28px;
        height: 28px;
      }

      // info = primary teal
      &--info {
        color: var(--color-primary);
        background: color-mix(in srgb, var(--color-primary) 12%, transparent);
      }

      // error / warning = error red
      &--error,
      &--warning {
        color: var(--color-error);
        background: color-mix(in srgb, var(--color-error) 12%, transparent);
      }
    }

    &__title {
      font-size: var(--text-lg);
      font-weight: 600;
      color: var(--color-text);
      text-align: center;
      line-height: 1.3;
    }

    &__text {
      font-size: var(--text-sm);
      color: var(--color-text-muted);
      text-align: center;
      max-width: 28ch;
      line-height: 1.5;
    }

    &__btn {
      margin-top: var(--space-2);
      min-height: 44px;
      min-width: 120px;
      cursor: pointer;
      border-radius: var(--radius-md);
      padding: var(--space-2) var(--space-6);
      background: var(--color-primary);
      color: var(--color-text-inverse);
      font-weight: 600;
      font-size: var(--text-sm);
      border: none;
      transition:
        background-color var(--transition-interactive),
        transform var(--transition-interactive);

      &:hover {
        background: var(--color-primary-hover);
      }

      &:active {
        background: var(--color-primary-active);
        transform: scale(0.97);
      }

      &:focus-visible {
        outline: 2px solid var(--color-primary);
        outline-offset: 3px;
      }
    }
  }

  // -----------------------------------------------
  // Login card
  // -----------------------------------------------
  .window {
    width: 300px;
    // Glassmorphism: frosted surface with blur(12px)
    background: color-mix(in srgb, var(--color-surface) 75%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-radius: var(--radius-lg);
    // Subtle glass edge: lighter top border, standard border elsewhere
    border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
    border-top-color: color-mix(in srgb, var(--color-text) 15%, transparent);
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
        // Input fields: slightly less transparent than the card for legibility
        background: color-mix(in srgb, var(--color-bg) 70%, transparent);
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

  // -----------------------------------------------
  // Transitions
  // -----------------------------------------------
  .dialog-fade-enter-active,
  .dialog-fade-leave-active {
    transition: opacity 220ms ease, transform 220ms cubic-bezier(0.16, 1, 0.3, 1);
  }
  .dialog-fade-enter,
  .dialog-fade-leave-to {
    opacity: 0;
    transform: scale(0.96);
  }

  .card-fade-enter-active,
  .card-fade-leave-active {
    transition: opacity 180ms ease;
  }
  .card-fade-enter,
  .card-fade-leave-to {
    opacity: 0;
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

    get systemDialog() {
      return this.$accessor.client.systemDialog
    }

    dismissDialog() {
      this.$accessor.client.setSystemDialog(null)
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
