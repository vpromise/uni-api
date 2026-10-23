package task_pricing_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/samber/lo"
)

type TaskPricingSetting struct {
	SoraSizeRatio map[string]float64 `json:"sora_size_ratio"`
}

var taskPricingSetting = TaskPricingSetting{
	SoraSizeRatio: map[string]float64{
		"1792x1024": 1.666667,
		"1024x1792": 1.666667,
	},
}

func init() {
	config.GlobalConfig.Register("task_pricing_setting", &taskPricingSetting)
}

func SoraSizeRatio(size string) float64 {
	if ratio, ok := taskPricingSetting.SoraSizeRatio[size]; ok && ratio > 0 {
		return ratio
	}
	return 1
}

func GetCopy() TaskPricingSetting {
	return TaskPricingSetting{
		SoraSizeRatio: lo.Assign(taskPricingSetting.SoraSizeRatio),
	}
}
