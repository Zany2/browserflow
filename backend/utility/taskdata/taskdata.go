package taskdata

import (
	"encoding/json"
	"strings"

	"github.com/Zany2/browserflow/backend/internal/model"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// DecodeJSONMap parses JSON object text. 解析 JSON 对象文本
func DecodeJSONMap(value string) (model.JSONMap, error) {
	result := model.JSONMap{}
	if strings.TrimSpace(value) == "" {
		return result, nil
	}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// EncodeJSONMap serializes map fields for jsonb columns. 序列化 JSONB 字段
func EncodeJSONMap(value model.JSONMap) (string, error) {
	if value == nil {
		return "{}", nil
	}
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// RecordTime converts database time values to GoFrame time. 转换数据库时间字段
func RecordTime(value any) *gtime.Time {
	if value == nil {
		return nil
	}
	t := gconv.Time(value)
	if t.IsZero() {
		return nil
	}
	return gtime.NewFromTime(t)
}

// NormalizeTriggerType keeps execution records in known trigger types. 规范化触发类型
func NormalizeTriggerType(triggerType string) string {
	triggerType = strings.TrimSpace(triggerType)
	switch triggerType {
	case "cron", "task_create", "skill", "system":
		return triggerType
	default:
		return "manual"
	}
}
