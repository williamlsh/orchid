package frontend

import (
	"net"
	"strconv"

	"github.com/ossm-org/orchid/pkg/apis/auth"
	"github.com/ossm-org/orchid/pkg/email"
	"github.com/ossm-org/orchid/pkg/logging"
	"github.com/ossm-org/orchid/services/cache"
	"github.com/ossm-org/orchid/services/frontend"
	"github.com/spf13/cobra"
)

var (
	host           string
	port           int
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
		frontendConfig.FrontendHostPort = net.JoinHostPort(host, strconv.Itoa(port))
		frontendConfig.AuthSecrets = authSecrets
		frontendConfig.Email = emailConfig

		logger := logging.NewLogger("", false)
		cache := cache.New(logger, cacheConfig)
		server := frontend.NewServer(logger, cache, frontendConfig)
		return server.Run()
	},
}

func init() {
	FrontendCmd.PersistentFlags().StringVarP(&host, "frontend-service-host", "h", "0.0.0.0", "Frontend service host")
	FrontendCmd.PersistentFlags().IntVarP(&port, "frontend-service-port", "p", 8080, "Frontend service port")

	FrontendCmd.PersistentFlags().StringVarP(&cacheConfig.Addr, "redis-addr", "r", "localhost:6379", "Redis server address")
	FrontendCmd.PersistentFlags().StringVarP(&cacheConfig.Passwd, "redis-passwd", "p", "", "Redis server password")

	FrontendCmd.PersistentFlags().StringVarP(&emailConfig.From, "email-from", "f", "abc@gmail.com", "Email from address")
	FrontendCmd.PersistentFlags().StringVarP(&emailConfig.Host, "smtp-server-host", "sh", "", "Smtp server host")
	FrontendCmd.PersistentFlags().IntVarP(&emailConfig.Port, "smtp-server-port", "sp", 25, "Smtp server port")
	FrontendCmd.PersistentFlags().StringVarP(&emailConfig.Username, "email-username", "eu", "abc", "Email username")
	FrontendCmd.PersistentFlags().StringVarP(&emailConfig.Passwd, "email-passwd", "ep", "123abc", "Email password")

	FrontendCmd.PersistentFlags().StringVarP(&authSecrets.AccessSecret, "auth-access-secret", "as", "123abc", "Authentication access secret")
	FrontendCmd.PersistentFlags().StringVarP(&authSecrets.RefreshSecret, "auth-refresh-secret", "rs", "123abc", "Authentication refresh secret")
}
