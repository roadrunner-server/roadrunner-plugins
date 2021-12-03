package v2_6 //nolint:stylecheck

import (
	"time"
)

type Env map[string]string

type Config struct {
	RPC struct {
		Listen string `yaml:"listen"`
	} `yaml:"rpc"`
	Server struct {
		OnInit struct {
			Command     string `yaml:"command"`
			ExecTimeout string `yaml:"exec_timeout"`
			Env         Env    `yaml:"env"`
		} `yaml:"on_init"`
		Command      string `yaml:"command"`
		User         string `yaml:"user"`
		Group        string `yaml:"group"`
		Env          Env    `yaml:"env"`
		Relay        string `yaml:"relay"`
		RelayTimeout string `yaml:"relay_timeout"`
	} `yaml:"server"`
	Logs struct {
		Mode       string `yaml:"mode"`
		Level      string `yaml:"level"`
		Encoding   string `yaml:"encoding"`
		LineEnding string `yaml:"line_ending"`
		Output     string `yaml:"output"`
		ErrOutput  string `yaml:"err_output"`
		Channels   struct {
			HTTP struct {
				Mode      string `yaml:"mode"`
				Level     string `yaml:"level"`
				Encoding  string `yaml:"encoding"`
				Output    string `yaml:"output"`
				ErrOutput string `yaml:"err_output"`
			} `yaml:"http"`
			Server struct {
				Mode      string `yaml:"mode"`
				Level     string `yaml:"level"`
				Encoding  string `yaml:"encoding"`
				Output    string `yaml:"output"`
				ErrOutput string `yaml:"err_output"`
			} `yaml:"server"`
			RPC struct {
				Mode      string `yaml:"mode"`
				Level     string `yaml:"level"`
				Encoding  string `yaml:"encoding"`
				Output    string `yaml:"output"`
				ErrOutput string `yaml:"err_output"`
			} `yaml:"rpc"`
		} `yaml:"channels"`
	} `yaml:"logs"`
	Temporal struct {
		Address   string `yaml:"address"`
		CacheSize int    `yaml:"cache_size"`
		Namespace string `yaml:"namespace"`
		Metrics   struct {
			Address string `yaml:"address"`
			Type    string `yaml:"type"`
			Prefix  string `yaml:"prefix"`
		} `yaml:"metrics"`
		Activities struct {
			Debug           bool   `yaml:"debug"`
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  string `yaml:"destroy_timeout"`
			Supervisor      struct {
				WatchTick       string `yaml:"watch_tick"`
				TTL             string `yaml:"ttl"`
				IdleTTL         string `yaml:"idle_ttl"`
				MaxWorkerMemory int    `yaml:"max_worker_memory"`
				ExecTTL         string `yaml:"exec_ttl"`
			} `yaml:"supervisor"`
		} `yaml:"activities"`
		Codec      string `yaml:"codec"`
		DebugLevel int    `yaml:"debug_level"`
	} `yaml:"temporal"`
	Kv struct {
		BoltdbSouth struct {
			Driver string `yaml:"driver"`
			Config struct {
				File        string `yaml:"file"`
				Permissions int    `yaml:"permissions"`
				Interval    int    `yaml:"interval"`
			} `yaml:"config"`
		} `yaml:"boltdb-south"`
		UsCentralKv struct {
			Driver string `yaml:"driver"`
			Config struct {
				Addr []string `yaml:"addr"`
			} `yaml:"config"`
		} `yaml:"us-central-kv"`
		FastKvFr struct {
			Driver string `yaml:"driver"`
			Config struct {
				Addrs            []string `yaml:"addrs"`
				MasterName       string   `yaml:"master_name"`
				Username         string   `yaml:"username"`
				Password         string   `yaml:"password"`
				DB               int      `yaml:"db"`
				SentinelPassword string   `yaml:"sentinel_password"`
				RouteByLatency   bool     `yaml:"route_by_latency"`
				RouteRandomly    bool     `yaml:"route_randomly"`
				DialTimeout      int      `yaml:"dial_timeout"`
				MaxRetries       int      `yaml:"max_retries"`
				MinRetryBackoff  int      `yaml:"min_retry_backoff"`
				MaxRetryBackoff  int      `yaml:"max_retry_backoff"`
				PoolSize         int      `yaml:"pool_size"`
				MinIdleConns     int      `yaml:"min_idle_conns"`
				MaxConnAge       int      `yaml:"max_conn_age"`
				ReadTimeout      int      `yaml:"read_timeout"`
				WriteTimeout     int      `yaml:"write_timeout"`
				PoolTimeout      int      `yaml:"pool_timeout"`
				IdleTimeout      int      `yaml:"idle_timeout"`
				IdleCheckFreq    int      `yaml:"idle_check_freq"`
				ReadOnly         bool     `yaml:"read_only"`
			} `yaml:"config"`
		} `yaml:"fast-kv-fr"`
		LocalMemory struct {
			Driver string `yaml:"driver"`
			Config struct {
				Interval int `yaml:"interval"`
			} `yaml:"config"`
		} `yaml:"local-memory"`
	} `yaml:"kv"`

	Services map[string]Service `mapstructure:"service"`

	HTTP struct {
		Address           string   `yaml:"address"`
		InternalErrorCode int      `yaml:"internal_error_code"`
		AccessLogs        bool     `yaml:"access_logs"`
		MaxRequestSize    int      `yaml:"max_request_size"`
		Middleware        []string `yaml:"middleware"`
		TrustedSubnets    []string `yaml:"trusted_subnets"`
		NewRelic          struct {
			AppName    string `yaml:"app_name"`
			LicenseKey string `yaml:"license_key"`
		} `yaml:"new_relic"`
		Uploads struct {
			Dir    string   `yaml:"dir"`
			Forbid []string `yaml:"forbid"`
			Allow  []string `yaml:"allow"`
		} `yaml:"uploads"`
		Headers struct {
			Cors struct {
				AllowedOrigin    string `yaml:"allowed_origin"`
				AllowedHeaders   string `yaml:"allowed_headers"`
				AllowedMethods   string `yaml:"allowed_methods"`
				AllowCredentials bool   `yaml:"allow_credentials"`
				ExposedHeaders   string `yaml:"exposed_headers"`
				MaxAge           int    `yaml:"max_age"`
			} `yaml:"cors"`
			Request struct {
				Input string `yaml:"input"`
			} `yaml:"request"`
			Response struct {
				XPoweredBy string `yaml:"X-Powered-By"`
			} `yaml:"response"`
		} `yaml:"headers"`
		Static struct {
			Dir           string   `yaml:"dir"`
			Forbid        []string `yaml:"forbid"`
			CalculateEtag bool     `yaml:"calculate_etag"`
			Weak          bool     `yaml:"weak"`
			Allow         []string `yaml:"allow"`
			Request       struct {
				Input string `yaml:"input"`
			} `yaml:"request"`
			Response struct {
				Output string `yaml:"output"`
			} `yaml:"response"`
		} `yaml:"static"`
		Pool struct {
			Debug           bool   `yaml:"debug"`
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  string `yaml:"destroy_timeout"`
			Supervisor      struct {
				WatchTick       string `yaml:"watch_tick"`
				TTL             string `yaml:"ttl"`
				IdleTTL         string `yaml:"idle_ttl"`
				MaxWorkerMemory int    `yaml:"max_worker_memory"`
				ExecTTL         string `yaml:"exec_ttl"`
			} `yaml:"supervisor"`
		} `yaml:"pool"`
		Ssl struct {
			Address string `yaml:"address"`
			Acme    struct {
				CertsDir              string   `yaml:"certs_dir"`
				Email                 string   `yaml:"email"`
				AltHTTPPort           int      `yaml:"alt_http_port"`
				AltTlsalpnPort        int      `yaml:"alt_tlsalpn_port"`
				ChallengeType         string   `yaml:"challenge_type"`
				UseProductionEndpoint bool     `yaml:"use_production_endpoint"`
				Domains               []string `yaml:"domains"`
			} `yaml:"acme"`
			Redirect bool   `yaml:"redirect"`
			Cert     string `yaml:"cert"`
			Key      string `yaml:"key"`
			RootCa   string `yaml:"root_ca"`
		} `yaml:"ssl"`
		Fcgi struct {
			Address string `yaml:"address"`
		} `yaml:"fcgi"`
		HTTP2 struct {
			H2C                  bool `yaml:"h2c"`
			MaxConcurrentStreams int  `yaml:"max_concurrent_streams"`
		} `yaml:"http2"`
	} `yaml:"http"`
	Redis struct {
		Addrs            []string `yaml:"addrs"`
		MasterName       string   `yaml:"master_name"`
		Username         string   `yaml:"username"`
		Password         string   `yaml:"password"`
		DB               int      `yaml:"db"`
		SentinelPassword string   `yaml:"sentinel_password"`
		RouteByLatency   bool     `yaml:"route_by_latency"`
		RouteRandomly    bool     `yaml:"route_randomly"`
		DialTimeout      int      `yaml:"dial_timeout"`
		MaxRetries       int      `yaml:"max_retries"`
		MinRetryBackoff  int      `yaml:"min_retry_backoff"`
		MaxRetryBackoff  int      `yaml:"max_retry_backoff"`
		PoolSize         int      `yaml:"pool_size"`
		MinIdleConns     int      `yaml:"min_idle_conns"`
		MaxConnAge       int      `yaml:"max_conn_age"`
		ReadTimeout      int      `yaml:"read_timeout"`
		WriteTimeout     int      `yaml:"write_timeout"`
		PoolTimeout      int      `yaml:"pool_timeout"`
		IdleTimeout      int      `yaml:"idle_timeout"`
		IdleCheckFreq    int      `yaml:"idle_check_freq"`
		ReadOnly         bool     `yaml:"read_only"`
	} `yaml:"redis"`
	Websockets struct {
		Broker        string `yaml:"broker"`
		AllowedOrigin string `yaml:"allowed_origin"`
		Path          string `yaml:"path"`
	} `yaml:"websockets"`
	Broadcast struct {
		Default struct {
			Driver string `yaml:"driver"`
			Config struct {
			} `yaml:"config"`
		} `yaml:"default"`
		DefaultRedis struct {
			Driver string `yaml:"driver"`
			Config struct {
				Addrs            []string `yaml:"addrs"`
				MasterName       string   `yaml:"master_name"`
				Username         string   `yaml:"username"`
				Password         string   `yaml:"password"`
				DB               int      `yaml:"db"`
				SentinelPassword string   `yaml:"sentinel_password"`
				RouteByLatency   bool     `yaml:"route_by_latency"`
				RouteRandomly    bool     `yaml:"route_randomly"`
				DialTimeout      int      `yaml:"dial_timeout"`
				MaxRetries       int      `yaml:"max_retries"`
				MinRetryBackoff  int      `yaml:"min_retry_backoff"`
				MaxRetryBackoff  int      `yaml:"max_retry_backoff"`
				PoolSize         int      `yaml:"pool_size"`
				MinIdleConns     int      `yaml:"min_idle_conns"`
				MaxConnAge       int      `yaml:"max_conn_age"`
				ReadTimeout      int      `yaml:"read_timeout"`
				WriteTimeout     int      `yaml:"write_timeout"`
				PoolTimeout      int      `yaml:"pool_timeout"`
				IdleTimeout      int      `yaml:"idle_timeout"`
				IdleCheckFreq    int      `yaml:"idle_check_freq"`
				ReadOnly         bool     `yaml:"read_only"`
			} `yaml:"config"`
		} `yaml:"default-redis"`
	} `yaml:"broadcast"`
	Metrics struct {
		Address string `yaml:"address"`
		Collect struct {
			AppMetric struct {
				Type       string    `yaml:"type"`
				Help       string    `yaml:"help"`
				Labels     []string  `yaml:"labels"`
				Buckets    []float64 `yaml:"buckets"`
				Objectives []struct {
					Num2 float64 `yaml:"2,omitempty"`
					One4 float64 `yaml:"1.4,omitempty"`
				} `yaml:"objectives"`
			} `yaml:"app_metric"`
		} `yaml:"collect"`
	} `yaml:"metrics"`
	Status struct {
		Address               string `yaml:"address"`
		UnavailableStatusCode int    `yaml:"unavailable_status_code"`
	} `yaml:"status"`
	Reload struct {
		Interval string   `yaml:"interval"`
		Patterns []string `yaml:"patterns"`
		Services struct {
			HTTP struct {
				Dirs      []string `yaml:"dirs"`
				Recursive bool     `yaml:"recursive"`
				Ignore    []string `yaml:"ignore"`
				Patterns  []string `yaml:"patterns"`
			} `yaml:"http"`
		} `yaml:"services"`
	} `yaml:"reload"`
	Nats struct {
		Addr string `yaml:"addr"`
	} `yaml:"nats"`
	Boltdb struct {
		Permissions int `yaml:"permissions"`
	} `yaml:"boltdb"`
	AMQP struct {
		Addr string `yaml:"addr"`
	} `yaml:"amqp"`
	Beanstalk struct {
		Addr    string `yaml:"addr"`
		Timeout string `yaml:"timeout"`
	} `yaml:"beanstalk"`
	Sqs struct {
		Key          string `yaml:"key"`
		Secret       string `yaml:"secret"`
		Region       string `yaml:"region"`
		SessionToken string `yaml:"session_token"`
		Endpoint     string `yaml:"endpoint"`
	} `yaml:"sqs"`
	Jobs struct {
		NumPollers   int `yaml:"num_pollers"`
		PipelineSize int `yaml:"pipeline_size"`
		Pool         struct {
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  string `yaml:"destroy_timeout"`
		} `yaml:"pool"`
		Pipelines map[string]interface{} `yaml:"pipelines"`
		Consume   []string               `yaml:"consume"`
	} `yaml:"jobs"`
	Grpc struct {
		Listen string   `yaml:"listen"`
		Proto  []string `yaml:"proto"`
		TLS    struct {
			Key            string `yaml:"key"`
			Cert           string `yaml:"cert"`
			RootCa         string `yaml:"root_ca"`
			ClientAuthType string `yaml:"client_auth_type"`
		} `yaml:"tls"`
		MaxSendMsgSize        int    `yaml:"max_send_msg_size"`
		MaxRecvMsgSize        int    `yaml:"max_recv_msg_size"`
		MaxConnectionIdle     string `yaml:"max_connection_idle"`
		MaxConnectionAge      string `yaml:"max_connection_age"`
		MaxConnectionAgeGrace string `yaml:"max_connection_age_grace"`
		MaxConcurrentStreams  int    `yaml:"max_concurrent_streams"`
		PingTime              string `yaml:"ping_time"`
		Timeout               string `yaml:"timeout"`
		Pool                  struct {
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  int    `yaml:"destroy_timeout"`
		} `yaml:"pool"`
	} `yaml:"grpc"`
	TCP struct {
		Servers struct {
			Server1 struct {
				Addr        string `yaml:"addr"`
				Delimiter   string `yaml:"delimiter"`
				ReadBufSize int    `yaml:"read_buf_size"`
			} `yaml:"server1"`
			Server2 struct {
				Addr        string `yaml:"addr"`
				ReadBufSize int    `yaml:"read_buf_size"`
			} `yaml:"server2"`
			Server3 struct {
				Addr        string `yaml:"addr"`
				Delimiter   string `yaml:"delimiter"`
				ReadBufSize int    `yaml:"read_buf_size"`
			} `yaml:"server3"`
		} `yaml:"servers"`
		Pool struct {
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  string `yaml:"destroy_timeout"`
		} `yaml:"pool"`
	} `yaml:"tcp"`
	Fileserver struct {
		Address           string `yaml:"address"`
		CalculateEtag     bool   `yaml:"calculate_etag"`
		Weak              bool   `yaml:"weak"`
		StreamRequestBody bool   `yaml:"stream_request_body"`
		Serve             []struct {
			Prefix        string `yaml:"prefix"`
			Root          string `yaml:"root"`
			Compress      bool   `yaml:"compress"`
			CacheDuration int    `yaml:"cache_duration"`
			MaxAge        int    `yaml:"max_age"`
			BytesRange    bool   `yaml:"bytes_range"`
		} `yaml:"serve"`
	} `yaml:"fileserver"`
	Endure struct {
		GracePeriod string `yaml:"grace_period"`
		PrintGraph  bool   `yaml:"print_graph"`
		LogLevel    string `yaml:"log_level"`
	} `yaml:"endure"`
}

// -----------

// Service represents particular service configuration
type Service struct {
	Command         string        `mapstructure:"command"`
	Output          string        `mapstructure:"log_output"`
	ProcessNum      int           `mapstructure:"process_num"`
	ExecTimeout     time.Duration `mapstructure:"exec_timeout"`
	RemainAfterExit bool          `mapstructure:"remain_after_exit"`
	RestartSec      uint64        `mapstructure:"restart_sec"`
	Env             Env           `mapstructure:"env"`
}
