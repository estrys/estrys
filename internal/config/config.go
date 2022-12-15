package config

import (
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Address                    string        `mapstructure:"address"`
	Domain                     *url.URL      `mapstructure:"-"`
	Token                      string        `mapstructure:"token"`
	LogLevel                   logrus.Level  `mapstructure:"-"`
	DBURL                      *url.URL      `mapstructure:"-"`
	RedisAddress               string        `mapstructure:"redis_address"`
	TwitterUserCacheTimeout    time.Duration `mapstructure:"-"`
	TwitterTweetCacheTimeout   time.Duration `mapstructure:"-"`
	DisableHTTPSignatureVerify bool          `mapstructure:"disable_http_signature_verify"`
	DisableEmbedWorker         bool          `mapstructure:"disable_embed_worker"`
	AllowedUsers               []string      `mapstructure:"allowed_users"`
	TwitterAllowedUsers        []string      `mapstructure:"twitter_allowed_users"`
	RunMigrations              bool          `mapstructure:"run_migrations"`
	SentryDSN                  string        `mapstructure:"sentry_dsn"`
}

type Loader interface {
	Load() error
	Get() Config
}

type configLoader struct {
	conf Config
}

func NewLoader() *configLoader {
	return &configLoader{}
}

func (l *configLoader) Get() Config {
	return l.conf
}

func (l *configLoader) Load() error {
	conf := &Config{}

	viper.SetEnvPrefix("estrys")
	viper.AutomaticEnv()

	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil && os.Getenv("TEST") != "true" {
		return errors.New("unable to read default config from .env")
	}
	viper.SetConfigFile(".env.local")
	_ = viper.MergeInConfig()

	err = viper.Unmarshal(conf)
	if err != nil {
		return errors.Wrap(err, "unable to deserialize configuration")
	}

	domainStr := viper.GetString("domain")
	if domainStr == "" {
		return errors.New("You must define a domain")
	}
	conf.Domain, err = url.Parse(domainStr)
	if err != nil {
		return errors.Wrap(err, "unable to parse domain")
	}

	logLevel := viper.GetString("log_level")
	if logLevel == "" {
		return errors.New("You must define a log level")
	}
	conf.LogLevel, err = logrus.ParseLevel(logLevel)
	if err != nil {
		return errors.Wrap(err, "unable to parse log level")
	}

	dbURL := viper.GetString("db_url")
	if dbURL == "" {
		return errors.New("You must define a database URL")
	}
	conf.DBURL, err = url.Parse(dbURL)
	if err != nil {
		return errors.Wrap(err, "unable to parse database URL")
	}

	twitterUserCacheDuration := viper.GetString("cache_twitter_user_ttl")
	if twitterUserCacheDuration == "" {
		return errors.New("You must define a twitter user cache duration")
	}
	conf.TwitterUserCacheTimeout, err = time.ParseDuration(twitterUserCacheDuration)
	if err != nil {
		return errors.Wrap(err, "unable to parse twitter cache duration")
	}

	tweetCacheTTL := viper.GetString("cache_tweet_ttl")
	if tweetCacheTTL != "" {
		conf.TwitterTweetCacheTimeout, err = time.ParseDuration(tweetCacheTTL)
		if err != nil {
			return errors.Wrap(err, "unable to parse tweets cache ttl duration")
		}
	}

	if conf.Token == "" {
		return errors.New("you need to configure a token")
	}

	l.conf = *conf
	return nil
}
