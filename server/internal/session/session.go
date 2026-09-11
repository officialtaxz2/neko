package session

import (
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
)

// client is expected to reconnect within 5 second
// if some unexpected websocket disconnect happens
const wsDelayedDuration = 5 * time.Second

type SessionCtx struct {
	id      string
	token   string
	logger  zerolog.Logger
	manager *SessionManagerCtx
	profile types.MemberProfile
	state   types.SessionState

	websocketPeer types.WebSocketPeer
	websocketMu   sync.Mutex

	// websocket delayed set connected events
	wsDelayedMu    sync.Mutex
	wsDelayedTimer *time.Timer

	webrtcPeer types.WebRTCPeer
	webrtcMu   sync.Mutex

	mediaDelivery types.MediaDelivery
	mediaMu       sync.Mutex
}

func (session *SessionCtx) ID() string {
	return session.id
}

func (session *SessionCtx) Profile() types.MemberProfile {
	return session.profile
}

func (session *SessionCtx) profileChanged() {
	if !session.profile.CanHost && session.IsHost() {
		session.ClearHost()
	}

	if (!session.profile.CanConnect || !session.profile.CanLogin || !session.profile.CanWatch) && session.state.IsWatching {
		// TODO: Needed for legacy implementation. Websocket must die before webrtc and deliver signal close message
		// otherwise webrtc destroy would trigger websocket reconnect. In case of kick event, webrtc destroy is called
		// before websocket destroy that delivers the information about the kick.
		time.AfterFunc(time.Second, func() {
			// The delivery may have been removed if the user disconnected while
			// waiting for the delayed teardown.
			if delivery := session.GetMediaDelivery(); delivery != nil {
				_ = delivery.Close()
			}
		})
	}

	if (!session.profile.CanConnect || !session.profile.CanLogin) && session.state.IsConnected {
		session.DestroyWebSocketPeer("profile changed")
	}

	// update receive-media paused state independently from its backend
	if delivery := session.GetMediaDelivery(); delivery != nil {
		_ = delivery.SetPaused(session.PrivateModeEnabled())
	}
}

func (session *SessionCtx) State() types.SessionState {
	return session.state
}

func (session *SessionCtx) IsHost() bool {
	return session.manager.isHost(session)
}

// only needed for legacy webrtc handler
func (session *SessionCtx) LegacyIsHost() bool {
	settings := session.manager.Settings()
	if !session.profile.IsInteractive() || !session.profile.CanHost || session.PrivateModeEnabled() {
		return false
	}
	if settings.LockedControls && !session.profile.IsAdmin {
		return false
	}
	return settings.ImplicitHosting || session.manager.isHost(session)
}

func (session *SessionCtx) SetAsHost() {
	session.manager.setHost(session, session)
}

func (session *SessionCtx) SetAsHostBy(bySession types.Session) {
	session.manager.setHost(bySession, session)
}

func (session *SessionCtx) ClearHost() {
	session.manager.setHost(session, nil)
}

func (session *SessionCtx) PrivateModeEnabled() bool {
	return session.manager.Settings().PrivateMode && !session.profile.IsAdmin
}

func (session *SessionCtx) SetCursor(cursor types.Cursor) {
	if session.profile.IsInteractive() && session.manager.Settings().InactiveCursors && session.profile.SendsInactiveCursor {
		session.manager.SetCursor(cursor, session)
	}
}

// ---
// websocket
// ---

// Connect WebSocket peer sets current peer and emits connected event. It also destroys the
// previous peer, if there was one. If the peer is already set, it will be ignored.
func (session *SessionCtx) ConnectWebSocketPeer(websocketPeer types.WebSocketPeer) {
	session.websocketMu.Lock()
	isCurrentPeer := websocketPeer == session.websocketPeer
	session.websocketPeer, websocketPeer = websocketPeer, session.websocketPeer
	session.websocketMu.Unlock()

	// ignore if already set
	if isCurrentPeer {
		return
	}

	session.logger.Info().Msg("set websocket connected")

	// update state
	now := time.Now()
	session.state.IsConnected = true
	session.state.ConnectedSince = &now
	session.state.NotConnectedSince = nil

	if session.profile.IsAdmin {
		session.manager.totalAdmins.Add(1)
		session.manager.lastAdminLeftAt.Store((*time.Time)(nil))
	} else {
		session.manager.totalUsers.Add(1)
		session.manager.lastUserLeftAt.Store((*time.Time)(nil))
	}

	session.manager.emmiter.Emit("connected", session)

	// if there is a previous peer, destroy it
	if websocketPeer != nil {
		websocketPeer.Destroy("connection replaced")
	}
}

