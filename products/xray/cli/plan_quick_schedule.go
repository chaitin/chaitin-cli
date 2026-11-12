package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	planTypeNow       = "NOW"
	planTypeClocked   = "CLOCKED"
	planTypeDay       = "DAY"
	planTypeWeek      = "WEEK"
	planTypeMonth     = "MONTH"
	planTypeMonthWeek = "MONTH_WEEK"
)

var validPlanTypes = map[string]struct{}{
	planTypeNow: {}, planTypeClocked: {}, planTypeDay: {},
	planTypeWeek: {}, planTypeMonth: {}, planTypeMonthWeek: {},
}

var weekdayNames = map[int]string{
	1: "MON", 2: "TUES", 3: "WED", 4: "THUR", 5: "FRI", 6: "SAT", 7: "SUN",
}

type quickScheduleInput struct {
	planType       string
	planTypeSet    bool
	execAt         string
	execAtSet      bool
	weekday        int
	weekdaySet     bool
	dayOfMonth     int
	dayOfMonthSet  bool
	weekOfMonth    int
	weekOfMonthSet bool
}

func registerQuickScheduleFlags(cmd *cobra.Command, defaultPlanType string) {
	cmd.Flags().String("plan-type", defaultPlanType, "计划类型: NOW, CLOCKED, DAY, WEEK, MONTH, MONTH_WEEK")
	cmd.Flags().String("exec-at", "", "触发时间；CLOCKED 使用带时区 RFC3339，周期任务使用 HH:MM 或 HH:MM:SS")
	cmd.Flags().Int("weekday", 0, "星期几，仅 WEEK/MONTH_WEEK 使用：1=周一 ... 7=周日")
	cmd.Flags().Int("day-of-month", 0, "每月第几日，仅 MONTH 使用，范围 1-31")
	cmd.Flags().Int("week-of-month", 0, "每月第几周，仅 MONTH_WEEK 使用，范围 1-4")
}

func readQuickScheduleInput(cmd *cobra.Command) (quickScheduleInput, error) {
	planType, err := cmd.Flags().GetString("plan-type")
	if err != nil {
		return quickScheduleInput{}, err
	}
	execAt, err := cmd.Flags().GetString("exec-at")
	if err != nil {
		return quickScheduleInput{}, err
	}
	weekday, err := cmd.Flags().GetInt("weekday")
	if err != nil {
		return quickScheduleInput{}, err
	}
	dayOfMonth, err := cmd.Flags().GetInt("day-of-month")
	if err != nil {
		return quickScheduleInput{}, err
	}
	weekOfMonth, err := cmd.Flags().GetInt("week-of-month")
	if err != nil {
		return quickScheduleInput{}, err
	}

	return quickScheduleInput{
		planType:       strings.ToUpper(strings.TrimSpace(planType)),
		planTypeSet:    cmd.Flags().Changed("plan-type"),
		execAt:         strings.TrimSpace(execAt),
		execAtSet:      cmd.Flags().Changed("exec-at"),
		weekday:        weekday,
		weekdaySet:     cmd.Flags().Changed("weekday"),
		dayOfMonth:     dayOfMonth,
		dayOfMonthSet:  cmd.Flags().Changed("day-of-month"),
		weekOfMonth:    weekOfMonth,
		weekOfMonthSet: cmd.Flags().Changed("week-of-month"),
	}, nil
}

func (in quickScheduleInput) changed() bool {
	return in.planTypeSet || in.execAtSet || in.weekdaySet || in.dayOfMonthSet || in.weekOfMonthSet
}

