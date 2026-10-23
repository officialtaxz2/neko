import { useStore } from 'zustand'
import { useNeko } from '@/state/provider'
import { mediaStore } from '@/state/media'

type MediaState = ReturnType<ReturnType<typeof mediaStore>['getState']>

// the stream's path for components: what was selected, what the server offers, how HLS is doing
export function useMedia<T>(select: (s: MediaState) => T) {
  return useStore(mediaStore(useNeko()), select)
}

export const modeLabel = (mode: string) => (mode === 'll-hls' ? 'LL-HLS' : 'HLS')
