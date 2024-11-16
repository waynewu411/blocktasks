package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const maxBodyLogSize = 4 * 1024

type debugTripper struct {
	lg           *zap.Logger
	roundTripper http.RoundTripper
}

func newDebugTripper(lg *zap.Logger) http.RoundTripper {
	return &debugTripper{lg: lg, roundTripper: http.DefaultTransport}
}

func (t *debugTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	requestId := uuid.New().String()
	start := time.Now()
	defer func() {
		reqBody := newLimitedBuffer(maxBodyLogSize)
		if req != nil && req.Body != nil {
			req.Body = io.NopCloser(io.TeeReader(req.Body, reqBody))
		}

		respBody := newLimitedBuffer(maxBodyLogSize)
		if resp != nil && resp.Body != nil {
			resp.Body = io.NopCloser(io.TeeReader(resp.Body, respBody))
		}

		t.lg.Debug(
			fmt.Sprintf(
				"%s %s:%s %d %v",
				requestId, req.Method, req.URL.String(), resp.StatusCode, time.Since(start),
			),
			zap.Any("request headers", req.Header),
			zap.String("request body", reqBody.String()),
			zap.Any("response headers", resp.Header),
			zap.String("response body", respBody.String()),
		)
	}()

	resp, err = t.roundTripper.RoundTrip(req)
	if err != nil {
		t.lg.Error("request error", zap.Error(err))
		return nil, err
	}

	return resp, nil
}
