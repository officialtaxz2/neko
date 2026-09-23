package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Media struct {
	WebCodecsWS MediaWebCodecsWS
	HLS         MediaHLS
}

type MediaWebCodecsWS struct {
	Enabled               bool
	AllowedOrigins        []string
	TrustedProxies        []string
	AllowInsecureLoopback bool
	MaxConnections        int
}

type MediaHLS struct {
	Enabled        bool
	AllowedOrigins []string
	TrustedProxies []string
	Modes          []string
	MaxLeases      int
	MaxRequests    int
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
	if err := viper.BindPFlag("media.webcodecs_ws.max_connections", cmd.PersistentFlags().Lookup("media.webcodecs_ws.max_connections")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("media.hls.enabled", false, "enable experimental HLS negotiation foundations (no media route in phase 1)")
	if err := viper.BindPFlag("media.hls.enabled", cmd.PersistentFlags().Lookup("media.hls.enabled")); err != nil {
		return err
	}
	cmd.PersistentFlags().StringSlice("media.hls.allowed_origins", []string{}, "exact HTTPS origins allowed to bootstrap experimental HLS playback")
	if err := viper.BindPFlag("media.hls.allowed_origins", cmd.PersistentFlags().Lookup("media.hls.allowed_origins")); err != nil {
		return err
	}
	cmd.PersistentFlags().StringSlice("media.hls.trusted_proxies", []string{}, "explicit trusted proxy IPs or CIDRs for experimental HLS forwarded headers")
	if err := viper.BindPFlag("media.hls.trusted_proxies", cmd.PersistentFlags().Lookup("media.hls.trusted_proxies")); err != nil {
		return err
	}
	cmd.PersistentFlags().StringSlice("media.hls.modes", []string{"hls", "ll-hls"}, "enabled experimental HLS modes (hls,ll-hls)")
	if err := viper.BindPFlag("media.hls.modes", cmd.PersistentFlags().Lookup("media.hls.modes")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("media.hls.max_leases", 128, "maximum concurrent experimental HLS playback leases (1-128)")
	if err := viper.BindPFlag("media.hls.max_leases", cmd.PersistentFlags().Lookup("media.hls.max_leases")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("media.hls.max_requests", 512, "maximum concurrent experimental HLS HTTP requests (1-512)")
	return viper.BindPFlag("media.hls.max_requests", cmd.PersistentFlags().Lookup("media.hls.max_requests"))
}

func (config *Media) Set() {
	config.WebCodecsWS.Enabled = viper.GetBool("media.webcodecs_ws.enabled")
	config.WebCodecsWS.AllowedOrigins = viper.GetStringSlice("media.webcodecs_ws.allowed_origins")
	config.WebCodecsWS.TrustedProxies = viper.GetStringSlice("media.webcodecs_ws.trusted_proxies")
	config.WebCodecsWS.AllowInsecureLoopback = viper.GetBool("media.webcodecs_ws.allow_insecure_loopback")
	config.WebCodecsWS.MaxConnections = viper.GetInt("media.webcodecs_ws.max_connections")
	config.HLS.Enabled = viper.GetBool("media.hls.enabled")
	config.HLS.AllowedOrigins = viper.GetStringSlice("media.hls.allowed_origins")
	config.HLS.TrustedProxies = viper.GetStringSlice("media.hls.trusted_proxies")
	config.HLS.Modes = viper.GetStringSlice("media.hls.modes")
	config.HLS.MaxLeases = viper.GetInt("media.hls.max_leases")
	config.HLS.MaxRequests = viper.GetInt("media.hls.max_requests")
}
