import { useEffect } from 'react'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { SlidersHorizontalIcon } from '@phosphor-icons/react'
import { useApp } from '@/state/provider'
import { remember } from '@/state/storage'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Settings } from '@/fork/settings/Settings'
import { useSlot } from '@/plugins'
import type { PluginTab } from '@/plugins/types'
import { EachHook } from '@/components/EachHook'

// after the plugins' tabs
const settingsTab: PluginTab = {
  id: 'settings',
  icon: SlidersHorizontalIcon,
  label: 'side.settings',
  component: Settings,
}
const always = () => true
const tabClass =
  'mr-0 flex min-w-0 flex-1 basis-0 flex-col items-center gap-1 rounded-control bg-transparent px-0.5 pt-2 pb-1.5 text-[11.5px] font-semibold text-text-muted transition-all duration-300 ease-fluid hover:bg-glass hover:text-pure data-[state=active]:bg-mint/12 data-[state=active]:text-mint'

// The side panel of the fork's design: upstream's panel (layout/Side.tsx) on glass, with the tabs as
// icon-over-label buttons that fill the width.
export function Side() {
  const tabs = [...useSlot('side.tab'), settingsTab]
  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <EachHook items={tabs} use={(tab) => tab.useVisible ?? always}>
      {(visible) => <Panel tabs={tabs.filter((_, i) => visible[i])} />}
    </EachHook>
  )
}

// the tabs available now
function Panel({ tabs }: { tabs: PluginTab[] }) {
  const { t } = useTranslation()
  const app = useApp()
  const wanted = useStore(app, (s) => s.tab)
  // a remembered tab that is not available now falls back to the first one
  const tab = tabs.some((x) => x.id === wanted) ? wanted : tabs[0].id
  const Page = tabs.find((x) => x.id === tab)!.component

  return (
    <Tabs value={tab} onValueChange={(id) => (app.setState({ tab: id }), remember('tab', id))} asChild>
      <aside
        className="tablet:w-full tablet:portrait:h-[60vh] tablet:landscape:h-screen flex max-h-full w-side max-w-full shrink-0 flex-col border-l border-white/[0.08] bg-bar-side shadow-side backdrop-blur-lg"
        data-testid="side"
      >
        <TabsList className="h-auto! gap-1 border-b border-white/[0.06] bg-transparent p-2.5" data-testid="tabs">
          {tabs.map(({ id, icon: Icon, label }) => (
            <TabsTrigger key={id} value={id} className={tabClass}>
              <Icon className="size-[15px]" aria-hidden />
              <span className="max-w-full shrink-0 truncate leading-tight">{t(label)}</span>
            </TabsTrigger>
          ))}
        </TabsList>
        <TabsContent value={tab} className="flex max-h-full grow overflow-auto pt-[5px]">
          <Page />
        </TabsContent>
      </aside>
    </Tabs>
  )
}
