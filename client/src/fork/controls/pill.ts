import { selectControlling, selectIsAdmin, selectSession, type State } from '@m1k1o/neko'

// what the control pill says: nobody controls (free), you do (mine), someone else does (other),
// an admin locked control (locked), or implicit hosting where a click into the picture takes it
export type PillState = 'free' | 'mine' | 'other' | 'locked' | 'implicit'

export function pillState(s: State): PillState {
  const controlLocked = s.settings.locked_controls && !selectIsAdmin(s)
  const hosting = selectControlling(s)
  if (s.settings.implicit_hosting) {
    if (controlLocked) return 'locked'
    return hosting ? 'mine' : 'implicit'
  }
  if (hosting) return 'mine'
  if (controlLocked) return 'locked'
  const host = s.control.host_id
  if (host && s.sessions[host]) return 'other'
  return 'free'
}

// only for a member who may control at all (not a view-only session)
export const pillVisible = (s: State) => !!selectSession(s)?.profile.can_host
