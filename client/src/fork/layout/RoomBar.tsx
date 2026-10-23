import { Emotes } from '@/features/emotes/Emotes'
import { MemberStrip } from '@/fork/room/MemberStrip'
import { ControlPill } from '@/fork/controls/ControlPill'
import { Playback } from '@/fork/controls/Playback'
import { RoomTools } from '@/fork/tools/RoomTools'

// The room bar of the fork's design, one row under the picture: the members on the left; the
// control pill and playback (play, sound, microphone) in the middle; the emotes, fullscreen and More
// on the right (the outer columns share the rest of the width, so the middle stays centred). Not
// shown on a phone, where the dock takes its place.
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
        <div className="flex items-center justify-center gap-3">
          <ControlPill separator />
          <Playback />
        </div>
        <div className="flex min-w-0 items-center justify-end gap-2">
          <Emotes />
          <RoomTools />
        </div>
      </div>
    </div>
  )
}
