package mediahls

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func readGolden(t *testing.T, name string) []byte { t.Helper(); data,err:=os.ReadFile("testdata/"+name); if err!=nil { t.Fatal(err) }; return data }

func TestMasterPlaylistGolden(t *testing.T) {
	playlist:=MasterPlaylist{AudioURI:"audio/index.m3u8",Variants:[]MasterVariant{
		{ID:"low",URI:"low/index.m3u8",Bandwidth:650000,AverageBandwidth:493000,Width:640,Height:360,FrameRate:15,VideoCodec:"avc1.64001f"},
		{ID:"high",URI:"high/index.m3u8",Bandwidth:4000000,AverageBandwidth:3128000,Width:1280,Height:720,FrameRate:25,VideoCodec:"avc1.64001f"},
		{ID:"medium",URI:"medium/index.m3u8",Bandwidth:1500000,AverageBandwidth:1228000,Width:854,Height:480,FrameRate:20,VideoCodec:"avc1.64001f"},
	}}
	actual,err:=playlist.Render(); if err!=nil { t.Fatal(err) }
	if expected:=readGolden(t,"master.m3u8"); !bytes.Equal(actual,expected) { t.Fatalf("master mismatch\nactual:\n%s\nexpected:\n%s",actual,expected) }
	text:=string(actual)
	for _,forbidden:=range []string{"http://","https://",testPublicID,"ticket","cookie"} { if strings.Contains(text,forbidden) { t.Fatalf("master contains %q",forbidden) } }
}

func goldenSegments() []Segment {
	start:=time.Date(2026,9,23,7,0,0,0,time.UTC)
	return []Segment{{URI:"seg-42.m4s",Duration:6,ProgramDateTime:start},{URI:"seg-43.m4s",Duration:6,ProgramDateTime:start.Add(6*time.Second)},{URI:"seg-44.m4s",Duration:6,ProgramDateTime:start.Add(12*time.Second)}}
}

func TestConventionalPlaylistGolden(t *testing.T) {
	actual,err:=(MediaPlaylist{Mode:ModeHLS,MediaSequence:42,DiscontinuitySequence:3,MapURI:"init-7.mp4",Segments:goldenSegments()}).Render()
	if err!=nil { t.Fatal(err) }
	if expected:=readGolden(t,"conventional.m3u8"); !bytes.Equal(actual,expected) { t.Fatalf("playlist mismatch\nactual:\n%s\nexpected:\n%s",actual,expected) }
}

func goldenParentParts(sequence int) []Part {
	parts:=make([]Part,0,6)
	for index:=0; index<6; index++ { parts=append(parts,Part{URI:fmt.Sprintf("part-%d-%d.m4s",sequence,index),Duration:1,Independent:index%2==0}) }
	return parts
}

func goldenLowLatencySegments() []Segment {
	segments:=goldenSegments()
	segments[1].Parts=goldenParentParts(43)
	segments[2].Parts=goldenParentParts(44)
	return segments
}

func TestLowLatencyPlaylistGolden(t *testing.T) {
	actual,err:=(MediaPlaylist{Mode:ModeLLHLS,MediaSequence:42,DiscontinuitySequence:3,MapURI:"init-7.mp4",Segments:goldenLowLatencySegments(),PartsProgramDateTime:time.Date(2026,9,23,7,0,18,0,time.UTC),Parts:[]Part{{URI:"part-45-0.m4s",Duration:1,Independent:true},{URI:"part-45-1.m4s",Duration:1},{URI:"part-45-2.m4s",Duration:1,Independent:true}},PreloadHint:"part-45-3.m4s",RenditionReports:[]RenditionReport{{URI:"../medium/index.m3u8",LastMSN:45,LastPart:2},{URI:"../low/index.m3u8",LastMSN:45,LastPart:2}}}).Render()
	if err!=nil { t.Fatal(err) }
	if expected:=readGolden(t,"low-latency.m3u8"); !bytes.Equal(actual,expected) { t.Fatalf("playlist mismatch\nactual:\n%s\nexpected:\n%s",actual,expected) }
}

func TestPlaylistRejectsAbsoluteCredentialsAndOversizedModel(t *testing.T) {
	if _,err:=(MasterPlaylist{AudioURI:"https://example/audio.m3u8",Variants:[]MasterVariant{{ID:"high",URI:"high/index.m3u8",Bandwidth:2,AverageBandwidth:1,Width:1,Height:1,FrameRate:1,VideoCodec:"avc1.64001f"}}}).Render(); err==nil { t.Fatal("absolute URI accepted") }
	if _,err:=(MediaPlaylist{Mode:ModeLLHLS,MediaSequence:1,MapURI:"init-1.mp4",Segments:goldenSegments(),PartsProgramDateTime:time.Now(),Parts:[]Part{{URI:"part-1-0.m4s",Duration:.5}},PreloadHint:"part-1-1.m4s"}).Render(); err==nil { t.Fatal("undersized normal part accepted") }
	if _,err:=(MediaPlaylist{Mode:ModeHLS,MediaSequence:1,MapURI:"seg-1.m4s",Segments:goldenSegments()}).Render(); err==nil { t.Fatal("media object accepted as init map") }
}

func TestPlaylistRejectsMisplacedParentParts(t *testing.T) {
	conventional:=goldenSegments(); conventional[2].Parts=goldenParentParts(44)
	if _,err:=(MediaPlaylist{Mode:ModeHLS,MediaSequence:42,MapURI:"init-7.mp4",Segments:conventional}).Render(); err==nil { t.Fatal("conventional playlist accepted parent parts") }
	incomplete:=goldenLowLatencySegments(); incomplete[2].Parts=incomplete[2].Parts[:5]
	if _,err:=(MediaPlaylist{Mode:ModeLLHLS,MediaSequence:42,MapURI:"init-7.mp4",Segments:incomplete,PartsProgramDateTime:time.Date(2026,9,23,7,0,18,0,time.UTC),PreloadHint:"part-45-0.m4s"}).Render(); err==nil { t.Fatal("incomplete parent parts accepted") }
}
