package activitypubclient

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/httpsig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/models"
	"github.com/estrys/estrys/internal/router/routes"
	"github.com/estrys/estrys/internal/router/urlgenerator"
)

type contextKey int

const (
	contextActivity contextKey = iota
)

type InboxNotAcceptedError struct {
	statusCode int
}

func (e *InboxNotAcceptedError) Error() string {
	return fmt.Sprintf("%d", e.statusCode)
}

type activityContext struct {
	user *models.User
}

type httpSigRoundTripper struct {
	mu           *sync.Mutex
	signer       httpsig.Signer
	urlGenerator urlgenerator.URLGenerator
}

func (t *httpSigRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	activityContext, ok := request.Context().Value(contextActivity).(activityContext)
	if !ok {
		return nil, errors.New("no context set for the request")
	}

	userPrivateKey, err := x509.ParsePKCS1PrivateKey(activityContext.user.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse user private key")
	}
	pubKeyID, err := t.urlGenerator.URL(
		routes.UserRoute,
		[]string{"username", activityContext.user.Username},
		urlgenerator.OptionAbsoluteURL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate user route")
	}

	buf, _ := io.ReadAll(request.Body)
	rdr2 := io.NopCloser(bytes.NewBuffer(buf))
	t.mu.Lock()
	err = t.signer.SignRequest(userPrivateKey, pubKeyID.String(), request, buf)
	t.mu.Unlock()
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign http request")
	}
	request.Body = rdr2

	return http.DefaultTransport.RoundTrip(request) //nolint:wrapcheck
}

type ActivityPubClient interface {
	PostInbox(ctx context.Context, to *models.Actor, from *models.User, act pub.Activity) error
}

type activityPubClient struct {
	client *http.Client
	log    logger.Logger
}

func NewClient(
	client *http.Client,
	log logger.Logger,
	urlGenerator urlgenerator.URLGenerator,
) (*activityPubClient, error) {
	signer, _, err := httpsig.NewSigner(
		[]httpsig.Algorithm{httpsig.RSA_SHA256},
		httpsig.DigestSha256,
		[]string{httpsig.RequestTarget, "date", "host", "digest"},
		httpsig.Signature,
		30, // seconds
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create an http request signer")
	}
	client.Transport = &httpSigRoundTripper{
		signer:       signer,
		urlGenerator: urlGenerator,
		mu:           &sync.Mutex{},
	}
	return &activityPubClient{
		client: client,
		log:    log,
	}, nil
}

func (c *activityPubClient) PostInbox(
	ctx context.Context,
	to *models.Actor,
	from *models.User,
	act pub.Activity,
) error {
	actorURL, err := url.Parse(to.URL)
	if err != nil {
		return errors.Wrap(err, "unable to decode actor url")
	}

	serializedActivity, err := streams.Serialize(act)
	if err != nil {
		return errors.Wrap(err, "unable to serialize activity")
	}
	bodyBuffer := new(bytes.Buffer)
	err = json.NewEncoder(bodyBuffer).Encode(serializedActivity)
	if err != nil {
		return errors.Wrap(err, "unable to encode serialized activity")
	}
	requestContext := context.WithValue(ctx, contextActivity, activityContext{from})
	inboxPath, _ := url.JoinPath(actorURL.String(), "/inbox")
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost, inboxPath, bodyBuffer)
	if err != nil {
		return errors.Wrap(err, "unable to create inbox request")
	}

	request.Header.Add("date", time.Now().Format(http.TimeFormat))
	request.Header.Add("host", actorURL.Host)
	request.Header.Add("content-type", "application/activity+json")

	response, err := c.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "unable to perform post to inbox request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(response.Body)
		logrus.WithField("response", string(respBody)).Trace("unable to post to inbox")
		return &InboxNotAcceptedError{statusCode: response.StatusCode}
	}

	return nil
}
