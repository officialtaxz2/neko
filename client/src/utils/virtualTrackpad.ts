// Why: Encapsulates all relative-cursor state so video.vue stays thin.
// Trackpad mode interprets touch as relative delta (like a laptop touchpad).
// Absolute touch mode in video.vue is NOT touched.

export const LONG_PRESS_DELAY_MS = 600
export const CLICK_THRESHOLD_PX = 10

interface CursorPosition {
  x: number
  y: number
}

let _srcW = 0
let _srcH = 0
let _virtualX = 0
let _virtualY = 0
let _initialised = false

let _lastTouchX = 0
let _lastTouchY = 0
let _touchTotalDistance = 0
let _longPressTimer: ReturnType<typeof setTimeout> | null = null
let _longPressFired = false

function _clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value))
}

/**
 * Call once (lazy, on first touchstart) to anchor the virtual cursor
 * at the centre of the source resolution.
 */
export function initVirtualCursor(srcW: number, srcH: number): void {
  _srcW = srcW
  _srcH = srcH
  if (!_initialised) {
    _virtualX = Math.round(srcW / 2)
    _virtualY = Math.round(srcH / 2)
    _initialised = true
  }
}

export function resetVirtualTrackpad(): void {
  _initialised = false
  _srcW = 0
  _srcH = 0
  _virtualX = 0
  _virtualY = 0
  _touchTotalDistance = 0
  _longPressFired = false
  if (_longPressTimer !== null) {
    clearTimeout(_longPressTimer)
    _longPressTimer = null
  }
}

export function getVirtualCursorPosition(): CursorPosition {
  return { x: _virtualX, y: _virtualY }
}

export function handleVirtualTouchStart(
  touch: Touch,
  onMouseMove: (x: number, y: number) => void,
  onMouseDown: (button: number) => void,
  onMouseUp: (button: number) => void,
): void {
  _lastTouchX = touch.clientX
  _lastTouchY = touch.clientY
  _touchTotalDistance = 0
  _longPressFired = false

  // Long-press → right-click
  _longPressTimer = setTimeout(() => {
    if (_touchTotalDistance <= CLICK_THRESHOLD_PX) {
      _longPressFired = true
      onMouseMove(_virtualX, _virtualY)
      onMouseDown(2)
      onMouseUp(2)
    }
  }, LONG_PRESS_DELAY_MS)

  onMouseMove(_virtualX, _virtualY)
}

export function handleVirtualTouchMove(
  touch: Touch,
  sensitivity: number,
  onMouseMove: (x: number, y: number) => void,
): void {
  const dx = (touch.clientX - _lastTouchX) * sensitivity
  const dy = (touch.clientY - _lastTouchY) * sensitivity

  _touchTotalDistance += Math.sqrt(dx * dx + dy * dy)

  // Cancel long-press if finger moved too far
  if (_touchTotalDistance > CLICK_THRESHOLD_PX && _longPressTimer !== null) {
    clearTimeout(_longPressTimer)
    _longPressTimer = null
  }

  _virtualX = _clamp(_virtualX + dx, 0, _srcW)
  _virtualY = _clamp(_virtualY + dy, 0, _srcH)

  _lastTouchX = touch.clientX
  _lastTouchY = touch.clientY

  onMouseMove(_virtualX, _virtualY)
}

export function handleVirtualTouchEnd(
  onMouseDown: (button: number) => void,
  onMouseUp: (button: number) => void,
): void {
  if (_longPressTimer !== null) {
    clearTimeout(_longPressTimer)
    _longPressTimer = null
  }

  // Tap → left-click (only if not a drag and long-press didn't already fire)
  if (_touchTotalDistance <= CLICK_THRESHOLD_PX && !_longPressFired) {
    onMouseDown(0)
    onMouseUp(0)
  }

  _touchTotalDistance = 0
  _longPressFired = false
}
