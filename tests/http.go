//nolint:forcetypeassert
package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/dic"
	internalerrors "github.com/estrys/estrys/internal/errors"
	dic_test "github.com/estrys/estrys/tests/dic"
)

type RequestOption interface {
	Value() any
}

type requestParams map[string]string
type RequestParams struct {
	Params requestParams
}

func (r RequestParams) Value() any {
	return r.Params
}

type RequestBody struct {
	Body io.Reader
}

func (r RequestBody) Value() any {
	return r.Body
}

type RequestBodyFromFile struct {
	FilePath string
}

func (r RequestBodyFromFile) Value() any {
	return r.FilePath
}

type RequestContext struct {
	Context context.Context //nolint:containedctx
}

func (r RequestContext) Value() any {
	return r.Context
}

type HTTPTestCase struct {
	Name           string
	GoldenFile     string
	StatusCode     int
	RequestOptions []RequestOption
	Mock           func(t *testing.T)
}

type HTTPTestSuite struct{}

func (s *HTTPTestSuite) RunHTTPCases(t *testing.T, handler internalerrors.ErrorAwareHTTPHandler, cases []HTTPTestCase) {
	t.Helper()
	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Mock != nil {
				testCase.Mock(t)
			}
			response, body := s.DoRequest(t, handler, testCase.RequestOptions...)
			defer response.Body.Close()
			require.Equal(t, testCase.StatusCode, response.StatusCode)
			if testCase.GoldenFile != "" {
				switch {
				case strings.HasSuffix(testCase.GoldenFile, ".json"):
					AssertJSONResponse(t, testCase.GoldenFile, string(body))
				case strings.HasSuffix(testCase.GoldenFile, ".html"):
					AssertHTMLResponse(t, testCase.GoldenFile, string(body))
				default:
					t.Errorf(
						"unable to match a response format for goldenfile: %s",
						testCase.GoldenFile,
					)
				}
			}
		})
	}
}

func (s *HTTPTestSuite) DoRequest(
	t *testing.T,
	handler internalerrors.ErrorAwareHTTPHandler,
	opts ...RequestOption,
) (*http.Response, []byte) {
	t.Helper()
	dic_test.BuildTestContainer(t)
	defer dic.ResetContainer()

	request := createFakeRequest(opts)

	responseRecorder := httptest.NewRecorder()
	internalerrors.HTTPErrorHandler(handler)(responseRecorder, request)
	response := responseRecorder.Result()
	body, err := io.ReadAll(responseRecorder.Result().Body)

	require.NoError(t, err)

	return response, body
}

func createFakeRequest(opts []RequestOption) *http.Request {
	var params requestParams
	var body io.ReadCloser
	var ctx context.Context
	for _, opt := range opts {
		if paramOp, ok := opt.(RequestParams); ok {
			params = paramOp.Value().(requestParams)
		}
		if bodyFileOp, ok := opt.(RequestBodyFromFile); ok {
			filePath := bodyFileOp.Value().(string)
			fileContent, err := os.ReadFile(GetGoldenFilePath(filePath))
			if err != nil {
				panic(errors.Wrap(err, "unable to open http test file input"))
			}
			body = io.NopCloser(bytes.NewReader(fileContent))
		}
		if bodyOp, ok := opt.(RequestBody); ok {
			body = bodyOp.Value().(io.ReadCloser)
		}
		if ctxOp, ok := opt.(RequestContext); ok {
			ctx = ctxOp.Value().(context.Context)
		}
	}
	req := &http.Request{
		Body: body,
	}
	req = req.WithContext(context.Background())
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	if params != nil {
		req = mux.SetURLVars(req, params)
	}
	return req
}
