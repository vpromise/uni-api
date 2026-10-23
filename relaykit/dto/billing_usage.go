package dto

import "strings"

const (
	BillingUsageSourceClaudeMessages = "claude_messages"

	BillingUsageSourceOAIChat      = "oai_chat"
	BillingUsageSourceOAIResponses = "oai_responses"

	BillingUsageSemanticAnthropic = "anthropic"

	BillingUsageSemanticOpenAI = "openai"
)

type BillingUsage struct {
	Source      string       `json:"source,omitempty"`
	Semantic    string       `json:"semantic,omitempty"`
	Estimated   bool         `json:"estimated,omitempty"`
	OpenAIUsage *Usage       `json:"openai_usage,omitempty"`
	ClaudeUsage *ClaudeUsage `json:"claude_usage,omitempty"`
}

func NewClaudeMessagesBillingUsage(usage *ClaudeUsage) *BillingUsage {
	if !HasClaudeUsageTokens(usage) {
		return nil
	}
	return &BillingUsage{
		Source:      BillingUsageSourceClaudeMessages,
		Semantic:    BillingUsageSemanticAnthropic,
		ClaudeUsage: cloneClaudeUsage(usage),
	}
}

func HasClaudeUsageTokens(usage *ClaudeUsage) bool {
	if usage == nil {
		return false
	}
	if usage.InputTokens != 0 ||
		usage.OutputTokens != 0 ||
		usage.CacheCreationInputTokens != 0 ||
		usage.CacheReadInputTokens != 0 ||
		usage.ClaudeCacheCreation5mTokens != 0 ||
		usage.ClaudeCacheCreation1hTokens != 0 {
		return true
	}
	if usage.CacheCreation != nil &&
		(usage.CacheCreation.Ephemeral5mInputTokens != 0 || usage.CacheCreation.Ephemeral1hInputTokens != 0) {
		return true
	}
	return false
}

func NewOpenAIChatBillingUsage(usage *Usage) *BillingUsage {
	return newOpenAIBillingUsage(BillingUsageSourceOAIChat, usage)
}

func NewOpenAIResponsesBillingUsage(usage *Usage) *BillingUsage {
	return newOpenAIBillingUsage(BillingUsageSourceOAIResponses, usage)
}

func newOpenAIBillingUsage(source string, usage *Usage) *BillingUsage {
	if !HasOpenAIUsageTokens(usage) {
		return nil
	}
	return &BillingUsage{
		Source:      source,
		Semantic:    BillingUsageSemanticOpenAI,
		OpenAIUsage: cloneOpenAIUsage(usage),
	}
}

func HasOpenAIUsageTokens(usage *Usage) bool {
	if usage == nil {
		return false
	}
	if usage.PromptTokens != 0 ||
		usage.CompletionTokens != 0 ||
		usage.TotalTokens != 0 ||
		usage.InputTokens != 0 ||
		usage.OutputTokens != 0 ||
		usage.PromptCacheHitTokens != 0 ||
		usage.ClaudeCacheCreation5mTokens != 0 ||
		usage.ClaudeCacheCreation1hTokens != 0 {
		return true
	}
	if usage.PromptTokensDetails.CachedTokensDetails.HasTokens() ||
		usage.PromptTokensDetails.CacheWriteTokensDetails.HasTokens() ||
		usage.PromptTokensDetails.CachedTokens != 0 ||
		usage.PromptTokensDetails.CachedCreationTokens != 0 ||
		usage.PromptTokensDetails.CacheWriteTokens != 0 ||
		usage.PromptTokensDetails.TextTokens != 0 ||
		usage.PromptTokensDetails.ImageTokens != 0 ||
		usage.PromptTokensDetails.AudioTokens != 0 {
		return true
	}
	if usage.CompletionTokenDetails.ReasoningTokens != 0 ||
		usage.CompletionTokenDetails.TextTokens != 0 ||
		usage.CompletionTokenDetails.ImageTokens != 0 ||
		usage.CompletionTokenDetails.AudioTokens != 0 {
		return true
	}
	if usage.InputTokensDetails == nil {
		return false
	}
	return usage.InputTokensDetails.CachedTokensDetails.HasTokens() ||
		usage.InputTokensDetails.CacheWriteTokensDetails.HasTokens() ||
		usage.InputTokensDetails.CachedTokens != 0 ||
		usage.InputTokensDetails.CachedCreationTokens != 0 ||
		usage.InputTokensDetails.CacheWriteTokens != 0 ||
		usage.InputTokensDetails.TextTokens != 0 ||
		usage.InputTokensDetails.ImageTokens != 0 ||
		usage.InputTokensDetails.AudioTokens != 0
}

