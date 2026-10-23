import { useEffect, useRef } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling, selectIsAdmin, selectSession } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import {
  ArrowsOutIcon,
  ClipboardIcon,
  KeyboardIcon,
  MonitorIcon,
  MouseIcon,
  PictureInPictureIcon,
  PlayCircleIcon,
  SlidersHorizontalIcon,
  SpeakerHighIcon,
} from '@phosphor-icons/react'
import { cn } from '@/lib/utils'
import { useActions } from '@/state/actions'
import { useNeko } from '@/state/provider'
import { stageStore } from '@/state/stage'
import { IconButton } from '@/components/IconButton'
import { DropdownMenu, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Popover, PopoverTrigger } from '@/components/ui/popover'
import { Emote } from '@/features/emotes/Emote'
import { KeyboardHelper } from '@/features/input/KeyboardHelper'
import { TrackpadCursor } from '@/features/input/TrackpadCursor'
import { useTouch } from '@/features/input/useTouch'
import { Resolution } from './Resolution'
import { Clipboard } from './Clipboard'
import { Player } from './Player'
import { useFullscreen } from './useFullscreen'
import { useClipboardSync, canReadClipboard } from './useClipboardSync'
import { usePip } from './usePip'

// the play / unmute button covering the video
const overlayClasses =
  'absolute inset-0 flex cursor-pointer items-center justify-center overflow-hidden rounded-none bg-black/20'

export function Stage({ hideControls, extraControls }: { hideControls: boolean; extraControls: boolean }) {
  const { t } = useTranslation()
  const { client, overlay, app } = useNeko()
  const actions = useActions()
  const { admin, hosting, hosted, implicit, lockedControls, video, canHost } = useStore(
    client.store,
    useShallow((s) => ({
      admin: selectIsAdmin(s),
      hosting: selectControlling(s),
      hosted: s.control.host_id !== null,
      implicit: s.settings.implicit_hosting,
      lockedControls: s.settings.locked_controls,
      video: s.video,
      // the fork: who may not host (a view-only session) gets no control request
      canHost: !!selectSession(s)?.profile.can_host,
    })),
  )
  const emotes = useStore(app, (s) => s.emotes)
  const keyboardHelper = useStore(app, (s) => s.keyboardHelper)
  const player = useRef<HTMLDivElement>(null)
  const { fullscreen, request: requestFullscreen } = useFullscreen(player)
  // the fork: the room bar and the More menu take this element to fullscreen (state/stage.ts)
  const neko = useNeko()
  useEffect(() => {
    const stage = stageStore(neko)
    stage.setState({ player: player.current })
    return () => stage.setState({ player: null })
  }, [neko])
  const syncClipboard = useClipboardSync()
  const pip = usePip()
  const touch = useTouch() // the fork: "force touch" counts too

  const controlLocked = lockedControls && !admin
  const { playing, playable, muted, mutedByAutoplay } = video
  // usually the extra controls are only shown on a phone
  const extra = extraControls ? '' : 'phone:block hidden'

  return (
    <div className="relative h-full w-full">
      <div ref={player} className="absolute inset-0 flex items-center justify-center bg-black">
        <div className="absolute inset-0" data-testid="player" onMouseEnter={syncClipboard}>
          <Player transport={client.transport} />
          {/* the fork: the trackpad's pointer, the special keys' bar (features/input) */}
          <TrackpadCursor />
          <KeyboardHelper />
          <div className="pointer-events-none absolute inset-0 overflow-hidden">
            {Object.entries(emotes).map(([id, type]) => (
              <Emote key={id} id={id} type={type} />
            ))}
          </div>
          {!playing && playable ? (
            <IconButton
              label="Play"
              className={overlayClasses}
              data-testid="player-overlay"
              onClick={() => (client.unmute(), client.play().catch(() => {}))}
            >
              <PlayCircleIcon className="size-[120px]" />
            </IconButton>
          ) : (
            // only for the mute that autoplay forced; a user who muted on purpose keeps the desktop
            playing &&
            muted &&
            mutedByAutoplay && (
              <IconButton
                label="Unmute"
                className={overlayClasses}
                data-testid="player-overlay"
                onClick={() => client.unmute()}
              >
                <SpeakerHighIcon className="size-[120px]" />
              </IconButton>
            )
          )}
        </div>
        {/* the fork: the buttons on the picture only without the room bar (embed); otherwise
            they are in the bar and its More menu */}
        {!fullscreen && !hideControls && extraControls && (
          <ul className="absolute top-[15px] right-5">
            <li className="mb-2.5 last:mb-0">
              <IconButton variant="video" label="Fullscreen" onClick={requestFullscreen}>
                <ArrowsOutIcon className="size-4" />
              </IconButton>
            </li>
            {admin && (
              <li className="mb-2.5 last:mb-0">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <IconButton variant="video" label="Screen resolution" data-testid="resolution-open">
                      <MonitorIcon className="size-4" />
                    </IconButton>
                  </DropdownMenuTrigger>
                  <Resolution />
                </DropdownMenu>
              </li>
            )}
            {canHost && !controlLocked && !implicit && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <IconButton
                  variant="video"
                  label={hosting ? t('controls.release') : t('controls.request')}
                  className={cn(
                    hosted && !hosting && 'text-style-error/40',
                    !hosted && !hosting && 'text-text-normal/40',
                  )}
                  onClick={() => playable && actions.toggleControl()}
                >
                  <MouseIcon className="size-4" />
                </IconButton>
              </li>
            )}
          </ul>
        )}
        {/* the fork: the buttons on the picture only without the room bar (embed); otherwise
            they are in the bar and its More menu */}
        {!fullscreen && !hideControls && extraControls && (
          <ul className="absolute right-5 bottom-[15px]">
            {hosting && !canReadClipboard && (
              <li className="mb-2.5 last:mb-0">
                {/* modal: the overlay takes the keyboard focus when the mouse enters the video,
                    which would close the popover and send its Ctrl+C to the remote */}
                <Popover modal>
                  <PopoverTrigger asChild>
                    <IconButton variant="video" label="Clipboard" data-testid="clipboard-open">
                      <ClipboardIcon className="size-4" />
                    </IconButton>
                  </PopoverTrigger>
                  <Clipboard />
                </Popover>
              </li>
            )}
            {pip.canPip && (
              <li className="mb-2.5 last:mb-0">
                <IconButton variant="video" label="Picture-in-Picture" onClick={pip.request}>
                  <PictureInPictureIcon className="size-4" />
                </IconButton>
              </li>
            )}
            {hosting && touch && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <IconButton
                  variant="video"
                  label={t('fork:input.keyboard_helper')}
                  data-testid="keyboard-helper-toggle"
                  aria-pressed={keyboardHelper}
                  onMouseDown={(e) => e.preventDefault()} // keeps an open on-screen keyboard open
                  onClick={() => app.setState({ keyboardHelper: !keyboardHelper })}
                >
                  <SlidersHorizontalIcon className="size-4" />
                </IconButton>
              </li>
            )}
            {hosting && touch && (
              <li className={cn('mb-2.5 last:mb-0', extra)}>
                <IconButton
                  variant="video"
                  label="Keyboard"
                  data-testid="keyboard-toggle"
                  onMouseDown={(e) => e.preventDefault()} // tapping the button must not take the focus the keyboard needs
                  onClick={() => overlay.mobileKeyboardToggle()}
                >
                  <KeyboardIcon className="size-4" />
                </IconButton>
              </li>
            )}
          </ul>
        )}
      </div>
    </div>
  )
}
