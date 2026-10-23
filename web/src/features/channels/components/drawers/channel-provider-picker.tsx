/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { Puzzle } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { resolveLocalizedText } from '@/lib/localized-text'

import type { TaskPluginOption } from '../../api'
import {
  CHANNEL_PROVIDER_PRESENTATION,
  CHANNEL_TYPE_OPTIONS,
} from '../../constants'
import type { ChannelProviderTarget } from '../../lib/channel-configuration'
import { getChannelPluginExtensions } from '../../lib/channel-plugin-extensions'
import { ChannelTypeLogo } from '../channel-type-badge'

type ChannelProviderPickerProps = {
  isCreating?: boolean
  plugins: TaskPluginOption[]
  currentProvider?: ChannelProviderTarget | null
  canBindPlugin: boolean
  loading: boolean
  failed: boolean
  disabled: boolean
  onRetry: () => void
  onSelect: (target: ChannelProviderTarget) => void
}

export function ChannelProviderPicker(props: ChannelProviderPickerProps) {
  const { t, i18n } = useTranslation()
  const [search, setSearch] = useState('')
  const availablePlugins =
    props.canBindPlugin && !props.loading && !props.failed ? props.plugins : []
  const options = CHANNEL_TYPE_OPTIONS.map((option) => {
    const extensions = getChannelPluginExtensions(
      option.value,
      availablePlugins
    )
    return {
      ...option,
      extensions,
      searchText: `${option.value} ${option.label} ${t(option.label)} ${extensions.map((plugin) => `${plugin.name} ${plugin.key}`).join(' ')}`,
    }
  }).filter((option) =>
    option.searchText
      .toLocaleLowerCase()
      .includes(search.trim().toLocaleLowerCase())
  )

  return (
    <Command shouldFilter={false} className='min-h-0 flex-1 p-4 sm:p-6'>
      <CommandInput
        autoFocus
        value={search}
        onValueChange={setSearch}
        placeholder={t('Search providers, plugins, or type numbers')}
        aria-label={t('Search providers, plugins, or type numbers')}
      />
      <CommandList className='mt-3 max-h-none'>
        <CommandEmpty>{t('No matching provider')}</CommandEmpty>
        <CommandGroup className='p-1 [&_[cmdk-group-items]]:grid [&_[cmdk-group-items]]:gap-2.5 md:[&_[cmdk-group-items]]:grid-cols-3'>
          {options.map((option) => {
            const current =
              props.currentProvider?.kind === 'builtin' &&
              props.currentProvider.type === option.value
            const presentation = CHANNEL_PROVIDER_PRESENTATION[option.value]
            const extensionSummary = option.extensions
              .map((plugin) =>
                [
                  plugin.name,
                  resolveLocalizedText(plugin.description, i18n.language),
                ]
                  .filter(Boolean)
                  .join(' · ')
              )
              .join(' / ')
            return (
              <CommandItem
                key={option.value}
                value={`builtin:${option.value}`}
                aria-label={`${t(option.label)} ${t('Built-in')} #${option.value}`}
                aria-current={current ? true : undefined}
                data-checked={current}
                disabled={props.disabled}
                onSelect={() =>
                  props.onSelect({ kind: 'builtin', type: option.value })
                }
                className='min-h-24 flex-col items-stretch gap-2.5 rounded-lg border p-3 [&>svg]:hidden'
              >
                <span className='flex items-center gap-2.5'>
                  <ChannelTypeLogo type={option.value} size={20} />
                  <span className='font-medium'>{t(option.label)}</span>
                </span>
                {presentation && (
                  <span className='text-muted-foreground text-xs'>
                    {t(presentation.descriptionKey)}
                  </span>
                )}
                <span className='flex items-center gap-2'>
                  <Badge variant={current ? 'default' : 'secondary'}>
                    {current ? t('Current') : t('Built-in')}
                  </Badge>
                  {extensionSummary && (
                    <span
                      aria-label={t('Plugin extensions')}
                      title={extensionSummary}
                      className='text-muted-foreground flex min-w-0 items-center gap-1 text-xs'
                    >
                      <Puzzle className='size-3' aria-hidden='true' />
                      <span className='truncate'>
                        {option.extensions
                          .map((plugin) => plugin.name)
                          .join(' · ')}
                      </span>
                    </span>
                  )}
                  <span className='text-muted-foreground ml-auto text-xs'>
                    #{option.value}
                  </span>
                </span>
              </CommandItem>
            )
          })}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
