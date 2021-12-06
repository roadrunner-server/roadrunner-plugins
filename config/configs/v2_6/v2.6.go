package v2_6 //nolint:stylecheck

import (
	"time"
)

type (
	Env      map[string]string
	Pipeline map[string]interface{}
)

type Config struct {
	RPC *struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"rpc"`
	// --------------
	Server *struct {
		OnInit *struct {
			Command     string `mapstructure:"command"`
			ExecTimeout string `mapstructure:"exec_timeout"`
			Env         Env    `mapstructure:"env"`
		} `mapstructure:"on_init"`
		Command      string `mapstructure:"command"`
		User         string `mapstructure:"user"`
		Group        string `mapstructure:"group"`
		Env          Env    `mapstructure:"env"`
		Relay        string `mapstructure:"relay"`
		RelayTimeout string `mapstructure:"relay_timeout"`
	} `mapstructure:"server"`
	// --------------
	Logs *struct {
		Mode        string            `mapstructure:"mode"`
		Level       string            `mapstructure:"level"`
		LineEnding  string            `mapstructure:"line_ending"`
		Encoding    string            `mapstructure:"encoding"`
		Output      []string          `mapstructure:"output"`
		ErrorOutput []string          `mapstructure:"errorOutput"`
		FileLogger  *FileLoggerConfig `mapstructure:"file_logger_options"`
		Channels    map[string]Logs   `mapstructure:"channels"`
	} `mapstructure:"logs"`
	// --------------
	Temporal *struct {
		Address   string `mapstructure:"address"`
		CacheSize int    `mapstructure:"cache_size"`
		Namespace string `mapstructure:"namespace"`
		Metrics   *struct {
			Address string `mapstructure:"address"`
			Type    string `mapstructure:"type"`
			Prefix  string `mapstructure:"prefix"`
		} `mapstructure:"metrics"`
		Activities *struct {
			Debug           bool   `mapstructure:"debug"`
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  string `mapstructure:"destroy_timeout"`
			Supervisor      *struct {
				WatchTick       string `mapstructure:"watch_tick"`
				TTL             string `mapstructure:"ttl"`
				IdleTTL         string `mapstructure:"idle_ttl"`
				MaxWorkerMemory int    `mapstructure:"max_worker_memory"`
				ExecTTL         string `mapstructure:"exec_ttl"`
			} `mapstructure:"supervisor"`
		} `mapstructure:"activities"`
		Codec      string `mapstructure:"codec"`
		DebugLevel int    `mapstructure:"debug_level"`
	} `mapstructure:"temporal"`
	// --------------
	Kv *map[string]interface{} `mapstructure:"kv"`
	// --------------
	Services *map[string]Service `mapstructure:"service"`
	// --------------
	HTTP *struct {
		Address           string   `mapstructure:"address"`
		InternalErrorCode int      `mapstructure:"internal_error_code"`
		AccessLogs        bool     `mapstructure:"access_logs"`
		MaxRequestSize    int      `mapstructure:"max_request_size"`
		Middleware        []string `mapstructure:"middleware"`
		TrustedSubnets    []string `mapstructure:"trusted_subnets"`
		NewRelic          *struct {
			AppName    string `mapstructure:"app_name"`
			LicenseKey string `mapstructure:"license_key"`
		} `mapstructure:"new_relic"`
		Uploads *struct {
			Dir    string   `mapstructure:"dir"`
			Forbid []string `mapstructure:"forbid"`
			Allow  []string `mapstructure:"allow"`
		} `mapstructure:"uploads"`
		Headers *struct {
			Cors *struct {
				AllowedOrigin    string `mapstructure:"allowed_origin"`
				AllowedHeaders   string `mapstructure:"allowed_headers"`
				AllowedMethods   string `mapstructure:"allowed_methods"`
				AllowCredentials bool   `mapstructure:"allow_credentials"`
				ExposedHeaders   string `mapstructure:"exposed_headers"`
				MaxAge           int    `mapstructure:"max_age"`
			} `mapstructure:"cors"`
			Request *struct {
				Input string `mapstructure:"input"`
			} `mapstructure:"request"`
			Response *struct {
				XPoweredBy string `mapstructure:"X-Powered-By"`
			} `mapstructure:"response"`
		} `mapstructure:"headers"`
		Static *struct {
			Dir           string   `mapstructure:"dir"`
			Forbid        []string `mapstructure:"forbid"`
			CalculateEtag bool     `mapstructure:"calculate_etag"`
			Weak          bool     `mapstructure:"weak"`
			Allow         []string `mapstructure:"allow"`
			Request       *struct {
				Input string `mapstructure:"input"`
			} `mapstructure:"request"`
			Response *struct {
				Output string `mapstructure:"output"`
			} `mapstructure:"response"`
		} `mapstructure:"static"`
		Pool *struct {
			Debug           bool   `mapstructure:"debug"`
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  string `mapstructure:"destroy_timeout"`
			Supervisor      *struct {
				WatchTick       string `mapstructure:"watch_tick"`
				TTL             string `mapstructure:"ttl"`
				IdleTTL         string `mapstructure:"idle_ttl"`
				MaxWorkerMemory int    `mapstructure:"max_worker_memory"`
				ExecTTL         string `mapstructure:"exec_ttl"`
			} `mapstructure:"supervisor"`
		} `mapstructure:"pool"`
		Ssl *struct {
			Address string `mapstructure:"address"`
			Acme    *struct {
				CertsDir              string   `mapstructure:"certs_dir"`
				Email                 string   `mapstructure:"email"`
				AltHTTPPort           int      `mapstructure:"alt_http_port"`
				AltTlsalpnPort        int      `mapstructure:"alt_tlsalpn_port"`
				ChallengeType         string   `mapstructure:"challenge_type"`
				UseProductionEndpoint bool     `mapstructure:"use_production_endpoint"`
				Domains               []string `mapstructure:"domains"`
			} `mapstructure:"acme"`
			Redirect bool   `mapstructure:"redirect"`
			Cert     string `mapstructure:"cert"`
			Key      string `mapstructure:"key"`
			RootCa   string `mapstructure:"root_ca"`
		} `mapstructure:"ssl"`
		Fcgi *struct {
			Address string `mapstructure:"address"`
		} `mapstructure:"fcgi"`
		HTTP2 *struct {
			H2C                  bool `mapstructure:"h2c"`
			MaxConcurrentStreams int  `mapstructure:"max_concurrent_streams"`
		} `mapstructure:"http2"`
	} `mapstructure:"http"`
	// --------------
	Redis *struct {
		Addrs            []string `mapstructure:"addrs"`
		MasterName       string   `mapstructure:"master_name"`
		Username         string   `mapstructure:"username"`
		Password         string   `mapstructure:"password"`
		DB               int      `mapstructure:"db"`
		SentinelPassword string   `mapstructure:"sentinel_password"`
		RouteByLatency   bool     `mapstructure:"route_by_latency"`
		RouteRandomly    bool     `mapstructure:"route_randomly"`
		DialTimeout      int      `mapstructure:"dial_timeout"`
		MaxRetries       int      `mapstructure:"max_retries"`
		MinRetryBackoff  int      `mapstructure:"min_retry_backoff"`
		MaxRetryBackoff  int      `mapstructure:"max_retry_backoff"`
		PoolSize         int      `mapstructure:"pool_size"`
		MinIdleConns     int      `mapstructure:"min_idle_conns"`
		MaxConnAge       int      `mapstructure:"max_conn_age"`
		ReadTimeout      int      `mapstructure:"read_timeout"`
		WriteTimeout     int      `mapstructure:"write_timeout"`
		PoolTimeout      int      `mapstructure:"pool_timeout"`
		IdleTimeout      int      `mapstructure:"idle_timeout"`
		IdleCheckFreq    int      `mapstructure:"idle_check_freq"`
		ReadOnly         bool     `mapstructure:"read_only"`
	} `mapstructure:"redis"`
	// --------------
	Websockets *struct {
		Broker        string `mapstructure:"broker"`
		AllowedOrigin string `mapstructure:"allowed_origin"`
		Path          string `mapstructure:"path"`
	} `mapstructure:"websockets"`
	// --------------
	Broadcast *map[string]interface{} `mapstructure:"broadcast"`
	// --------------
	Metrics *struct {
		Address string               `mapstructure:"address"`
		Collect map[string]Collector `mapstructure:"collect"`
	} `mapstructure:"metrics"`
	// --------------
	Status *struct {
		Address               string `mapstructure:"address"`
		UnavailableStatusCode int    `mapstructure:"unavailable_status_code"`
	} `mapstructure:"status"`
	// --------------
	Reload *struct {
		Interval time.Duration            `mapstructure:"interval"`
		Patterns []string                 `mapstructure:"patterns"`
		Plugins  map[string]ServiceConfig `mapstructure:"services"`
	} `mapstructure:"reload"`
	// --------------
	Nats *struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"nats"`
	// --------------
	Boltdb *struct {
		Permissions int `mapstructure:"permissions"`
	} `mapstructure:"boltdb"`
	// --------------
	AMQP *struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"amqp"`
	// --------------
	Beanstalk *struct {
		Addr    string `mapstructure:"addr"`
		Timeout string `mapstructure:"timeout"`
	} `mapstructure:"beanstalk"`
	// --------------
	Sqs *struct {
		Key          string `mapstructure:"key"`
		Secret       string `mapstructure:"secret"`
		Region       string `mapstructure:"region"`
		SessionToken string `mapstructure:"session_token"`
		Endpoint     string `mapstructure:"endpoint"`
	} `mapstructure:"sqs"`
	// --------------
	Jobs *struct {
		NumPollers   int `mapstructure:"num_pollers"`
		PipelineSize int `mapstructure:"pipeline_size"`
		Pool         *struct {
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  string `mapstructure:"destroy_timeout"`
		} `mapstructure:"pool"`
		Pipelines map[string]*Pipeline `mapstructure:"pipelines"`
		Consume   []string             `mapstructure:"consume"`
	} `mapstructure:"jobs"`
	// --------------
	Grpc *struct {
		Listen string   `mapstructure:"listen"`
		Proto  []string `mapstructure:"proto"`
		TLS    *struct {
			Key            string `mapstructure:"key"`
			Cert           string `mapstructure:"cert"`
			RootCa         string `mapstructure:"root_ca"`
			ClientAuthType string `mapstructure:"client_auth_type"`
		} `mapstructure:"tls"`
		MaxSendMsgSize        int    `mapstructure:"max_send_msg_size"`
		MaxRecvMsgSize        int    `mapstructure:"max_recv_msg_size"`
		MaxConnectionIdle     string `mapstructure:"max_connection_idle"`
		MaxConnectionAge      string `mapstructure:"max_connection_age"`
		MaxConnectionAgeGrace string `mapstructure:"max_connection_age_grace"`
		MaxConcurrentStreams  int    `mapstructure:"max_concurrent_streams"`
		PingTime              string `mapstructure:"ping_time"`
		Timeout               string `mapstructure:"timeout"`
		Pool                  *struct {
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  int    `mapstructure:"destroy_timeout"`
		} `mapstructure:"pool"`
	} `mapstructure:"grpc"`
	// --------------
	TCP *struct {
		Servers        map[string]*Server `mapstructure:"servers"`
		ReadBufferSize int                `mapstructure:"read_buf_size"`
		Pool           *struct {
			Debug           bool   `mapstructure:"debug"`
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  string `mapstructure:"destroy_timeout"`
			Supervisor      *struct {
				WatchTick       string `mapstructure:"watch_tick"`
				TTL             string `mapstructure:"ttl"`
				IdleTTL         string `mapstructure:"idle_ttl"`
				MaxWorkerMemory int    `mapstructure:"max_worker_memory"`
				ExecTTL         string `mapstructure:"exec_ttl"`
			} `mapstructure:"supervisor"`
		} `mapstructure:"pool"`
	} `mapstructure:"tcp"`
	// --------------
	Fileserver *struct {
		Address           string `mapstructure:"address"`
		CalculateEtag     bool   `mapstructure:"calculate_etag"`
		Weak              bool   `mapstructure:"weak"`
		StreamRequestBody bool   `mapstructure:"stream_request_body"`
		Serve             []*struct {
			Prefix        string `mapstructure:"prefix"`
			Root          string `mapstructure:"root"`
			Compress      bool   `mapstructure:"compress"`
			CacheDuration int    `mapstructure:"cache_duration"`
			MaxAge        int    `mapstructure:"max_age"`
			BytesRange    bool   `mapstructure:"bytes_range"`
		} `mapstructure:"serve"`
	} `mapstructure:"fileserver"`
	// --------------
	Endure *struct {
		GracePeriod string `mapstructure:"grace_period"`
		PrintGraph  bool   `mapstructure:"print_graph"`
		LogLevel    string `mapstructure:"log_level"`
	} `mapstructure:"endure"`
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

// ChannelConfig configures loggers per channel.
type ChannelConfig struct {
	Channels map[string]Config `mapstructure:"channels"`
}

// FileLoggerConfig *structure represents configuration for the file logger
type FileLoggerConfig struct {
	LogOutput  string `mapstructure:"log_output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type Logs struct {
	Mode        string            `mapstructure:"mode"`
	Level       string            `mapstructure:"level"`
	LineEnding  string            `mapstructure:"line_ending"`
	Encoding    string            `mapstructure:"encoding"`
	Output      []string          `mapstructure:"output"`
	ErrorOutput []string          `mapstructure:"errorOutput"`
	FileLogger  *FileLoggerConfig `mapstructure:"file_logger_options"`
}

// Collector describes single application specific metric.
type Collector struct {
	Namespace  string              `json:"namespace"`
	Subsystem  string              `json:"subsystem"`
	Type       string              `json:"type"`
	Help       string              `json:"help"`
	Labels     []string            `json:"labels"`
	Buckets    []float64           `json:"buckets"`
	Objectives map[float64]float64 `json:"objectives"`
}

type ServiceConfig struct {
	Recursive bool     `mapstructure:"recursive"`
	Patterns  []string `mapstructure:"patterns"`
	Dirs      []string `mapstructure:"dirs"`
	Ignore    []string `mapstructure:"ignore"`
}

type Server struct {
	Addr       string `mapstructure:"addr"`
	Delimiter  string `mapstructure:"delimiter"`
	delimBytes []byte
}
