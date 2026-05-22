package cronexpr

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var predefinedExpressions = map[string]struct{}{
	"@yearly":   {},
	"@annually": {},
	"@monthly":  {},
	"@weekly":   {},
	"@daily":    {},
	"@midnight": {},
	"@hourly":   {},
}

var monthNames = map[string]int{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

var weekNames = map[string]int{
	"sun": 0,
	"mon": 1,
	"tue": 2,
	"wed": 3,
	"thu": 4,
	"fri": 5,
	"sat": 6,
}

type fieldRange struct {
	min           int
	max           int
	names         map[string]int
	allowQuestion bool
}

// Normalize adapts five-field cron expressions to GoFrame gcron seconds-first format.
func Normalize(expression string) string {
	expression = strings.TrimSpace(expression)
	if len(strings.Fields(expression)) == 5 {
		return "0 " + expression
	}
	return expression
}

// Validate checks supported GoFrame cron expressions before they reach gcron.
func Validate(expression string) error {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil
	}

	lowerExpression := strings.ToLower(expression)
	if _, ok := predefinedExpressions[lowerExpression]; ok {
		return nil
	}
	if strings.HasPrefix(lowerExpression, "@every ") {
		durationText := strings.TrimSpace(expression[len("@every "):])
		if durationText == "" {
			return fmt.Errorf("Cron 表达式 @every 缺少时间间隔")
		}
		if _, err := time.ParseDuration(durationText); err != nil {
			return fmt.Errorf("Cron 表达式 @every 时间间隔无效")
		}
		return nil
	}

	parts := strings.Fields(expression)
	if len(parts) != 5 && len(parts) != 6 {
		return fmt.Errorf("Cron 表达式必须是 5 段或 6 段")
	}
	if len(parts) == 5 {
		parts = append([]string{"0"}, parts...)
	}

	ranges := []fieldRange{
		{min: 0, max: 59},
		{min: 0, max: 59},
		{min: 0, max: 23},
		{min: 1, max: 31, allowQuestion: true},
		{min: 1, max: 12, names: monthNames},
		{min: 0, max: 6, names: weekNames, allowQuestion: true},
	}
	for index, part := range parts {
		if !validateField(part, ranges[index]) {
			return fmt.Errorf("Cron 表达式第 %d 段无效", index+1)
		}
	}
	return nil
}

func validateField(part string, field fieldRange) bool {
	part = strings.TrimSpace(part)
	if part == "" {
		return false
	}
	if part == "*" {
		return true
	}
	if field.allowQuestion && part == "?" {
		return true
	}

	for _, item := range strings.Split(part, ",") {
		if !validateListItem(item, field) {
			return false
		}
	}
	return true
}

func validateListItem(item string, field fieldRange) bool {
	item = strings.TrimSpace(item)
	if item == "" {
		return false
	}

	parts := strings.Split(item, "/")
	if len(parts) > 2 {
		return false
	}
	base := parts[0]
	if len(parts) == 2 {
		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return false
		}
	}
	if base == "*" {
		return true
	}
	if field.allowQuestion && base == "?" {
		return true
	}
	if strings.Contains(base, "-") {
		rangeParts := strings.Split(base, "-")
		if len(rangeParts) != 2 || rangeParts[0] == "" || rangeParts[1] == "" {
			return false
		}
		start, ok := parseValue(rangeParts[0], field)
		if !ok {
			return false
		}
		end, ok := parseValue(rangeParts[1], field)
		return ok && start <= end
	}

	_, ok := parseValue(base, field)
	return ok
}

func parseValue(value string, field fieldRange) (int, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if field.names != nil {
		if namedValue, ok := field.names[value]; ok {
			return namedValue, true
		}
	}

	numberValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	if numberValue < field.min || numberValue > field.max {
		return 0, false
	}
	return numberValue, true
}
