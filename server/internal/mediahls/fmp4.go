package mediahls

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var ErrInvalidCodecConfig = errors.New("invalid HLS codec configuration")

type fragmentSample struct {
	data     []byte
	dts      uint64
	pts      uint64
	duration uint32
	keyframe bool
}

type fragmentMuxer struct {
	trackID   uint32
	timescale uint32
	sequence  uint32
}

func newFragmentMuxer(timescale uint32) *fragmentMuxer {
	return &fragmentMuxer{trackID: 1, timescale: timescale}
}

func (muxer *fragmentMuxer) fragment(samples []fragmentSample) ([]byte, error) {
	if muxer == nil || muxer.timescale == 0 || len(samples) == 0 {
		return nil, ErrInvalidObject
	}
	var payload bytes.Buffer
	for _, sample := range samples {
		if len(sample.data) == 0 || sample.duration == 0 || sample.pts < sample.dts || uint64(len(sample.data)) > uint64(^uint32(0)) {
			return nil, ErrInvalidObject
		}
		payload.Write(sample.data)
	}
	muxer.sequence++
	trun := muxer.trackRun(samples, 0)
	traf := mp4Box("traf",
		fullBox("tfhd", 0, 0x020000, u32(muxer.trackID)),
		fullBox("tfdt", 1, 0, u64(samples[0].dts)),
		trun,
	)
	moof := mp4Box("moof", fullBox("mfhd", 0, 0, u32(muxer.sequence)), traf)
	dataOffset := int32(len(moof) + 8)
	traf = mp4Box("traf",
		fullBox("tfhd", 0, 0x020000, u32(muxer.trackID)),
		fullBox("tfdt", 1, 0, u64(samples[0].dts)),
		muxer.trackRun(samples, dataOffset),
	)
	moof = mp4Box("moof", fullBox("mfhd", 0, 0, u32(muxer.sequence)), traf)
	return append(moof, mp4Box("mdat", payload.Bytes())...), nil
}

func (muxer *fragmentMuxer) trackRun(samples []fragmentSample, dataOffset int32) []byte {
	var payload bytes.Buffer
	payload.Write(u32(uint32(len(samples))))
	payload.Write(i32(dataOffset))
	for _, sample := range samples {
		payload.Write(u32(sample.duration))
		payload.Write(u32(uint32(len(sample.data))))
		if sample.keyframe {
			payload.Write(u32(0x02000000))
		} else {
			payload.Write(u32(0x01010000))
		}
		payload.Write(u32(uint32(sample.pts - sample.dts)))
	}
	return fullBox("trun", 0, 0x000f01, payload.Bytes())
}

func buildVideoInit(width, height uint32, codecConfig []byte) ([]byte, error) {
	if width == 0 || height == 0 || width > uint32(^uint16(0)) || height > uint32(^uint16(0)) ||
		len(codecConfig) < 7 || codecConfig[0] != 1 || codecConfig[1] != 100 || codecConfig[2] != 0 || codecConfig[3] != 31 || !validAVCDecoderConfiguration(codecConfig) {
		return nil, ErrInvalidCodecConfig
	}
	entry := make([]byte, 78)
	binary.BigEndian.PutUint16(entry[6:8], 1)
	binary.BigEndian.PutUint16(entry[24:26], uint16(width))
	binary.BigEndian.PutUint16(entry[26:28], uint16(height))
	binary.BigEndian.PutUint32(entry[28:32], 0x00480000)
	binary.BigEndian.PutUint32(entry[32:36], 0x00480000)
	binary.BigEndian.PutUint16(entry[40:42], 1)
	binary.BigEndian.PutUint16(entry[74:76], 0x0018)
	binary.BigEndian.PutUint16(entry[76:78], 0xffff)
	stsd := append(u32(1), mp4Box("avc1", entry, mp4Box("avcC", codecConfig))...)
	return buildInit(false, width, height, 90_000, fullBox("stsd", 0, 0, stsd)), nil
}

func validAVCDecoderConfiguration(config []byte) bool {
	if len(config) < 7 || config[4]&0x03 != 0x03 {
		return false
	}
	offset := 6
	spsCount := int(config[5] & 0x1f)
	if spsCount == 0 {
		return false
	}
	readNALUnits := func(count int) bool {
		for index := 0; index < count; index++ {
			if offset+2 > len(config) {
				return false
			}
			length := int(binary.BigEndian.Uint16(config[offset : offset+2]))
			offset += 2
			if length == 0 || offset+length > len(config) {
				return false
			}
			offset += length
		}
		return true
	}
	if !readNALUnits(spsCount) || offset >= len(config) {
		return false
	}
	ppsCount := int(config[offset])
	offset++
	return ppsCount > 0 && readNALUnits(ppsCount)
}

func buildAudioInit(codecConfig []byte) ([]byte, error) {
	if len(codecConfig) < 2 || codecConfig[0]>>3 != 2 || ((codecConfig[0]&0x07)<<1)|(codecConfig[1]>>7) != 3 || (codecConfig[1]>>3)&0x0f != 2 {
		return nil, ErrInvalidCodecConfig
	}
	entry := make([]byte, 28)
	binary.BigEndian.PutUint16(entry[6:8], 1)
	binary.BigEndian.PutUint16(entry[16:18], 2)
	binary.BigEndian.PutUint16(entry[18:20], 16)
	binary.BigEndian.PutUint32(entry[24:28], 48_000<<16)
	esds := fullBox("esds", 0, 0, esDescriptor(codecConfig))
	stsd := append(u32(1), mp4Box("mp4a", entry, esds)...)
	return buildInit(true, 0, 0, 48_000, fullBox("stsd", 0, 0, stsd)), nil
}

