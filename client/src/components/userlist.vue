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
        @contextmenu.stop.prevent="onContext($event, { member: m })"
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
          <button
            class="action-btn"
            title="Ignore"
            aria-label="Ignore user"
            @click="onContext($event, { member: m })"
          >
            <i class="fas fa-eye-slash" aria-hidden="true" />
          </button>
          <button
            class="action-btn"
            title="Mute"
            aria-label="Mute user"
            @click="onContext($event, { member: m })"
          >
            <i class="fas fa-microphone-slash" aria-hidden="true" />
          </button>
          <button
            class="action-btn"
            title="Give Controls"
            aria-label="Give controls to user"
            @click="onContext($event, { member: m })"
          >
            <i class="fas fa-gamepad" aria-hidden="true" />
          </button>
          <button
            class="action-btn action-btn--destructive"
            title="Kick"
            aria-label="Kick user"
            @click="onContext($event, { member: m })"
          >
            <i class="fas fa-user-times" aria-hidden="true" />
          </button>
          <button
            class="action-btn action-btn--destructive"
            title="Ban IP"
            aria-label="Ban IP address"
            @click="onContext($event, { member: m })"
          >
            <i class="fas fa-ban" aria-hidden="true" />
          </button>
        </div>
      </li>
    </ul>

    <neko-context ref="context" />
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
  import { Component, Ref, Vue } from 'vue-property-decorator'

  import Context from './context.vue'
  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-userlist',
    components: {
      'neko-context': Context,
      'neko-avatar': Avatar,
    },
  })
  export default class extends Vue {
    @Ref('context') readonly _context!: any

    get self() {
      return this.$accessor.user.member
    }

    get id() {
      return this.$accessor.user.id
    }

    get host() {
      return this.$accessor.remote.id
    }

    get members() {
      return this.$accessor.user.members
    }

    get otherMembers() {
      return this.members.filter((m: any) => m.id !== this.id && m.connected)
    }

    get loading() {
      return this.$accessor.connecting
    }

    // Admin or current host may moderate
    get canModerate() {
      return (
        this.$accessor.user.admin ||
        this.$accessor.remote.id === this.$accessor.user.id
      )
    }

    onContext(event: MouseEvent, data: any) {
      this._context.open(event, data)
    }
  }
</script>
