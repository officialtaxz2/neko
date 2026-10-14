package websocket

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type WebSocketPeerCtx struct {
	logger     zerolog.Logger
	connection *websocket.Conn
	writer     *utils.WebSocketWriter
}

func newPeer(logger zerolog.Logger, connection *websocket.Conn) *WebSocketPeerCtx {
	return &WebSocketPeerCtx{
		logger:     logger.With().Str("submodule", "peer").Logger(),
		connection: connection,
		writer:     utils.NewWebSocketWriter(connection),
	}
}

func (peer *WebSocketPeerCtx) Send(event string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		peer.logger.Err(err).Str("event", event).Msg("message marshalling has failed")
		return
	}

	err = peer.writer.SendJSON(types.WebSocketMessage{
		Event:   event,
		Payload: raw,
	})

	if err != nil {
		if e := errors.Unwrap(err); e != nil {
			err = e // unwrap if possible
		}
		peer.logger.Warn().Err(err).Str("event", event).Msg("send message error")
		return
	}

	// log events if not ignored
	if !slices.Contains(nologEvents, event) {
		if len(raw) > maxPayloadLogLength {
			raw = []byte("<truncated>")
		}

		peer.logger.Debug().
			Str("address", peer.connection.RemoteAddr().String()).
			Str("event", event).
			Str("payload", string(raw)).
			Msg("sending message to client")
	}
}

func (peer *WebSocketPeerCtx) Ping() error {
	// application level heartbeat
	if err := peer.writer.SendJSON(types.WebSocketMessage{
		Event: event.SYSTEM_HEARTBEAT,
	}); err != nil {
		return err
	}

	return peer.writer.Send(websocket.PingMessage, nil)
}

func (peer *WebSocketPeerCtx) Destroy(reason string) {
	raw, err := json.Marshal(message.SystemDisconnect{Message: reason})
	if err != nil {
		peer.writer.Close()
		return
	}
	peer.writer.CloseAfterJSON(types.WebSocketMessage{
		Event: event.SYSTEM_DISCONNECT,
		Payload: raw,
	})
}
