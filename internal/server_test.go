package internal

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/estrys/estrys/internal/dic"
	internalerrors "github.com/estrys/estrys/internal/errors"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/logger/mocks"
)

const httpTestServerListenAddr = "localhost:11337"

type HTTPServerTestSuite struct {
	suite.Suite
}

func TestHTTPServerTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPServerTestSuite))
}

func (suite *HTTPServerTestSuite) SetupTest() {
	dic.ResetContainer()
}

func (suite *HTTPServerTestSuite) TestErrorHandler() {
	t := suite.T()
	_ = dic.Register[logger.Logger](mocks.NewNullLogger())
	r := mux.NewRouter()
	r.Path("/").HandlerFunc(internalerrors.HTTPErrorHandler(func(w http.ResponseWriter, req *http.Request) error {
		return internalerrors.New("", http.StatusNotFound).WithUserMessage("test not found")
	}))
	_ = dic.Register[*mux.Router](r)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	testURL, _ := url.JoinPath("http://", httpTestServerListenAddr)
	req, _ := http.NewRequest(http.MethodGet, testURL, nil)
	httpClient := http.Client{}

	var wg sync.WaitGroup
	var serverErr error
	go func() {
		wg.Add(1)
		serverErr = StartServer(ctx, Config{Address: httpTestServerListenAddr})
		wg.Done()
	}()

	// We need to wait for the server to be ready, we do not have a way to sync on that so far
	// Increase this value if the test is flaky
	time.Sleep(3 * time.Second)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("content-type"))
	bodyContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, `{"error":"test not found"}`, string(bodyContent))
	cancelFunc()
	wg.Wait()
	require.NoError(t, serverErr)
}

func (suite *HTTPServerTestSuite) TestPanic() {
	t := suite.T()
	_ = dic.Register[logger.Logger](mocks.NewNullLogger())
	r := mux.NewRouter()
	r.Path("/").HandlerFunc(internalerrors.HTTPErrorHandler(func(w http.ResponseWriter, req *http.Request) error {
		panic("test")
	}))
	_ = dic.Register[*mux.Router](r)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	testURL, _ := url.JoinPath("http://", httpTestServerListenAddr)
	req, _ := http.NewRequest(http.MethodGet, testURL, nil)
	httpClient := http.Client{}

	var wg sync.WaitGroup
	var serverErr error
	go func() {
		wg.Add(1)
		serverErr = StartServer(ctx, Config{Address: httpTestServerListenAddr})
		wg.Done()
	}()

	// We need to wait for the server to be ready, we do not have a way to sync on that so far
	// Increase this value if the test is flaky
	time.Sleep(3 * time.Second)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	bodyContent, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, ``, string(bodyContent))
	cancelFunc()
	wg.Wait()
	require.NoError(t, serverErr)
}
