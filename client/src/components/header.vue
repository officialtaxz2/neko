<template>
  <div class="header">
    <a href="https://github.com/officialtaxz2/neko" title="Github repository" target="_blank" class="neko">
      <img src="@/assets/images/logo.svg" alt="n.eko" />
      <span><b>n</b>.eko</span>
    </a>
    <ul class="menu">
      <li>
        <i
          :class="[{ disabled: !admin }, { locked: isLocked('control') }, 'fas', 'fa-mouse']"
          @click="toggleLock('control')"
          v-tooltip="{
            content: lockedTooltip('control'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        />
      </li>
      <li>
        <i
          :class="[{ disabled: !admin }, { locked: isLocked('login') }, locked ? 'fa-lock' : 'fa-lock-open', 'fas']"
          @click="toggleLock('login')"
          v-tooltip="{
            content: lockedTooltip('login'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        />
      </li>
      <li v-if="fileTransfer">
        <i
          :class="[{ disabled: !admin }, { locked: isLocked('file_transfer') }, 'fas', 'fa-file']"
          @click="toggleLock('file_transfer')"
          v-tooltip="{
            content: lockedTooltip('file_transfer'),
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        />
      </li>

      <!-- Theme toggle (dark / light) -->
      <li>
        <button
          class="theme-toggle"
          :aria-label="currentTheme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="$emit('toggle-theme')"
          v-tooltip="{
            content: currentTheme === 'dark' ? 'Light mode' : 'Dark mode',
            placement: 'bottom',
            offset: 5,
            boundariesElement: 'body',
            delay: { show: 300, hide: 100 },
          }"
        >
          <!-- Sun icon (shown in dark mode → switch to light) -->
          <svg v-if="currentTheme === 'dark'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="5"/>
            <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
          </svg>
          <!-- Moon icon (shown in light mode → switch to dark) -->
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </button>
      </li>

      <li>
        <span v-if="showBadge" class="badge">&bull;</span>
        <i class="fas fa-bars toggle" @click="toggleMenu" />
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  .header {
    flex: 1;
    display: flex;
    flex-direction: row;
    align-items: center;
    // Gradient-System: diagonaler Teal-Tint vom Logo-Bereich (links) zur reinen bg-Farbe (rechts)
    // 135deg erzeugt einen Diagonalverlauf von oben-links nach unten-rechts
    background: linear-gradient(
      135deg,
      color-mix(in srgb, var(--color-primary) 7%, var(--color-bg)) 0%,
      var(--color-bg) 55%
    );
    // Glassmorphism-Konsistenz: Topbar soll mit dem Sidebar-Glas harmonieren
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border-bottom: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent);

    .neko {
      flex: 1;
      display: flex;
      justify-content: flex-start;
      align-items: center;
      width: 150px;
      margin-left: var(--space-5);
      color: var(--color-text);
      text-decoration: none;
      transition: color var(--transition-interactive);

      &:hover {
        color: var(--color-primary);
      }

      img {
        display: block;
        float: left;
        height: 30px;
        margin-right: var(--space-2);
      }

      span {
        font-size: 30px;
        line-height: 30px;
        font-family: var(--font-display);

        b {
          font-weight: 900;
        }
      }
    }

    .menu {
      justify-self: flex-end;
      margin-right: var(--space-3);
      white-space: nowrap;
      display: flex;
      align-items: center;

      li {
        display: inline-flex;
        align-items: center;
        margin-right: var(--space-2);

        i {
          display: block;
          width: 30px;
          height: 30px;
          text-align: center;
          line-height: 32px;
          border-radius: var(--radius-md);
          cursor: pointer;
          color: var(--color-text-muted);
          // Micro-animation: smooth color + scale on hover (Kat. 1)
          transition:
            color            var(--transition-interactive),
            background-color var(--transition-interactive),
            transform        var(--transition-fast);

          &:hover:not(.disabled) {
            color: var(--color-text);
            background: var(--color-surface-offset);
            transform: scale(1.08);
          }

          &:active:not(.disabled) {
            transform: scale(0.95);
          }
        }

        .disabled {
          cursor: default;
          opacity: 0.5;
        }

        .locked {
          color: var(--color-error);
          opacity: 0.7;
        }

        // Hamburger menu toggle
        .toggle {
          background: color-mix(in srgb, var(--color-surface-2) 80%, transparent);
          border: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent);
          color: var(--color-text-muted);

          &:hover {
            background: var(--color-surface-offset);
            color: var(--color-text);
            border-color: var(--color-primary);
          }
        }

        // Dark / Light theme toggle button
        .theme-toggle {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          width: 30px;
          height: 30px;
          border-radius: var(--radius-md);
          background: color-mix(in srgb, var(--color-surface-2) 80%, transparent);
          border: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent);
          color: var(--color-text-muted);
          cursor: pointer;
          padding: 0;
          // Micro-animation (Kat. 1)
          transition:
            color            var(--transition-interactive),
            background-color var(--transition-interactive),
            border-color     var(--transition-interactive),
            transform        var(--transition-fast);

          &:hover {
            color: var(--color-primary);
            background: var(--color-primary-highlight);
            border-color: var(--color-primary);
            transform: scale(1.08);
          }

          &:active {
            transform: scale(0.92);
          }

          svg {
            pointer-events: none;
          }
        }

        .badge {
          position: absolute;
          background: var(--color-error);
          color: var(--color-text-inverse);
          font-weight: bold;
          font-size: 1.25em;
          border-radius: var(--radius-full);
          width: 20px;
          height: 20px;
          text-align: center;
          line-height: 20px;
          pointer-events: none;
          transform: translate(-50%, -25%) scale(1);
          box-shadow: 0 0 0 0 rgba(0, 0, 0, 1);
          animation: badger-pulse 2s infinite;
        }

        @keyframes badger-pulse {
          0% {
            transform: translate(-50%, -25%) scale(0.85);
            box-shadow: 0 0 0 0 rgba(0, 0, 0, 0.7);
          }
          70% {
            transform: translate(-50%, -25%) scale(1);
            box-shadow: 0 0 0 10px rgba(0, 0, 0, 0);
          }
          100% {
            transform: translate(-50%, -25%) scale(0.85);
            box-shadow: 0 0 0 0 rgba(0, 0, 0, 0);
          }
        }
      }
    }
  }
