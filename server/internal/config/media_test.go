package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestMediaHLSDefaultsRemainOffAndBounded(t *testing.T){
	viper.Reset();defer viper.Reset();command:=&cobra.Command{Use:"test"};configuration:=Media{};if err:=configuration.Init(command);err!=nil{t.Fatal(err)};configuration.Set()
	if configuration.HLS.Enabled{t.Fatal("HLS defaulted on")}
	if configuration.HLS.MaxLeases!=128||configuration.HLS.MaxRequests!=512{t.Fatalf("HLS limits = %#v",configuration.HLS)}
	if len(configuration.HLS.Modes)!=2||configuration.HLS.Modes[0]!="hls"||configuration.HLS.Modes[1]!="ll-hls"{t.Fatalf("HLS modes = %#v",configuration.HLS.Modes)}
	if len(configuration.HLS.AllowedOrigins)!=0||len(configuration.HLS.TrustedProxies)!=0{t.Fatalf("HLS trust defaults = %#v",configuration.HLS)}
}
