package qcc

import (
	"strings"
	"testing"
)

func TestResolveArea(t *testing.T) {
	want := Area{ProvinceCode: "ZJ", Code: "330782"}
	for _, input := range []string{"义乌市", "浙江省 金华市 义乌市", "330782"} {
		t.Run(input, func(t *testing.T) {
			got, err := ResolveArea(input)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("ResolveArea(%q) = %#v, want %#v", input, got, want)
			}
		})
	}
}

func TestResolveAreaSuggestsSimilarName(t *testing.T) {
	_, err := ResolveArea("义乌")
	if err == nil || !strings.Contains(err.Error(), `did you mean "义乌市"?`) {
		t.Fatalf("ResolveArea() error = %v, want area suggestion", err)
	}
}
