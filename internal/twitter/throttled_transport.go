package twitter

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/estrys/estrys/internal/logger"
)

type ThrottledTransport struct {
	log            logger.Logger
	wrap           http.RoundTripper
	waitUntilReset map[string]time.Duration
	ratelimiters   map[string]*rate.Limiter
}

func (c *ThrottledTransport) getRequestKey(request *http.Request) string {
	switch {
	case strings.HasPrefix(request.URL.Path, `/2/users/by/username`):
		return "user_by_username"
	case regexp.MustCompile(`^/2/users/\d+/tweets$`).MatchString(request.URL.Path):
		return "user_tweets"
	case strings.HasPrefix(request.URL.Path, `/2/users`):
		return "user_lookup"
	case strings.HasPrefix(request.URL.Path, `/2/tweets`):
		return "tweets_lookup"
	}
	return ""
}

func (c *ThrottledTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	requestKey := c.getRequestKey(request)
	if requestKey == "" {
		// That mean that the route is not supported by the rate limiter
		return c.wrap.RoundTrip(request)
	}
	if durationToWait, needToWait := c.waitUntilReset[requestKey]; needToWait {
		c.log.WithField("seconds_to_wait", durationToWait.Seconds()).
			Trace("No calls remaining, waiting to the next reset")
		time.Sleep(durationToWait)
	}
	if c.ratelimiters[requestKey] != nil {
		err := c.ratelimiters[requestKey].Wait(request.Context())
		if err != nil {
			return nil, err
		}
	}
	resp, err := c.wrap.RoundTrip(request)
	if resp != nil {
		resetTimestampHeader := resp.Header.Get("x-rate-limit-reset")
		resetTimestamp, timestampErr := strconv.ParseInt(resetTimestampHeader, 10, 64)
		remainingCallsHeader := resp.Header.Get("x-rate-limit-remaining")
		remainingCalls, remainingCallsErr := strconv.ParseFloat(remainingCallsHeader, 64)
		if timestampErr == nil && remainingCallsErr == nil {
			resetDate := time.Unix(resetTimestamp, 0)
			durationUntilReset := time.Until(resetDate)
			if remainingCalls == 0 {
				c.waitUntilReset[requestKey] = durationUntilReset
				return resp, err
			}
			delete(c.waitUntilReset, requestKey)
			newRate := remainingCalls / durationUntilReset.Seconds()
			c.log.WithFields(logrus.Fields{
				"rate":                newRate,
				"request_key":         requestKey,
				"remaining_calls":     remainingCalls,
				"seconds_until_reset": durationUntilReset.Seconds(),
			}).Trace("updated new rate")
			c.ratelimiters[requestKey] = rate.NewLimiter(rate.Limit(newRate), 1)
		}
	}
	return resp, err
}

func NewThrottledTransport(wrap http.RoundTripper, log logger.Logger) *ThrottledTransport {
	return &ThrottledTransport{
		log:            log,
		wrap:           wrap,
		ratelimiters:   make(map[string]*rate.Limiter),
		waitUntilReset: make(map[string]time.Duration),
	}
}
