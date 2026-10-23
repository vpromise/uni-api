package types

type RelayFormat string

const (
	RelayFormatOpenAI RelayFormat = "openai"
	RelayFormatClaude             = "claude"

	RelayFormatOpenAIResponses           = "openai_responses"
	RelayFormatOpenAIResponsesCompaction = "openai_responses_compaction"
	RelayFormatOpenAIAlphaSearch         = "openai_alpha_search"

	RelayFormatOpenAIAudio    = "openai_audio"
	RelayFormatOpenAIImage    = "openai_image"
	RelayFormatOpenAIRealtime = "openai_realtime"

	RelayFormatEmbedding = "embedding"

	RelayFormatTask = "task"
)
