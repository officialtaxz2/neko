import { mutationTree } from 'typed-vuex'
import type { HLSMode } from '~/neko/messages'

export const namespaced = true
export type HLSStatus = 'off' | 'idle' | 'negotiating' | 'connecting' | 'streaming' | 'paused' | 'terminal'

export const state = () => ({
  selected: undefined as HLSMode | undefined,
  availableModes: [] as HLSMode[],
  status: 'off' as HLSStatus,
  detail: '',
  player: '' as '' | 'native' | 'mse',
})

export const mutations = mutationTree(state, {
  select(state, mode?: HLSMode) {
    state.selected = mode
    state.status = mode ? 'idle' : 'off'
    state.detail = ''
    state.player = ''
  },
  advertise(state, modes: HLSMode[]) {
    state.availableModes = modes.slice()
  },
  setStatus(state, { status, detail = '' }: { status: HLSStatus; detail?: string }) {
    state.status = status
    state.detail = detail
  },
  setPlayer(state, player: '' | 'native' | 'mse') { state.player = player },
  reset(state) {
    state.availableModes = []
    state.status = state.selected ? 'idle' : 'off'
    state.detail = ''
    state.player = ''
  },
})
