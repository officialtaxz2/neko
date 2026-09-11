package types

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMediaBackendNotFound      = errors.New("media backend not found")
	ErrMediaBackendAlreadyExists = errors.New("media backend already exists")
	ErrMediaDeliveryNotAllowed   = errors.New("media delivery not allowed")
	ErrMediaSourceNotFound       = errors.New("media source not found")
	ErrMediaSubscriptionClosed   = errors.New("media subscription closed")
)

type MediaKind string

const (
	MediaKindAudio MediaKind = "audio"
	MediaKindVideo MediaKind = "video"
)

type MediaCodec struct {
	Name       string
	MIMEType   string
	ClockRate  uint32
	Channels   uint16
	Parameters map[string]string
	Config     []byte
}

// MediaSource describes one encoded source. Bitrate is the currently measured
// rate and NominalBitrate is the optional configured selection reference.
// Generation changes whenever the underlying capture pipeline is recreated or
// its format changes.
type MediaSource struct {
	ID                   string
	Kind                 MediaKind
	Codec                MediaCodec
	Width                uint32
	Height               uint32
	FrameRateNumerator   uint32
	FrameRateDenominator uint32
	Bitrate              uint64
	NominalBitrate       uint64
	Generation           uint64
}

// CloneMediaSource protects the provider's immutable-publication rule for
// codec maps and initialization buffers.
func CloneMediaSource(source MediaSource) MediaSource {
	if source.Codec.Parameters != nil {
		source.Codec.Parameters = cloneStringMap(source.Codec.Parameters)
	}
	if source.Codec.Config != nil {
		source.Codec.Config = append([]byte(nil), source.Codec.Config...)
	}
	return source
}

func cloneStringMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

type EncodedMediaUnit struct {
	Generation uint64
	Sequence   uint64
	PTS        time.Duration
	DTS        time.Duration
	DTSValid   bool
	Duration   time.Duration
	Keyframe   bool
	// CapturedAt preserves the timestamp consumed by the compatibility WebRTC
	// sender. Cross-backend presentation uses PTS/DTS, never this arrival time.
	CapturedAt time.Time
	// Data is immutable for the lifetime of the event. Consumers that mutate or
	// retain it beyond event processing must copy it.
	Data []byte
}

type MediaEventType string

const (
	MediaEventTypeFormat        MediaEventType = "format"
	MediaEventTypeUnit          MediaEventType = "unit"
	MediaEventTypeDiscontinuity MediaEventType = "discontinuity"
	MediaEventTypeEnd           MediaEventType = "end"
)

type MediaDiscontinuity struct {
	Generation uint64
	Reason     string
}

type MediaEvent struct {
	Type          MediaEventType
	Source        MediaSource
	Unit          EncodedMediaUnit
	Discontinuity MediaDiscontinuity
}

type MediaSelectorType int

const (
	MediaSelectorTypeExact MediaSelectorType = iota
	MediaSelectorTypeNearest
	MediaSelectorTypeLower
	MediaSelectorTypeHigher
)

func (selector MediaSelectorType) String() string {
	switch selector {
	case MediaSelectorTypeExact:
		return "exact"
	case MediaSelectorTypeNearest:
		return "nearest"
	case MediaSelectorTypeLower:
		return "lower"
	case MediaSelectorTypeHigher:
		return "higher"
	default:
		return fmt.Sprintf("%d", int(selector))
	}
}

func (selector *MediaSelectorType) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "exact", "":
		*selector = MediaSelectorTypeExact
	case "nearest":
		*selector = MediaSelectorTypeNearest
	case "lower":
		*selector = MediaSelectorTypeLower
	case "higher":
		*selector = MediaSelectorTypeHigher
	default:
		return fmt.Errorf("invalid stream selector type: %s", string(text))
	}
	return nil
}

func (selector MediaSelectorType) MarshalText() ([]byte, error) {
	return []byte(selector.String()), nil
}

type MediaSelector struct {
	Type    MediaSelectorType `json:"type"`
	ID      string            `json:"id"`
	Bitrate uint64            `json:"bitrate"`
}