</style>

<script lang="ts">
  import { Component, Vue, Prop } from 'vue-property-decorator'
  import { AdminLockResource } from '~/neko/messages'

  @Component({ name: 'neko-header' })
  export default class extends Vue {
    @Prop({ type: String, default: 'dark' }) readonly currentTheme!: string

    get admin() {
      return this.$accessor.user.admin
    }

    get locked() {
      return this.$accessor.locked
    }

    get side() {
      return this.$accessor.client.side
    }

    get texts() {
      return this.$accessor.chat.texts
    }

    get showBadge() {
      return !this.side && this.readTexts != this.texts
    }

    get fileTransfer() {
      return this.$accessor.remote.fileTransfer
    }

    toggleLock(resource: AdminLockResource) {
      this.$accessor.toggleLock(resource)
    }

    isLocked(resource: AdminLockResource): boolean {
      return this.$accessor.isLocked(resource)
    }

    readTexts: number = 0
    toggleMenu() {
      this.$accessor.client.toggleSide()
      this.readTexts = this.texts
    }

    lockedTooltip(resource: AdminLockResource) {
      if (this.admin) {
        return this.$t(`locks.${resource}.` + (this.isLocked(resource) ? `unlock` : `lock`))
      }
      return this.$t(`locks.${resource}.` + (this.isLocked(resource) ? `locked` : `unlocked`))
    }
  }
</script>
