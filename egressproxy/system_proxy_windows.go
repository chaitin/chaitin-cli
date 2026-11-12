//go:build windows

package egressproxy

import (
	"context"
	"os/exec"
	"time"
)

func readPlatformSystemProxy() systemProxySettings {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`).Output()
	if err != nil {
		return systemProxySettings{}
	}
	return parseWindowsProxy(string(out))
}
