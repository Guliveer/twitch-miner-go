package gql

import (
	"context"
	"errors"
	"testing"
)

func TestIsRetryableGQLError(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "service timeout", msg: "service timeout", want: true},
		{name: "temporarily unavailable", msg: "Temporarily unavailable", want: true},
		{name: "timed out", msg: "upstream timed out", want: true},
		{name: "integrity failure", msg: "failed integrity check", want: true},
		{name: "validation error", msg: "field does not exist", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableGQLError(tt.msg); got != tt.want {
				t.Fatalf("isRetryableGQLError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestIsTransientError(t *testing.T) {
	if !IsTransientError(wrapTransientGQLError("VideoPlayerStreamInfoOverlayChannel", "service timeout")) {
		t.Fatal("expected retryable GQL error to be transient")
	}

	if !IsTransientError(context.DeadlineExceeded) {
		t.Fatal("expected deadline exceeded to be transient")
	}

	if IsTransientError(errors.New("permanent failure")) {
		t.Fatal("did not expect arbitrary errors to be transient")
	}
}
