package mediaws

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func FuzzParseRecord(f *testing.F) {
	raw, err := os.ReadFile("testdata/neko_media_v1_golden.json")
	if err == nil {
		fixture := goldenFixture{}
		if json.Unmarshal(raw, &fixture) == nil {
			for _, item := range fixture.Records {
				if encoded, err := hex.DecodeString(item.Hex); err == nil {
					f.Add(encoded)
				}
			}
		}
	}
	f.Add([]byte("NEKO"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseRecord(data)
	})
}
