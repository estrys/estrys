package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-fed/httpsig"

	"github.com/estrys/estrys/internal/config"
	"github.com/estrys/estrys/internal/crypto"
	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/logger"
)

type contextKey int

const (
	HTTPSignature contextKey = iota
)

func IsRequestSigned(req *http.Request) bool {
	if isSigned, isBool := req.Context().Value(HTTPSignature).(*bool); isBool && isSigned != nil && *isSigned {
		return true
	}
	return false
}

func setSentryContext(request *http.Request, context map[string]any) {
	sentry.GetHubFromContext(request.Context()).ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext("httpsig", context)
	})
}

func HTTPSigMiddleware(next http.Handler) http.Handler {
	log := dic.GetService[logger.Logger]()
	conf := dic.GetService[config.Config]()
	keyManager := dic.GetService[crypto.KeyManager]()

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var signed bool
		sentryContext := map[string]any{
			"signed": &signed,
		}
		signedContext := context.WithValue(request.Context(), HTTPSignature, &signed)
		request = request.WithContext(signedContext)
		if conf.DisableHTTPSignatureVerify {
			signed = true
		}

		sig := request.Header.Get("signature")
		sentryContext["signature"] = sig
		if sig == "" {
			setSentryContext(request, sentryContext)
			next.ServeHTTP(responseWriter, request)
			return
		}

		verifier, err := httpsig.NewVerifier(request)
		if err != nil {
			sentryContext["error"] = err.Error()
			log.Error(err)
			setSentryContext(request, sentryContext)
			next.ServeHTTP(responseWriter, request)
			return
		}

		sentryContext["key_id"] = verifier.KeyId()
		log.WithField("keyId", verifier.KeyId()).Trace("request signature check starting")
		pubKey, err := keyManager.FetchKey(signedContext, verifier.KeyId())
		if err != nil {
			sentryContext["error"] = err.Error()
			setSentryContext(request, sentryContext)
			log.WithError(err).Error("unable to fetch keyManager")
			next.ServeHTTP(responseWriter, request)
			return
		}

		// Algorithm should not be retrieved from the signature header
		// https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-10#section-2.5
		var algo httpsig.Algorithm
		if _, isEcDsa := pubKey.(*ecdsa.PublicKey); isEcDsa {
			algo = httpsig.ECDSA_SHA256
		}
		if _, isEd := pubKey.(*ed25519.PublicKey); isEd {
			algo = httpsig.ED25519
		}
		if _, isRsa := pubKey.(*rsa.PublicKey); isRsa {
			algo = httpsig.RSA_SHA256
		}

		sentryContext["algo"] = algo
		err = verifier.Verify(pubKey, algo)

		if err != nil {
			sentryContext["error"] = err.Error()
			setSentryContext(request, sentryContext)
			log.WithField("algo", algo).WithError(err).Warn("unable to verify http signature")
			next.ServeHTTP(responseWriter, request)
			return
		}

		log.WithField("keyId", verifier.KeyId()).Debug("http signature valid")
		signed = true

		setSentryContext(request, sentryContext)
		next.ServeHTTP(responseWriter, request)
	})
}
