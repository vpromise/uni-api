package common

import "github.com/QuantumNous/new-api/constant"

func ChannelType2APIType(channelType int) (int, bool) {
	switch channelType {
	case constant.ChannelTypeOpenAI:
		return constant.APITypeOpenAI, true
	case constant.ChannelTypeAnthropic:
		return constant.APITypeAnthropic, true
	case constant.ChannelTypeCodex:
		return constant.APITypeCodex, true
	default:
		return -1, false
	}
}

func SupportsResponsesCompact(channelType, apiType int) bool {
	return apiType == constant.APITypeOpenAI || apiType == constant.APITypeCodex
}
