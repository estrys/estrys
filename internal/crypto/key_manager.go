package crypto

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"

	_http "github.com/estrys/estrys/internal/http"
	"github.com/estrys/estrys/internal/logger"
)

type KeyManager interface {
	GenerateKey() (*rsa.PrivateKey, error)
	FormatPubKey(crypto.PublicKey) string
	FetchKey(ctx context.Context, id string) (crypto.PublicKey, error)
}

type rsaKeyManager struct {
	log    logger.Logger
	client _http.Client
	cache  *lru.Cache[string, crypto.PublicKey]
}

type keyResponse struct {
	PublicKey struct {
		PublicKeyPEM string `json:"publicKeyPem"` //nolint:tagliatelle
		Owner        string `json:"owner"`
	} `json:"publicKey"` //nolint:tagliatelle
}

func NewKeyManager(log logger.Logger, client _http.Client) *rsaKeyManager {
	cache, _ := lru.New[string, crypto.PublicKey](1000)
	return &rsaKeyManager{
		log:    log,
		client: client,
		cache:  cache,
	}
}

func (k *rsaKeyManager) FetchKey(ctx context.Context, id string) (crypto.PublicKey, error) {
	if key, cacheHit := k.cache.Get(id); cacheHit {
		return key, nil
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, id, nil)
	req.Header.Set("accept", "application/activity+json")
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting user key")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error while reading response body")
	}
	var publicKeyResponse keyResponse
	err = json.Unmarshal(body, &publicKeyResponse)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode user json")
	}
	k.log.WithField("owner", publicKeyResponse.PublicKey.Owner).
		WithField("key", publicKeyResponse.PublicKey.PublicKeyPEM).
		Trace("success fetching key")

	pemBlock, _ := pem.Decode([]byte(publicKeyResponse.PublicKey.PublicKeyPEM))

	if pemBlock == nil {
		return nil, errors.New("unable to decode key PEM")
	}

	k.log.WithField("key", publicKeyResponse.PublicKey.PublicKeyPEM).Debug("fetched key")

	key, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse public key")
	}
	k.cache.Add(id, key)
	return key, nil
}

func (k *rsaKeyManager) GenerateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	return privateKey, errors.Wrap(err, "rsa key generation failed")
}

func (k *rsaKeyManager) FormatPubKey(key crypto.PublicKey) string {
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return ""
	}
	pemData := string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: keyBytes,
			},
		),
	)

	return pemData
}
