import { useStore } from 'zustand'
import { useNeko } from '@/state/provider'

// the fork: whether this device gets the touch controls (the on-screen keyboard, the special keys,
// the trackpad): a touch screen without a mouse, or any device with "force touch" on
export function useTouch() {
  const { client, app } = useNeko()
  const forced = useStore(app, (s) => s.settings.force_touch)
  return forced || client.isTouchDevice
}
