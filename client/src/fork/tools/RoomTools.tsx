import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowsOutIcon, DotsThreeIcon } from '@phosphor-icons/react'
import { useNeko } from '@/state/provider'
import { requestFullscreen } from '@/state/stage'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { barButton } from '@/fork/controls/styles'
import { MoreTools } from './MoreTools'
import { useTools } from './useTools'

// The desktop's room tools at the right of the bar: fullscreen right there, everything else in the
// More popover above it (the phone has the same tools in its sheet).
export function RoomTools() {
  const { t } = useTranslation()
  const neko = useNeko()
  const [open, setOpen] = useState(false)
  const ran = useRef(false) // a tool ran: the focus stays where it put it (the on-screen keyboard)
  const { input, view } = useTools(false)
  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        className={barButton}
        aria-label={t('fork:controls.fullscreen')}
        title={t('fork:controls.fullscreen')}
        data-testid="fullscreen"
        onClick={() => requestFullscreen(neko)}
      >
        <ArrowsOutIcon className="size-4" />
      </button>
      {input.length + view.length > 0 && (
        <Popover open={open} onOpenChange={(o) => ((ran.current = false), setOpen(o))}>
          <PopoverTrigger asChild>
            <button
              type="button"
              className={barButton}
              aria-label={t('fork:controls.tools')}
              title={open ? undefined : t('fork:controls.tools')}
              data-testid="more-open"
            >
              <DotsThreeIcon className="size-5" weight="bold" />
            </button>
          </PopoverTrigger>
          <PopoverContent
            side="top"
            align="end"
            sideOffset={10}
            className="w-[340px] max-w-[calc(100vw-24px)] rounded-outer border border-glass-line bg-sheet p-3 shadow-sheet"
            aria-label={t('fork:controls.tools')}
            data-testid="more-popover"
            onCloseAutoFocus={(e) => ran.current && e.preventDefault()}
          >
            <MoreTools
              done={() => {
                ran.current = true
                setOpen(false)
              }}
            />
          </PopoverContent>
        </Popover>
      )}
    </div>
  )
}
