import type { CSSProperties } from 'react'
import { Toaster as Sonner } from 'sonner'

// The toasts of the fork's design (upstream's components/ui/sonner.tsx, restyled): glass cards at the
// top left under the header, a coloured left edge and title per kind, full width on a small phone.
export function Toaster() {
  return (
    <Sonner
      position="top-left"
      offset={{ top: 60, left: 20 }}
      mobileOffset={{ top: 15, left: 10, right: 10 }}
      gap={10}
      expand
      visibleToasts={9}
      style={{ '--width': '340px' } as CSSProperties}
      icons={{ info: null, success: null, warning: null, error: null }}
      toastOptions={{
        unstyled: true,
        classNames: {
          toast:
            'w-full rounded-inner border border-l-4 px-4 py-3.5 text-[12px] leading-normal text-[#cbd5e1] shadow-[0_10px_30px_-5px_rgba(0,0,0,0.3),inset_0_1px_0_0_rgba(255,255,255,0.05)] backdrop-blur-md data-[type=default]:border-white/[0.08] data-[type=default]:border-l-[#6366f1] data-[type=default]:bg-[rgba(18,20,29,0.85)]',
          title: 'text-[13px] leading-[1.4] font-semibold tracking-[-0.0125em] text-white',
          description: 'opacity-80',
          info: 'border-[rgba(59,130,246,0.15)] border-l-[#3b82f6] bg-[rgba(15,23,42,0.88)] [&_[data-title]]:text-[#93c5fd]',
          success:
            'border-[rgba(34,197,94,0.15)] border-l-[#22c55e] bg-[rgba(16,24,21,0.88)] [&_[data-title]]:text-[#86efac]',
          warning:
            'border-[rgba(234,179,8,0.15)] border-l-[#eab308] bg-[rgba(28,23,15,0.88)] [&_[data-title]]:text-[#fde047]',
          error:
            'border-[rgba(239,68,68,0.15)] border-l-[#ef4444] bg-[rgba(28,15,15,0.88)] [&_[data-title]]:text-[#fca5a5]',
        },
      }}
    />
  )
}