// SelectMediaSource applies the ordered source-selection semantics used by the
// existing adaptive WebRTC path. Sources must be supplied highest-to-lowest.
func SelectMediaSource(sources []MediaSource, selector MediaSelector) (MediaSource, bool) {
	if len(sources) == 0 {
		return MediaSource{}, false
	}

	if selector.ID != "" {
		for index, source := range sources {
			if source.ID != selector.ID {
				continue
			}
			switch selector.Type {
			case MediaSelectorTypeLower:
				if index+1 < len(sources) {
					return CloneMediaSource(sources[index+1]), true
				}
			case MediaSelectorTypeHigher:
				if index > 0 {
					return CloneMediaSource(sources[index-1]), true
				}
			default:
				return CloneMediaSource(source), true
			}
			return MediaSource{}, false
		}
		return MediaSource{}, false
	}

	if selector.Bitrate == 0 {
		return MediaSource{}, false
	}

	available := make([]MediaSource, 0, len(sources))
	for _, source := range sources {
		if source.Bitrate != 0 {
			available = append(available, source)
		}
	}

	if selector.Type == MediaSelectorTypeNearest {
		if len(available) == 0 {
			// Preserve the current selector's fallback to the first configured
			// (highest-priority) source when no measured rate exists yet.
			return CloneMediaSource(sources[0]), true
		}
		var below *MediaSource
		var above *MediaSource
		for index := range available {
			source := &available[index]
			if source.Bitrate <= selector.Bitrate {
				if below == nil || source.Bitrate > below.Bitrate {
					below = source
				}
			} else if above == nil || source.Bitrate < above.Bitrate {
				above = source
			}
		}
		// Preserve the legacy bias toward a stream that fits below the target,
		// even when an above-target stream is numerically closer.
		if below != nil {
			return CloneMediaSource(*below), true
		}
		return CloneMediaSource(*above), true
	}

	switch selector.Type {
	case MediaSelectorTypeLower:
		for index := len(sources) - 1; index >= 0; index-- {
			source := sources[index]
			if source.Bitrate != 0 && source.Bitrate < selector.Bitrate {
				return CloneMediaSource(source), true
			}
		}
	case MediaSelectorTypeHigher:
		for _, source := range sources {
			if source.Bitrate != 0 && source.Bitrate > selector.Bitrate {
				return CloneMediaSource(source), true
			}
		}
	default:
		for _, source := range sources {
			if source.Bitrate == selector.Bitrate {
				return CloneMediaSource(source), true
			}
		}
	}

	return MediaSource{}, false
}

type MediaOverflowPolicy string

const (
	MediaOverflowDropNewest MediaOverflowPolicy = "drop_newest"
)

type MediaDrop struct {
	Backend  string
	Kind     MediaKind
	SourceID string
	Reason   string
}

type MediaSubscriptionObserver interface {
	OnMediaSubscriptionDrop(MediaDrop)
}

type SourceSubscriptionRequest struct {
	Kind           MediaKind
	Selector       MediaSelector
	QueueCapacity  int
	OverflowPolicy MediaOverflowPolicy
	Backend        string
	Observer       MediaSubscriptionObserver
}

type EncodedMediaProvider interface {
	Sources(kind MediaKind) []MediaSource
	Subscribe(context.Context, SourceSubscriptionRequest) (MediaSubscription, error)
}

type MediaSubscription interface {
	ID() string
	Source() MediaSource
	Events() <-chan MediaEvent
	Switch(context.Context, MediaSelector) error
	SetPaused(bool) error
	Close() error
}

type MediaBackendCapabilities struct {
	ReceiveAudio        bool
	ReceiveVideo        bool
	ServerSideSelection bool
	NativeAdaptive      bool
	PublishMedia        bool
}

type MediaBackendDescriptor struct {
	Name         string
	Capabilities MediaBackendCapabilities
}

type MediaDeliveryRequest struct {
	Backend string
	Audio   bool
	Video   bool
}

type MediaDeliveryState string

const (
	MediaDeliveryStateOpening MediaDeliveryState = "opening"
	MediaDeliveryStateActive  MediaDeliveryState = "active"
	MediaDeliveryStateFailed  MediaDeliveryState = "failed"
	MediaDeliveryStateClosed  MediaDeliveryState = "closed"
)

// MediaLease is a backend-scoped, receive-only grant. It deliberately exposes
// no login/share credential and no member, control, plugin or publish method.
type MediaLease interface {
	ID() string
	SessionID() string
	Backend() string
	SetState(MediaDeliveryState) bool
	Valid() bool
}

type MediaDelivery interface {
	ID() string
	SessionID() string
	Backend() string
	SetPaused(bool) error
	Close() error
	Done() <-chan struct{}
}

type MediaBackend interface {
	Name() string
	Capabilities() MediaBackendCapabilities
	Open(context.Context, MediaLease, MediaDeliveryRequest) (MediaDelivery, error)
}

type MediaDeliveryManager interface {
	Register(MediaBackend) error
	Backends() []MediaBackendDescriptor
	Open(context.Context, Session, MediaDeliveryRequest) (MediaDelivery, error)
	CloseSession(sessionID string)
	Shutdown() error
}
