package cli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/products/xray/models"
	"github.com/spf13/cobra"
)

func TestBuildPlanSetting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   quickScheduleInput
		want    map[string]any
		wantErr string
	}{
		{
			name:  "now",
			input: quickScheduleInput{planType: planTypeNow},
			want:  map[string]any{"enabled": true, "planType": planTypeNow},
		},
		{
			name:  "clocked",
			input: quickScheduleInput{planType: planTypeClocked, execAt: "2026-08-05T10:30:00+08:00", execAtSet: true},
			want:  map[string]any{"enabled": true, "planType": planTypeClocked, "startTime": "2026-08-05T10:30:00+08:00"},
		},
		{
			name:  "day normalizes time",
			input: quickScheduleInput{planType: planTypeDay, execAt: "10:30", execAtSet: true},
			want:  map[string]any{"enabled": true, "planType": planTypeDay, "execTime": "10:30:00"},
		},
		{
			name:  "week maps friday",
			input: quickScheduleInput{planType: planTypeWeek, execAt: "10:30", execAtSet: true, weekday: 5, weekdaySet: true},
			want: map[string]any{"enabled": true, "planType": planTypeWeek, "appoints": []any{
				map[string]any{"weekday": "FRI", "execTime": "10:30:00"},
			}},
		},
		{
			name:  "month",
			input: quickScheduleInput{planType: planTypeMonth, execAt: "10:30:45", execAtSet: true, dayOfMonth: 15, dayOfMonthSet: true},
			want: map[string]any{"enabled": true, "planType": planTypeMonth, "appoints": []any{
				map[string]any{"day": 15, "execTime": "10:30:45"},
			}},
		},
		{
			name: "month week maps sunday",
			input: quickScheduleInput{
				planType: planTypeMonthWeek, execAt: "10:30", execAtSet: true,
				weekday: 7, weekdaySet: true, weekOfMonth: 4, weekOfMonthSet: true,
			},
			want: map[string]any{"enabled": true, "planType": planTypeMonthWeek, "appoints": []any{
				map[string]any{"week": 4, "weekday": "SUN", "execTime": "10:30:00"},
			}},
		},
		{
			name:    "weekday out of range",
			input:   quickScheduleInput{planType: planTypeWeek, execAt: "10:30", weekday: 8, weekdaySet: true},
			wantErr: "between 1 and 7",
		},
		{
			name:    "month day out of range",
			input:   quickScheduleInput{planType: planTypeMonth, execAt: "10:30", dayOfMonth: 32, dayOfMonthSet: true},
			wantErr: "between 1 and 31",
		},
		{
			name:    "now rejects schedule",
			input:   quickScheduleInput{planType: planTypeNow, execAt: "10:30", execAtSet: true},
			wantErr: "does not accept",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := buildPlanSetting(tt.input, nil)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("buildPlanSetting() error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildPlanSetting() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildPlanSetting() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestBuildPlanSettingPreservesExistingScheduleValues(t *testing.T) {
	t.Parallel()
	existing := map[string]any{
		"enabled":  true,
		"planType": planTypeWeek,
		"appoints": []any{map[string]any{"weekday": "MON", "execTime": "09:00:00"}},
	}
	got, err := buildPlanSetting(quickScheduleInput{execAt: "11:30", execAtSet: true}, existing)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"enabled":  true,
		"planType": planTypeWeek,
		"appoints": []any{map[string]any{"weekday": "MON", "execTime": "11:30:00"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildPlanSetting() = %#v, want %#v", got, want)
	}
}

func TestSupportedPlanTypes(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"properties": map[string]any{
		"planSetting": map[string]any{"properties": map[string]any{
			"planType": map[string]any{"enum": []any{"DAY", "WEEK", "MONTH"}},
		}},
	}}
	want := []string{"DAY", "WEEK", "MONTH"}
	if got := supportedPlanTypes(schema); !reflect.DeepEqual(got, want) {
		t.Fatalf("supportedPlanTypes() = %v, want %v", got, want)
	}
	if err := validateSupportedPlanType("NOW", want); err == nil {
		t.Fatal("validateSupportedPlanType() expected an error")
	}
}

func TestQuickCommandHelp(t *testing.T) {
	t.Parallel()
	create, err := makeOperationPlanCreateQuickCmd()
	if err != nil {
		t.Fatal(err)
	}
	update, err := makeOperationPlanUpdateQuickCmd()
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []*cobra.Command{create, update} {
		text := cmd.Long + cmd.Flags().FlagUsages()
		for _, expected := range []string{"--exec-at", "--weekday", "1=周一", "7=周日", "--day-of-month", "--week-of-month", "MONTH_WEEK", "示例"} {
			if !strings.Contains(text, expected) {
				t.Errorf("%s help does not contain %q", cmd.Name(), expected)
			}
		}
	}
}

func TestQuickCommandsAreRegistered(t *testing.T) {
	t.Parallel()
	planCommand, err := makeGroupOfOperationsPlanCmd()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"PostPlanCreateQuick", "PostPlanUpdateQuick"} {
		command, _, err := planCommand.Find([]string{name})
		if err != nil || command == nil || command.Name() != name {
			t.Fatalf("plan command %q is not registered", name)
		}
	}
}

func TestUpdatePlanBodyMarshalsBasicSettingAsObject(t *testing.T) {
	t.Parallel()
	id := int64(42)
	execRightNow := false
	body, err := json.Marshal(models.UpdatePlanBody{
		ID:           &id,
		ExecRightNow: &execRightNow,
		BasicSetting: map[string]any{"planSetting": map[string]any{"planType": planTypeDay}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), `"basic_setting":"`) {
		t.Fatalf("basic_setting was encoded as a JSON string: %s", body)
	}
}
