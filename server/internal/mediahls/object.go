package mediahls

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

var (
	ErrInvalidGeneration = errors.New("invalid HLS generation")
	ErrInvalidObject     = errors.New("invalid HLS media object")
	ErrObjectExists      = errors.New("HLS media object already exists")
	ErrObjectLimit       = errors.New("HLS media object limit exceeded")
)

type RenditionFormat struct {
	ID string
	Codec string
	Width uint32
	Height uint32
	FrameRateNumerator uint32
	FrameRateDenominator uint32
	Timescale uint32
}

type Generation struct {
	ID uint64
	DiscontinuitySequence uint64
	StartedAt time.Time
	Formats []RenditionFormat
}

func NewGeneration(id, discontinuitySequence uint64, startedAt time.Time, formats []RenditionFormat) (Generation, error) {
	if id == 0 || startedAt.IsZero() || len(formats) == 0 { return Generation{}, ErrInvalidGeneration }
	seen := map[string]struct{}{}
	copyFormats := slices.Clone(formats)
	for _, format := range copyFormats {
		if format.ID != "audio" && format.ID != "high" && format.ID != "medium" && format.ID != "low" { return Generation{}, ErrInvalidGeneration }
		if format.Codec == "" || format.Timescale == 0 { return Generation{}, ErrInvalidGeneration }
		if _, exists := seen[format.ID]; exists { return Generation{}, ErrInvalidGeneration }
		seen[format.ID] = struct{}{}
		if format.ID == "audio" {
			if format.Codec != "mp4a.40.2" || format.Width != 0 || format.Height != 0 || format.Timescale != 48_000 { return Generation{}, ErrInvalidGeneration }
		} else if format.Codec != "avc1.64001f" || format.Width == 0 || format.Height == 0 || format.FrameRateNumerator == 0 || format.FrameRateDenominator == 0 || format.Timescale != 90_000 {
			return Generation{}, ErrInvalidGeneration
		}
	}
	return Generation{ID: id, DiscontinuitySequence: discontinuitySequence, StartedAt: startedAt, Formats: copyFormats}, nil
}

type ObjectKind uint8
const ( ObjectInit ObjectKind = iota + 1; ObjectPart; ObjectSegment )

type MediaObject struct {
	uri string
	variant string
	kind ObjectKind
	generation uint64
	sequence uint64
	part uint64
	contentType string
	data []byte
}

func NewMediaObject(uri, variant string, kind ObjectKind, generation, sequence, part uint64, contentType string, data []byte) (MediaObject, error) {
	if uri == "" || (variant != "audio" && variant != "high" && variant != "medium" && variant != "low") || generation == 0 || len(data) == 0 || (contentType != "video/mp4" && contentType != "audio/mp4") || (variant == "audio") != (contentType == "audio/mp4") { return MediaObject{}, ErrInvalidObject }
	maximum := 0
	expectedURI := ""
	switch kind {
	case ObjectInit:
		maximum = MaximumInitBytes
		if sequence != 0 || part != 0 { return MediaObject{}, ErrInvalidObject }
		expectedURI = fmt.Sprintf("init-%d.mp4", generation)
	case ObjectPart:
		maximum = MaximumPartBytes
		if sequence == 0 { return MediaObject{}, ErrInvalidObject }
		expectedURI = fmt.Sprintf("part-%d-%d.m4s", sequence, part)
	case ObjectSegment:
		maximum = MaximumSegmentBytes
		if sequence == 0 || part != 0 { return MediaObject{}, ErrInvalidObject }
		expectedURI = fmt.Sprintf("seg-%d.m4s", sequence)
	default:
		return MediaObject{}, ErrInvalidObject
	}
	if uri != expectedURI { return MediaObject{}, ErrInvalidObject }
	if len(data) > maximum { return MediaObject{}, ErrObjectLimit }
	return MediaObject{uri: uri, variant: variant, kind: kind, generation: generation, sequence: sequence, part: part, contentType: contentType, data: slices.Clone(data)}, nil
}

func (object MediaObject) URI() string { return object.uri }
func (object MediaObject) Variant() string { return object.variant }
func (object MediaObject) Kind() ObjectKind { return object.kind }
func (object MediaObject) Generation() uint64 { return object.generation }
func (object MediaObject) ContentType() string { return object.contentType }
func (object MediaObject) Size() int { return len(object.data) }
func (object MediaObject) Bytes() []byte { return slices.Clone(object.data) }

type objectCountKey struct { variant string; kind ObjectKind }
type objectKey struct { variant string; uri string }

type ObjectStore struct { mu sync.RWMutex; objects map[objectKey]MediaObject; counts map[objectCountKey]int; retained int }

func NewObjectStore() *ObjectStore { return &ObjectStore{objects: map[objectKey]MediaObject{}, counts: map[objectCountKey]int{}} }

func (store *ObjectStore) Publish(object MediaObject) error {
	if object.uri == "" || len(object.data) == 0 { return ErrInvalidObject }
	store.mu.Lock(); defer store.mu.Unlock()
	storageKey := objectKey{variant: object.variant, uri: object.uri}
	if _, exists := store.objects[storageKey]; exists { return ErrObjectExists }
	if store.retained+len(object.data) > MaximumRetainedBytes { return ErrObjectLimit }
	countKey := objectCountKey{variant: object.variant, kind: object.kind}
	maximum := MaximumRetainedParents
	if object.kind == ObjectPart { maximum = MaximumRetainedParts }
	if object.kind == ObjectInit { maximum = MaximumRetainedInits }
	if store.counts[countKey] >= maximum { return ErrObjectLimit }
	store.objects[storageKey] = object
	store.counts[countKey]++
	store.retained += len(object.data)
	return nil
}

func (store *ObjectStore) Get(variant, uri string) (MediaObject, bool) {
	store.mu.RLock(); defer store.mu.RUnlock()
	object, ok := store.objects[objectKey{variant: variant, uri: uri}]
	if !ok { return MediaObject{}, false }
	object.data = slices.Clone(object.data)
	return object, true
}

func (store *ObjectStore) Remove(variant, uri string) bool {
	store.mu.Lock(); defer store.mu.Unlock()
	key := objectKey{variant: variant, uri: uri}
	object, ok := store.objects[key]
	if !ok { return false }
	delete(store.objects, key); store.retained -= len(object.data); countKey:=objectCountKey{variant:object.variant,kind:object.kind}; store.counts[countKey]--; if store.counts[countKey]==0{delete(store.counts,countKey)}; return true
}

func (store *ObjectStore) RetainedBytes() int { store.mu.RLock(); defer store.mu.RUnlock(); return store.retained }

func (object MediaObject) String() string { return fmt.Sprintf("%s/%s", object.variant, object.uri) }
