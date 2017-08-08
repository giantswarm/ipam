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

	tests := []struct {
		config               func() Config
		expectedErrorHandler func(error) bool
	}{
		// Test that the default config, with a network set,
		// returns a new IPAM service.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Network = testNetwork
				return c
			},
		},

		// Test that a config with a nil logger returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Logger = nil
				c.Network = testNetwork
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
				return c
			},
			expectedErrorHandler: IsInvalidConfig,
		},

		// Test that a config with an empty network returns an invalid config error.
		{
			config: func() Config {
				c := DefaultConfig()
				c.Network = nil
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

// TestKey tests the key function.
func TestKey(t *testing.T) {
	tests := []struct {
		network     string
		expectedKey string
	}{
		{
			network:     "10.4.0.0/16",
			expectedKey: "/ipam/subnet/10.4.0.0-16",
		},
		{
			network:     "192.168.1.0/24",
			expectedKey: "/ipam/subnet/192.168.1.0-24",
		},
	}

	for index, test := range tests {
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		returnedKey := key(*network)

		if returnedKey != test.expectedKey {
			t.Fatalf(
				"%v: returned key did not match expected key.\nexpected: %v\nreturned: %v\n",
				index,
				test.expectedKey,
				returnedKey,
			)
		}
	}
}

// TestNewSubnetAndDeleteSubnet tests that NewSubnet and DeleteSubnet methods work together correctly.
func TestNewSubnetAndDeleteSubnet(t *testing.T) {
	type step struct {
		// add is true if we create a subnet, false if we delete one.
		add bool
		// mask is the mask to use if creating a new subnet.
		mask int
		// subnetToDelete is the subnet to delete, if we are deleting one.
		subnetToDelete string
		// expectedSubnet is the subnet we expect, if we are creating one.
		expectedSubnet string
		// expectedErrorHandler is the error handler we expect in case of error,
		// if an error should be returned.
		expectedErrorHandler func(error) bool
	}

	tests := []struct {
		network string
		steps   []step
	}{
		// Test that adding a single subnet returns the correct subnet.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},

		// Test that adding three subnets returns the correct subnets.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.1.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.2.0/24",
				},
			},
		},

		// Test that adding two subnets with different mask sizes,
		// returns the correct subnets.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           25,
					expectedSubnet: "10.4.0.0/25",
				},
				{
					add:            true,
					mask:           26,
					expectedSubnet: "10.4.0.128/26",
				},
			},
		},

		// Test adding a network that is too large returns an error.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:                  true,
					mask:                 15,
					expectedErrorHandler: IsMaskTooBig,
				},
			},
		},

		// Test that adding a subnet, deleting it, and adding another subnet,
		// works correctly.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            false,
					subnetToDelete: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},

		// Test that adding two subnets, deleting the first one, then
		// adding a third larger subnet, and then a fourth of the original size,
		// works correctly.
		{
			network: "10.4.0.0/16",
			steps: []step{
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.1.0/24",
				},
				{
					add:            false,
					subnetToDelete: "10.4.0.0/24",
				},
				{
					add:            true,
					mask:           23,
					expectedSubnet: "10.4.2.0/23",
				},
				{
					add:            true,
					mask:           24,
					expectedSubnet: "10.4.0.0/24",
				},
			},
		},
	}

	for index, test := range tests {
		// Parse network.
		_, network, err := net.ParseCIDR(test.network)
		if err != nil {
			t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
		}

		// Create a new IPAM service.
		config := DefaultConfig()
		config.Network = network

		service, err := New(config)
		if err != nil {
			t.Fatalf("%v: error returned creating ipam service: %v", index, err)
		}

		for _, step := range test.steps {
			if step.add {
				mask := net.CIDRMask(step.mask, 32)

				returnedSubnet, err := service.NewSubnet(mask)
				if err != nil {
					if step.expectedErrorHandler != nil {
						if !step.expectedErrorHandler(err) {
							t.Fatalf("%v: incorrect error returned creating new subnet: %v", index, err)
						}
					} else {
						t.Fatalf("%v: unexpected error returned creating new subnet: %v", index, err)
					}
				} else {
					_, expectedSubnet, err := net.ParseCIDR(step.expectedSubnet)
					if err != nil {
						t.Fatalf("%v: error returned parsing expected subnet: %v", index, err)
					}

					if !returnedSubnet.IP.Equal(expectedSubnet.IP) || !bytes.Equal(returnedSubnet.Mask, expectedSubnet.Mask) {
						t.Fatalf(
							"%v: returned subnet did not match expected.\nexpected: %v\nreturned: %v\n",
							index,
							*expectedSubnet,
							returnedSubnet,
						)
					}
				}
			} else {
				_, subnetToDelete, err := net.ParseCIDR(step.subnetToDelete)
				if err != nil {
					t.Fatalf("%v: error returned parsing network cidr: %v", index, err)
				}

				if err := service.DeleteSubnet(*subnetToDelete); err != nil {
					if !step.expectedErrorHandler(err) {
						t.Fatalf("%v: unexpected error returned creating new subnet: %v", index, err)
					}
				}
			}
		}
	}
}
