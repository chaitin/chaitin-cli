package egressproxy

import "testing"

func TestParseScutilProxy(t *testing.T) {
	settings := parseScutilProxy(`
<dictionary> {
  ExceptionsList : <array> {
    0 : *.local
    1 : 169.254/16
    2 : localhost
  }
  HTTPEnable : 1
  HTTPPort : 7893
  HTTPProxy : 127.0.0.1
  HTTPSEnable : 1
  HTTPSPort : 7893
  HTTPSProxy : 127.0.0.1
  SOCKSEnable : 1
  SOCKSProxy : 127.0.0.1
}
`)
	if settings.httpProxy != "http://127.0.0.1:7893" || settings.httpsProxy != "http://127.0.0.1:7893" {
		t.Fatalf("settings = %+v", settings)
	}
	if !hostBypassed("foo.local", settings.bypass) || !hostBypassed("169.254.1.1", settings.bypass) {
		t.Fatalf("bypass = %v, want local and link-local excluded", settings.bypass)
	}
	if hostBypassed("product.example", settings.bypass) {
		t.Fatal("product host was bypassed")
	}
}

func TestParseWindowsProxy(t *testing.T) {
	settings := parseWindowsProxy(`
    ProxyEnable    REG_DWORD    0x1
    ProxyOverride    REG_SZ    <local>;localhost
    ProxyServer    REG_SZ    http=127.0.0.1:7890;https=127.0.0.1:7891
`)
	if settings.httpProxy != "http://127.0.0.1:7890" || settings.httpsProxy != "http://127.0.0.1:7891" {
		t.Fatalf("settings = %+v", settings)
	}
	if !hostBypassed("intranet", settings.bypass) {
		t.Fatal("simple hostname was not bypassed")
	}
}

func TestParseWindowsProxyDisabled(t *testing.T) {
	settings := parseWindowsProxy("    ProxyEnable    REG_DWORD    0x0\n    ProxyServer    REG_SZ    127.0.0.1:7890\n")
	if settings.httpProxy != "" || settings.httpsProxy != "" {
		t.Fatalf("settings = %+v, want disabled", settings)
	}
}

func TestParseGsettingsProxy(t *testing.T) {
	settings := parseGsettingsProxy("'manual'", "'127.0.0.1'", "7893", "'127.0.0.1'", "uint32 7894", "['localhost', '127.0.0.0/8']")
	if settings.httpProxy != "http://127.0.0.1:7893" || settings.httpsProxy != "http://127.0.0.1:7894" {
		t.Fatalf("settings = %+v", settings)
	}
	if !hostBypassed("127.0.0.1", settings.bypass) {
		t.Fatalf("bypass = %v, want loopback excluded", settings.bypass)
	}
	if got := parseGsettingsProxy("'none'", "'127.0.0.1'", "80", "", "", ""); got.httpProxy != "" {
		t.Fatalf("disabled settings = %+v", got)
	}
}
