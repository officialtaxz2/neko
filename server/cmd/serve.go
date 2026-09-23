package cmd

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/internal/api"
	"github.com/m1k1o/neko/server/internal/capture"
	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/desktop"
	"github.com/m1k1o/neko/server/internal/http"
	mediadelivery "github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/internal/mediahls"
	"github.com/m1k1o/neko/server/internal/mediaws"
	"github.com/m1k1o/neko/server/internal/member"
	"github.com/m1k1o/neko/server/internal/plugins"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/internal/webrtc"
	"github.com/m1k1o/neko/server/internal/websocket"
	"github.com/m1k1o/neko/server/pkg/types"
)

func init() {
	service := serve{}

	command := &cobra.Command{
		Use:    "serve",
		Short:  "serve neko streaming server",
		Long:   `serve neko streaming server`,
		PreRun: service.PreRun,
		Run:    service.Run,
	}

	if err := service.Init(command); err != nil {
		log.Panic().Err(err).Msg("unable to initialize configuration")
	}

	root.AddCommand(command)
}

type serve struct {
	logger zerolog.Logger

	configs struct {
		Desktop config.Desktop
		Capture config.Capture
		WebRTC  config.WebRTC
		Member  config.Member
		Media   config.Media
		Session config.Session
		Plugins config.Plugins
		Server  config.Server
	}

	managers struct {
		desktop           *desktop.DesktopManagerCtx
		capture           *capture.CaptureManagerCtx
		media             *mediadelivery.ManagerCtx
		webRTC            *webrtc.WebRTCManagerCtx
		member            *member.MemberManagerCtx
		session           *session.SessionManagerCtx
		webSocket         *websocket.WebSocketManagerCtx
		mediaWS           *mediaws.Negotiator
		mediaWSBackend    *mediaws.Backend
		mediaWSController *mediaws.Controller
		mediaHLS           *mediahls.Negotiator
		mediaHLSBackend    *mediahls.Backend
		mediaHLSController *mediahls.Controller
		plugins           *plugins.ManagerCtx
		api               *api.ApiManagerCtx
		http              *http.HttpManagerCtx
	}
}

func (c *serve) Init(cmd *cobra.Command) error {
	if err := c.configs.Desktop.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Capture.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.WebRTC.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Member.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Media.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Session.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Plugins.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Server.Init(cmd); err != nil {
		return err
	}

	// legacy if explicitly enabled or if unspecified and legacy config is found
	if viper.GetBool("legacy") || !viper.IsSet("legacy") {
		if err := c.configs.Desktop.InitV2(cmd); err != nil {
			return err
		}
		if err := c.configs.Capture.InitV2(cmd); err != nil {
			return err
		}
		if err := c.configs.WebRTC.InitV2(cmd); err != nil {
			return err
		}
		if err := c.configs.Member.InitV2(cmd); err != nil {
			return err
		}
		if err := c.configs.Session.InitV2(cmd); err != nil {
			return err
		}
		if err := c.configs.Server.InitV2(cmd); err != nil {
			return err
		}
	}

	return nil
}

func (c *serve) PreRun(cmd *cobra.Command, args []string) {
	c.logger = log.With().Str("service", "neko").Logger()

	c.configs.Desktop.Set()
	c.configs.Capture.Set()
	c.configs.WebRTC.Set()
	c.configs.Member.Set()
	c.configs.Media.Set()
	c.configs.Session.Set()
	c.configs.Plugins.Set()
	c.configs.Server.Set()

	// legacy if explicitly enabled or if unspecified and legacy config is found
	if viper.GetBool("legacy") || !viper.IsSet("legacy") {
		c.configs.Desktop.SetV2()
		c.configs.Capture.SetV2()
		c.configs.WebRTC.SetV2()
		c.configs.Member.SetV2()
		c.configs.Session.SetV2()
		c.configs.Server.SetV2()
	}
}

