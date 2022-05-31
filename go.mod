module github.com/spiral/roadrunner-plugins/v2

go 1.17

require (
	github.com/Shopify/toxiproxy v2.1.4+incompatible
	github.com/aws/aws-sdk-go-v2 v1.12.0
	github.com/aws/aws-sdk-go-v2/config v1.12.0
	github.com/aws/aws-sdk-go-v2/credentials v1.7.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.15.0
	github.com/aws/smithy-go v1.9.1
	github.com/beanstalkd/go-beanstalk v0.1.0
	github.com/bradfitz/gomemcache v0.0.0-20220106215444-fb4bf637b56d
	github.com/caddyserver/certmagic v0.15.2
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/emicklei/proto v1.9.1
	github.com/fatih/color v1.13.0
	github.com/go-redis/redis/v8 v8.11.4
	github.com/gobwas/ws v1.1.0
	github.com/gofiber/fiber/v2 v2.24.0
	github.com/google/uuid v1.3.0
	github.com/klauspost/compress v1.15.4
	github.com/mholt/acmez v1.0.1
	github.com/nats-io/nats.go v1.13.0
	github.com/newrelic/go-agent/v3 v3.15.2
	github.com/prometheus/client_golang v1.12.2
	github.com/rabbitmq/amqp091-go v1.3.0
	github.com/roadrunner-server/api/v2 v2.17.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/spf13/viper v1.10.1
	// spiral
	github.com/spiral/endure v1.1.0
	github.com/spiral/errors v1.0.12
	github.com/spiral/goridge/v3 v3.2.7
	github.com/spiral/roadrunner/v2 v2.7.1
	// spiral
	github.com/stretchr/testify v1.7.1
	github.com/yookoala/gofast v0.6.0
	go.etcd.io/bbolt v1.3.6
	go.uber.org/zap v1.21.0
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/sys v0.0.0-20220513210249-45d2b4557a2a
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.6.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.13.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/go-version v1.4.0
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miekg/dns v1.1.45 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nats-server/v2 v2.6.1 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spiral/tcplisten v1.0.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.36.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20220112215332-a9c7c0acf9f2 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

require (
	github.com/goccy/go-json v0.9.7 // indirect
	github.com/roadrunner-server/goridge/v3 v3.4.1 // indirect
	go.buf.build/protocolbuffers/go/roadrunner-server/api v1.2.3 // indirect
)
