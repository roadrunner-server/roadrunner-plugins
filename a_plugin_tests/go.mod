module github.com/spiral/roadrunner-plugins/a_plugin_tests

go 1.15

require (
	github.com/alicebob/miniredis/v2 v2.14.1
	github.com/go-redis/redis/v8 v8.4.4
	github.com/golang/mock v1.4.4
	github.com/json-iterator/go v1.1.10
	github.com/prometheus/client_golang v1.9.0
	github.com/spiral/endure v1.0.0-beta20
	github.com/spiral/errors v1.0.6
	github.com/spiral/goridge/v3 v3.0.0-beta8
	github.com/spiral/roadrunner-plugins/checker v0.0.3
	github.com/spiral/roadrunner-plugins/config v0.0.3
	github.com/spiral/roadrunner-plugins/gzip v0.0.3
	github.com/spiral/roadrunner-plugins/headers v0.0.2
	github.com/spiral/roadrunner-plugins/http v0.0.3
	github.com/spiral/roadrunner-plugins/informer v0.0.2
	github.com/spiral/roadrunner-plugins/logger v0.0.3
	github.com/spiral/roadrunner-plugins/metrics v0.0.3
	github.com/spiral/roadrunner-plugins/redis v0.0.0-00010101000000-000000000000
	github.com/spiral/roadrunner-plugins/reload v0.0.0-00010101000000-000000000000
	github.com/spiral/roadrunner-plugins/resetter v0.0.2
	github.com/spiral/roadrunner-plugins/rpc v0.0.2
	github.com/spiral/roadrunner-plugins/server v0.0.2
	github.com/spiral/roadrunner/v2 v2.0.0-alpha26
	github.com/stretchr/testify v1.6.1
	github.com/yookoala/gofast v0.4.0
)

// ONLY FOR TESTS
replace (
	github.com/spiral/roadrunner-plugins/checker => ../checker
	github.com/spiral/roadrunner-plugins/config => ../config
	github.com/spiral/roadrunner-plugins/gzip => ../gzip
	github.com/spiral/roadrunner-plugins/headers => ../headers
	github.com/spiral/roadrunner-plugins/http => ../http
	github.com/spiral/roadrunner-plugins/informer => ../informer
	github.com/spiral/roadrunner-plugins/logger => ../logger
	github.com/spiral/roadrunner-plugins/metrics => ../metrics
	github.com/spiral/roadrunner-plugins/redis => ../redis
	github.com/spiral/roadrunner-plugins/reload => ../reload
	github.com/spiral/roadrunner-plugins/resetter => ../resetter
	github.com/spiral/roadrunner-plugins/rpc => ../rpc
	github.com/spiral/roadrunner-plugins/server => ../server
)
