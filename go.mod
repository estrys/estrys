module github.com/estrys/estrys

go 1.19

// Remove once https://github.com/go-fed/activity/pull/160 is nerged qnd update go-fed/activity
replace github.com/go-fed/activity => github.com/birdlephant/activity v0.0.0-20221204152203-733d1fd88157

require (
	github.com/friendsofgo/errors v0.9.2
	github.com/g8rswimmer/go-twitter/v2 v2.1.4
	github.com/getsentry/sentry-go v0.16.0
	github.com/go-fed/activity v1.0.1-0.20220119073622-b14b50eecad0
	github.com/go-fed/httpsig v1.1.0
	github.com/go-redis/redis/v9 v9.0.0-rc.2
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/golang-lru/v2 v2.0.1
	github.com/hibiken/asynq v0.23.0
	github.com/jackc/pgx/v5 v5.1.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.1
	github.com/vmihailenco/msgpack/v5 v5.3.5
	github.com/volatiletech/sqlboiler/v4 v4.13.0
	github.com/volatiletech/strmangle v0.0.4
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/lib/pq v1.10.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/volatiletech/inflect v0.0.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20220926161630-eccd6366d1be // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.2.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
