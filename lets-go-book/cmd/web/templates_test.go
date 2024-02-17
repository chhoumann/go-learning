package main

import (
	"testing"
	"time"

	"github.com/chhoumann/snippetbox/internal/assert"
)

func TestHumanDate(t *testing.T) {

	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2024, time.February, 17, 13, 46, 0, 0, time.UTC),
			want: "17 Feb 2024 at 13:46",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2024, time.February, 17, 13, 46, 0, 0, time.FixedZone("CET", 60*60)),
			want: "17 Feb 2024 at 12:46", // CET is 1 hour ahead of UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanDate(tt.tm)

			assert.Equal(t, got, tt.want)
		})
	}
}
