import { Controls } from '@/features/controls/Controls'
import { Emotes } from '@/features/emotes/Emotes'
import { MemberStrip } from '@/fork/room/MemberStrip'

// The room bar of the fork's design, one row under the picture: the members on the left, the
// controls in the middle, the emotes on the right (the outer columns share the rest of the width, so
// the controls stay centred). Not shown on a phone, where the dock takes its place.
export function RoomBar() {
  return (
    <div
      className="phone:hidden z-10 flex h-controls max-w-full shrink-0 border-t border-glass-border bg-bar-room shadow-bar-up"
      data-testid="room-bar"
    >
      <div className="grid min-w-0 flex-1 grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-4 px-4">
        <div className="flex min-w-0 items-center">
          <MemberStrip />
        </div>
        <div className="flex items-center justify-center">
          <Controls />
        </div>
        <div className="flex min-w-0 items-center justify-end gap-2.5">
          <Emotes />
        </div>
      </div>
    </div>
  )
}
