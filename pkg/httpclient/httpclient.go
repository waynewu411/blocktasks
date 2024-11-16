package httpclient

import (
	"net/http"

	"github.com/waynewu411/blocktasks/pkg/config"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type httpClient struct {
	lg          *zap.Logger
	client      *http.Client
	rateLimiter *rate.Limiter
}

func NewHttpClient(lg *zap.Logger, cfg config.HttpClientConfig) HttpClient {
	c := &httpClient{lg: lg, client: &http.Client{}}
	if cfg.RateLimit > 0 {
		c.rateLimiter = rate.NewLimiter(rate.Limit(cfg.RateLimit), 1)
	}
	if cfg.DebugEnabled {
		c.client.Transport = newDebugTripper(lg)
	} else {
		c.client.Transport = http.DefaultTransport
	}
	return c
}

func (t *httpClient) Do(req *http.Request) (*http.Response, error) {
	if t.rateLimiter != nil {
		if err := t.rateLimiter.Wait(req.Context()); err != nil {
			t.lg.Error("rate limiter wait error", zap.Error(err))
			return nil, err
		}
	}
	return t.client.Do(req)
}
