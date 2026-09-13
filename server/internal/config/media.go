package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Media struct {
	WebCodecsWS MediaWebCodecsWS
}

type MediaWebCodecsWS struct {
	Enabled               bool
	AllowedOrigins        []string
	TrustedProxies        []string
	AllowInsecureLoopback bool
	MaxConnections        int
}

func (Media) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("media.webcodecs_ws.enabled", false, "enable experimental WebCodecs media WebSocket negotiation")
	if err := viper.BindPFlag("media.webcodecs_ws.enabled", cmd.PersistentFlags().Lookup("media.webcodecs_ws.enabled")); err != nil {
		return err
	}
	cmd.PersistentFlags().StringSlice("media.webcodecs_ws.allowed_origins", []string{}, "exact allowed origins for the experimental media WebSocket; empty allows exact same-origin only")
	if err := viper.BindPFlag("media.webcodecs_ws.allowed_origins", cmd.PersistentFlags().Lookup("media.webcodecs_ws.allowed_origins")); err != nil {
		return err
	}
	cmd.PersistentFlags().StringSlice("media.webcodecs_ws.trusted_proxies", []string{"127.0.0.0/8", "::1/128"}, "trusted proxy IPs or CIDRs for media WebSocket forwarded headers")
	if err := viper.BindPFlag("media.webcodecs_ws.trusted_proxies", cmd.PersistentFlags().Lookup("media.webcodecs_ws.trusted_proxies")); err != nil {
		return err
	}
	cmd.PersistentFlags().Bool("media.webcodecs_ws.allow_insecure_loopback", false, "allow cleartext media WebSockets only for an explicit direct loopback development endpoint")
	if err := viper.BindPFlag("media.webcodecs_ws.allow_insecure_loopback", cmd.PersistentFlags().Lookup("media.webcodecs_ws.allow_insecure_loopback")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("media.webcodecs_ws.max_connections", 128, "maximum concurrent experimental media WebSockets (1-128)")
	return viper.BindPFlag("media.webcodecs_ws.max_connections", cmd.PersistentFlags().Lookup("media.webcodecs_ws.max_connections"))
}

func (config *Media) Set() {
	config.WebCodecsWS.Enabled = viper.GetBool("media.webcodecs_ws.enabled")
	config.WebCodecsWS.AllowedOrigins = viper.GetStringSlice("media.webcodecs_ws.allowed_origins")
	config.WebCodecsWS.TrustedProxies = viper.GetStringSlice("media.webcodecs_ws.trusted_proxies")
	config.WebCodecsWS.AllowInsecureLoopback = viper.GetBool("media.webcodecs_ws.allow_insecure_loopback")
	config.WebCodecsWS.MaxConnections = viper.GetInt("media.webcodecs_ws.max_connections")
}
