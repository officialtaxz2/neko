import { getterTree, mutationTree, actionTree } from 'typed-vuex'
import { get, set } from '~/utils/localstorage'
import { EVENT } from '~/neko/events'
import { accessor } from '~/store'

export const namespaced = true

interface KeyboardLayouts {
  [code: string]: string
}

export const state = () => {
  return {
    // Clamp persisted scroll to the new max of 50 so users who had a
    // value > 50 saved from the old range don't end up out-of-bounds.
    scroll: Math.min(get<number>('scroll', 10), 50),
    scroll_invert: get<boolean>('scroll_invert', true),
    trackpad_mode: get<boolean>('trackpad_mode', false),
    autoplay: get<boolean>('autoplay', true),
    ignore_emotes: get<boolean>('ignore_emotes', false),
    chat_sound: get<boolean>('chat_sound', true),
    keyboard_layout: get<string>('keyboard_layout', 'us'),

    keyboard_layouts_list: {} as KeyboardLayouts,

    broadcast_is_active: false,
    broadcast_url: '',
  }
}

export const getters = getterTree(state, {
  trackpad_mode: (state) => state.trackpad_mode,
})

export const mutations = mutationTree(state, {
  setScroll(state, scroll: number) {
    // Enforce range 1-50 at mutation level
    const clamped = Math.max(1, Math.min(scroll, 50))
    state.scroll = clamped
    set('scroll', clamped)
  },

  setInvert(state, value: boolean) {
    state.scroll_invert = value
    set('scroll_invert', value)
  },

  setTrackpadMode(state, value: boolean) {
    state.trackpad_mode = value
    set('trackpad_mode', value)
  },

  setAutoplay(state, value: boolean) {
    state.autoplay = value
    set('autoplay', value)
  },

  setIgnore(state, value: boolean) {
    state.ignore_emotes = value
    set('ignore_emotes', value)
  },

  setSound(state, value: boolean) {
    state.chat_sound = value
    set('chat_sound', value)
  },

  setKeyboardLayout(state, value: string) {
    state.keyboard_layout = value
    set('keyboard_layout', value)
  },

  setKeyboardLayoutsList(state, value: KeyboardLayouts) {
    state.keyboard_layouts_list = value
  },
  setBroadcastStatus(state, { url, isActive }) {
    state.broadcast_url = url
    state.broadcast_is_active = isActive
  },
})

export const actions = actionTree(
  { state, getters, mutations },
  {
    async initialise() {
      try {
        const req = await $http.get<KeyboardLayouts>('keyboard_layouts.json')
        accessor.settings.setKeyboardLayoutsList(req.data)
      } catch (err: any) {
        console.error(err)
      }
    },

    broadcastStatus(store, { url, isActive }) {
      accessor.settings.setBroadcastStatus({ url, isActive })
    },
    broadcastCreate(store, url: string) {
      $client.sendMessage(EVENT.BROADCAST.CREATE, { url })
    },
    broadcastDestroy() {
      $client.sendMessage(EVENT.BROADCAST.DESTROY)
    },
  },
)
