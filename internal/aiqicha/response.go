package aiqicha

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Based on the response interceptor in the official vendor bundle (module 26500):
// https://xinpub.cdn.bcebos.com/aiqicha/static/1793259024/js/vendor.92cf53d085f.js
// Only data values are encoded, including digits inside IDs, URLs and dates;
// object keys and envelope fields such as status must remain unchanged.
func decodeResponse(body io.Reader, result any) error {
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	var envelope map[string]any
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode aiqicha JSON response: %w", err)
	}
	var digits string
	switch envelope["ddw"] {
	case nil, json.Number("0"):
	case json.Number("1"):
		digits = "0123547698"
	case json.Number("2"):
		digits = "0123689457"
	default:
		return fmt.Errorf("aiqicha: unsupported ddw encoding %v", envelope["ddw"])
	}
	if digits != "" {
		envelope["data"] = decodeDigits(envelope["data"], digits)
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("decode aiqicha data: %w", err)
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("decode aiqicha result: %w", err)
	}
	return nil
}

func decodeDigits(value any, digits string) any {
	decode := func(text string) string {
		return strings.Map(func(character rune) rune {
			if character >= '0' && character <= '9' {
				return rune(digits[character-'0'])
			}
			return character
		}, text)
	}
	switch value := value.(type) {
	case string:
		return decode(value)
	case json.Number:
		return json.Number(decode(string(value)))
	case []any:
		for i := range value {
			value[i] = decodeDigits(value[i], digits)
		}
		return value
	case map[string]any:
		for key := range value {
			value[key] = decodeDigits(value[key], digits)
		}
		return value
	default:
		return value
	}
}
