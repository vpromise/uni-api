package constant

const (
	ChannelTypeUnknown   = 0
	ChannelTypeOpenAI    = 1
	ChannelTypeAnthropic = 14
	ChannelTypeCodex     = 57
	// Task plugins execute on supported upstream channels, not standalone channels.
	ChannelTypeTaskPlugin = 61
	ChannelTypeDummy      = 64
)

var ChannelBaseURLs = []string{
	ChannelTypeOpenAI:    "https://api.openai.com",
	ChannelTypeAnthropic: "https://api.anthropic.com",
	ChannelTypeCodex:     "https://chatgpt.com",
}

func GetChannelBaseURL(channelType int) string {
	if !IsSupportedChannelType(channelType) {
		return ""
	}
	return ChannelBaseURLs[channelType]
}

var ChannelTypeNames = map[int]string{
	ChannelTypeOpenAI:    "OpenAI",
	ChannelTypeAnthropic: "Anthropic",
	ChannelTypeCodex:     "ChatGPT Subscription (Codex)",
}

func GetChannelTypeName(channelType int) string {
	if name, ok := ChannelTypeNames[channelType]; ok {
		return name
	}
	return "Unknown"
}

func IsSupportedChannelType(channelType int) bool {
	switch channelType {
	case ChannelTypeOpenAI, ChannelTypeAnthropic, ChannelTypeCodex:
		return true
	default:
		return false
	}
}
