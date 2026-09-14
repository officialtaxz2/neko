import { mutationTree } from 'typed-vuex'

export const namespaced = true

export type WebCodecsMediaStatus =
  | 'off'
  | 'idle'
  | 'negotiating'
  | 'connecting'
  | 'streaming'
  | 'reconnecting'
  | 'terminal'

export const state = () => ({
  selected: false,
  status: 'off' as WebCodecsMediaStatus,
  detail: '',
  retryAttempt: 0,
  audioEnabled: false,
})

export const mutations = mutationTree(state, {
  select(state, selected: boolean) {
    state.selected = selected
    state.status = selected ? 'idle' : 'off'
    state.detail = ''
    state.retryAttempt = 0
    state.audioEnabled = false
  },

  setStatus(
    state,
    { status, detail = '', retryAttempt = 0 }: { status: WebCodecsMediaStatus; detail?: string; retryAttempt?: number },
  ) {
    state.status = status
    state.detail = detail
    state.retryAttempt = retryAttempt
  },

  setAudioEnabled(state, enabled: boolean) {
    state.audioEnabled = enabled
  },

  reset(state) {
    state.status = state.selected ? 'idle' : 'off'
    state.detail = ''
    state.retryAttempt = 0
    state.audioEnabled = false
  },
})