// CloneBillingUsageWithEstimatedCompletion preserves the original upstream
// billing dialect and fills a missing completion count without rebuilding the
// payload from a converted, potentially lossy Usage value.
func CloneBillingUsageWithEstimatedCompletion(usage *BillingUsage, completionTokens int) *BillingUsage {
	clone := CloneBillingUsage(usage)
	if clone == nil || completionTokens <= 0 {
		return clone
	}

	updated := false
	switch {
	case clone.OpenAIUsage != nil:
		openAIUsage := clone.OpenAIUsage
		if openAIUsage.CompletionTokens == 0 && openAIUsage.OutputTokens == 0 {
			openAIUsage.CompletionTokens = completionTokens
			openAIUsage.OutputTokens = completionTokens
			inputTokens := openAIUsage.PromptTokens
			if inputTokens == 0 {
				inputTokens = openAIUsage.InputTokens
			}
			if totalTokens := inputTokens + completionTokens; openAIUsage.TotalTokens < totalTokens {
				openAIUsage.TotalTokens = totalTokens
			}
			updated = true
		}
	case clone.ClaudeUsage != nil:
		if clone.ClaudeUsage.OutputTokens == 0 {
			clone.ClaudeUsage.OutputTokens = completionTokens
			updated = true
		}

	}
	if updated {
		clone.Estimated = true
	}
	return clone
}

func CloneBillingUsage(usage *BillingUsage) *BillingUsage {
	if usage == nil {
		return nil
	}
	clone := *usage
	clone.OpenAIUsage = cloneOpenAIUsage(usage.OpenAIUsage)
	clone.ClaudeUsage = cloneClaudeUsage(usage.ClaudeUsage)

	return &clone
}

// CanonicalUsage decodes the original provider usage carried across relay
// hops into the shared accounting shape. The BillingUsage snapshot remains the
// source of truth and is cloned onto the returned value for further relays.
func (usage *BillingUsage) CanonicalUsage() (*Usage, bool) {
	if usage == nil {
		return nil, false
	}
	source := strings.TrimSpace(usage.Source)
	semantic := strings.TrimSpace(usage.Semantic)

	// A structurally recognized but all-zero payload must not become the
	// settlement source of truth; rejecting it lets settlement fall back to a
	// non-zero top-level usage.
	if HasOpenAIUsageTokens(usage.OpenAIUsage) &&
		(strings.EqualFold(source, BillingUsageSourceOAIChat) ||
			strings.EqualFold(source, BillingUsageSourceOAIResponses) ||
			strings.EqualFold(semantic, BillingUsageSemanticOpenAI)) {
		return usage.canonicalOpenAIUsage(), true
	}
	if HasClaudeUsageTokens(usage.ClaudeUsage) &&
		(strings.EqualFold(source, BillingUsageSourceClaudeMessages) ||
			strings.EqualFold(semantic, BillingUsageSemanticAnthropic)) {
		return usage.canonicalClaudeUsage(), true
	}

	return nil, false
}

