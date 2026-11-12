package client

import (
	"net/http"

	"github.com/chaitin/chaitin-cli/egressproxy"
)

// init runs after client.go because of the file name, so the generated
// default client already exists. Wrapping here keeps the generated client untouched.
func init() {
	if defaultClient == nil || defaultClient.httpClient == nil {
		return
	}
	transport, ok := defaultClient.httpClient.Transport.(*http.Transport)
	if !ok {
		return
	}
	egressproxy.Wrap(transport)
}