// Disconnect WebSocket peer sets current peer to nil and emits disconnected event. It also
// allows for a delayed disconnect. That means, the peer will not be disconnected immediately,
// but after a delay. If the peer is connected again before the delay, the disconnect will be
// cancelled.
//
// If the peer is not the current peer or the peer is nil, it will be ignored.
func (session *SessionCtx) DisconnectWebSocketPeer(websocketPeer types.WebSocketPeer, delayed bool) {
	session.websocketMu.Lock()
	isCurrentPeer := websocketPeer == session.websocketPeer && websocketPeer != nil
	session.websocketMu.Unlock()

	// ignore if not current peer
	if !isCurrentPeer {
		return
	}

	//
	// ws delayed
	//

	var wsDelayedTimer *time.Timer

	if delayed {
		wsDelayedTimer = time.AfterFunc(wsDelayedDuration, func() {
			session.DisconnectWebSocketPeer(websocketPeer, false)
		})
	}

	session.wsDelayedMu.Lock()
	if session.wsDelayedTimer != nil {
		session.wsDelayedTimer.Stop()
	}
	session.wsDelayedTimer = wsDelayedTimer
	session.wsDelayedMu.Unlock()

	if delayed {
		session.logger.Info().Msg("delayed websocket disconnected")
		return
	}

	//
	// not delayed
	//

	session.logger.Info().Msg("set websocket disconnected")

	now := time.Now()
	session.state.IsConnected = false
	session.state.ConnectedSince = nil
	session.state.NotConnectedSince = &now

	if session.profile.IsAdmin {
		if session.manager.totalAdmins.Add(-1) == 0 {
			session.manager.lastAdminLeftAt.Store(&now)
		}
	} else {
		if session.manager.totalUsers.Add(-1) == 0 {
			session.manager.lastUserLeftAt.Store(&now)
		}
	}

	session.manager.emmiter.Emit("disconnected", session)

	session.websocketMu.Lock()
	if websocketPeer == session.websocketPeer {
		session.websocketPeer = nil
	}
	session.websocketMu.Unlock()
}

// Destroy WebSocket peer disconnects the peer and destroys it. It ensures that the peer is
// disconnected immediately even though normal flow would be to disconnect it delayed.
func (session *SessionCtx) DestroyWebSocketPeer(reason string) {
	session.websocketMu.Lock()
	peer := session.websocketPeer
	session.websocketMu.Unlock()

	if peer == nil {
		return
	}

	// disconnect peer first, so that it is not used anymore
	session.DisconnectWebSocketPeer(peer, false)

	// destroy it afterwards
	peer.Destroy(reason)
}

// Send event to websocket peer.
func (session *SessionCtx) Send(event string, payload any) {
	session.websocketMu.Lock()
	peer := session.websocketPeer
	session.websocketMu.Unlock()

	if peer != nil {
		peer.Send(event, payload)
	}
}

// ---
// webrtc
// ---

// Set webrtc peer and destroy the old one, if there is old one.
func (session *SessionCtx) SetWebRTCPeer(webrtcPeer types.WebRTCPeer) {
	session.webrtcMu.Lock()
	session.webrtcPeer, webrtcPeer = webrtcPeer, session.webrtcPeer
	session.webrtcMu.Unlock()

	if webrtcPeer != nil && webrtcPeer != session.webrtcPeer {
		webrtcPeer.Destroy()
	}
}

// Set if current webrtc peer is connected or not. Since there might be lefover calls from
// webrtc peer, that are not used anymore, we need to check if the webrtc peer is still the
// same as the one we are setting the connected state for.
//
// If webrtc peer is disconnected, we don't expect it to be reconnected, so we set it to nil
// and send a signal close to the client. New connection is expected to use a new webrtc peer.
func (session *SessionCtx) SetWebRTCConnected(webrtcPeer types.WebRTCPeer, connected bool) {
	session.webrtcMu.Lock()
	isCurrentPeer := webrtcPeer == session.webrtcPeer
	if isCurrentPeer && !connected {
		session.webrtcPeer = nil
	}
	session.webrtcMu.Unlock()

	if !isCurrentPeer {
		return
	}
	if delivery, ok := webrtcPeer.(types.MediaDelivery); ok {
		if session.SetMediaDeliveryActive(delivery, connected) && !connected {
			session.Send(event.SIGNAL_CLOSE, nil)
		}
		return
	}

	session.logger.Info().
		Bool("connected", connected).
		Msg("set webrtc connected")

	// update state
	session.state.IsWatching = connected
	if now := time.Now(); connected {
		session.state.WatchingSince = &now
		session.state.NotWatchingSince = nil
	} else {
		session.state.WatchingSince = nil
		session.state.NotWatchingSince = &now
	}

	session.manager.emmiter.Emit("state_changed", session)

	if connected {
		return
	}

	session.Send(event.SIGNAL_CLOSE, nil)
}

// Get current WebRTC peer. Nil if not connected.
func (session *SessionCtx) GetWebRTCPeer() types.WebRTCPeer {
	session.webrtcMu.Lock()
	defer session.webrtcMu.Unlock()

	return session.webrtcPeer
}

// ---
// backend-neutral receive media
// ---

func (session *SessionCtx) SetMediaDelivery(delivery types.MediaDelivery) {
	session.mediaMu.Lock()
	previous := session.mediaDelivery
	session.mediaDelivery = delivery
	session.mediaMu.Unlock()

	if previous != nil && previous != delivery {
		_ = previous.Close()
	}
}

func (session *SessionCtx) SetMediaDeliveryActive(delivery types.MediaDelivery, active bool) bool {
	session.mediaMu.Lock()
	isCurrent := delivery != nil && delivery == session.mediaDelivery
	if isCurrent && !active {
		session.mediaDelivery = nil
	}
	session.mediaMu.Unlock()
	if !isCurrent {
		return false
	}

	session.logger.Info().
		Str("backend", delivery.Backend()).
		Bool("active", active).
		Msg("set media delivery active")

	session.state.IsWatching = active
	if now := time.Now(); active {
		session.state.WatchingSince = &now
		session.state.NotWatchingSince = nil
	} else {
		session.state.WatchingSince = nil
		session.state.NotWatchingSince = &now
	}
	session.manager.emmiter.Emit("state_changed", session)
	return true
}

func (session *SessionCtx) GetMediaDelivery() types.MediaDelivery {
	session.mediaMu.Lock()
	defer session.mediaMu.Unlock()
	return session.mediaDelivery
}
