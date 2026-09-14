package mediaws

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type controllerTestDeliveries struct{}

func (*controllerTestDeliveries) Register(types.MediaBackend) error { return nil }
func (*controllerTestDeliveries) Backends() []types.MediaBackendDescriptor { return nil }
func (*controllerTestDeliveries) Open(context.Context, types.Session, types.MediaDeliveryRequest) (types.MediaDelivery, error) {
	return nil, errors.New("unexpected delivery open")
}
func (*controllerTestDeliveries) CloseSession(string) {}
func (*controllerTestDeliveries) Shutdown() error     { return nil }

type controllerTestPeer struct{}

func (*controllerTestPeer) Send(string, any)  {}
func (*controllerTestPeer) Ping() error        { return nil }
func (*controllerTestPeer) Destroy(string)     {}

func newControllerTest(t *testing.T, configOverride ControllerConfig) (*Controller, *session.SessionManagerCtx, *TicketStore, time.Time) {
	t.Helper()
	sessions := session.New(&config.Session{})
	tickets := NewTicketStore()
	controller, err := NewController(sessions, &controllerTestDeliveries{}, tickets, configOverride)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0)
	controller.now = func() time.Time { return now }
	return controller, sessions, tickets, now
}

func attachmentRequest(ticket, origin string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "https://neko.example/api/media/ws", nil)
	request.RemoteAddr = "192.0.2.10:4321"
	request.Header.Set("Origin", origin)
	request.Header.Set("Sec-WebSocket-Protocol", ProtocolName+", neko.media.ticket."+ticket)
	return request
}

func httpErrorCode(t *testing.T, err error) int {
	t.Helper()
	var httpError *utils.HTTPError
	if !errors.As(err, &httpError) {
		t.Fatalf("error = %v, want *utils.HTTPError", err)
	}
	return httpError.Code
}

func TestAttachmentProtocolsAreExactAndOrdered(t *testing.T) {
	ticket := "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFB"
	valid := []string{ProtocolName, "neko.media.ticket." + ticket}
	if parsed, status := parseAttachmentProtocols(valid); status != 0 || parsed != ticket {
		t.Fatalf("valid protocols = %q/%d", parsed, status)
	}
	cases := []struct {
		values []string
		status int
	}{
		{nil, http.StatusUpgradeRequired},
		{[]string{"other, neko.media.ticket." + ticket}, http.StatusUpgradeRequired},
		{[]string{ProtocolName}, http.StatusBadRequest},
		{[]string{ProtocolName + ", neko.media.ticket.short"}, http.StatusBadRequest},
		{[]string{ProtocolName + ", neko.media.ticket." + ticket + ", extra"}, http.StatusBadRequest},
	}
	for _, testCase := range cases {
		if _, status := parseAttachmentProtocols(testCase.values); status != testCase.status {
			t.Fatalf("protocols %q status = %d, want %d", testCase.values, status, testCase.status)
		}
	}
}

func TestAttachmentOpenErrorsUseBoundedCloseCodes(t *testing.T) {
	for _, testCase := range []struct {
		err    error
		code   int
		reason string
	}{
		{types.ErrMediaDeliveryNotAllowed, 4401, "attachment_invalid"},
		{types.ErrMediaSourceNotFound, 4406, "unsupported_format"},
		{ErrCodecUnsupported, 4406, "unsupported_format"},
		{errors.New("backend failed"), 4500, "backend_open_failed"},
	} {
		code, reason := attachmentOpenClose(testCase.err)
		if code != testCase.code || reason != testCase.reason {
			t.Fatalf("attachment error %v = %d/%q, want %d/%q", testCase.err, code, reason, testCase.code, testCase.reason)
		}
	}
}

func TestControllerPreUpgradeBoundaryAndTicketConsumption(t *testing.T) {
	controller, sessions, tickets, now := newControllerTest(t, ControllerConfig{})
	viewer, _, err := sessions.Create("viewer", types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: true})
	if err != nil {
		t.Fatal(err)
	}
	viewer.ConnectWebSocketPeer(&controllerTestPeer{})

	issue := func() TicketOffer {
		offer, err := tickets.Issue(testBinding(viewer.ID()), now)
		if err != nil {
			t.Fatal(err)
		}
		return offer
	}

	offer := issue()
	request := attachmentRequest(offer.Ticket, "https://wrong.example")
	if code := httpErrorCode(t, controller.Handle(httptest.NewRecorder(), request)); code != http.StatusForbidden {
		t.Fatalf("bad-origin status = %d", code)
	}
	if !tickets.Pending(viewer.ID(), now) {
		t.Fatal("origin rejection consumed the ticket")
	}

	request = attachmentRequest(offer.Ticket, "https://neko.example")
	request.URL.RawQuery = "ticket=forbidden"
	if code := httpErrorCode(t, controller.Handle(httptest.NewRecorder(), request)); code != http.StatusBadRequest {
		t.Fatalf("query status = %d", code)
	}
	if !tickets.Pending(viewer.ID(), now) {
		t.Fatal("query rejection consumed the ticket")
	}

	request = attachmentRequest(offer.Ticket, "https://neko.example")
	recorder := httptest.NewRecorder()
	if err := controller.Handle(recorder, request); err != nil {
		t.Fatalf("valid pre-upgrade request error = %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("malformed WebSocket upgrade status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if tickets.Pending(viewer.ID(), now) {
		t.Fatal("valid pre-upgrade request did not atomically redeem the ticket")
	}
	if code := httpErrorCode(t, controller.Handle(httptest.NewRecorder(), attachmentRequest(offer.Ticket, "https://neko.example"))); code != http.StatusGone {
		t.Fatalf("ticket replay status = %d", code)
	}

	unknown := "QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFC"
	if code := httpErrorCode(t, controller.Handle(httptest.NewRecorder(), attachmentRequest(unknown, "https://neko.example"))); code != http.StatusUnauthorized {
		t.Fatalf("unknown ticket status = %d", code)
	}

	offer = issue()
	if err := sessions.Update(viewer.ID(), types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: false}); err != nil {
		t.Fatal(err)
	}
	if code := httpErrorCode(t, controller.Handle(httptest.NewRecorder(), attachmentRequest(offer.Ticket, "https://neko.example"))); code != http.StatusForbidden {
		t.Fatalf("CanWatch=false status = %d", code)
	}
}

