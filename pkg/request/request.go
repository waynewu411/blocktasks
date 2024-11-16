package request

//go:generate mockgen -destination=mock_request.go -package=request github.com/waynewu411/blocktasks/pkg/request Request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/waynewu411/blocktasks/pkg/httpclient"
	"go.uber.org/zap"
)

type Request interface {
	MakeRequest(method string, apiUrl string, queryParams map[string]string, reqBody string) ([]byte, error)
}

type request struct {
	lg             *zap.Logger
	httpClient     httpclient.HttpClient
	requestHeaders map[string]string
}

type Option func(*request)

func WithHttpClient(client httpclient.HttpClient) Option {
	return func(r *request) {
		r.httpClient = client
	}
}

func WithRequestHeaders(headers map[string]string) Option {
	return func(r *request) {
		r.requestHeaders = headers
	}
}

func NewRequest(lg *zap.Logger, opts ...Option) Request {
	r := &request{
		lg:         lg,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (t *request) MakeRequest(method string, apiUrl string, queryParams map[string]string, reqBody string) ([]byte, error) {
	requestUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	query := requestUrl.Query()
	for key, value := range queryParams {
		query.Set(key, value)
	}
	requestUrl.RawQuery = query.Encode()

	req, err := http.NewRequest(method, requestUrl.String(), strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range t.requestHeaders {
		req.Header.Set(key, value)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to make request: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
