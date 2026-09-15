export const MEDIA_BACKEND_WEBRTC = 'webrtc'
export const MEDIA_BACKEND_WEBCODECS = 'webcodecs-ws'
export const MEDIA_BACKEND_STORAGE_KEY = 'media_backend'

/**
 * @typedef {'webrtc' | 'webcodecs-ws'} MediaBackendPreference
 */

/**
 * Invalid and absent persisted values deliberately resolve to WebRTC.
 *
 * @param {unknown} value
 * @returns {MediaBackendPreference}
 */
export function normalizeMediaBackendPreference(value) {
  return value === MEDIA_BACKEND_WEBCODECS ? MEDIA_BACKEND_WEBCODECS : MEDIA_BACKEND_WEBRTC
}

/**
 * @param {Pick<Storage, 'getItem'> | undefined} [storage]
 * @returns {MediaBackendPreference}
 */
export function readMediaBackendPreference(storage) {
  try {
    const target = storage || (typeof window !== 'undefined' ? window.localStorage : undefined)
    return normalizeMediaBackendPreference(target?.getItem(MEDIA_BACKEND_STORAGE_KEY))
  } catch {
    return MEDIA_BACKEND_WEBRTC
  }
}

/**
 * Persist a normalized choice. Storage denial is non-fatal; the next page load
 * will safely return to the WebRTC default.
 *
 * @param {unknown} value
 * @param {Pick<Storage, 'setItem'> | undefined} [storage]
 * @returns {boolean}
 */
export function writeMediaBackendPreference(value, storage) {
  try {
    const target = storage || (typeof window !== 'undefined' ? window.localStorage : undefined)
    if (!target) return false
    target.setItem(MEDIA_BACKEND_STORAGE_KEY, normalizeMediaBackendPreference(value))
    return true
  } catch {
    return false
  }
}

/**
 * The existing exact query remains a stateless diagnostic override. Any
 * present but non-exact media selector stays on the safe WebRTC path instead
 * of activating a stored experimental preference.
 *
 * @param {string} search
 * @param {unknown} storedPreference
 * @returns {{ backend: MediaBackendPreference, overridden: boolean, invalidOverride: boolean }}
 */
export function resolveMediaBackendSelection(search, storedPreference) {
  const preference = normalizeMediaBackendPreference(storedPreference)
  const parameters = new URLSearchParams(typeof search === 'string' ? search : '')
  const media = parameters.getAll('media')

  if (media.length === 0) {
    return { backend: preference, overridden: false, invalidOverride: false }
  }

  if (media.length === 1 && media[0] === MEDIA_BACKEND_WEBCODECS) {
    return { backend: MEDIA_BACKEND_WEBCODECS, overridden: true, invalidOverride: false }
  }

  return { backend: MEDIA_BACKEND_WEBRTC, overridden: true, invalidOverride: true }
}

/**
 * Selection changes are persisted and then re-enter the normal startup path.
 * Only the diagnostic media selector is removed; all other query parameters
 * and the view-only fragment are preserved.
 *
 * @param {string} href
 * @returns {string}
 */
export function mediaBackendNavigationURL(href) {
  const url = new URL(href)
  url.searchParams.delete('media')
  return url.toString()
}

/**
 * Healthy WebCodecs playback uses the compact status treatment. Negotiation,
 * recovery and terminal states remain prominent and actionable.
 *
 * @param {string} status
 */
export function useCompactMediaStatus(status) {
  return status === 'streaming'
}
