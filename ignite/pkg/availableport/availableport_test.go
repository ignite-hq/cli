package availableport_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ignite/cli/ignite/pkg/availableport"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name    string
		n       uint
		options []availableport.Options
		err     error
	}{
		{
			name: "test 10 ports",
			n:    10,
		},
		{
			name: "invalid port range",
			n:    10,
			options: []availableport.Options{
				availableport.WithMinPort(5),
				availableport.WithMaxPort(1),
			},
			err: fmt.Errorf("invalid ports range: max < min (1 < 5)"),
		},
		{
			name: "invalid maximum port range",
			n:    10,
			options: []availableport.Options{
				availableport.WithMinPort(55001),
				availableport.WithMaxPort(1),
			},
			err: fmt.Errorf("invalid ports range: max < min (1 < 55001)"),
		},
		{
			name: "only invalid maximum port range",
			n:    10,
			options: []availableport.Options{
				availableport.WithMaxPort(43999),
			},
			err: fmt.Errorf("invalid ports range: max < min (43999 < 44000)"),
		},
		{
			name: "with randomizer",
			n:    100,
			options: []availableport.Options{
				availableport.WithRandomizer(rand.New(rand.NewSource(2023))),
				availableport.WithMinPort(100),
				availableport.WithMaxPort(200),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := availableport.Find(tt.n, tt.options...)
			if tt.err != nil {
				require.Equal(t, tt.err, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, int(tt.n))

			seen := make(map[uint]struct{})
			for _, val := range got {
				_, ok := seen[val]
				require.Falsef(t, ok, "duplicated port %d", val)
				seen[val] = struct{}{}
			}
		})
	}
}
