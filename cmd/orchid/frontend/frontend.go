package frontend

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/cache"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/pkg/logging"
	"github.com/ossm-org/orchid/pkg/storage"
	"github.com/ossm-org/orchid/services/frontend"
)

var (
	logLevel       string
	logDevelopment bool
	frontendHost   string
	frontendPort   int

	pgUser    string
	pgPass    string
	pgDbname  string
	pgHost    string
	pgPort    int
	pgSslmode string
	pgMaxConn int

	cacheConfig    cache.ConfigOptions
	authSecrets    auth.ConfigOptions
	emailConfig    email.ConfigOptions
	frontendConfig frontend.ConfigOptions
	storageConfig  storage.ConfigOptions
)

// Cmd runs frontend service.
var Cmd = &cobra.Command{
	Use:   "frontend",
	Short: "Starts Frontend service",
	Long:  `Starts Frontend service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d",
			pgUser, pgPass, pgHost, pgPort, pgDbname, pgSslmode, pgMaxConn)

		frontendConfig.FrontendHostPort = net.JoinHostPort(frontendHost, strconv.Itoa(frontendPort))
		frontendConfig.AuthSecrets = authSecrets
		frontendConfig.Email = emailConfig

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := logging.NewLogger(logLevel, logDevelopment)
		defer logger.Sync()

		ctx = logging.WithLogger(ctx, logger)

		cache := cache.New(ctx, &cacheConfig)
		defer cache.Client.Close()

		storage := storage.New(ctx, storageConfig)

		db := database.New(ctx, dsn)
		defer db.Pool.Close()

		if err := db.Migrate(ctx); err != nil {
			return err
		}
		if err := db.IsSchemaUpToDate(ctx); err != nil {
			return err
		}

		server := frontend.NewServer(logger, cache, db, storage, frontendConfig)
		return server.Run()
	},
}

func init() {
	Cmd.PersistentFlags().StringVar(&frontendHost, "frontend-service-host", "0.0.0.0", "Frontend service host")
	Cmd.PersistentFlags().IntVar(&frontendPort, "frontend-service-port", 8080, "Frontend service port")

	Cmd.PersistentFlags().StringVar(&cacheConfig.Addr, "redis-addr", "localhost:6379", "Redis server address")
	Cmd.PersistentFlags().StringVar(&cacheConfig.Passwd, "redis-passwd", "", "Redis server password")

	Cmd.PersistentFlags().StringVar(&emailConfig.From, "email-from", "abc@gmail.com", "Email from address")
	Cmd.PersistentFlags().StringVar(&emailConfig.Host, "smtp-server-host", "", "Smtp server host")
	Cmd.PersistentFlags().IntVar(&emailConfig.Port, "smtp-server-port", 25, "Smtp server port")
	Cmd.PersistentFlags().StringVar(&emailConfig.Username, "email-username", "abc", "Email username")
	Cmd.PersistentFlags().StringVar(&emailConfig.Passwd, "email-passwd", "123abc", "Email password")

	Cmd.PersistentFlags().StringVar(&authSecrets.AccessSecret, "auth-access-secret", "123abc", "Authentication access secret")
	Cmd.PersistentFlags().StringVar(&authSecrets.RefreshSecret, "auth-refresh-secret", "123abc", "Authentication refresh secret")

	Cmd.PersistentFlags().StringVar(&pgUser, "pg-user", "", "postgreSQL database username")
	Cmd.PersistentFlags().StringVar(&pgPass, "pg-passwd", "", "postgreSQL database password")
	Cmd.PersistentFlags().StringVar(&pgDbname, "pg-dbname", "", "postgreSQL database name")
	Cmd.PersistentFlags().StringVar(&pgHost, "pg-host", "", "postgreSQL database host")
	Cmd.PersistentFlags().IntVar(&pgPort, "pg-port", 5432, "postgreSQL database port")
	Cmd.PersistentFlags().StringVar(&pgSslmode, "pg-sslmode", "disable", "postgreSQL database sslmode")
	Cmd.PersistentFlags().IntVar(&pgMaxConn, "pg-pool-max-conn", 10, "postgreSQL database pool max connections")

	Cmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Log level")
	Cmd.PersistentFlags().BoolVar(&logDevelopment, "log-development", false, "Log development")

	Cmd.PersistentFlags().StringVar(&storageConfig.Endpoint, "minio-endpoint", "127.0.0.1:9000", "Minio endpoint")
	Cmd.PersistentFlags().StringVar(&storageConfig.ID, "minio-id", "", "Minio ID")
	Cmd.PersistentFlags().StringVar(&storageConfig.Secret, "minio-secret", "", "Minio secret")
	Cmd.PersistentFlags().BoolVar(&storageConfig.Secure, "minio-enable-secure", false, "Enable minio secure connection")

	rand.Seed(int64(time.Now().Nanosecond()))
}
