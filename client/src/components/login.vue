<template>
  <div class="login-page">
    <!-- System Dialog Overlay: shown above login card when a system event fires -->
    <transition name="dialog-fade">
      <div
        v-if="systemDialog"
        class="system-dialog"
        role="dialog"
        aria-modal="true"
        :aria-label="systemDialog.title"
      >
        <div class="system-dialog__card">
          <div
            class="system-dialog__icon"
            :class="`system-dialog__icon--${systemDialog.type}`"
          >
            <svg v-if="systemDialog.type === 'info'" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 8v1M12 11v5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 8v5M12 16v1" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </div>
          <h2 class="system-dialog__title">{{ systemDialog.title }}</h2>
          <p v-if="systemDialog.text" class="system-dialog__text">{{ systemDialog.text }}</p>
          <button class="system-dialog__btn" type="button" @click="dismissDialog" autofocus>
            {{ $t('connection.button_confirm') }}
          </button>
        </div>
      </div>
    </transition>

    <!-- Login card: hidden while system dialog is visible -->
    <transition name="card-fade">
      <div class="window" v-if="!systemDialog">
        <!-- Logo / About trigger -->
        <button
          type="button"
          class="logo"
          @click.stop.prevent="about"
          :title="$t('connect.login_title')"
          aria-label="About n.eko"
        >
          <img src="@/assets/images/logo.svg" alt="n.eko" width="90" height="90" loading="lazy" />
          <!-- Kinetic wordmark: each character wrapped for staggered weight-pulse -->
          <span class="neko-logo" aria-hidden="true">
            <span>n</span><span>.</span><span>e</span><span>k</span><span>o</span>
          </span>
        </button>

        <!-- Connecting: bounce loader -->
        <div v-if="connecting" class="loader" role="status" :aria-label="$t('connect.connect')">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
        </div>

        <!-- Login form -->
        <form v-else class="form" @submit.stop.prevent="login" novalidate>
          <p class="form__title">
            <span v-if="!autoPassword">{{ $t('connect.login_title') }}</span>
            <span v-else>{{ $t('connect.invitation_title') }}</span>
          </p>

          <!-- Displayname field -->
          <div class="form__field">
            <label for="login-displayname" class="form__label">{{ $t('connect.displayname') }}</label>
            <input
              id="login-displayname"
              type="text"
              class="form__input"
              v-model="displayname"
              autocomplete="username"
              :aria-describedby="fieldError === 'displayname' ? 'error-displayname' : undefined"
              :aria-invalid="fieldError === 'displayname'"
            />
            <p
              v-if="fieldError === 'displayname'"
              id="error-displayname"
              class="form__error"
              role="alert"
            >
              {{ $t('connect.empty_displayname') }}
            </p>
          </div>

          <!-- Password field -->
          <div v-if="!autoPassword" class="form__field">
            <label for="login-password" class="form__label">{{ $t('connect.password') }}</label>
            <input
              id="login-password"
              type="password"
              class="form__input"
              v-model="password"
              autocomplete="current-password"
            />
          </div>

          <button type="submit" class="form__submit">
            {{ $t('connect.connect') }}
          </button>
        </form>
      </div>
    </transition>
  </div>
</template>

