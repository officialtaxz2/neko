package mediahls

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestGenerationRequiresCompleteDeterministicFormats(t *testing.T) {
	formats:=[]RenditionFormat{{ID:"audio",Codec:"mp4a.40.2",Timescale:48000},{ID:"high",Codec:"avc1.64001f",Width:1280,Height:720,FrameRateNumerator:25,FrameRateDenominator:1,Timescale:90000}}
	generation,err:=NewGeneration(7,3,time.Unix(1_700_000_000,0),formats)
	if err!=nil||generation.ID!=7 { t.Fatalf("generation = %#v, %v",generation,err) }
	formats[0].Codec="mutated"
	if generation.Formats[0].Codec!="mp4a.40.2" { t.Fatal("generation retained caller slice") }
	if _,err:=NewGeneration(8,3,time.Now(),[]RenditionFormat{{ID:"high",Codec:"avc1.64001f",Timescale:90000}}); !errors.Is(err,ErrInvalidGeneration) { t.Fatalf("incomplete format error = %v",err) }
}

func TestMediaObjectsAreImmutableUniqueAndBounded(t *testing.T) {
	original:=[]byte{1,2,3,4}
	object,err:=NewMediaObject("seg-42.m4s","high",ObjectSegment,7,42,0,"video/mp4",original)
	if err!=nil { t.Fatal(err) }
	original[0]=9
	if object.Bytes()[0]!=1 { t.Fatal("object retained caller bytes") }
	returned:=object.Bytes(); returned[0]=8
	if object.Bytes()[0]!=1 { t.Fatal("object exposed mutable bytes") }
	store:=NewObjectStore()
	if err:=store.Publish(object);err!=nil { t.Fatal(err) }
	if err:=store.Publish(object);!errors.Is(err,ErrObjectExists) { t.Fatalf("duplicate error = %v",err) }
	stored,ok:=store.Get("high","seg-42.m4s"); if !ok||stored.Size()!=4||store.RetainedBytes()!=4 { t.Fatalf("stored = %#v, ok=%v bytes=%d",stored,ok,store.RetainedBytes()) }
	if _,ok:=store.Get("medium","seg-42.m4s");ok{t.Fatal("object escaped its rendition")}
	medium,err:=NewMediaObject("seg-42.m4s","medium",ObjectSegment,7,42,0,"video/mp4",[]byte{5});if err!=nil{t.Fatal(err)};if err:=store.Publish(medium);err!=nil{t.Fatalf("same URI in another rendition rejected: %v",err)}
	if _,err:=NewMediaObject("part-42-0.m4s","high",ObjectPart,7,42,0,"video/mp4",make([]byte,MaximumPartBytes+1));!errors.Is(err,ErrObjectLimit) { t.Fatalf("size error = %v",err) }
	for index:=1;index<MaximumRetainedParents;index++{sequence:=42+uint64(index);next,err:=NewMediaObject(fmt.Sprintf("seg-%d.m4s",sequence),"high",ObjectSegment,7,sequence,0,"video/mp4",[]byte{1});if err!=nil{t.Fatal(err)};if err:=store.Publish(next);err!=nil{t.Fatal(err)}}
	overflow,err:=NewMediaObject("seg-99.m4s","high",ObjectSegment,7,99,0,"video/mp4",[]byte{1});if err!=nil{t.Fatal(err)};if err:=store.Publish(overflow);!errors.Is(err,ErrObjectLimit){t.Fatalf("retention error = %v",err)}
}