func buildInit(audio bool, width, height, timescale uint32, stsd []byte) []byte {
	volume := uint16(0)
	handler := "vide"
	handlerName := []byte("VideoHandler\x00")
	mediaHeader := fullBox("vmhd", 0, 1, make([]byte, 8))
	if audio {
		volume = 0x0100
		handler = "soun"
		handlerName = []byte("SoundHandler\x00")
		mediaHeader = fullBox("smhd", 0, 0, make([]byte, 4))
	}
	matrix := []byte{
		0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0x40, 0, 0, 0,
	}
	mvhdPayload := append(make([]byte, 8), u32(1_000)...)
	mvhdPayload = append(mvhdPayload, make([]byte, 4)...)
	mvhdPayload = append(mvhdPayload, u32(0x00010000)...)
	mvhdPayload = append(mvhdPayload, u16(0x0100)...)
	mvhdPayload = append(mvhdPayload, make([]byte, 10)...)
	mvhdPayload = append(mvhdPayload, matrix...)
	mvhdPayload = append(mvhdPayload, make([]byte, 24)...)
	mvhdPayload = append(mvhdPayload, u32(2)...)

	tkhdPayload := append(make([]byte, 8), u32(1)...)
	tkhdPayload = append(tkhdPayload, make([]byte, 4+4+8)...)
	tkhdPayload = append(tkhdPayload, u16(0)...)
	tkhdPayload = append(tkhdPayload, u16(0)...)
	tkhdPayload = append(tkhdPayload, u16(volume)...)
	tkhdPayload = append(tkhdPayload, u16(0)...)
	tkhdPayload = append(tkhdPayload, matrix...)
	tkhdPayload = append(tkhdPayload, u32(width<<16)...)
	tkhdPayload = append(tkhdPayload, u32(height<<16)...)

	mdhdPayload := append(make([]byte, 8), u32(timescale)...)
	mdhdPayload = append(mdhdPayload, make([]byte, 4)...)
	mdhdPayload = append(mdhdPayload, u16(0x55c4)...)
	mdhdPayload = append(mdhdPayload, u16(0)...)
	hdlrPayload := append(make([]byte, 4), []byte(handler)...)
	hdlrPayload = append(hdlrPayload, make([]byte, 12)...)
	hdlrPayload = append(hdlrPayload, handlerName...)

	url := fullBox("url ", 0, 1)
	dref := fullBox("dref", 0, 0, u32(1), url)
	dinf := mp4Box("dinf", dref)
	stbl := mp4Box("stbl", stsd, fullBox("stts", 0, 0, u32(0)), fullBox("stsc", 0, 0, u32(0)), fullBox("stsz", 0, 0, u32(0), u32(0)), fullBox("stco", 0, 0, u32(0)))
	minf := mp4Box("minf", mediaHeader, dinf, stbl)
	mdia := mp4Box("mdia", fullBox("mdhd", 0, 0, mdhdPayload), fullBox("hdlr", 0, 0, hdlrPayload), minf)
	trak := mp4Box("trak", fullBox("tkhd", 0, 7, tkhdPayload), mdia)
	trex := fullBox("trex", 0, 0, u32(1), u32(1), u32(0), u32(0), u32(0))
	moov := mp4Box("moov", fullBox("mvhd", 0, 0, mvhdPayload), trak, mp4Box("mvex", trex))
	ftyp := mp4Box("ftyp", []byte("iso6"), u32(1), []byte("iso6cmfcmp41"))
	return append(ftyp, moov...)
}

func esDescriptor(config []byte) []byte {
	decoderSpecific := descriptor(0x05, config)
	decoderPayload := []byte{0x40, 0x15, 0, 0, 0}
	decoderPayload = append(decoderPayload, u32(128_000)...)
	decoderPayload = append(decoderPayload, u32(128_000)...)
	decoderPayload = append(decoderPayload, decoderSpecific...)
	esPayload := []byte{0, 1, 0}
	esPayload = append(esPayload, descriptor(0x04, decoderPayload)...)
	esPayload = append(esPayload, descriptor(0x06, []byte{0x02})...)
	return descriptor(0x03, esPayload)
}

func descriptor(tag byte, payload []byte) []byte {
	length := len(payload)
	encoded := []byte{byte(length & 0x7f)}
	for length >>= 7; length > 0; length >>= 7 {
		encoded = append([]byte{byte(length&0x7f) | 0x80}, encoded...)
	}
	result := []byte{tag}
	result = append(result, encoded...)
	return append(result, payload...)
}

func mp4Box(kind string, payloads ...[]byte) []byte {
	size := 8
	for _, payload := range payloads {
		size += len(payload)
	}
	result := make([]byte, 8, size)
	binary.BigEndian.PutUint32(result[:4], uint32(size))
	copy(result[4:8], kind)
	for _, payload := range payloads {
		result = append(result, payload...)
	}
	return result
}

func fullBox(kind string, version byte, flags uint32, payloads ...[]byte) []byte {
	header := []byte{version, byte(flags >> 16), byte(flags >> 8), byte(flags)}
	return mp4Box(kind, append([][]byte{header}, payloads...)...)
}

func u16(value uint16) []byte { result := make([]byte, 2); binary.BigEndian.PutUint16(result, value); return result }
func u32(value uint32) []byte { result := make([]byte, 4); binary.BigEndian.PutUint32(result, value); return result }
func i32(value int32) []byte { return u32(uint32(value)) }
func u64(value uint64) []byte { result := make([]byte, 8); binary.BigEndian.PutUint64(result, value); return result }
