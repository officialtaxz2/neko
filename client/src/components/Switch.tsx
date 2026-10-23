import type { ComponentProps } from 'react'
import { cn } from '@/lib/utils'

// the fork: an on/off switch, a checkbox underneath (Space toggles it, the label names it). Off is
// an outlined track with a light knob, on is mint with a dark knob, so the state reads from the
// colour and not only from the knob's side.
export function Switch({ className, ...props }: { label: string } & Omit<ComponentProps<'input'>, 'type'>) {
  const { label, ...input } = props
  return (
    <label className={cn('relative inline-block h-6 w-11 flex-none cursor-pointer', className)}>
      <input type="checkbox" className="peer absolute h-0 w-0 opacity-0" aria-label={label} {...input} />
      <span className="absolute inset-0 rounded-full border border-glass-line bg-white/[0.06] transition-all duration-300 ease-fluid peer-checked:border-transparent peer-checked:bg-mint peer-checked:shadow-[0_0_10px_rgba(38,230,180,0.3)] peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-mint peer-disabled:cursor-default peer-disabled:opacity-50" />
      <span className="pointer-events-none absolute top-1 left-1 size-4 rounded-full bg-pure shadow-[0_2px_4px_rgba(0,0,0,0.3)] transition-all duration-300 ease-fluid peer-checked:translate-x-5 peer-checked:bg-[#050508]" />
    </label>
  )
}
