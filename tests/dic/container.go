package dic_test

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/estrys/estrys/internal/dic"
	"github.com/estrys/estrys/internal/dic/container"
	"github.com/estrys/estrys/internal/logger"
	"github.com/estrys/estrys/internal/logger/mocks"
)

func BuildTestContainer(t *testing.T) {
	t.Helper()
	require.NoError(t, os.Setenv("TEST", "true"))
	viper.Set("domain", "https://example.com")

	// Following vars are set just to pass the config check but they are not gonna be used
	viper.Set("log_level", "debug")
	// Also make sure those variables are invalid to avoid reaching real backend
	viper.Set("token", "token")
	viper.Set("db_url", "foobar")
	viper.Set("redis_address", "foobar")
	viper.Set("cache_twitter_user_ttl", "10s")

	// Override here services for tests
	require.NoError(t, dic.Register[logger.Logger](mocks.NewNullLogger()))

	require.NoError(t, container.BuildContainer())
}
