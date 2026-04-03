package main

import "testing"

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "default path",
			args: []string{"safeline", "stats", "overview"},
			want: defaultConfigPathFromCWD(),
		},
		{
			name: "long flag with value",
			args: []string{"--config", "/tmp/a.yaml", "safeline"},
			want: "/tmp/a.yaml",
		},
		{
			name: "long flag inline",
			args: []string{"--config=/tmp/a.yaml", "safeline"},
			want: "/tmp/a.yaml",
		},
		{
			name: "short flag with value",
			args: []string{"-c", "/tmp/a.yaml", "safeline"},
			want: "/tmp/a.yaml",
		},
		{
			name: "short flag inline",
			args: []string{"-c/tmp/a.yaml", "safeline"},
			want: "/tmp/a.yaml",
		},
		{
			name:    "missing value",
			args:    []string{"-c"},
			wantErr: true,
		},
		{
			name: "stop at double dash",
			args: []string{"safeline", "--", "--config", "/tmp/a.yaml"},
			want: defaultConfigPathFromCWD(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := resolveConfigPath(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveConfigPath() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveConfigPath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
