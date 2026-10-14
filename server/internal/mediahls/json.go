package mediahls

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type BootstrapPayload struct {
	Ticket string `json:"ticket"`
}

var ErrBootstrapBodyTooLarge = errors.New("HLS bootstrap body exceeds 2 KiB")

func DecodeBootstrapPayload(data []byte) (BootstrapPayload, error) {
	if len(data) > MaximumBootstrapBody {
		return BootstrapPayload{}, ErrBootstrapBodyTooLarge
	}
	if len(data) == 0 {
		return BootstrapPayload{}, ErrTicketMalformed
	}
	payload := BootstrapPayload{}
	if err := DecodeStrictJSON(data, &payload); err != nil {
		return BootstrapPayload{}, err
	}
	if err := RequireJSONFields(data, "ticket"); err != nil {
		return BootstrapPayload{}, err
	}
	if err := ValidateTicketSyntax(payload.Ticket); err != nil {
		return BootstrapPayload{}, err
	}
	return payload, nil
}

func DecodeStrictJSON(data []byte, out any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func RequireJSONFields(data []byte, required ...string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range required {
		if _, ok := fields[field]; !ok {
			return fmt.Errorf("missing required JSON field %q", field)
		}
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return errors.New("invalid JSON object key")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = struct{}{}
				if err := walk(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walk(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		default:
			return errors.New("invalid JSON delimiter")
		}
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}
