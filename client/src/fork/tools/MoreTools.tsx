import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { DropdownMenu, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Resolution } from '@/features/video/Resolution'
import { useTools, type Tool } from './useTools'

const tile =
  'flex min-h-[72px] cursor-pointer flex-col items-center justify-center gap-[7px] rounded-inner border border-white/[0.07] bg-white/[0.04] px-1.5 py-2.5 text-text-normal transition-all duration-300 ease-fluid hover:bg-white/[0.08] hover:text-pure [&_svg]:size-[18px]'
const group =
  'mx-0.5 mt-1 mb-2 text-[10.5px] font-bold tracking-[0.08em] text-text-muted uppercase [&:not(:first-child)]:mt-3.5'

// The More menu's tiles in two groups, three a row. `done` runs after a tool (closes the menu).
export function MoreTools({ sheet = false, done }: { sheet?: boolean; done: () => void }) {
  const { t } = useTranslation()
  const { input, view } = useTools(sheet)
  const button = (tool: Tool) => (
    <button
      key={tool.id}
      type="button"
      className={cn(tile, tool.on && 'border-mint/40 text-mint hover:text-mint')}
      aria-pressed={tool.on}
      data-testid={`tool-${tool.id}`}
      // the on-screen keyboard needs the focus the overlay takes, not this button
      onMouseDown={(e) => e.preventDefault()}
      onClick={() => {
        tool.run()
        if (!tool.menu) done()
      }}
    >
      {tool.icon}
      <span className="text-center text-[12px] leading-tight break-words hyphens-auto">{tool.label}</span>
    </button>
  )
  const render = (tool: Tool) =>
    tool.menu === 'resolution' ? (
      <DropdownMenu key={tool.id}>
        <DropdownMenuTrigger asChild>{button(tool)}</DropdownMenuTrigger>
        <Resolution side={sheet ? 'top' : 'left'} />
      </DropdownMenu>
    ) : (
      button(tool)
    )
  return (
    <div data-testid="more-tools">
      {input.length > 0 && (
        <>
          <h4 className={group}>{t('fork:controls.input_group')}</h4>
          <div className="grid grid-cols-3 gap-2">{input.map(render)}</div>
        </>
      )}
      {view.length > 0 && (
        <>
          <h4 className={group}>{t('fork:controls.view_group')}</h4>
          <div className="grid grid-cols-3 gap-2">{view.map(render)}</div>
        </>
      )}
    </div>
  )
}
