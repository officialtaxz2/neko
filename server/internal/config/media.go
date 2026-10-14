package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Media struct {
	HLS MediaHLS
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
	cmd.PersistentFlags().Bool("media.hls.enabled", false, "enable experimental shared passive HLS/LL-HLS delivery")
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
	config.HLS.Enabled = viper.GetBool("media.hls.enabled")
	config.HLS.AllowedOrigins = viper.GetStringSlice("media.hls.allowed_origins")
	config.HLS.TrustedProxies = viper.GetStringSlice("media.hls.trusted_proxies")
	config.HLS.Modes = viper.GetStringSlice("media.hls.modes")
	config.HLS.MaxLeases = viper.GetInt("media.hls.max_leases")
	config.HLS.MaxRequests = viper.GetInt("media.hls.max_requests")
}
