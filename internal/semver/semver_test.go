package semver_test

import (
	"testing"

	"ttp.sh/frost/internal/semver"
)

func TestVersion(t *testing.T) {
	testCases := []struct {
		desc     string
		expected string
	}{
		{
			desc:     "1.0.0",
			expected: "1.0.0.0",
		},
		{
			desc:     "v1.0.0",
			expected: "1.0.0.0",
		},
		{
			desc:     "dev-master",
			expected: "9999999-dev",
		},
		{
			desc:     "2",
			expected: "2.0.0.0",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			v, err := semver.NewVersion(tC.desc)
			if err != nil {
				t.Errorf("got error: %v", err)
			} else if v.String() != tC.expected {

				t.Errorf("expected: %s got %s", tC.expected, v.String())
			}
		})
	}
}
