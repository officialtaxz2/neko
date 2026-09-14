import { RECONNECT_DELAYS_MS, reconnectDelayForAttempt } from '../recovery.js'

export const MEDIA_RETRY_DELAYS_MS = RECONNECT_DELAYS_MS

export function isMediaRetryCloseCode(code) {
  return code === 4413 || code === 4500
}

export function mediaRetryDelayForAttempt(attempt) {
  return reconnectDelayForAttempt(attempt, MEDIA_RETRY_DELAYS_MS)
}
