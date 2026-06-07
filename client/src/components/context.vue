<template>
  <vue-context class="context" ref="context">
    <template slot-scope="child" v-if="child.data">
      <li class="header">
        <div class="user">
          <neko-avatar class="avatar" :seed="child.data.member.displayname" :size="25" />
          <strong>{{ child.data.member.displayname }}</strong>
        </div>
      </li>
      <li class="seperator" />
      <li>
        <span @click="ignore(child.data.member)" v-if="!child.data.member.ignored">{{ $t('context.ignore') }}</span>
        <span @click="unignore(child.data.member)" v-else>{{ $t('context.unignore') }}</span>
      </li>

      <template v-if="admin">
        <li>
          <span @click="mute(child.data.member)" v-if="!child.data.member.muted">{{ $t('context.mute') }}</span>
          <span @click="unmute(child.data.member)" v-else>{{ $t('context.unmute') }}</span>
        </li>
        <li v-if="child.data.member.id === host && !implicitHosting">
          <span @click="adminRelease(child.data.member)">{{ $t('context.release') }}</span>
        </li>
        <li v-if="child.data.member.id === host && !implicitHosting">
          <span @click="adminControl(child.data.member)">{{ $t('context.take') }}</span>
        </li>
        <li>
          <span v-if="child.data.member.id !== host && !implicitHosting" @click="adminGive(child.data.member)">{{
            $t('context.give')
          }}</span>
        </li>
      </template>
      <template v-else>
        <li v-if="hosting && !implicitHosting">
          <span @click="give(child.data.member)">{{ $t('context.give') }}</span>
        </li>
      </template>

      <template v-if="admin && !child.data.member.admin">
        <li class="seperator" />
        <li>
          <span @click="kick(child.data.member)" style="color: #f04747">{{ $t('context.kick') }}</span>
        </li>
        <li>
          <span @click="ban(child.data.member)" style="color: #f04747">{{ $t('context.ban') }}</span>
        </li>
      </template>
    </template>
  </vue-context>
</template>

<style lang="scss" scoped>
  .context {
    background-color: var(--bg-obsidian-floating);
    backdrop-filter: blur(16px);
    background-clip: padding-box;
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-inner);
    display: block;
    margin: 0;
    padding: 6px;
    min-width: 160px;
    z-index: 1500;
    position: fixed;
    list-style: none;
    box-sizing: border-box;
    max-height: calc(100% - 50px);
    overflow-y: auto;
    color: var(--text-subtle);
    user-select: none;
    box-shadow: var(--elevation-medium);

    > li {
      margin: 2px 0;
      position: relative;
      align-content: center;

      &.header {
        .user {
          display: flex;
          flex-direction: row;
          align-items: center;
          padding: 6px 8px;
          gap: 8px;

          .avatar {
            width: 26px;
            height: 26px;
            border-radius: 50%;
            border: 1px solid rgba(255, 255, 255, 0.1);
          }

          strong {
            line-height: normal;
            font-size: 13.5px;
            font-weight: 700;
            color: var(--text-pure);
            max-width: 200px;
            text-overflow: ellipsis;
            overflow: hidden;
          }
        }
      }

      &.seperator {
        height: 1px;
        background: rgba(255, 255, 255, 0.08);
        margin: 6px 4px;
      }

      > span {
        cursor: pointer;
        display: block;
        padding: 6px 12px;
        font-size: 13.5px;
        font-weight: 500;
        text-decoration: none;
        white-space: nowrap;
        background-color: transparent;
        border-radius: var(--radius-button);
        transition: var(--transition-fluid);

        &:hover,
        &:focus {
          text-decoration: none;
          background-color: rgba(255, 255, 255, 0.05);
          color: var(--color-cyber-mint);
          text-shadow: 0 0 6px var(--color-cyber-mint-glow);
        }

        &:focus {
          outline: 0;
        }
      }
    }

    &:focus {
      outline: 0;
    }
  }
</style>

<script lang="ts">
  import { Component, Ref, Vue } from 'vue-property-decorator'
  import { Member } from '~/neko/types'

  // @ts-ignore
  import { VueContext } from 'vue-context'
  import Avatar from './avatar.vue'

  @Component({
    name: 'neko-context',
    components: {
      'vue-context': VueContext,
      'neko-avatar': Avatar,
    },
  })
  export default class NekoContext extends Vue {
    @Ref('context') readonly context!: any

    get admin() {
      return this.$accessor.user.admin
    }

    get hosting() {
      return this.$accessor.remote.hosting
    }

    get host() {
      return this.$accessor.remote.id
    }

    get implicitHosting() {
      return this.$accessor.remote.implicitHosting
    }

    open(event: MouseEvent, data: any) {
      this.context.open(event, data)
    }

    async kick(member: Member) {
      const value = await this.$swal({
        title: this.$t('context.confirm.kick_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.kick_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })

      if (value) {
        this.$accessor.user.kick(member)
      }
    }

    async ban(member: Member) {
      const value = await this.$swal({
        title: this.$t('context.confirm.ban_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.ban_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })

      if (value) {
        this.$accessor.user.ban(member)
      }
    }

    async mute(member: Member) {
      const value = await this.$swal({
        title: this.$t('context.confirm.mute_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.mute_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })

      if (value) {
        this.$accessor.user.mute(member)
      }
    }

    async unmute(member: Member) {
      const value = await this.$swal({
        title: this.$t('context.confirm.unmute_title', { name: member.displayname }) as string,
        text: this.$t('context.confirm.unmute_text', { name: member.displayname }) as string,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: this.$t('context.confirm.button_yes') as string,
        cancelButtonText: this.$t('context.confirm.button_cancel') as string,
      })

      if (value) {
        this.$accessor.user.unmute(member)
      }
    }

    adminRelease() {
      this.$accessor.remote.adminRelease()
    }

    adminControl() {
      this.$accessor.remote.adminControl()
    }

    adminGive(member: Member) {
      this.$accessor.remote.adminGive(member)
    }

    give(member: Member) {
      this.$accessor.remote.give(member)
    }

    ignore(member: Member) {
      this.$accessor.user.setIgnored({ id: member.id, ignored: true })
    }

    unignore(member: Member) {
      this.$accessor.user.setIgnored({ id: member.id, ignored: false })
    }
  }
</script>
