package cli

import (
	"context"
	"errors"
	"testing"
)

func TestSources(t *testing.T) {
	tests := []struct {
		name    searchSourceName
		src     search
		wantErr error
	}{
		{name: searchSourceQCC, src: newSearchQCC(), wantErr: errQCCSearchNotImplemented},
		{name: searchSourceAiqicha, src: newSearchAiqicha(), wantErr: errAiqichaSearchNotImplemented},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			_, err := tt.src.SearchEnterprises(context.Background(), "百度")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SearchEnterprises() error = %v, want %v", err, tt.wantErr)
			}

			_, err = tt.src.SearchPeople(context.Background(), "李彦宏")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SearchPeople() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