func buildPlanSetting(in quickScheduleInput, existing map[string]any) (map[string]any, error) {
	planType := in.planType
	if planType == "" && existing != nil {
		planType, _ = existing["planType"].(string)
		planType = strings.ToUpper(planType)
	}
	if _, ok := validPlanTypes[planType]; !ok {
		return nil, fmt.Errorf("invalid plan-type %q; expected NOW, CLOCKED, DAY, WEEK, MONTH, or MONTH_WEEK", planType)
	}

	setting := map[string]any{"enabled": true, "planType": planType}
	switch planType {
	case planTypeNow:
		if in.execAtSet || in.weekdaySet || in.dayOfMonthSet || in.weekOfMonthSet {
			return nil, fmt.Errorf("plan-type NOW does not accept scheduling flags")
		}
	case planTypeClocked:
		if in.weekdaySet || in.dayOfMonthSet || in.weekOfMonthSet {
			return nil, fmt.Errorf("plan-type CLOCKED only accepts --exec-at")
		}
		runAt := in.execAt
		if runAt == "" && existing != nil {
			runAt, _ = existing["startTime"].(string)
		}
		parsed, err := time.Parse(time.RFC3339, runAt)
		if err != nil {
			return nil, fmt.Errorf("--exec-at must be RFC3339 with timezone for CLOCKED: %w", err)
		}
		setting["startTime"] = parsed.Format(time.RFC3339)
	case planTypeDay:
		if in.weekdaySet || in.dayOfMonthSet || in.weekOfMonthSet {
			return nil, fmt.Errorf("plan-type DAY only accepts --exec-at")
		}
		execAt, err := resolveExecTime(in, existing, nil)
		if err != nil {
			return nil, err
		}
		setting["execTime"] = execAt
	case planTypeWeek:
		if in.dayOfMonthSet || in.weekOfMonthSet {
			return nil, fmt.Errorf("plan-type WEEK only accepts --exec-at and --weekday")
		}
		appoint := existingAppoint(existing)
		execAt, err := resolveExecTime(in, nil, appoint)
		if err != nil {
			return nil, err
		}
		weekday, err := resolveInt(in.weekday, in.weekdaySet, appoint, "weekday")
		if err != nil || weekday < 1 || weekday > 7 {
			return nil, fmt.Errorf("--weekday is required and must be between 1 and 7")
		}
		setting["appoints"] = []any{map[string]any{"weekday": weekdayNames[weekday], "execTime": execAt}}
	case planTypeMonth:
		if in.weekdaySet || in.weekOfMonthSet {
			return nil, fmt.Errorf("plan-type MONTH only accepts --exec-at and --day-of-month")
		}
		appoint := existingAppoint(existing)
		execAt, err := resolveExecTime(in, nil, appoint)
		if err != nil {
			return nil, err
		}
		day, err := resolveInt(in.dayOfMonth, in.dayOfMonthSet, appoint, "day")
		if err != nil || day < 1 || day > 31 {
			return nil, fmt.Errorf("--day-of-month is required and must be between 1 and 31")
		}
		setting["appoints"] = []any{map[string]any{"day": day, "execTime": execAt}}
	case planTypeMonthWeek:
		if in.dayOfMonthSet {
			return nil, fmt.Errorf("plan-type MONTH_WEEK only accepts --exec-at, --weekday, and --week-of-month")
		}
		appoint := existingAppoint(existing)
		execAt, err := resolveExecTime(in, nil, appoint)
		if err != nil {
			return nil, err
		}
		weekday, err := resolveInt(in.weekday, in.weekdaySet, appoint, "weekday")
		if err != nil || weekday < 1 || weekday > 7 {
			return nil, fmt.Errorf("--weekday is required and must be between 1 and 7")
		}
		week, err := resolveInt(in.weekOfMonth, in.weekOfMonthSet, appoint, "week")
		if err != nil || week < 1 || week > 4 {
			return nil, fmt.Errorf("--week-of-month is required and must be between 1 and 4")
		}
		setting["appoints"] = []any{map[string]any{
			"week": week, "weekday": weekdayNames[weekday], "execTime": execAt,
		}}
	}
	return setting, nil
}

func resolveExecTime(in quickScheduleInput, setting, appoint map[string]any) (string, error) {
	value := in.execAt
	if value == "" && appoint != nil {
		value, _ = appoint["execTime"].(string)
	}
	if value == "" && setting != nil {
		value, _ = setting["execTime"].(string)
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("15:04:05"), nil
		}
	}
	return "", fmt.Errorf("--exec-at is required and must use HH:MM or HH:MM:SS")
}

func existingAppoint(setting map[string]any) map[string]any {
	if setting == nil {
		return nil
	}
	items, ok := setting["appoints"].([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	appoint, _ := items[0].(map[string]any)
	return appoint
}

func resolveInt(value int, changed bool, existing map[string]any, key string) (int, error) {
	if changed {
		return value, nil
	}
	if existing == nil {
		return 0, fmt.Errorf("missing %s", key)
	}
	switch raw := existing[key].(type) {
	case float64:
		return int(raw), nil
	case int:
		return raw, nil
	case string:
		for number, name := range weekdayNames {
			if raw == name {
				return number, nil
			}
		}
	}
	return 0, fmt.Errorf("missing %s", key)
}

func validateSupportedPlanType(planType string, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	for _, item := range allowed {
		if strings.EqualFold(item, planType) {
			return nil
		}
	}
	sort.Strings(allowed)
	return fmt.Errorf("plan-type %s is not supported by this template; supported types: %s", planType, strings.Join(allowed, ", "))
}

func supportedPlanTypes(schema any) []string {
	root, ok := schema.(map[string]any)
	if !ok {
		return nil
	}
	properties, _ := root["properties"].(map[string]any)
	planSetting, _ := properties["planSetting"].(map[string]any)
	planProperties, _ := planSetting["properties"].(map[string]any)
	planType, _ := planProperties["planType"].(map[string]any)
	values, _ := planType["enum"].([]any)
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}
