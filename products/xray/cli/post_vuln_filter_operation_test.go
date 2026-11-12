package cli

import (
	"strings"
	"testing"

	"github.com/chaitin/chaitin-cli/products/xray/client/vulnerability"
)

func TestPostVulnFilterBodyAcceptsXprocessID(t *testing.T) {
	t.Parallel()
	cmd, err := makeOperationVulnerabilityPostVulnFilterCmd()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.ParseFlags([]string{`--body={"limit":1,"offset":0,"xprocess_id":1745}`}); err != nil {
		t.Fatal(err)
	}

	params := vulnerability.NewPostVulnFilterParams()
	if err, _ := retrieveOperationVulnerabilityPostVulnFilterBodyFlag(params, "", cmd); err != nil {
		t.Fatal(err)
	}
	if params.Body == nil || params.Body.XprocessID != 1745 {
		t.Fatalf("xprocess_id = %v, want 1745", params.Body)
	}
}

func TestPostVulnFilterBodyRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	cmd, err := makeOperationVulnerabilityPostVulnFilterCmd()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.ParseFlags([]string{`--body={"limit":1,"offset":0,"xprocess_ids":[1745]}`}); err != nil {
		t.Fatal(err)
	}

	params := vulnerability.NewPostVulnFilterParams()
	err, _ = retrieveOperationVulnerabilityPostVulnFilterBodyFlag(params, "", cmd)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %v, want unknown field error", err)
	}
}

func TestPostVulnFilterXprocessIDFlag(t *testing.T) {
	t.Parallel()
	cmd, err := makeOperationVulnerabilityPostVulnFilterCmd()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.ParseFlags([]string{"--filterVulnBody.xprocess_id=1745"}); err != nil {
		t.Fatal(err)
	}

	params := vulnerability.NewPostVulnFilterParams()
	if err, _ := retrieveOperationVulnerabilityPostVulnFilterBodyFlag(params, "", cmd); err != nil {
		t.Fatal(err)
	}
	if params.Body == nil || params.Body.XprocessID != 1745 {
		t.Fatalf("xprocess_id = %v, want 1745", params.Body)
	}
}
