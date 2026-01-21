package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	AWS        AWSConfig
	Services   ServicesConfig
	Security   SecurityConfig
	Features   FeaturesConfig
	Observability ObservabilityConfig
}

type ServerConfig struct {
	Port        string
	Environment string
	ServiceName string
	Version     string
	GracefulShutdownTimeout time.Duration
}

type AWSConfig struct {
	Region      string
	AccountID   string
	
	// DynamoDB
	DynamoDB DynamoDBConfig
	
	// S3
	S3 S3Config
	
	// SQS
	SQS SQSConfig
	
	// IoT Core
	IoT IoTConfig
}

type DynamoDBConfig struct {
	Endpoint        string // For LocalStack
	TableJobs       string
	TableDevices    string
}

type S3Config struct {
	Endpoint            string // For LocalStack
	BucketImages        string
	PresignedURLExpiry  time.Duration
}

type SQSConfig struct {
	Endpoint              string // For LocalStack
	QueueURLClassification string
	QueueURLDLQ           string
}

type IoTConfig struct {
	Endpoint string
}

type ServicesConfig struct {
	Classifier ClassifierServiceConfig
	Decision   DecisionServiceConfig
	
	UseServiceDiscovery bool
	DiscoveryNamespace  string
}

type ClassifierServiceConfig struct {
	URL     string
	Timeout time.Duration
}

type DecisionServiceConfig struct {
	URL     string
	Timeout time.Duration
}

type SecurityConfig struct {
	CognitoUserPoolID string
	CognitoClientID   string
	JWTSecret         string
	
	RateLimit RateLimitConfig
	CircuitBreaker CircuitBreakerConfig
}

type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

type CircuitBreakerConfig struct {
	Timeout     time.Duration
	MaxRequests uint32
	Interval    time.Duration
}

type FeaturesConfig struct {
	EnableAsyncClassification bool
	EnableCache               bool
	EnableTracing             bool
}

type ObservabilityConfig struct {
	LogLevel     string
	LogFormat    string
	EnableMetrics bool
	MetricsPort  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:                    getEnv("PORT", "8080"),
			Environment:             getEnv("APP_ENV", "development"),
			ServiceName:             getEnv("SERVICE_NAME", "orchestrator"),
			Version:                 getEnv("VERSION", "1.0.0"),
			GracefulShutdownTimeout: getDurationEnv("GRACEFUL_SHUTDOWN_TIMEOUT", "30s"),
		},
		AWS: AWSConfig{
			Region:    getEnv("AWS_REGION", "us-east-1"),
			AccountID: getEnv("AWS_ACCOUNT_ID", ""),
			DynamoDB: DynamoDBConfig{
				Endpoint:     getEnv("DYNAMODB_ENDPOINT", ""),
				TableJobs:    getEnv("DYNAMODB_TABLE_JOBS", "smart-bin-dev-jobs"),
				TableDevices: getEnv("DYNAMODB_TABLE_DEVICES", "smart-bin-dev-devices"),
			},
			S3: S3Config{
				Endpoint:           getEnv("S3_ENDPOINT", ""),
				BucketImages:       getEnv("S3_BUCKET_IMAGES", "smart-bin-dev-images"),
				PresignedURLExpiry: getDurationEnv("S3_PRESIGNED_URL_EXPIRY", "15m"),
			},
			SQS: SQSConfig{
				Endpoint:               getEnv("SQS_ENDPOINT", ""),
				QueueURLClassification: getEnv("SQS_QUEUE_URL_CLASSIFICATION", ""),
				QueueURLDLQ:            getEnv("SQS_QUEUE_URL_DLQ", ""),
			},
			IoT: IoTConfig{
				Endpoint: getEnv("IOT_ENDPOINT", ""),
			},
		},
		Services: ServicesConfig{
			Classifier: ClassifierServiceConfig{
				URL:     getEnv("CLASSIFIER_SERVICE_URL", "http://localhost:8081"),
				Timeout: getDurationEnv("CLASSIFIER_TIMEOUT", "60s"),
			},
			Decision: DecisionServiceConfig{
				URL:     getEnv("DECISION_SERVICE_URL", "http://localhost:8082"),
				Timeout: getDurationEnv("DECISION_TIMEOUT", "10s"),
			},
			UseServiceDiscovery: getBoolEnv("USE_SERVICE_DISCOVERY", false),
			DiscoveryNamespace:  getEnv("SERVICE_DISCOVERY_NAMESPACE", "smart-bin.local"),
		},
		Security: SecurityConfig{
			CognitoUserPoolID: getEnv("COGNITO_USER_POOL_ID", ""),
			CognitoClientID:   getEnv("COGNITO_CLIENT_ID", ""),
			JWTSecret:         getEnv("JWT_SECRET", ""),
			RateLimit: RateLimitConfig{
				Requests: getIntEnv("RATE_LIMIT_REQUESTS", 100),
				Window:   getDurationEnv("RATE_LIMIT_WINDOW", "1m"),
			},
			CircuitBreaker: CircuitBreakerConfig{
				Timeout:     getDurationEnv("CIRCUIT_BREAKER_TIMEOUT", "30s"),
				MaxRequests: uint32(getIntEnv("CIRCUIT_BREAKER_MAX_REQUESTS", 3)),
				Interval:    getDurationEnv("CIRCUIT_BREAKER_INTERVAL", "60s"),
			},
		},
		Features: FeaturesConfig{
			EnableAsyncClassification: getBoolEnv("ENABLE_ASYNC_CLASSIFICATION", true),
			EnableCache:               getBoolEnv("ENABLE_CACHE", false),
			EnableTracing:             getBoolEnv("ENABLE_TRACING", false),
		},
		Observability: ObservabilityConfig{
			LogLevel:      getEnv("LOG_LEVEL", "info"),
			LogFormat:     getEnv("LOG_FORMAT", "json"),
			EnableMetrics: getBoolEnv("ENABLE_METRICS", true),
			MetricsPort:   getEnv("METRICS_PORT", "9090"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	
	if c.AWS.Region == "" {
		return fmt.Errorf("AWS_REGION is required")
	}
	
	if c.AWS.DynamoDB.TableJobs == "" {
		return fmt.Errorf("DYNAMODB_TABLE_JOBS is required")
	}
	
	if c.AWS.S3.BucketImages == "" {
		return fmt.Errorf("S3_BUCKET_IMAGES is required")
	}
	
	if c.Services.Classifier.URL == "" {
		return fmt.Errorf("CLASSIFIER_SERVICE_URL is required")
	}
	
	if c.Services.Decision.URL == "" {
		return fmt.Errorf("DECISION_SERVICE_URL is required")
	}
	
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development" || c.Server.Environment == "dev"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production" || c.Server.Environment == "prod"
}

// UseLocalStack returns true if configured to use LocalStack
func (c *Config) UseLocalStack() bool {
	return c.AWS.DynamoDB.Endpoint != "" || c.AWS.S3.Endpoint != ""
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// Fallback to default if parsing fails
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}