func TestControllerTransportAddressAndOriginPolicy(t *testing.T) {
	controller, _, _, _ := newControllerTest(t, ControllerConfig{TrustedProxies: []string{"127.0.0.0/8"}})
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/media/ws", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	if secure, _ := controller.secureTransport(request, parseAddress(request.RemoteAddr), true); secure {
		t.Fatal("cleartext loopback was allowed without the explicit development option")
	}

	controller.allowInsecureLoopback = true
	if secure, scheme := controller.secureTransport(request, parseAddress(request.RemoteAddr), true); !secure || scheme != "http" {
		t.Fatalf("explicit loopback transport = %v/%q", secure, scheme)
	}
	request.Header.Set("X-Forwarded-For", "203.0.113.7")
	if secure, _ := controller.secureTransport(request, parseAddress(request.RemoteAddr), true); secure {
		t.Fatal("forwarded cleartext request was treated as direct loopback development")
	}

	request = httptest.NewRequest(http.MethodGet, "http://neko.example/api/media/ws", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("X-Forwarded-For", "203.0.113.7, 127.0.0.1")
	address, peer, trusted := controller.clientAddress(request)
	if address != "203.0.113.7" || !trusted || !peer.IsLoopback() {
		t.Fatalf("trusted proxy address = %q/%v/%v", address, peer, trusted)
	}
	if secure, scheme := controller.secureTransport(request, peer, trusted); !secure || scheme != "https" {
		t.Fatalf("trusted proxy transport = %v/%q", secure, scheme)
	}
	if !controller.allowOrigin("https://neko.example", "https", "neko.example") || controller.allowOrigin("https://NEKO.example", "https", "neko.example") {
		t.Fatal("same-origin comparison was not exact")
	}

	request.RemoteAddr = "198.51.100.8:1234"
	address, _, trusted = controller.clientAddress(request)
	if address != "198.51.100.8" || trusted {
		t.Fatalf("untrusted forwarded address = %q/%v", address, trusted)
	}
}

func TestControllerAttachAndInvalidAttemptBounds(t *testing.T) {
	controller, _, _, now := newControllerTest(t, ControllerConfig{MaxConnections: 2})
	if status := controller.reserve("viewer"); status != 0 {
		t.Fatalf("first reservation status = %d", status)
	}
	controller.markAttached("viewer")
	if status := controller.reserve("viewer"); status != 0 {
		t.Fatalf("replacement reservation status = %d", status)
	}
	controller.releaseConnection()
	if status := controller.reserve("viewer"); status != http.StatusConflict {
		t.Fatalf("overlapping replacement status = %d", status)
	}
	if status := controller.reserve("other"); status != 0 {
		t.Fatalf("second session reservation status = %d", status)
	}
	if status := controller.reserve("third"); status != http.StatusTooManyRequests {
		t.Fatalf("connection-limit status = %d", status)
	}
	controller.releaseReservation("viewer")
	controller.releaseReservation("other")

	for attempt := 0; attempt < InvalidAttachesPerMinute; attempt++ {
		if controller.invalidBlocked("203.0.113.9", now) {
			t.Fatalf("invalid attempt %d blocked before the fixed limit", attempt)
		}
		controller.recordInvalid("203.0.113.9", now)
	}
	if !controller.invalidBlocked("203.0.113.9", now) {
		t.Fatal("invalid attachment limit was not enforced")
	}
	if controller.invalidBlocked("203.0.113.9", now.Add(time.Minute)) {
		t.Fatal("invalid attachment window did not reset")
	}
}

func TestControllerRejectsInvalidConfiguration(t *testing.T) {
	for _, controllerConfig := range []ControllerConfig{
		{MaxConnections: MaximumConnections + 1},
		{AllowedOrigins: []string{"https://neko.example/"}},
		{AllowedOrigins: []string{"https://neko.example?"}},
		{AllowedOrigins: []string{"HTTPS://neko.example"}},
		{TrustedProxies: []string{"not-an-address"}},
	} {
		if _, err := NewController(session.New(&config.Session{}), &controllerTestDeliveries{}, NewTicketStore(), controllerConfig); err == nil {
			t.Fatalf("invalid controller configuration accepted: %#v", controllerConfig)
		}
	}
}
