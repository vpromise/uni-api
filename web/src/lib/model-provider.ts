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

type ModelProvider = {
  readonly name: string
  readonly icon: string
  readonly label?: string
}

type ModelProviderRule = ModelProvider & {
  keywords?: readonly string[]
  pattern?: RegExp
  fallbackKeywords?: readonly string[]
  fallbackPattern?: RegExp
}

const MODEL_PROVIDER_RULES: readonly ModelProviderRule[] = [
  {
    name: 'OpenAI',
    icon: 'OpenAI.Color',
    fallbackKeywords: [
      'text-embedding-',
      'omni-moderation',
      'dall-e',
      'whisper',
      'tts-',
    ],
    fallbackPattern: /\bo[134](?:-|$)/,
    keywords: [
      'openai/',
      'openai.',
      'gpt-',
      'chatgpt-',
      'codex-',
      'dall-e-',
      'whisper-',
      'omni-moderation-',
      'text-moderation-',
      'text-embedding-ada-',
      'text-embedding-3-',
      'text-ada-',
      'text-babbage-',
      'text-curie-',
      'davinci-',
      'babbage-',
      'computer-use-preview',
      'sora',
    ],
    pattern: /(?:^|[/.:])(?:o(?:1|3|4)(?=$|[-.:])|tts-)/,
  },
  {
    name: 'Anthropic',
    icon: 'Claude.Color',
    label: 'Claude',
    keywords: ['anthropic', 'claude'],
  },
]

export function resolveModelProvider(modelName: string): ModelProvider | null {
  const model = modelName.trim().toLowerCase()

  return (
    MODEL_PROVIDER_RULES.find(
      (rule) =>
        rule.keywords?.some((keyword) => model.includes(keyword)) ||
        rule.pattern?.test(model)
    ) ??
    MODEL_PROVIDER_RULES.find(
      (rule) =>
        rule.fallbackKeywords?.some((keyword) => model.includes(keyword)) ||
        rule.fallbackPattern?.test(model)
    ) ??
    null
  )
}
