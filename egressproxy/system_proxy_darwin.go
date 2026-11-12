//go:build darwin

package egressproxy

import (
	"context"
	"os/exec"
	"time"
)

func readPlatformSystemProxy() systemProxySettings {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "scutil", "--proxy").Output()
	if err != nil {
		return systemProxySettings{}
	}
	return parseScutilProxy(string(out))
}
