package httpclient

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/waynewu411/blocktasks/pkg/config"
	"go.uber.org/zap/zaptest"
)

func Test_HttpClientRateLimiter(t *testing.T) {
	httpClientCfg := config.HttpClientConfig{
		RateLimit:    1,
		DebugEnabled: true,
	}

	hc := NewHttpClient(zaptest.NewLogger(t), &httpClientCfg)

	req, err := http.NewRequest(http.MethodGet, "https://google.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := hc.Do(req)
	assert.NoError(t, err)
	resp.Body.Close()

	// create a context with very short timeout
	// so that we can confirm this request is pending on
	// rateLimiter.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://google.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = hc.Do(req)
	assert.Error(t, err)
}