func (usage *BillingUsage) canonicalOpenAIUsage() *Usage {
	canonical := cloneOpenAIUsage(usage.OpenAIUsage)
	if canonical.InputTokensDetails != nil {
		// InputTokensDetails fills fields that PromptTokensDetails omitted;
		// existing PromptTokensDetails values stay canonical on overlap.
		filled := *canonical.InputTokensDetails
		mergeInputTokenDetails(&filled, canonical.PromptTokensDetails)
		canonical.PromptTokensDetails = filled
	}
	if canonical.PromptTokensDetails.CachedTokens == 0 && canonical.PromptCacheHitTokens > 0 {
		canonical.PromptTokensDetails.CachedTokens = canonical.PromptCacheHitTokens
	}
	if canonical.PromptTokens == 0 && canonical.InputTokens > 0 {
		canonical.PromptTokens = canonical.InputTokens
	}
	if canonical.CompletionTokens == 0 && canonical.OutputTokens > 0 {
		canonical.CompletionTokens = canonical.OutputTokens
	}
	if canonical.InputTokens == 0 && canonical.PromptTokens > 0 {
		canonical.InputTokens = canonical.PromptTokens
	}
	if canonical.OutputTokens == 0 && canonical.CompletionTokens > 0 {
		canonical.OutputTokens = canonical.CompletionTokens
	}
	if canonical.TotalTokens == 0 {
		canonical.TotalTokens = canonical.PromptTokens + canonical.CompletionTokens
	}
	canonical.UsageSemantic = BillingUsageSemanticOpenAI
	canonical.UsageSource = usage.Source
	canonical.BillingUsage = CloneBillingUsage(usage)
	return canonical
}

func (usage *BillingUsage) canonicalClaudeUsage() *Usage {
	claudeUsage := usage.ClaudeUsage
	// Flat legacy fields are a fallback only when this snapshot never carried
	// a CacheCreation sub-object. Presence (non-nil), not zero vs non-zero,
	// is the discriminator — a later sub-object that zeros 1h must win.
	var cacheCreation5m, cacheCreation1h int
	if claudeUsage.CacheCreation != nil {
		cacheCreation5m = claudeUsage.GetCacheCreation5mTokens()
		cacheCreation1h = claudeUsage.GetCacheCreation1hTokens()
	} else {
		cacheCreation5m = claudeUsage.ClaudeCacheCreation5mTokens
		cacheCreation1h = claudeUsage.ClaudeCacheCreation1hTokens
	}

	canonical := &Usage{
		PromptTokens:                claudeUsage.InputTokens,
		CompletionTokens:            claudeUsage.OutputTokens,
		TotalTokens:                 claudeUsage.InputTokens + claudeUsage.OutputTokens,
		InputTokens:                 claudeUsage.InputTokens + claudeUsage.CacheReadInputTokens + claudeUsage.CacheCreationInputTokens,
		OutputTokens:                claudeUsage.OutputTokens,
		UsageSemantic:               BillingUsageSemanticAnthropic,
		UsageSource:                 BillingUsageSourceClaudeMessages,
		BillingUsage:                CloneBillingUsage(usage),
		ClaudeCacheCreation5mTokens: cacheCreation5m,
		ClaudeCacheCreation1hTokens: cacheCreation1h,
	}
	canonical.PromptTokensDetails.CachedTokens = claudeUsage.CacheReadInputTokens
	canonical.PromptTokensDetails.CachedCreationTokens = claudeUsage.CacheCreationInputTokens
	return canonical
}

func cloneOpenAIUsage(usage *Usage) *Usage {
	if usage == nil {
		return nil
	}
	clone := *usage
	clone.BillingUsage = nil
	clone.PromptTokensDetails = usage.PromptTokensDetails.Clone()
	if usage.InputTokensDetails != nil {
		inputTokensDetails := usage.InputTokensDetails.Clone()
		clone.InputTokensDetails = &inputTokensDetails
	}
	return &clone
}

func cloneClaudeUsage(usage *ClaudeUsage) *ClaudeUsage {
	if usage == nil {
		return nil
	}
	clone := *usage
	clone.BillingUsage = nil
	if usage.CacheCreation != nil {
		cacheCreation := *usage.CacheCreation
		clone.CacheCreation = &cacheCreation
	}
	if usage.ServerToolUse != nil {
		serverToolUse := *usage.ServerToolUse
		clone.ServerToolUse = &serverToolUse
	}
	return &clone
}
