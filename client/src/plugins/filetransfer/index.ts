import { FileIcon } from '@phosphor-icons/react'
import { event } from '@/state/bus'
import { t } from '@/i18n'
import type { Plugin } from '@/plugins/types'
import { Files } from './Files'
import { FileLock } from './FileLock'
import { store } from './store'
import { useAllowed, useEnabled, locked, toggleLock, refresh } from './actions'

export const filetransfer: Plugin = {
  id: 'filetransfer',
  ns: 'files',
  tab: { id: 'files', icon: FileIcon, label: 'files:tab', component: Files, useVisible: useAllowed },
  topBar: [{ id: 'filetransfer-lock', component: FileLock }],
  // the fork: the same lock as a switch in the settings, as the Vue client had it
  roomLocks: [
    {
      id: 'file_transfer',
      label: 'files:locks.locked',
      replaces: 'filetransfer-lock',
      useAvailable: useEnabled,
      locked,
      toggle: toggleLock,
    },
  ],
  onEvent(neko, ...[event, payload]) {
    if (event === 'filetransfer/update') store(neko).setState({ files: payload })
  },
  init(neko) {
    const { client, bus } = neko
    // the list is requested once the session is known; a lock change becomes an event line
    let last: boolean | null = null
    client.store.subscribe(
      (s) => s.session_id,
      (id) => {
        if (!id) return
        last = locked(client.state.settings)
        refresh(neko)
      },
    )
    client.events.on('room.settings.updated', (next, id) => {
      const now = locked(next)
      if (last !== null && id && now !== last) event(neko, id, t(`files:locks.notif_${now ? 'locked' : 'unlocked'}`))
      last = now
    })
    bus.on('logout', () => store(neko).setState({ uploads: [] }))
  },
}
