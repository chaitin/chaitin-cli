package client

import (
	"encoding/json"
	"testing"

	"github.com/spf13/pflag"
)

type boolParamSample struct {
	Name    []string `json:"name"`
	Online  bool     `json:"is_online"`
	Collect bool     `json:"enable_net_conn_collect"`
	Force   bool     `json:"force"`
}

func newSampleFlagSet(p *boolParamSample) *pflag.FlagSet {
	fs := pflag.NewFlagSet("sample", pflag.ContinueOnError)
	fs.StringSliceVar(&p.Name, "name", nil, "name")
	fs.BoolVar(&p.Online, "is-online", false, "online")
	fs.BoolVar(&p.Collect, "enable-net-conn-collect", false, "collect")
	fs.BoolVar(&p.Force, "force", false, "force")
	return fs
}

func marshalForTest(t *testing.T, fs *pflag.FlagSet, params interface{}) map[string]interface{} {
	t.Helper()

	raw, err := marshalParamsWithUnsetOmitted(params, unsetBoolJSONKeys(fs, params))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

func TestUnsetBoolsAreOmitted(t *testing.T) {
	params := &boolParamSample{}
	fs := newSampleFlagSet(params)

	out := marshalForTest(t, fs, params)

	for _, key := range []string{"is_online", "enable_net_conn_collect", "force"} {
		if _, ok := out[key]; ok {
			t.Errorf("%s should be omitted when its flag was never set", key)
		}
	}
}

func TestExplicitFalseIsSent(t *testing.T) {
	params := &boolParamSample{}
	fs := newSampleFlagSet(params)
	if err := fs.Set("is-online", "false"); err != nil {
		t.Fatal(err)
	}

	out := marshalForTest(t, fs, params)

	value, ok := out["is_online"]
	if !ok {
		t.Fatal("is_online must be sent when explicitly set to false")
	}
	if value != false {
		t.Errorf("is_online = %v, want false", value)
	}
}

func TestExplicitTrueIsSent(t *testing.T) {
	params := &boolParamSample{}
	fs := newSampleFlagSet(params)
	if err := fs.Set("is-online", "true"); err != nil {
		t.Fatal(err)
	}

	out := marshalForTest(t, fs, params)

	if out["is_online"] != true {
		t.Errorf("is_online = %v, want true", out["is_online"])
	}
}

func TestNonBoolFieldsUntouched(t *testing.T) {
	params := &boolParamSample{}
	fs := newSampleFlagSet(params)
	if err := fs.Set("name", "a,b"); err != nil {
		t.Fatal(err)
	}

	out := marshalForTest(t, fs, params)

	if out["name"] == nil {
		t.Error("name should still be sent")
	}
}

func TestNoFlagsKeepsParamsAsIs(t *testing.T) {
	params := &boolParamSample{Online: true}

	raw, err := marshalParamsWithUnsetOmitted(params, unsetBoolJSONKeys(nil, params))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out["is_online"] != true {
		t.Error("params must be sent unchanged when no flag set is available")
	}
}

func TestNonStructParamsPassThrough(t *testing.T) {
	fs := pflag.NewFlagSet("sample", pflag.ContinueOnError)
	fs.Bool("force", false, "force")

	raw, err := marshalParamsWithUnsetOmitted([]string{"a"}, unsetBoolJSONKeys(fs, []string{"a"}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(raw) != `["a"]` {
		t.Errorf("non-object params = %s, want [\"a\"]", raw)
	}
}
