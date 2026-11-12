package models

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-openapi/strfmt"
)

func TestFilterVulnBodyOmitsUnsetSlices(t *testing.T) {
	t.Parallel()
	limit, offset := int64(3), int64(0)
	body := FilterVulnBody{
		LimitAndOffsetFilter: LimitAndOffsetFilter{Limit: &limit, Offset: &offset},
		ProjectID:            1,
		Status:               []VulnStatusEnum{},
		XprocessID:           1745,
	}

	assertJSONEquals(t, body, map[string]any{
		"limit":       float64(3),
		"offset":      float64(0),
		"project_id":  float64(1),
		"xprocess_id": float64(1745),
	})
}

func TestFilterReportOmitsUnsetSlices(t *testing.T) {
	t.Parallel()
	limit, offset := int64(10), int64(0)
	body := FilterReport{CommonFilter: CommonFilter{
		LimitAndOffsetFilter: LimitAndOffsetFilter{Limit: &limit, Offset: &offset},
	}}

	assertJSONEquals(t, body, map[string]any{"limit": float64(10), "offset": float64(0)})
}

func TestPostCustompocUpdateParamsBodyOptionalSlices(t *testing.T) {
	t.Parallel()
	id := strfmt.UUID("00000000-0000-0000-0000-000000000001")

	t.Run("unset lists are omitted", func(t *testing.T) {
		assertJSONEquals(t, PostCustompocUpdateParamsBody{CustomPocID: &id}, map[string]any{
			"custom_poc_id": string(id),
		})
	})
	t.Run("explicit empty lists are preserved", func(t *testing.T) {
		assertJSONEquals(t, PostCustompocUpdateParamsBody{
			CustomPocID: &id,
			Exposures:   []string{},
			Tags:        []strfmt.UUID{},
		}, map[string]any{
			"custom_poc_id": string(id),
			"exposures":     []any{},
			"tags":          []any{},
		})
	})
	t.Run("populated lists are preserved", func(t *testing.T) {
		tag := strfmt.UUID("00000000-0000-0000-0000-000000000002")
		assertJSONEquals(t, PostCustompocUpdateParamsBody{
			CustomPocID: &id,
			Exposures:   []string{"CVE-2026-0001"},
			Tags:        []strfmt.UUID{tag},
		}, map[string]any{
			"custom_poc_id": string(id),
			"exposures":     []any{"CVE-2026-0001"},
			"tags":          []any{string(tag)},
		})
	})
}

func TestDeleteCustompocParamsBodyRemainsRequired(t *testing.T) {
	t.Parallel()
	if err := (&DeleteCustompocParamsBody{}).Validate(strfmt.Default); err == nil {
		t.Fatal("expected custom_poc_id_list to remain required")
	}
}

func assertJSONEquals(t *testing.T, value any, want map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON = %s, want %#v", encoded, want)
	}
}
