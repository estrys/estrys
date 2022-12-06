package crypto

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"testing/iotest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httpmock "github.com/estrys/estrys/internal/http/mocks"
	loggermock "github.com/estrys/estrys/internal/logger/mocks"
)

func Test_rsaKeyManager_FetchKey(t *testing.T) {
	fakeLogger := loggermock.NewNullLogger()
	fakeUrl := "https://example.com/key"

	tests := []struct {
		name string
		id   string
		want crypto.PublicKey
		mock func(client *httpmock.Client)
		err  error
	}{
		{
			name: "rsa",
			id:   fakeUrl,
			want: func() crypto.PublicKey {
				pemBlock, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtQKn/lAs285puIPRoWQv\n0lQpA4wzoqWt2YRcXy7O3qbllb4dkX7XmG6nJluuWOpkS5E4cajLzvrtRq1MFzOW\ndWgPmIbO4uf4S8ByhvLMR3ytvp+iEynckI9XLva9ObnUXxV7ovLD94hlES400lhS\n46DY/Dt26mrEkHiZqoV5JfTKsS1Pa8MhHZ8NuBXholL75cf8UdWgjHqBZj+jcjht\nCDTLnU2N0i1MjowTcOXdcNocC4iZLUGPCjHZHRQoP/CD+JnI8sHVQ1Iw8OH0Kgcy\ne2wv8hAiAcPIqSFl76KOH6VJpBbIsG6azyaFz5/qu15MyJGqcaZv1Ct52cQdBIk8\nwQIDAQAB\n-----END PUBLIC KEY-----\n"))
				expectedPublicKey, _ := x509.ParsePKIXPublicKey(pemBlock.Bytes)
				return expectedPublicKey
			}(),
		},
		{
			name: "ed25519",
			id:   fakeUrl,
			want: func() crypto.PublicKey {
				pemBlock, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMCowBQYDK2VwAyEAIVYwccn41LcXRnrrX9+mfIHgg7XviPGon6WUhbmN32M=\n-----END PUBLIC KEY-----\n"))
				expectedPublicKey, _ := x509.ParsePKIXPublicKey(pemBlock.Bytes)
				return expectedPublicKey
			}(),
		},
		{
			name: "no_key",
			id:   fakeUrl,
			err:  errors.New("unable to decode key PEM"),
		},
		{
			name: "not_json",
			id:   fakeUrl,
			err:  errors.New("unable to decode user json"),
		},
		{
			name: "invalid_private_key",
			id:   fakeUrl,
			err:  errors.New("unable to parse public key"),
		},
		{
			name: "error during http call",
			id:   fakeUrl,
			mock: func(client *httpmock.Client) {
				client.On("Do", mock.Anything).Once().Return(nil, errors.New("fatal error"))
			},
			err: errors.New("fatal error"),
		},
		{
			name: "error reading response",
			id:   fakeUrl,
			mock: func(client *httpmock.Client) {
				client.On("Do", mock.Anything).Once().Return(
					&http.Response{
						Body: io.NopCloser(iotest.ErrReader(errors.New("EOF"))),
					}, nil)
			},
			err: errors.New("error while reading response body: EOF"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			httpClient := httpmock.NewClient(t)

			if tt.mock == nil {
				httpClient.On("Do", mock.MatchedBy(func(r *http.Request) bool {
					expectedCalledURL, _ := url.Parse(fakeUrl)
					return r.Method == http.MethodGet && expectedCalledURL.String() == r.URL.String()
				})).Once().Return(func() (*http.Response, error) {
					data, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", tt.name))
					if err != nil {
						return nil, err
					}
					resp := http.Response{
						Body: io.NopCloser(bytes.NewReader(data)),
					}
					return &resp, nil
				}())
			} else {
				tt.mock(httpClient)
			}

			k := NewKeyManager(
				fakeLogger,
				httpClient,
			)

			got, err := k.FetchKey(context.Background(), tt.id)
			if tt.err == nil {
				require.Nil(t, err)
			} else {
				require.ErrorContains(t, err, tt.err.Error())
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_rsaKeyManager_FormatPubKey(t *testing.T) {
	keyManager := NewKeyManager(
		nil,
		nil,
	)

	keyBytes, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+ytTPYSJGWBWLiSL8oUoHbFks\nZhMneshS06URryUlmTg3TuS/qO27yCMe8NOWTqweU4KV0hTxbFkd8X/64brZUzWN\nTIeMDOoKnCd0OqK+MkrKK2fSdl+5t5goWZx/IYnopQOcfkEpxADmfqElS+67vdvc\nqDFSCzpz1cTaE5CwJwIDAQAB\n-----END PUBLIC KEY-----\n"))
	require.NotNil(t, keyBytes)
	pubKey, err := x509.ParsePKIXPublicKey(keyBytes.Bytes)
	require.NoError(t, err)
	keyStr := keyManager.FormatPubKey(pubKey)
	require.Equal(
		t,
		`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+ytTPYSJGWBWLiSL8oUoHbFks
ZhMneshS06URryUlmTg3TuS/qO27yCMe8NOWTqweU4KV0hTxbFkd8X/64brZUzWN
TIeMDOoKnCd0OqK+MkrKK2fSdl+5t5goWZx/IYnopQOcfkEpxADmfqElS+67vdvc
qDFSCzpz1cTaE5CwJwIDAQAB
-----END PUBLIC KEY-----
`,
		keyStr,
	)
}
