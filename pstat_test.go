package main

import (
	"reflect"
	"testing"
)

func assertEqual[T any](t *testing.T, expected, current T) {
	t.Helper()

	ok := reflect.DeepEqual(expected, current)
	if ok {
		return
	}

	t.Fatalf("expected: %+v, current %+v:", expected, current)
}

func TestIOData(t *testing.T) {
	tests := []struct {
		name    string
		cmd     []string
		assertF func(t *testing.T, s stats)
	}{
		{
			name: "timestamp not zero and elapsed between 0 and 1",
			cmd:  []string{"sleep", "0"},
			assertF: func(t *testing.T, s stats) {
				assertEqual(t, true, func() bool {
					elapsed := s.End - s.Start

					isTimestampInExpectedRange := true &&
						s.End != 0 &&
						elapsed >= 0 &&
						elapsed <= 1

					if !isTimestampInExpectedRange {
						t.Fatalf("unexpected " +
							"timestamp range")
					}

					return true
				}())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statsData := pstat(t.Context(), test.cmd)
			test.assertF(t, statsData)
		})
	}
}
