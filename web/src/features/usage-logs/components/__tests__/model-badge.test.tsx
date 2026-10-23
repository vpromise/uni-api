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
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import {
  categorizeModels,
  getModelCategory,
} from '@/features/channels/lib/model-categories'
import { getLobeIconNames } from '@/lib/lobe-icon'

import { ModelBadge } from '../model-badge'
const providers = [
  {
    category: 'OpenAI',
    models: [
      'codex-auto-review',
      'codex-mini-latest',
      ' CODEX-AUTO-REVIEW ',
      'openai/codex-auto-review',
      'openai.custom',
      'gpt-5-codex',
      'chatgpt-4o-latest',
      'dall-e',
      'whisper',
      'omni-moderation',
      'text-moderation-latest',
      'text-embedding-custom',
      'text-embedding-ada-002',
      'text-embedding-3-large',
      'text-ada-001',
      'text-babbage-001',
      'text-curie-001',
      'davinci-002',
      'babbage-002',
      'computer-use-preview',
      'sora-2',
      'tts-1',
      'vendor-tts-1',
      'o1',
      'o3-mini',
      'o4-mini',
      'openai/o3:latest',
      'o3.preview',
      'vendor-o3',
    ],
  },
  {
    category: 'Anthropic',
    label: 'Claude',
    models: ['claude-sonnet-4', 'anthropic/claude', 'claude'],
  },
]
it.each(
  providers.flatMap((provider) =>
    provider.models.map((model) => ({
      model,
      category: provider.category,
      label: provider.label ?? provider.category,
    }))
  )
)(
  'groups $model under $category and displays its provider icon',
  async ({ model, category, label }) => {
    expect(getModelCategory(model)).toBe(category)
    render(<ModelBadge modelName={model} />)
    expect(screen.getByText(model.trim())).toBeVisible()
    const icon = screen.getByLabelText(label)
    expect(icon).toBeVisible()
    await waitFor(() => expect(icon.querySelector('svg, img')).not.toBeNull())
  }
)
it.each([
  '',
  'custom-model',
  'o2',
  'o30',
  'prefix_o3',
  'vendor-o3:latest',
  'wan',
  'aqa-extra',
  't2v-02',
])('keeps unknown model %s in Other with the dot fallback', (model) => {
  expect(getModelCategory(model)).toBe('Other')
  const { container } = render(<ModelBadge modelName={model} />)
  expect(container.querySelector('svg, img')).toBeNull()
  expect(container.querySelector('[aria-hidden="true"]')).not.toBeNull()
})
it('preserves model names and input order when grouping models', () => {
  expect(
    categorizeModels([
      ' CODEX-AUTO-REVIEW ',
      'custom-model',
      'gpt-5',
      'spark',
      'codex-mini-latest',
    ])
  ).toEqual({
    OpenAI: [' CODEX-AUTO-REVIEW ', 'gpt-5', 'codex-mini-latest'],
    Other: ['custom-model', 'spark'],
  })
})
it('shows the OpenAI icon in the mobile model button', () => {
  render(
    <ModelBadge modelName='codex-auto-review' wrapText onInspect={vi.fn()} />
  )
  expect(
    screen.getByRole('button', {
      name: 'Model: codex-auto-review',
    })
  ).toContainElement(screen.getByLabelText('OpenAI'))
})
it('makes the Wan icon available to model icon selectors', () => {
  expect(getLobeIconNames()).toContain('Wan')
})
it('opens the mismatch evidence with the keyboard and shows all three models', async () => {
  const user = userEvent.setup()
  const returned =
    'unexpected-provider-model-with-a-long-dated-version-2026-09-17'
  render(
    <ModelBadge
      modelName='requested-model'
      responseModel={{
        requested_model: 'requested-model',
        upstream_model: 'mapped-model',
        returned_model: returned,
      }}
    />
  )
  expect(screen.getByText(`Response model: ${returned}`)).toBeVisible()
  await user.tab()
  expect(
    screen.getByRole('button', {
      name: `Model: requested-model, Response model: ${returned}`,
    })
  ).toHaveFocus()
  await user.keyboard('{Enter}')
  expect(await screen.findByText(returned)).toBeVisible()
  expect(screen.getByText('mapped-model')).toBeVisible()
  expect(screen.getByText('Request Model')).toBeVisible()
  expect(screen.getByText('Upstream Model')).toBeVisible()
  expect(screen.getByText('Response Model')).toBeVisible()
  expect(
    screen.getByText(/this warning alone does not prove model substitution/)
  ).toBeVisible()
})
it.each([false, true])(
  'copies the model without opening details when there is no mapping or difference (response observed: %s)',
  async (observed) => {
    const user = userEvent.setup()
    const copy = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue()
    render(
      <ModelBadge
        modelName='requested-model'
        responseModel={
          observed
            ? {
                requested_model: 'requested-model',
                upstream_model: 'requested-model',
                returned_model: 'requested-model',
              }
            : undefined
        }
      />
    )
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
    await user.click(screen.getByText('requested-model'))
    expect(copy).toHaveBeenCalledWith('requested-model')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(screen.queryByText('Response Model')).not.toBeInTheDocument()
  }
)
it('keeps mapped model details available when the response matches the upstream model', async () => {
  const user = userEvent.setup()
  render(
    <ModelBadge
      modelName='requested-model'
      responseModel={{
        requested_model: 'requested-model',
        upstream_model: 'mapped-model',
        returned_model: 'mapped-model',
      }}
    />
  )
  await user.click(
    screen.getByRole('button', {
      name: 'Model: requested-model',
    })
  )
  expect(await screen.findByText('Response Model')).toBeVisible()
  expect(screen.queryByText(/^Response model:/)).not.toBeInTheDocument()
})
it.each([
  'requested-model-2026-09-17',
  'REQUESTED-MODEL',
  'mapped-model-2026-09-17',
  'MAPPED-MODEL',
  'deepseek/requested-model',
  'accounts/vendor/models/MAPPED-MODEL',
])(
  'keeps the compatible response %s in the popover without a list annotation',
  async (returned) => {
    const user = userEvent.setup()
    render(
      <ModelBadge
        modelName='requested-model'
        responseModel={{
          requested_model: 'requested-model',
          upstream_model: 'mapped-model',
          returned_model: returned,
        }}
      />
    )
    expect(screen.queryByText(returned)).not.toBeInTheDocument()
    expect(screen.queryByText(/^Response model:/)).not.toBeInTheDocument()
    const trigger = screen.getByRole('button', {
      name: 'Model: requested-model',
    })
    expect(trigger).toHaveTextContent(/^requested-model$/)
    await user.click(trigger)
    expect(await screen.findByText(returned)).toBeVisible()
    expect(screen.queryByText(/^Response model:/)).not.toBeInTheDocument()
    expect(
      screen.queryByText(/this warning alone does not prove model substitution/)
    ).not.toBeInTheDocument()
  }
)
