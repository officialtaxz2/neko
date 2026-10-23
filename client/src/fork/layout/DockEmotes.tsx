import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { SmileyIcon } from '@phosphor-icons/react'
import { useActions } from '@/state/actions'
import { selectMuted } from '@/state/app'
import { EMOTES } from '@/state/emotes'
import { useClient } from '@/state/provider'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { dockButton } from '@/fork/controls/styles'
import '@/features/emotes/sprites.css'

// The phone's emotes: one dock button, every emote in a grid above it, sized for a finger.
export function DockEmotes() {
  const { t } = useTranslation()
  const client = useClient()
  const actions = useActions()
  const muted = useStore(client.store, selectMuted)
  if (muted) return null
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button type="button" className={dockButton} aria-label={t('fork:controls.emotes')} data-testid="dock-emotes">
          <SmileyIcon className="size-5" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        side="top"
        align="end"
        sideOffset={8}
        className="grid w-[min(236px,calc(100vw-24px))] grid-cols-6 gap-0.5 rounded-inner border border-glass-line bg-sheet p-2 shadow-sheet"
        data-testid="dock-emotes-menu"
      >
        {EMOTES.map((e) => (
          <DropdownMenuItem
            key={e}
            className="flex size-9 items-center justify-center p-0 text-[24px]"
            aria-label={e}
            onSelect={(ev) => (ev.preventDefault(), actions.sendEmote(e))}
          >
            <div className={`emote ${e}`} data-emote={e} />
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
