<template>
  <div class="login-layout">
    <!-- Left: Branding Panel (desktop only) -->
    <div class="login-branding-panel" aria-hidden="true">
      <!-- Spotlight + dot grid via ::before / ::after pseudo-elements in CSS -->
      <div class="branding-content">
        <!-- Wordmark (non-interactive copy — purely decorative) -->
        <div class="branding-wordmark">
          <img src="@/assets/images/logo.svg" alt="" width="64" height="64" loading="lazy" />
          <span class="branding-logo" aria-hidden="true">
            <span>n</span><span>.</span><span>e</span><span>k</span><span>o</span>
          </span>
        </div>

        <p class="branding-tagline">
          Your browser.<br />Your session.
        </p>

        <ul class="branding-pills" role="list">
          <li>
            <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" width="14" height="14">
              <rect x="3" y="11" width="18" height="11" rx="2" stroke="currentColor" stroke-width="1.5"/>
              <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            Self-hosted
          </li>
          <li>
            <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" width="14" height="14">
              <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.5"/>
              <path d="M12 6v6l4 2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
            Open Source
          </li>
        </ul>
      </div>
    </div>

    <!-- Right: Login page (existing, unchanged logic) -->
    <div class="login-page" ref="loginPage">
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

          <!-- About link (replaces the logo button that was above the form) -->
          <button
            type="button"
            class="about-link"
            @click.stop.prevent="about"
            :title="$t('connect.login_title')"
            aria-label="About n.eko"
          >n.eko</button>
        </div>
      </transition>
    </div>
  </div>
</template>

<style lang="scss" scoped>
  // -----------------------------------------------
  // Split-screen layout (Phase 3.1)
  // -----------------------------------------------
  .login-layout {
    display: grid;
    // Left branding panel takes half; right login panel is sized to its content
    // so no dead empty space appears to the right of the card.
    grid-template-columns: 1fr auto;
    min-height: 100dvh;
  }

  @media (max-width: 768px) {
    .login-layout { grid-template-columns: 1fr; }
    .login-branding-panel { display: none; }
  }

  // -----------------------------------------------
  // Left: Branding Panel
  // -----------------------------------------------
  .login-branding-panel {
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: var(--space-16);
    background: var(--color-surface);
    position: relative;
    overflow: hidden;

    // Layer 1: spotlight gradient centred on panel
    &::before {
      content: '';
      position: absolute;
      inset: 0;
      background-image:
        radial-gradient(
          ellipse 80% 70% at 50% 50%,
          color-mix(in srgb, var(--color-primary) 10%, transparent) 0%,
          transparent 70%
        );
      pointer-events: none;
    }

    // Layer 2: dot grid
    &::after {
      content: '';
      position: absolute;
      inset: 0;
      background-image: radial-gradient(
        circle,
        color-mix(in srgb, var(--color-primary) 20%, transparent) 1.5px,
        transparent 1.5px
      );
      background-size: 28px 28px;
      opacity: 0.04;
      pointer-events: none;
    }
  }

  .branding-content {
    position: relative; // above pseudo-elements
    z-index: 1;
    display: flex;
    flex-direction: column;
    gap: var(--space-8);
  }

  .branding-wordmark {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: var(--space-3);

    img {
      width: 64px;
      height: 64px;
    }
  }

  // Decorative wordmark on branding panel — same weight-pulse as login card
  @keyframes weight-pulse {
    0%   { font-variation-settings: 'wght' 900; }
    100% { font-variation-settings: 'wght' 400; }
  }

  .branding-logo {
    font-family: var(--font-display);
    font-size: var(--text-xl);
    line-height: 1;
    color: var(--color-text);

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

  .branding-tagline {
    font-family: var(--font-display);
    font-size: var(--text-xl);
    font-weight: 600;
    color: var(--color-text);
    line-height: 1.25;
    max-width: 16ch;
  }

  .branding-pills {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    gap: var(--space-2);
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      display: inline-flex;
      align-items: center;
      gap: var(--space-1);
      padding: var(--space-1) var(--space-3);
      border-radius: var(--radius-full);
      background: color-mix(in srgb, var(--color-surface-offset) 80%, transparent);
      border: 1px solid color-mix(in srgb, var(--color-border) 60%, transparent);
      font-family: var(--font-body);
      font-size: var(--text-xs);
      color: var(--color-text-muted);

      svg { flex-shrink: 0; }
    }
  }

  // -----------------------------------------------
  // Right: Login page wrapper
  // -----------------------------------------------
  .login-page {
    position: relative;
    // Ensure the right panel is at least wide enough for the card + generous padding.
    // Without a min-width the "auto" column would hug the card tightly on narrow viewports.
    min-width: min(460px, 50vw);
    background-color: var(--color-bg);
    // Spotlight + dot grid (right panel only)
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
    background-position:
      center,
      calc(50% + var(--mx, 0) * 12px) calc(50% + var(--my, 0) * 12px);
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

      &:hover { background: var(--color-primary-hover); }

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
  // Login card (Glassmorphism + parallax tilt)
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
    perspective: 1000px;
    transform:
      rotateY(calc(var(--mx, 0) * 4deg))
      rotateX(calc(var(--my, 0) * -4deg));
    transition: transform 80ms linear;
    will-change: transform;
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
  // About link (replaces the removed logo button)
  // -----------------------------------------------
  .about-link {
    display: block;
    width: 100%;
    margin-top: var(--space-3);
    padding: var(--space-1) 0;
    text-align: center;
    font-family: var(--font-body);
    font-size: var(--text-xs);
    color: var(--color-text-faint);
    background: none;
    border: none;
    cursor: pointer;
    transition: color var(--transition-interactive);

    &:hover { color: var(--color-text-muted); }

    &:focus-visible {
      outline: 2px solid var(--color-primary);
      outline-offset: 3px;
      border-radius: var(--radius-sm);
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

    .branding-logo span {
      animation: none;
      font-variation-settings: 'wght' 400;
    }

    .window {
      transform: none !important;
      transition: none !important;
    }

    .login-page {
      background-position: center, center !important;
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
    private _parallaxHandler: ((e: MouseEvent) => void) | null = null

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

      // Parallax on right panel only — guarded by prefers-reduced-motion
      const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
      if (!prefersReduced) {
        const el = this.$refs.loginPage as HTMLElement
        this._parallaxHandler = (e: MouseEvent) => {
          const x = e.clientX / window.innerWidth  - 0.5
          const y = e.clientY / window.innerHeight - 0.5
          el.style.setProperty('--mx', String(x))
          el.style.setProperty('--my', String(y))
        }
        el.addEventListener('mousemove', this._parallaxHandler)
      }
    }

    beforeDestroy() {
      if (this._parallaxHandler) {
        const el = this.$refs.loginPage as HTMLElement
        el.removeEventListener('mousemove', this._parallaxHandler)
        this._parallaxHandler = null
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
