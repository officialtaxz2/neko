export const RECONNECT_DELAYS_MS = Object.freeze([1000, 2000, 5000, 10000])

/**
 * @typedef {Object} ReconnectContext
 * @property {boolean} previouslyConnected
 * @property {boolean} credentialsAvailable
 * @property {boolean} supported
 * @property {boolean} demo
 * @property {boolean} suppressed
 */

/**
 * Application-level reconnect is only safe after a real session was connected.
 * Initial login failures and explicit/server-directed disconnects stay manual.
 *
 * @param {ReconnectContext} context
 */
export function shouldReconnect(context) {
  return (
    context.previouslyConnected &&
    context.credentialsAvailable &&
    context.supported &&
    !context.demo &&
    !context.suppressed
  )
}

/**
 * Return the delay for a zero-based reconnect attempt, or null once the
 * bounded schedule is exhausted (or the input is invalid).
 *
 * @param {number} attempt
 * @param {readonly number[]} [delays]
 * @returns {number | null}
 */
export function reconnectDelayForAttempt(attempt, delays = RECONNECT_DELAYS_MS) {
  if (!Number.isInteger(attempt) || attempt < 0 || attempt >= delays.length) {
    return null
  }

  return delays[attempt]
}
