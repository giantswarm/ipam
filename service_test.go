package ipam

import "testing"

// TestNew tests the New method.
func TestNew(t *testing.T) {
	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config returns a new IPAM service.
		{
			config: DefaultConfig,
		},

		// Test that a config with a nil logger returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Logger = nil
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with a nil storage returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Storage = nil
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that an empty config returns an invalid config error.
		{
			config:               func() Config { return Config{} },
			expectedErrorHandler: IsInvalidConfig,
		},
	}

	for index, test := range tests {
		service, err := New(test.config())

		if err == nil && test.expectedErrorHandler != nil {
			t.Fatalf("%v: expected error not returned", index)
		}
		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned: %v", index, err)
			} else {
				if !test.expectedErrorHandler(err) {
					t.Fatalf("%v: incorrect error returned: %v", index, err)
				}
			}
		} else {
			if service == nil {
				t.Fatalf("%v: service is nil", index)
			}
		}
	}
}
