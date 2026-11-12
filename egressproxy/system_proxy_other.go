//go:build !darwin && !windows

package egressproxy

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func readPlatformSystemProxy() systemProxySettings {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := exec.LookPath("gsettings"); err != nil {
		return systemProxySettings{}
	}
	mode := gsettings(ctx, "get", "org.gnome.system.proxy", "mode")
	if strings.Trim(strings.TrimSpace(mode), "'") != "manual" {
		return systemProxySettings{}
	}
	return parseGsettingsProxy(
		mode,
		gsettings(ctx, "get", "org.gnome.system.proxy.http", "host"),
		gsettings(ctx, "get", "org.gnome.system.proxy.http", "port"),
		gsettings(ctx, "get", "org.gnome.system.proxy.https", "host"),
		gsettings(ctx, "get", "org.gnome.system.proxy.https", "port"),
		gsettings(ctx, "get", "org.gnome.system.proxy", "ignore-hosts"),
	)
}

func gsettings(ctx context.Context, args ...string) string {
	out, err := exec.CommandContext(ctx, "gsettings", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
