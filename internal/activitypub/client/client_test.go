package activitypubclient_test

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/httpsig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	activitypubclient "github.com/estrys/estrys/internal/activitypub/client"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger/mocks"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/router/urlgenerator"
	dic_test "github.com/estrys/estrys/tests/dic"
)

var (
	fakePrivateKey, _    = os.ReadFile(path.Join("testdata", "key.pem"))
	decodedPrivKeyPem, _ = pem.Decode(fakePrivateKey)
	pubKey, _            = x509.ParsePKCS1PrivateKey(decodedPrivKeyPem.Bytes)
)

func Test_activityPubClient_PostInbox(t *testing.T) {

	var fakeServerURL string

	tests := []struct {
		name        string
		to          func() *models.Actor
		from        *models.User
		act         pub.Activity
		handlerFunc func(*testing.T) http.HandlerFunc
		err         string
	}{
		{
			name: "ok",
			to: func() *models.Actor {
				return &models.Actor{
					URL: fakeServerURL + "/users/fake-actor",
				}
			},
			from: &models.User{
				Username:   "fake-username",
				PrivateKey: decodedPrivKeyPem.Bytes,
			},
			act: streams.NewActivityStreamsAccept(),
			handlerFunc: func(t *testing.T) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					// Check that request sounds ok
					assert.Equal(t, "/users/fake-actor/inbox", request.URL.String())
					assert.Equal(t, http.MethodPost, request.Method)
					assert.Equal(t, "application/activity+json", request.Header.Get("content-type"))

					// Check that response signature is valid
					verifier, err := httpsig.NewVerifier(request)
					assert.NoError(t, err)
					assert.NoError(t, verifier.Verify(pubKey.Public(), httpsig.RSA_SHA256))

					// Check the response body
					reqBody, err := io.ReadAll(request.Body)
					reqBodyStruct := map[string]interface{}{}
					assert.NoError(t, json.Unmarshal(reqBody, &reqBodyStruct))
					assert.NoError(t, err)
					expectedBody, err := os.ReadFile(path.Join("testdata/expected_requests/follow_inbox.json"))
					expectedBodyStruct := map[string]interface{}{}
					assert.NoError(t, json.Unmarshal(expectedBody, &expectedBodyStruct))
					assert.Equal(t, expectedBodyStruct, reqBodyStruct)

					writer.WriteHeader(http.StatusAccepted)
				}
			},
		},
		{
			name: "401",
			to: func() *models.Actor {
				return &models.Actor{
					URL: fakeServerURL + "/users/fake-actor",
				}
			},
			from: &models.User{
				Username:   "fake-username",
				PrivateKey: decodedPrivKeyPem.Bytes,
			},
			act: streams.NewActivityStreamsAccept(),
			handlerFunc: func(t *testing.T) http.HandlerFunc {
				return func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusUnauthorized)
				}
			},
			err: "error posting to inbox 401",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fakeServer := httptest.NewServer(tt.handlerFunc(t))
			fakeServerURL = fakeServer.URL
			defer fakeServer.Close()

			dic_test.BuildTestContainer(t)
			defer dic.ResetContainer()

			c, err := activitypubclient.NewActivityPubClient(
				fakeServer.Client(),
				mocks.NewNullLogger(),
				dic.GetService[urlgenerator.URLGenerator](),
			)
			require.NoError(t, err)

			err = c.PostInbox(
				context.TODO(),
				tt.to(),
				tt.from,
				tt.act,
			)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
