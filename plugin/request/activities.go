package request

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var Activities = (*activities)(nil)

type activities struct {
	client *http.Client
}

type Assert struct {
	StatusCodes []int  `json:"status_code,omitempty"`
	Path        string `json:"path,omitempty"`
	Value       []any  `json:"value,omitempty"`
}

type JSONRequest struct {
	Method  string              `json:"method,omitempty"`
	URL     string              `json:"url,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
	Assert  Assert              `json:"assert,omitempty"`
}

func (r *activities) DoJSON(ctx context.Context, request JSONRequest) (any, *cadence.CustomError) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity-start", zap.Any("request", request))

	if request.Headers == nil {
		request.Headers = make(map[string][]string)
	}
	if request.Assert.StatusCodes == nil {
		request.Assert.StatusCodes = []int{200, 201, 202}
	}

	var err error
	var bb []byte
	var req *http.Request
	var res *http.Response

	if request.Body != nil {
		if bb, err = jsoniter.Marshal(request.Body); err != nil {
			logger.Error("activity-error", ext.ZapError(err)...)
			return nil, cadence.NewCustomError(yarpcerrors.CodeInvalidArgument.String(), err.Error())
		}
	}

	header := http.Header(request.Headers)

	if header.Get("content-length") == "" {
		cl := fmt.Sprintf("%d", len(bb))
		header.Set("content-length", cl)
	}
	if header.Get("content-type") == "" {
		header.Set("content-type", "application/json; charset=utf-8")
	}
	if header.Get("accept") == "" {
		header.Set("accept", "application/json")
	}
	if header.Get("accept-charset") == "" {
		header.Set("accept-charset", "utf-8")
	}

	if req, err = createRequest(ctx, request.Method, request.URL, header, bb); err != nil {
		logger.Error("activity-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(yarpcerrors.CodeInvalidArgument.String(), err.Error())
	}

	if res, err = r.client.Do(req); err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			logger = logger.With(
				zap.String("url_err_op", urlErr.Op),
				zap.String("url_err_url", urlErr.URL),
				zap.String("url_err_err", urlErr.Err.Error()),
				zap.String("url_err_err_type", fmt.Sprintf("%T", urlErr.Err)),
			)
		}
		logger.Error("activity-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(yarpcerrors.CodeUnknown.String(), err.Error())
	}

	var expectedStatusCode = false
	for _, code := range request.Assert.StatusCodes {
		if code == res.StatusCode {
			expectedStatusCode = true
			break
		}
	}
	if !expectedStatusCode {
		details := fmt.Sprintf("bad http response status code: expected: %v, got: %d", request.Assert.StatusCodes, res.StatusCode)
		logger.Error("activity-error", zap.String("details", details))
		return nil, cadence.NewCustomError(strconv.Itoa(res.StatusCode), details)
	}

	var _res any
	if err := jsoniter.NewDecoder(res.Body).Decode(&_res); err != nil {
		logger.Error("activity-error", ext.ZapError(err)...)
		code := "400" // bad-request https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400
		details := fmt.Sprintf("http response body is not a json: %s", err.Error())
		return nil, cadence.NewCustomError(code, details)
	}

	if request.Assert.Path != "" {
		value, err := ext.JP[any](_res, request.Assert.Path)
		if err != nil {
			// 412 - precondition-failed https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/412
			return nil, cadence.NewCustomError("412", _res)
		}
		if request.Assert.Value != nil {
			var found bool
			for _, v := range request.Assert.Value {
				if v == value {
					found = true
					break
				}
			}
			if !found {
				// 412 - precondition-failed https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/412
				return nil, cadence.NewCustomError("412", _res)
			}
		}
	}
	return _res, nil
}

func (r *activities) Do(
	ctx context.Context,
	method string,
	url string,
	headers map[string][]string,
	body []byte,
) ([]byte, *cadence.CustomError) {
	logger := activity.GetLogger(ctx)
	logger.Info(
		"activity-start",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("body_len", len(body)),
	)
	if req, err := createRequest(ctx, method, url, headers, body); err != nil {
		logger.Error("activity-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(yarpcerrors.CodeInvalidArgument.String(), err.Error())
	} else {
		return do(ctx, r.client, req)
	}
}

func do(ctx context.Context, client *http.Client, req *http.Request) ([]byte, *cadence.CustomError) {
	logger := activity.GetLogger(ctx)
	res, err := client.Do(req)
	if err != nil {
		logger.Error("activity-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(yarpcerrors.CodeUnknown.String(), err.Error())
	}
	var buf bytes.Buffer
	if err := res.Write(&buf); err != nil {
		logger.Error("activity-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(yarpcerrors.CodeInternal.String(), err.Error())
	}
	return buf.Bytes(), nil
}

func createRequest(ctx context.Context, method string, url string, headers http.Header, body []byte) (*http.Request, error) {
	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, br)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	return req, nil
}
