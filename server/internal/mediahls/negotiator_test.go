package mediahls

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	sessionmanager "github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

type negotiationPeer struct { events map[string][]any }
func newNegotiationPeer()*negotiationPeer{return &negotiationPeer{events:map[string][]any{}}}
func(peer *negotiationPeer)Send(eventName string,payload any){peer.events[eventName]=append(peer.events[eventName],payload)}
func(*negotiationPeer)Ping()error{return nil}
func(*negotiationPeer)Destroy(string){}
func(peer *negotiationPeer)last(eventName string)(any,bool){items:=peer.events[eventName];if len(items)==0{return nil,false};return items[len(items)-1],true}

func newNegotiationFixture(t *testing.T,canWatch bool)(*Negotiator,*sessionmanager.SessionManagerCtx,types.Session,*negotiationPeer,time.Time){
	t.Helper();sessions:=sessionmanager.New(&config.Session{});viewer,_,err:=sessions.Create("viewer",types.MemberProfile{Name:"viewer",CanLogin:true,CanConnect:true,CanWatch:canWatch});if err!=nil{t.Fatal(err)}
	peer:=newNegotiationPeer();viewer.ConnectWebSocketPeer(peer);now:=time.Unix(1_700_000_000,0)
	negotiator,err:=NewNegotiator(sessions,nil,Config{AllowedOrigins:[]string{"https://neko.example"},Modes:[]string{ModeHLS,ModeLLHLS},MaxLeases:MaximumLeases,MaxRequests:MaximumRequests});if err!=nil{t.Fatal(err)};negotiator.now=func()time.Time{return now}
	return negotiator,sessions,viewer,peer,now
}

func TestNegotiatorAdvertisesFixedCapabilitiesWithoutOpeningMedia(t *testing.T){
	negotiator,_,viewer,peer,_:=newNegotiationFixture(t,true)
	if !negotiator.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CAPABILITIES_REQUEST,Payload:json.RawMessage(`{"version":1,"mode":"ll-hls"}`)}){t.Fatal("event not handled")}
	raw,ok:=peer.last(event.MEDIA_HLS_CAPABILITIES);if !ok{t.Fatal("capabilities not sent")};capabilities,ok:=raw.(message.MediaHLSCapabilities);if !ok{t.Fatalf("type = %T",raw)}
	if capabilities.Backend!=BackendName||len(capabilities.Modes)!=2||len(capabilities.Variants)!=3||capabilities.Limits.MaxLeases!=MaximumLeases||capabilities.VideoCodec!="h264-high-3.1"{t.Fatalf("capabilities = %#v",capabilities)}
}

func TestNegotiatorCreatesBoundTicketAndRevokesAuthorization(t *testing.T){
	negotiator,sessions,viewer,peer,now:=newNegotiationFixture(t,true);request:=json.RawMessage(`{"version":1,"backend":"hls","mode":"ll-hls"}`)
	negotiator.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CREATE,Payload:request});raw,ok:=peer.last(event.MEDIA_HLS_OFFER);if !ok{t.Fatal("offer not sent")};offer:=raw.(message.MediaHLSOffer)
	if offer.Path!=BootstrapPath||offer.Mode!=ModeLLHLS||offer.ExpiresInMS!=TicketLifetime.Milliseconds(){t.Fatalf("offer = %#v",offer)}
	binding,err:=negotiator.Tickets().Redeem(offer.Ticket,now);if err!=nil{t.Fatal(err)};if binding.SessionID!=viewer.ID()||binding.Mode!=ModeLLHLS{t.Fatalf("binding = %#v",binding)}
	negotiator.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CREATE,Payload:request});raw,_=peer.last(event.MEDIA_HLS_OFFER);offer=raw.(message.MediaHLSOffer);profile:=viewer.Profile();profile.CanWatch=false;if err:=sessions.Update(viewer.ID(),profile);err!=nil{t.Fatal(err)}
	if _,err:=negotiator.Tickets().Redeem(offer.Ticket,now);!errors.Is(err,ErrTicketGone){t.Fatalf("revoked error = %v",err)}
}

func TestNegotiatorStrictPayloadAuthorizationAndRate(t *testing.T){
	denied,_,viewer,peer,_:=newNegotiationFixture(t,false);denied.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CAPABILITIES_REQUEST,Payload:json.RawMessage(`{"version":1,"mode":"hls"}`)});if _,ok:=peer.last(event.MEDIA_HLS_CAPABILITIES);ok{t.Fatal("non-watcher got capabilities")}
	negotiator,_,viewer,peer,_:=newNegotiationFixture(t,true);valid:=json.RawMessage(`{"version":1,"backend":"hls","mode":"hls"}`);for attempt:=0;attempt<BootstrapCreateBurst+1;attempt++{negotiator.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CREATE,Payload:valid})};if count:=len(peer.events[event.MEDIA_HLS_OFFER]);count!=BootstrapCreateBurst{t.Fatalf("offers = %d",count)}
	before:=len(peer.events[event.MEDIA_HLS_OFFER]);negotiator.Handler(viewer,types.WebSocketMessage{Event:event.MEDIA_HLS_CREATE,Payload:json.RawMessage(`{"version":1,"backend":"hls","mode":"hls","ticket":"forbidden"}`)});if len(peer.events[event.MEDIA_HLS_OFFER])!=before{t.Fatal("unknown field accepted")}
}

func TestNormalizeConfigRejectsUnsafeOrAmbiguousValues(t *testing.T){
	for _,candidate:=range []Config{{AllowedOrigins:nil,Modes:[]string{ModeHLS}},{AllowedOrigins:[]string{"http://neko.example"},Modes:[]string{ModeHLS}},{AllowedOrigins:[]string{"https://neko.example"},Modes:[]string{ModeHLS,ModeHLS}},{AllowedOrigins:[]string{"https://neko.example"},Modes:[]string{"auto"}}}{
		if _,err:=NormalizeConfig(candidate);err==nil{t.Fatalf("accepted config %#v",candidate)}
	}
}