func (c *serve) Start(cmd *cobra.Command) {
	c.managers.session = session.New(
		&c.configs.Session,
	)

	c.managers.member = member.New(
		c.managers.session,
		&c.configs.Member,
	)

	if err := c.managers.member.Connect(); err != nil {
		c.logger.Panic().Err(err).Msg("unable to connect to member manager")
	}

	c.managers.desktop = desktop.New(
		&c.configs.Desktop,
	)
	c.managers.desktop.Start()

	c.managers.capture = capture.New(
		c.managers.desktop,
		&c.configs.Capture,
	)
	c.managers.capture.Start()

	c.managers.media = mediadelivery.New(
		c.managers.session,
	)

	c.managers.webRTC = webrtc.New(
		c.managers.desktop,
		c.managers.capture,
		c.managers.media,
		&c.configs.WebRTC,
	)
	if err := c.managers.media.Register(c.managers.webRTC); err != nil {
		c.logger.Panic().Err(err).Msg("unable to register WebRTC media backend")
	}
	c.managers.webRTC.Start()

	c.managers.webSocket = websocket.New(
		c.managers.session,
		c.managers.desktop,
		c.managers.capture,
		c.managers.webRTC,
	)
	if c.configs.Media.WebCodecsWS.Enabled {
		tickets := mediaws.NewTicketStore()
		c.managers.mediaWS = mediaws.NewNegotiator(c.managers.session, c.managers.capture.Media(), tickets)
		c.managers.webSocket.AddHandler(c.managers.mediaWS.Handler)
		c.managers.mediaWSBackend = mediaws.NewBackend(c.managers.capture.Media())
		if err := c.managers.media.Register(c.managers.mediaWSBackend); err != nil {
			c.logger.Panic().Err(err).Msg("unable to register WebCodecs media WebSocket backend")
		}
		controller, err := mediaws.NewController(c.managers.session, c.managers.media, tickets, mediaws.ControllerConfig{
			AllowedOrigins:        c.configs.Media.WebCodecsWS.AllowedOrigins,
			TrustedProxies:        c.configs.Media.WebCodecsWS.TrustedProxies,
			AllowInsecureLoopback: c.configs.Media.WebCodecsWS.AllowInsecureLoopback,
			MaxConnections:        c.configs.Media.WebCodecsWS.MaxConnections,
		})
		if err != nil {
			c.logger.Panic().Err(err).Msg("unable to configure WebCodecs media WebSocket route")
		}
		c.managers.mediaWSController = controller
	}
	if c.configs.Media.HLS.Enabled {
		hlsConfig := mediahls.Config{
			AllowedOrigins: c.configs.Media.HLS.AllowedOrigins,
			TrustedProxies: c.configs.Media.HLS.TrustedProxies,
			Modes:          c.configs.Media.HLS.Modes,
			MaxLeases:      c.configs.Media.HLS.MaxLeases,
			MaxRequests:    c.configs.Media.HLS.MaxRequests,
		}
		tickets := mediahls.NewTicketStore()
		leases, err := mediahls.NewLeaseStore(hlsConfig.MaxLeases, hlsConfig.MaxRequests)
		if err != nil {
			c.logger.Panic().Err(err).Msg("unable to configure HLS lease store")
		}
		backend, err := mediahls.NewBackend(c.managers.capture.Media(), leases)
		if err != nil {
			c.logger.Panic().Err(err).Msg("unable to configure HLS packager backend")
		}
		if err := c.managers.media.Register(backend); err != nil {
			c.logger.Panic().Err(err).Msg("unable to register HLS media backend")
		}
		negotiator, err := mediahls.NewNegotiator(c.managers.session, tickets, hlsConfig)
		if err != nil {
			c.logger.Panic().Err(err).Msg("unable to configure HLS negotiation")
		}
		controller, err := mediahls.NewController(c.managers.session, c.managers.media, tickets, leases, backend, hlsConfig)
		if err != nil {
			c.logger.Panic().Err(err).Msg("unable to configure HLS HTTP delivery")
		}
		c.managers.mediaHLS = negotiator
		c.managers.mediaHLSBackend = backend
		c.managers.mediaHLSController = controller
		c.managers.webSocket.AddHandler(c.managers.mediaHLS.Handler)
	}
	c.managers.webSocket.Start()

	c.managers.api = api.New(
		c.managers.session,
		c.managers.member,
		c.managers.desktop,
		c.managers.capture,
	)

	c.managers.plugins = plugins.New(
		&c.configs.Plugins,
	)

	// init and set configuration now
	// this means it won't be in --help
	c.managers.plugins.InitConfigs(cmd)
	c.managers.plugins.SetConfigs()

	c.managers.plugins.Start(
		c.managers.session,
		c.managers.webSocket,
		c.managers.api,
	)

	var mediaWebSocketHandler types.RouterHandler
	if c.managers.mediaWSController != nil {
		mediaWebSocketHandler = c.managers.mediaWSController.Handle
	}
	var mediaHLSRouter func(types.Router)
	if c.managers.mediaHLSController != nil {
		mediaHLSRouter = c.managers.mediaHLSController.Route
	}
	c.managers.http = http.New(
		c.managers.webSocket,
		c.managers.api,
		&c.configs.Server,
		mediaWebSocketHandler,
		mediaHLSRouter,
	)
	c.managers.http.Start()
}

func (c *serve) Shutdown() {
	var err error

	err = c.managers.http.Shutdown()
	c.logger.Err(err).Msg("http manager shutdown")

	err = c.managers.plugins.Shutdown()
	c.logger.Err(err).Msg("plugins manager shutdown")

	err = c.managers.webSocket.Shutdown()
	c.logger.Err(err).Msg("websocket manager shutdown")

	err = c.managers.media.Shutdown()
	c.logger.Err(err).Msg("media delivery manager shutdown")
	if c.managers.mediaHLSBackend != nil {
		c.managers.mediaHLSBackend.Shutdown()
		c.logger.Info().Msg("HLS packager backend shutdown")
	}

	err = c.managers.webRTC.Shutdown()
	c.logger.Err(err).Msg("webrtc manager shutdown")

	err = c.managers.capture.Shutdown()
	c.logger.Err(err).Msg("capture manager shutdown")

	err = c.managers.desktop.Shutdown()
	c.logger.Err(err).Msg("desktop manager shutdown")

	err = c.managers.member.Disconnect()
	c.logger.Err(err).Msg("member manager disconnect")
}

func (c *serve) Run(cmd *cobra.Command, args []string) {
	c.logger.Info().Msg("starting neko server")
	c.Start(cmd)
	c.logger.Info().Msg("neko ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	c.logger.Warn().Msgf("received %s, attempting graceful shutdown", sig)
	c.Shutdown()
	c.logger.Info().Msg("shutdown complete")
}
