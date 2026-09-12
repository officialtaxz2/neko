package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Media struct {
	WebCodecsWS MediaWebCodecsWS
}

type MediaWebCodecsWS struct {
	Enabled bool
}

func (Media) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("media.webcodecs_ws.enabled", false, "enable experimental WebCodecs media WebSocket negotiation")
	return viper.BindPFlag("media.webcodecs_ws.enabled", cmd.PersistentFlags().Lookup("media.webcodecs_ws.enabled"))
}

func (config *Media) Set() {
	config.WebCodecsWS.Enabled = viper.GetBool("media.webcodecs_ws.enabled")
}