<style lang="scss" scoped>
  // -----------------------------------------------
  // Root: Vollbild-Page (kein Overlay mehr)
  // -----------------------------------------------
  .login-page {
    position: fixed;
    inset: 0;
    background-color: var(--color-bg);
    // Layer 1: radial Spotlight-Gradient
    // Layer 2: dot-grid pattern
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
    display: flex;
    justify-content: center;
    align-items: center;
  }

  // -----------------------------------------------
  // System Dialog
  // -----------------------------------------------
  .system-dialog {
    position: absolute;
    inset: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    background: color-mix(in srgb, var(--color-bg) 60%, transparent);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    z-index: 10;

    &__card {
      width: min(340px, calc(100vw - var(--space-8)));
      padding: var(--space-8) var(--space-6);
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: var(--space-4);
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
      border: 2px solid currentColor;
      opacity: 0.9;

      svg {
        width: 28px;
        height: 28px;
      }

      &--info {
        color: var(--color-primary);
        background: color-mix(in srgb, var(--color-primary) 12%, transparent);
      }

      &--error,
      &--warning {
        color: var(--color-error);
        background: color-mix(in srgb, var(--color-error) 12%, transparent);
      }
    }

    &__title {
      font-family: var(--font-display);
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
      font-family: var(--font-body);
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
  // Login card (Glassmorphism)
  // -----------------------------------------------
  .window {
    width: min(300px, calc(100vw - var(--space-8)));
    background: color-mix(in srgb, var(--color-surface) 75%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-radius: var(--radius-lg);
    border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
    border-top-color: color-mix(in srgb, var(--color-text) 15%, transparent);
    padding: var(--space-3);
    box-shadow: var(--shadow-lg);
  }

  // -----------------------------------------------
  // Logo button
  // -----------------------------------------------
  .logo {
    width: 100%;
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    cursor: pointer;
    padding: var(--space-2) 0;
    background: none;
    border: none;
    border-radius: var(--radius-md);
    transition: opacity var(--transition-interactive);

    &:hover  { opacity: 0.8; }
    &:active { opacity: 0.6; }

    &:focus-visible {
      outline: 2px solid var(--color-primary);
      outline-offset: 3px;
    }

    img {
      height: 90px;
      width: auto;
      margin-right: var(--space-2);
    }
  }

  // -----------------------------------------------
  // Kinetic wordmark
  // -----------------------------------------------
  @keyframes weight-pulse {
    0%   { font-variation-settings: 'wght' 900; }
    100% { font-variation-settings: 'wght' 400; }
  }

  .neko-logo {
    font-family: var(--font-display);
    font-size: var(--text-xl);
    line-height: 1;
    color: var(--color-text);
    // Prevent the wordmark from being read as 5 separate letters by AT
    // (aria-hidden="true" on the parent handles this)

    span {
      display: inline-block;
      animation: weight-pulse 600ms ease-out both;

      &:nth-child(1) { animation-delay:   0ms; }
      &:nth-child(2) { animation-delay:  40ms; }
      &:nth-child(3) { animation-delay:  80ms; }
      &:nth-child(4) { animation-delay: 120ms; }
      &:nth-child(5) { animation-delay: 160ms; }
    }
  }

  // -----------------------------------------------
  // Login form
  // -----------------------------------------------
  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-top: var(--space-1);

    &__title {
      text-align: center;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--color-text-muted);
      font-family: var(--font-body);
      font-size: var(--text-sm);
      line-height: 1.6;
      margin-bottom: var(--space-1);
    }

    &__field {
      display: flex;
      flex-direction: column;
      gap: var(--space-1);
    }

    &__label {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      font-weight: 500;
      color: var(--color-text-muted);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    &__input {
      min-height: 44px;
      border: 1px solid transparent;
      padding: var(--space-2) var(--space-3);
      border-radius: var(--radius-md);
      background: color-mix(in srgb, var(--color-bg) 70%, transparent);
      color: var(--color-text);
      font-family: var(--font-body);
      font-size: var(--text-sm);
      width: 100%;
      transition:
        border-color     var(--transition-interactive),
        background-color var(--transition-interactive),
        box-shadow       var(--transition-interactive);

      &:focus {
        border-color: var(--color-primary);
        outline: none;
        box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
      }

      &[aria-invalid="true"] { border-color: var(--color-error); }

      &::placeholder { color: var(--color-text-faint); }

      &::selection { background: var(--color-primary-highlight); }
    }

    &__error {
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--color-error);
      line-height: 1.4;
    }

    &__submit {
      min-height: 44px;
      cursor: pointer;
      border-radius: var(--radius-md);
      padding: var(--space-2) var(--space-4);
      background: var(--color-primary);
      color: var(--color-text-inverse);
      text-align: center;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      font-family: var(--font-body);
      font-weight: 700;
      font-size: var(--text-sm);
      border: none;
      width: 100%;
      margin-top: var(--space-1);
      transition:
        background-color var(--transition-interactive),
        transform        var(--transition-fast);

      &:hover  { background: var(--color-primary-hover); }

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
  // Bounce loader
  // -----------------------------------------------
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
    }

    .bounce2 { animation-delay: -1s; }
  }

  @keyframes bounce {
    0%, 100% { transform: scale(0); }
    50%       { transform: scale(1); }
  }

  // -----------------------------------------------
  // prefers-reduced-motion
  // -----------------------------------------------
  @media (prefers-reduced-motion: reduce) {
    .bounce1,
    .bounce2 {
      animation: none;
      transform: scale(0.7);
      opacity: 0.4;
    }

    .dialog-fade-enter-active,
    .dialog-fade-leave-active,
    .card-fade-enter-active,
    .card-fade-leave-active {
      transition: none;
    }

    // Static weight — no animation
    .neko-logo span {
      animation: none;
      font-variation-settings: 'wght' 400;
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
    transition: opacity var(--transition-interactive);
  }
  .card-fade-enter,
  .card-fade-leave-to {
    opacity: 0;
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'

  @Component({ name: 'neko-login-page' })
  export default class extends Vue {
    private autoPassword: string | null = new URL(location.href).searchParams.get('pwd')

    displayname = ''
    password = ''
    fieldError: string | null = null

    mounted() {
      let password = this.$accessor.password
      if (this.autoPassword !== null) {
        this.removeUrlParam('pwd')
        password = this.autoPassword
      }

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

    get connecting(): boolean {
      return this.$accessor.connecting
    }

    get systemDialog() {
      return this.$accessor.client.systemDialog
    }

    dismissDialog() {
      this.$accessor.client.setSystemDialog(null)
    }

    login() {
      this.fieldError = null

      if (this.displayname.trim() === '') {
        this.fieldError = 'displayname'
        return
      }

      const password = this.autoPassword !== null ? this.autoPassword : this.password
      this.$accessor.login({ displayname: this.displayname, password })
      this.autoPassword = null
    }

    about() {
      this.$accessor.client.toggleAbout()
    }

    private removeUrlParam(param: string) {
      const url = document.location.href
      const urlparts = url.split('?')
      if (urlparts.length < 2) return

      const urlBase = urlparts.shift()!
      const queryString = urlparts.join('?')
      const prefix = encodeURIComponent(param) + '='
      const pars = queryString.split(/[&;]/g).filter((p) => p.lastIndexOf(prefix, 0) === -1)

      window.history.pushState('', document.title, urlBase + (pars.length > 0 ? '?' + pars.join('&') : ''))
    }
  }
</script>
