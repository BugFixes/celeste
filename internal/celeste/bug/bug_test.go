package bug_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/bugfixes/celeste/internal/celeste/bug"
)

func TestConvertLevelFromString(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	defer func() {
		_ = sugar.Sync()
	}()

	tests := []struct {
		name    string
		request string
		expect  int
	}{
		{
			name:    "log string",
			request: "log",
			expect:  bug.GetLevelLog(),
		},
		{
			name:    "log int",
			request: "1",
			expect:  bug.GetLevelLog(),
		},
		{
			name:    "info string",
			request: "info",
			expect:  bug.GetLevelInfo(),
		},
		{
			name:    "info int",
			request: "2",
			expect:  bug.GetLevelInfo(),
		},
		{
			name:    "error string",
			request: "error",
			expect:  bug.GetLevelError(),
		},
		{
			name:    "error int",
			request: "3",
			expect:  bug.GetLevelError(),
		},
		{
			name:    "unknown string",
			request: "bob",
			expect:  bug.GetLevelUnknown(),
		},
		{
			name:    "unknown int",
			request: "99",
			expect:  bug.GetLevelUnknown(),
		},
		{
			name:    "blank level",
			request: "",
			expect:  bug.GetLevelUnknown(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := bug.ConvertLevelFromString(test.request, sugar)

			if passed := assert.Equal(t, test.expect, resp); !passed {
				t.Errorf("lookup expect: %v, got: %v", test.expect, resp)
			}
		})
	}
}
