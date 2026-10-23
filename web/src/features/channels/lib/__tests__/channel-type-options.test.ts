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
import { describe, expect, test } from 'vitest'

import {
  CHANNEL_TYPE_OPTIONS,
  channelTypeOptionsForTaskPluginBind,
} from '../../constants'
describe('supported upstream channel options', () => {
  test.each([true, false])(
    'offers only the three supported integrations with bind permission %s',
    (canBind) => {
      expect(
        channelTypeOptionsForTaskPluginBind(canBind).map(
          (option) => option.value
        )
      ).toEqual([1, 14, 57])
      expect(CHANNEL_TYPE_OPTIONS.map((option) => option.label)).toEqual([
        'OpenAI',
        'Anthropic',
        'ChatGPT Subscription (Codex)',
      ])
    }
  )
})
