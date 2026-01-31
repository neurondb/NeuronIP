'use client'

import {
  HomeIcon,
  Cog6ToothIcon,
  Bars3Icon,
  QuestionMarkCircleIcon,
  DocumentTextIcon,
} from '@heroicons/react/24/outline'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/CommandPalette'
import { ROUTE_ARCHETYPE_MAP } from '@/lib/pageArchetypes'
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'
import { useAppStore } from '@/lib/store/useAppStore'
import { getRouteMeta } from '@/lib/pageArchetypes'

const PAGE_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  '/': HomeIcon,
  default: DocumentTextIcon,
}

function getPageIcon(path: string) {
  return PAGE_ICONS[path] ?? PAGE_ICONS.default
}

export default function CommandPaletteK() {
  const [open, setOpen] = useState(false)
  const router = useRouter()
  const pathname = usePathname()
  const { toggleSidebar, recentPaths, pinnedPaths } = useAppStore()

  useKeyboardShortcuts([
    {
      key: 'k',
      metaKey: true,
      action: () => setOpen((o) => !o),
    },
  ])

  useEffect(() => {
    const onOpen = () => setOpen(true)
    window.addEventListener('open-command-palette', onOpen)
    return () => window.removeEventListener('open-command-palette', onOpen)
  }, [])

  const run = (fn: () => void) => {
    fn()
    setOpen(false)
  }

  const pages = Object.values(ROUTE_ARCHETYPE_MAP)
  const recents = recentPaths
    .filter((p) => p !== pathname)
    .filter((p) => !pinnedPaths.includes(p))
    .slice(0, 5)
    .map((path) => getRouteMeta(path) ?? { path, title: path, archetype: 'list-detail' as const })
  const pinned = pinnedPaths
    .slice(0, 5)
    .map((path) => getRouteMeta(path) ?? { path, title: path, archetype: 'list-detail' as const })

  return (
    <CommandDialog open={open} onOpenChange={setOpen}>
      <CommandInput placeholder="Search pages, actions…" />
      <CommandList>
        <CommandEmpty>No results.</CommandEmpty>

        {pinned.length > 0 && (
          <CommandGroup heading="Pinned">
            {pinned.map((meta) => {
              const Icon = getPageIcon(meta.path)
              return (
                <CommandItem
                  key={meta.path}
                  onSelect={() =>
                    run(() => router.push(meta.path))
                  }
                >
                  <Icon className="mr-2 h-4 w-4" />
                  {meta.title}
                </CommandItem>
              )
            })}
          </CommandGroup>
        )}

        {recents.length > 0 && (
          <CommandGroup heading="Recent">
            {recents.map((meta) => {
              const Icon = getPageIcon(meta.path)
              return (
                <CommandItem
                  key={meta.path}
                  onSelect={() =>
                    run(() => router.push(meta.path))
                  }
                >
                  <Icon className="mr-2 h-4 w-4" />
                  {meta.title}
                </CommandItem>
              )
            })}
          </CommandGroup>
        )}

        {(pinned.length > 0 || recents.length > 0) && <CommandSeparator />}

        <CommandGroup heading="Pages">
          {pages.map((meta) => {
            const Icon = getPageIcon(meta.path)
            return (
              <CommandItem
                key={meta.path}
                onSelect={() =>
                  run(() => router.push(meta.path))
                }
              >
                <Icon className="mr-2 h-4 w-4" />
                {meta.title}
              </CommandItem>
            )
          })}
        </CommandGroup>

        <CommandSeparator />

        <CommandGroup heading="Actions">
          <CommandItem
            onSelect={() =>
              run(() => router.push('/'))
            }
          >
            <HomeIcon className="mr-2 h-4 w-4" />
            Go to Dashboard
          </CommandItem>
          <CommandItem
            onSelect={() =>
              run(() => toggleSidebar())
            }
          >
            <Bars3Icon className="mr-2 h-4 w-4" />
            Toggle sidebar
          </CommandItem>
          <CommandItem
            onSelect={() => {
              setOpen(false)
              window.dispatchEvent(new CustomEvent('open-shortcuts-modal'))
            }}
          >
            <QuestionMarkCircleIcon className="mr-2 h-4 w-4" />
            Show keyboard shortcuts
          </CommandItem>
          <CommandItem
            onSelect={() =>
              run(() => router.push('/settings'))
            }
          >
            <Cog6ToothIcon className="mr-2 h-4 w-4" />
            Open Settings
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
