package main

import (
	"net"
	"testing"
)

// TestIPNetEqual tests the IPNetEqual function.
func TestIPNetEqual(t *testing.T) {
	tests := []struct {
		a     string
		b     string
		equal bool
	}{
		{
			a:     "10.0.0.0/8",
			b:     "10.0.0.0/8",
			equal: true,
		},
		{
			a:     "10.1.0.0/8",
			b:     "10.0.0.0/8",
			equal: true,
		},
		{
			a:     "10.0.0.0/8",
			b:     "10.0.0.0/16",
			equal: false,
		},
		{
			a:     "10.1.0.0/24",
			b:     "10.2.0.0/24",
			equal: false,
		},
	}

	for index, test := range tests {
		_, aIPNet, _ := net.ParseCIDR(test.a)
		_, bIPNet, _ := net.ParseCIDR(test.b)

		returnedEqual := IPNetEqual(*aIPNet, *bIPNet)

		if returnedEqual != test.equal {
			t.Fatalf(
				"%v: unexpected equal returned.\na: %v, b: %v\nexpected: %v, returned: %v",
				index,
				test.a,
				test.b,
				test.equal,
				returnedEqual,
			)
		}
	}
}

// TestNext tests the Next function.
func TestNext(t *testing.T) {
	tests := []struct {
		network         string
		expectedNetwork string
	}{
		// Test a simple case.
		{
			network:         "10.4.0.0/24",
			expectedNetwork: "10.4.1.0/24",
		},

		// Test another simple case.
		{
			network:         "10.4.1.0/24",
			expectedNetwork: "10.4.2.0/24",
		},

		// Test a case with a mask that splits octet boundaries.
		{
			network:         "10.4.0.0/26",
			expectedNetwork: "10.4.0.64/26",
		},

		// Test another case with a non-standard mask.
		{
			network:         "10.4.1.1/14",
			expectedNetwork: "10.8.0.0/14",
		},

		// Test giving IP that is inside a network.
		{
			network:         "10.4.255.0/24",
			expectedNetwork: "10.5.0.0/24",
		},

		// Test that we don't panic if at the end of the space.
		{
			network:         "255.255.255.0/24",
			expectedNetwork: "0.0.0.0/24",
		},
	}

	for index, test := range tests {
		_, network, _ := net.ParseCIDR(test.network)
		returnedNetwork := Next(*network)

		_, expected, _ := net.ParseCIDR(test.expectedNetwork)
		if !IPNetEqual(returnedNetwork, *expected) {
			t.Fatalf(
				"%v: unexpected network returned. \nexpected: %s (%#v, %#v) \nreturned: %s (%#v, %#v)",
				index,

				expected.String(),
				expected.IP,
				expected.Mask,

				returnedNetwork.String(),
				returnedNetwork.IP,
				returnedNetwork.Mask,
			)
		}
	}
}

// TestFree tests the Free function.
func TestFree(t *testing.T) {
	tests := []struct {
		network         string
		mask            int
		existing        []string
		expectedNetwork string
		expectedError   error
	}{
		// Test that a network with no existing subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{},
			expectedNetwork: "10.4.0.0/24",
			expectedError:   nil,
		},

		// Test that a network with one existing subnet returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.0.0/24"},
			expectedNetwork: "10.4.1.0/24",
			expectedError:   nil,
		},

		// Test that a network with two existing (non-fragmented) subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.0.0/24", "10.4.1.0/24"},
			expectedNetwork: "10.4.2.0/24",
			expectedError:   nil,
		},

		// Test that a network with an existing subnet, that is fragmented,
		// and can fit one network before, returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.1.0/24"},
			expectedNetwork: "10.4.0.0/24",
			expectedError:   nil,
		},

		// Test that a network with no existing subnets returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{},
			expectedNetwork: "10.4.0.0/26",
			expectedError:   nil,
		},

		// Test that a network with one existing subnet returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		// {
		// 	network:         "10.4.0.0/24",
		// 	mask:            26,
		// 	existing:        []string{"10.4.0.0/26"},
		// 	expectedNetwork: "10.4.0.64/26",
		// 	expectedError:   nil,
		// },
	}

	for index, test := range tests {
		_, network, _ := net.ParseCIDR(test.network)
		mask := net.CIDRMask(test.mask, 32)

		existing := []net.IPNet{}
		for _, e := range test.existing {
			_, n, _ := net.ParseCIDR(e)
			existing = append(existing, *n)
		}

		returnedNetwork, err := Free(*network, mask, existing)

		if err != test.expectedError {
			t.Fatalf("%v: unexpected error returned.\nexpected: %v, returned: %v", index, test.expectedError, err)
		}

		_, expected, _ := net.ParseCIDR(test.expectedNetwork)
		if !IPNetEqual(returnedNetwork, *expected) {
			t.Fatalf(
				"%v: unexpected network returned. \nexpected: %s (%#v, %#v) \nreturned: %s (%#v, %#v)",
				index,

				expected.String(),
				expected.IP,
				expected.Mask,

				returnedNetwork.String(),
				returnedNetwork.IP,
				returnedNetwork.Mask,
			)
		}
	}
}
