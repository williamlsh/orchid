package frontend

import (
	"fmt"
	"net"
	"strconv"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/database"
	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/pkg/logging"
	"github.com/ossm-org/orchid/services/cache"
	"github.com/ossm-org/orchid/services/frontend"
	"github.com/spf13/cobra"
)

var (
	frontendHost string
	frontendPort int

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
)

// FrontendCmd runs frontend service.
var FrontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Starts Frontend service",
	Long:  `Starts Frontend service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d",
			pgUser, pgPass, pgHost, pgPort, pgDbname, pgSslmode, pgMaxConn)

		frontendConfig.FrontendHostPort = net.JoinHostPort(frontendHost, strconv.Itoa(frontendPort))
		frontendConfig.AuthSecrets = authSecrets
		frontendConfig.Email = emailConfig

		logger := logging.NewLogger("", false)
		cache := cache.New(logger, cacheConfig)
		db := database.New(logger, dsn)
		server := frontend.NewServer(logger, cache, db, frontendConfig)
		return server.Run()
	},
}

func init() {
	FrontendCmd.PersistentFlags().StringVar(&frontendHost, "frontend-service-host", "0.0.0.0", "Frontend service host")
	FrontendCmd.PersistentFlags().IntVar(&frontendPort, "frontend-service-port", 8080, "Frontend service port")

	FrontendCmd.PersistentFlags().StringVar(&cacheConfig.Addr, "redis-addr", "localhost:6379", "Redis server address")
	FrontendCmd.PersistentFlags().StringVar(&cacheConfig.Passwd, "redis-passwd", "", "Redis server password")

	FrontendCmd.PersistentFlags().StringVar(&emailConfig.From, "email-from", "abc@gmail.com", "Email from address")
	FrontendCmd.PersistentFlags().StringVar(&emailConfig.Host, "smtp-server-host", "", "Smtp server host")
	FrontendCmd.PersistentFlags().IntVar(&emailConfig.Port, "smtp-server-port", 25, "Smtp server port")
	FrontendCmd.PersistentFlags().StringVar(&emailConfig.Username, "email-username", "abc", "Email username")
	FrontendCmd.PersistentFlags().StringVar(&emailConfig.Passwd, "email-passwd", "123abc", "Email password")

	FrontendCmd.PersistentFlags().StringVar(&authSecrets.AccessSecret, "auth-access-secret", "123abc", "Authentication access secret")
	FrontendCmd.PersistentFlags().StringVar(&authSecrets.RefreshSecret, "auth-refresh-secret", "123abc", "Authentication refresh secret")

	FrontendCmd.PersistentFlags().StringVar(&pgUser, "pg-user", "", "postgreSQL database username")
	FrontendCmd.PersistentFlags().StringVar(&pgPass, "pg-passwd", "", "postgreSQL database password")
	FrontendCmd.PersistentFlags().StringVar(&pgDbname, "pg-dbname", "", "postgreSQL database name")
	FrontendCmd.PersistentFlags().StringVar(&pgHost, "pg-host", "", "postgreSQL database host")
	FrontendCmd.PersistentFlags().IntVar(&pgPort, "pg-port", 5432, "postgreSQL database port")
	FrontendCmd.PersistentFlags().StringVar(&pgSslmode, "pg-sslmode", "disable", "postgreSQL database sslmode")
	FrontendCmd.PersistentFlags().IntVar(&pgMaxConn, "pg-pool-max-conn", 10, "postgreSQL database pool max connections")
}
