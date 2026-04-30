<template>
  <div class="neko-userlist">
    <!-- ── Skeleton: connecting ──────────────────────────────────────── -->
    <ul v-if="loading" class="userlist userlist--skeleton" aria-hidden="true">
      <li v-for="n in 4" :key="'sk' + n" class="userlist-row userlist-row--skeleton">
        <div class="skeleton sk-avatar"></div>
        <div class="skeleton sk-name"></div>
      </li>
    </ul>

    <!-- ── Real user list ────────────────────────────────────────────── -->
    <ul v-else class="userlist" role="list">
      <!-- Self -->
      <li v-if="self" class="userlist-row userlist-row--self">
        <neko-avatar class="user-avatar" :seed="self.displayname" :size="28" />
        <span class="user-name">{{ self.displayname }}</span>
        <span class="user-you-badge">{{ $t('you') || 'You' }}</span>
      </li>

      <!-- Other connected members -->
      <li
        v-for="(m, index) in otherMembers"
        :key="m.id || index"
        class="userlist-row"
        :class="{ 'is-host': m.id === host, 'is-admin': m.admin }"
        @contextmenu.stop.prevent
      >
        <neko-avatar class="user-avatar" :seed="m.displayname" :size="28" />
        <span class="user-name">{{ m.displayname }}</span>

        <!-- Host crown -->
        <i
          v-if="m.id === host"
          class="fas fa-crown user-crown"
          aria-label="Host"
        />

        <!-- Moderation actions: only rendered for admin/host, revealed on hover -->
        <div v-if="canModerate" class="user-actions" @click.stop>
          <!-- Ignore / Unignore -->
          <button
            class="action-btn"
            :title="m.ignored ? 'Unignore' : 'Ignore'"
            :aria-label="m.ignored ? 'Unignore user' : 'Ignore user'"
            @click="toggleIgnore(m)"
          >
            <i :class="m.ignored ? 'fas fa-eye' : 'fas fa-eye-slash'" aria-hidden="true" />
          </button>

          <!-- Mute / Unmute (admin only) -->
          <button
            v-if="isAdmin"
            class="action-btn"
            :title="m.muted ? 'Unmute' : 'Mute'"
            :aria-label="m.muted ? 'Unmute user' : 'Mute user'"
            @click="toggleMute(m)"
          >
            <i :class="m.muted ? 'fas fa-microphone' : 'fas fa-microphone-slash'" aria-hidden="true" />
          </button>

          <!-- Give Controls: hidden if member already has host or implicit hosting active -->
          <button
            v-if="!implicitHosting && m.id !== host"
            class="action-btn"
            title="Give Controls"
            aria-label="Give controls to user"
            @click="give(m)"
          >
            <i class="fas fa-gamepad" aria-hidden="true" />
          </button>

          <!-- Kick (admin only, non-admin target) -->
          <button
            v-if="isAdmin && !m.admin"
            class="action-btn action-btn--destructive"
            title="Kick"
            aria-label="Kick user"
            @click="kick(m)"
          >
            <i class="fas fa-user-times" aria-hidden="true" />
          </button>

          <!-- Ban IP (admin only, non-admin target) -->
          <button
            v-if="isAdmin && !m.admin"
            class="action-btn action-btn--destructive"
            title="Ban IP"
            aria-label="Ban IP address"
            @click="ban(m)"
          >
            <i class="fas fa-ban" aria-hidden="true" />
          </button>
        </div>
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
  // ── Shimmer ──────────────────────────────────────────────────────────────
  @keyframes shimmer {
    0%   { background-position: -200% 0; }
    100% { background-position:  200% 0; }
  }

  .skeleton {
    background: linear-gradient(
      90deg,
      var(--color-surface-offset)  25%,
      var(--color-surface-dynamic) 50%,
      var(--color-surface-offset)  75%
    );
    background-size: 200% 100%;
    animation: shimmer 1.5s ease-in-out infinite;
    border-radius: var(--radius-sm);

    @media (prefers-reduced-motion: reduce) {
      animation: none;
    }
  }

  .sk-avatar {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-full);
    flex-shrink: 0;
  }

  .sk-name {
    flex: 1;
    height: 0.875rem;
    border-radius: var(--radius-sm);
    max-width: 140px;
  }

  // ── Root container: fills parent panel (flex:1 from side.vue) ────────────
  .neko-userlist {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    min-height: 0;
    scrollbar-width: thin;
    scrollbar-color: var(--color-surface-offset) transparent;

    &::-webkit-scrollbar       { width: 4px; }
    &::-webkit-scrollbar-track { background: transparent; }
    &::-webkit-scrollbar-thumb {
      background: var(--color-surface-offset);
      border-radius: var(--radius-full);

      &:hover { background: var(--color-surface-dynamic); }
    }
  }

  // ── List ─────────────────────────────────────────────────────────────────
  .userlist {
    list-style: none;
    padding: var(--space-2) 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0;

    &--skeleton {
      padding: var(--space-3) 0;
    }
  }

  // ── Row ──────────────────────────────────────────────────────────────────
  .userlist-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    min-height: 44px; // touch-target floor
    border-radius: var(--radius-md);
    margin: 0 var(--space-2);
    cursor: default;
    transition: background-color var(--transition-interactive);

    &:hover {
      background: color-mix(in srgb, var(--color-surface-offset) 55%, transparent);

      // Reveal action icons on hover
      .user-actions { opacity: 1; pointer-events: auto; }
    }

    // Self row: slightly subdued, no hover effect
    &--self {
      cursor: default;

      .user-name { color: var(--color-text-muted); }

      &:hover { background: transparent; }
    }

    // Skeleton row: no hover
    &--skeleton { cursor: default; &:hover { background: transparent; } }
  }

  // ── Avatar ───────────────────────────────────────────────────────────────
  .user-avatar {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-full);
    overflow: hidden;
    flex-shrink: 0;
  }

  // ── Name ─────────────────────────────────────────────────────────────────
  .user-name {
    flex: 1;
    font-size: var(--text-sm);
    font-family: var(--font-body);
    font-weight: 500;
    color: var(--color-text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }

  // ── "You" badge (self row) ───────────────────────────────────────────────
  .user-you-badge {
    font-size: var(--text-xs);
    font-family: var(--font-body);
    font-weight: 600;
    color: var(--color-text-faint);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    flex-shrink: 0;
  }

  // ── Host crown icon ──────────────────────────────────────────────────────
  .user-crown {
    font-size: var(--text-xs);
    color: var(--color-primary);
    flex-shrink: 0;
    transition: color var(--transition-interactive);
  }

  // ── Moderation action icons ───────────────────────────────────────────────
  .user-actions {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    flex-shrink: 0;
    // Hidden by default; shown on row hover via parent :hover rule
    opacity: 0;
    pointer-events: none;
    transition: opacity var(--transition-interactive);
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    // Touch-target: min 44×44px
    width: 28px;
    height: 28px;
    min-width: 44px;
    min-height: 44px;
    padding: 0;
    background: none;
    border: none;
    border-radius: var(--radius-md);
    color: var(--color-text-muted);
    cursor: pointer;
    transition:
      color            var(--transition-interactive),
      background-color var(--transition-interactive),
      transform        var(--transition-interactive);

    i {
      font-size: var(--text-xs);
      pointer-events: none;
      transition: transform var(--transition-interactive);
    }

    &:hover {
      background: color-mix(in srgb, var(--color-surface-dynamic) 70%, transparent);
      color: var(--color-text);

      i { transform: scale(1.18); }
    }

    &:active i { transform: scale(0.88); }

    &:focus-visible {
      outline: 2px solid var(--color-primary);
      outline-offset: -2px;
    }

    // Destructive actions (Kick, Ban): red on hover
    &--destructive:hover {
      color: var(--color-error);
      background: color-mix(in srgb, var(--color-error) 10%, transparent);
    }
  }
</style>

<script lang="ts">
  import { Component, Vue } from 'vue-property-decorator'
  import { Member } from '~/neko/types'

  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-userlist',
    components: {
      'neko-avatar': Avatar,
    },
  })
  export default class extends Vue {
    get self() {
      return this.$accessor.user.member
    }

    get id() {
      return this.$accessor.user.id
    }

    get host() {
      return this.$accessor.remote.id
    }

    get isAdmin() {
      return this.$accessor.user.admin
    }

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
    }

    get members() {
      return this.$accessor.user.members
    }

    get otherMembers() {
      // members is a { [id: string]: Member } dict — Object.values() converts to Member[]
      return Object.values(this.members).filter((m: any) => m.id !== this.id && m.connected)
    }

    get loading() {
      return this.$accessor.connecting
    }

    // Admin or current host may see moderation actions
    get canModerate() {
      return (
        this.$accessor.user.admin ||
        this.$accessor.remote.id === this.$accessor.user.id
      )
    }

    toggleIgnore(member: Member) {
      this.$accessor.user.setIgnored({ id: member.id, ignored: !member.ignored })
    }

    async toggleMute(member: Member) {
      if (member.muted) {
        const ok = await this.$swal({
          title: this.$t('context.confirm.unmute_title', { name: member.displayname }) as string,
          text: this.$t('context.confirm.unmute_text', { name: member.displayname }) as string,
          icon: 'warning',
          showCancelButton: true,
          confirmButtonText: this.$t('context.confirm.button_yes') as string,
          cancelButtonText: this.$t('context.confirm.button_cancel') as string,
        })
        if (ok) this.$accessor.user.unmute(member)
      } else {
        const ok = await this.$swal({
          title: this.$t('context.confirm.mute_title', { name: member.displayname }) as string,
          text: this.$t('context.confirm.mute_text', { name: member.displayname }) as string,
          icon: 'warning',
          showCancelButton: true,
          confirmButtonText: this.$t('context.confirm.button_yes') as string,
          cancelButtonText: this.$t('context.confirm.button_cancel') as string,
        })
        if (ok) this.$accessor.user.mute(member)
      }
    }

    give(member: Member) {
      // Admin uses privileged give; non-admin host uses standard give
      if (this.$accessor.user.admin) {
        this.$accessor.remote.adminGive(member)
      } else {
        this.$accessor.remote.give(member)
      }
    }

    async kick(member: Member) {
      const ok = await this.$swal({
        title: this.$t('context.confirm.kick_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.kick_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })
      if (ok) this.$accessor.user.kick(member)
    }

    async ban(member: Member) {
      const ok = await this.$swal({
        title: this.$t('context.confirm.ban_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.ban_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })
      if (ok) this.$accessor.user.ban(member)
    }
  }
</script>
