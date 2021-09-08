package redis

import (
	"context"
	"github.com/go-redis/redis/v7"
	"time"
)

type Client struct {
	*redis.Client
}

type Options struct {
	Username string `defval:"root"`
	Password string `defval:"111111" null:""`
	// max retry,  not retry = -1,  default, single=0 cluster=8
	MaxRetries int
	// timeout
	DialTimeout  time.Duration `defval:"60s"`
	ReadTimeout  time.Duration `defval:"300s"`
	WriteTimeout time.Duration `defval:"300s"`
}

type Config struct {
	Addr string `defval:"127.0.0.1:6379"`
	DB   int    `defval:"0"`
	Options
}

// NewRedisClient simplify GET SET DEL operation
func NewRedisClient(cfg Config) (*Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
	sCmd := cli.Ping()
	if sCmd.Err() != nil {
		return nil, sCmd.Err()
	}
	return &Client{
		Client: cli,
	}, nil
}

// Get return String value
func (c *Client) Get(key string) (interface{}, error) {
	return c.GetContext(context.Background(), key)
}

func (c *Client) GetContext(ctx context.Context, key string) (interface{}, error) {
	strCmd := c.Client.WithContext(ctx).Get(key)
	if err := strCmd.Err(); err != nil {
		return nil, err
	}
	return strCmd.Val(), nil
}

func (c *Client) Set(key string, value interface{}) error {
	return c.SetContext(context.Background(), key, value, -1)
}

func (c *Client) SetWithTTL(key string, value interface{}, expiration time.Duration) error {
	return c.SetContext(context.Background(), key, value, expiration)
}

func (c *Client) SetContext(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	sCmd := c.Client.WithContext(ctx).Set(key, value, expiration)
	return sCmd.Err()
}

func (c *Client) Del(key string) error {
	return c.DelContext(context.Background(), key)
}

func (c *Client) DelContext(ctx context.Context, key string) error {
	return c.Client.WithContext(ctx).Del(key).Err()
}

type ClusterConfig struct {
	Addrs []string
	Options
}

type ClusterClient struct {
	*redis.ClusterClient
}

func NewRedisClusterClient(config ClusterConfig) (*ClusterClient, error) {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Username:     config.Username,
		Password:     config.Password,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	sCmd := cli.Ping()
	if sCmd.Err() != nil {
		return nil, sCmd.Err()
	}
	return &ClusterClient{
		ClusterClient: cli,
	}, nil
}
