package ipam

import (
	"bytes"
	"net"
	"testing"
)

// TestNew tests the New function.
func TestNew(t *testing.T) {
	// testNetwork is a network to set for testing.
	_, testNetwork, _ := net.ParseCIDR("10.4.0.0/16")
	// testMask is a mask to set for testing.
	testMask := net.CIDRMask(16, 32)

	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config, with a network and mask set,
		// returns a new IPAM service.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Network = testNetwork
				c.Mask = &testMask
				return c
			},
		},

		// Test that a config with a nil logger returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Logger = nil
				c.Network = testNetwork
				c.Mask = &testMask
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with a nil storage returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Storage = nil
				c.Network = testNetwork
				c.Mask = &testMask
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with an empty network returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Network = nil
				c.Mask = &testMask
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with an empty mask returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Network = testNetwork
				c.Mask = nil
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

// TestNewSubnet tests the NewSubnet method.
func TestNewSubnet(t *testing.T) {
	tests := []struct {
		network              string
		mask                 int
		expectedSubnets      []string
		expectedErrorHandler func(error) bool
	}{
		// Test that the first subnet is returned correctly.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			expectedSubnets: []string{"10.4.0.0/24"},
		},

		// Test that two subnets are returned correctly.
		{
			network: "10.4.0.0/16",
			mask:    24,
			expectedSubnets: []string{
				"10.4.0.0/24",
				"10.4.1.0/24",
			},
		},

		// Test that three subnets are returned correctly, with a different mask.
		{
			network: "10.4.0.0/18",
			mask:    25,
			expectedSubnets: []string{
				"10.4.0.0/25",
				"10.4.0.128/25",
				"10.4.1.0/25",
			},
		},

		// Test that errors are returned correctly.
		{
			network: "10.4.0.0/16",
			mask:    15,
			// expect one subnet, even though we're expecting an error, so we actually call `NewSubnet`,
			expectedSubnets:      []string{"10.4.0.0/16"},
			expectedErrorHandler: IsMaskTooBig,
		},
	}

	for index, test := range tests {
		// Parse network and mask.
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		mask := net.CIDRMask(test.mask, 32)

		// Create a new IPAM service.
		config := DefaultConfig()
		config.Network = network
		config.Mask = &mask

		service, err := New(config)
		if err != nil {
			t.Fatalf("%v: error returned creating ipam service: %v", index, err)
		}

		// Parse expected subnets.
		expectedSubnets := []net.IPNet{}
		for _, expectedSubnet := range test.expectedSubnets {
			_, subnet, err := net.ParseCIDR(expectedSubnet)
			if err != nil {
				t.Fatalf("%v: error returned parsing expected subnet: %v", index, err)
			}
			expectedSubnets = append(expectedSubnets, *subnet)
		}

		// For each expected subnet, test that it is what the IPAM service returns.
		for _, expectedSubnet := range expectedSubnets {
			returnedSubnet, err := service.NewSubnet()

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
				if !returnedSubnet.IP.Equal(expectedSubnet.IP) || !bytes.Equal(returnedSubnet.Mask, expectedSubnet.Mask) {
					t.Fatalf(
						"%v: returned subnet did not match expected.\nexpected: %v\nreturned: %v\n",
						index,
						expectedSubnet,
						returnedSubnet,
					)
				}
			}
		}
	}
}
