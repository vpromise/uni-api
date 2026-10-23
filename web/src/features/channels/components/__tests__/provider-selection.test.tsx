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
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, test, vi } from 'vitest'

import { ChannelProviderPicker } from '../drawers/channel-provider-picker'

const defaults = {
  plugins: [],
  canBindPlugin: true,
  loading: false,
  failed: false,
  disabled: false,
  onRetry: vi.fn(),
}

test('offers only the three retained upstreams', async () => {
  const onSelect = vi.fn()
  render(<ChannelProviderPicker {...defaults} onSelect={onSelect} />)
  const options = screen.getAllByRole('option')
  expect(options).toHaveLength(3)
  await userEvent.click(screen.getByRole('option', { name: /#14$/ }))
  expect(onSelect).toHaveBeenCalledWith({ kind: 'builtin', type: 14 })
})

test('unknown channel numbers cannot create another provider', async () => {
  render(<ChannelProviderPicker {...defaults} onSelect={vi.fn()} />)
  await userEvent.type(screen.getByRole('combobox'), '999')
  expect(screen.queryAllByRole('option')).toHaveLength(0)
})

test('Sora extension search selects the OpenAI connection', async () => {
  const onSelect = vi.fn()
  render(
    <ChannelProviderPicker
      {...defaults}
      plugins={[
        {
          key: 'sora',
          name: 'Sora Video',
          channelTypes: [1],
          models: ['sora-2'],
        },
      ]}
      onSelect={onSelect}
    />
  )
  await userEvent.type(screen.getByRole('combobox'), 'sora')
  expect(screen.getAllByRole('option')).toHaveLength(1)
  await userEvent.click(screen.getByRole('option'))
  expect(onSelect).toHaveBeenCalledWith({ kind: 'builtin', type: 1 })
})
