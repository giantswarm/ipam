package main

import (
	"bytes"
	"net"
	"testing"
)

func ipNetEqual(a, b net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

// TestIPToDecimal tests the IPToDecimal function.
func TestIPToDecimal(t *testing.T) {
	tests := []struct {
		ip              string
		expectedDecimal uint32
	}{
		{
			ip:              "10.4.0.0",
			expectedDecimal: 168034304,
		},
	}

	for index, test := range tests {
		ip := net.ParseIP(test.ip)

		returnedDecimal := IPToDecimal(ip)

		if returnedDecimal != test.expectedDecimal {
			t.Fatalf(
				"%v: unexpected decimal returned.\nexpected: %v, returned: %v",
				index,
				test.expectedDecimal,
				returnedDecimal,
			)
		}
	}
}

// TestDecimalToIP tests the DecimalToIP function.
func TestDecimalToIP(t *testing.T) {
	tests := []struct {
		decimal    uint32
		expectedIP string
	}{
		{
			decimal:    168034304,
			expectedIP: "10.4.0.0",
		},
	}

	for index, test := range tests {
		returnedIP := DecimalToIP(test.decimal)

		expectedIP := net.ParseIP(test.expectedIP)
		if !returnedIP.Equal(expectedIP) {
			t.Fatalf(
				"%v: unexpected decimal returned.\nexpected: %v, returned: %v",
				index,
				expectedIP,
				returnedIP,
			)
		}
	}
}

// TestSize tests the Size function.
func TestSize(t *testing.T) {
	tests := []struct {
		mask         int
		expectedSize uint32
	}{
		{
			mask:         23,
			expectedSize: 512,
		},

		{
			mask:         24,
			expectedSize: 256,
		},

		{
			mask:         25,
			expectedSize: 128,
		},
	}

	for index, test := range tests {
		mask := net.CIDRMask(test.mask, 32)

		returnedSize := Size(mask)

		if returnedSize != test.expectedSize {
			t.Fatalf(
				"%v: unexpected size returned.\nexpected: %v, returned: %v",
				index,
				test.expectedSize,
				returnedSize,
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
		if !ipNetEqual(returnedNetwork, *expected) {
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

// TestSpace tests the Space function.
func TestSpace(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		mask     int
		expected bool
	}{
		// Test a simple case.
		{
			a:        "10.4.0.0/24",
			b:        "10.4.1.0/24",
			mask:     24,
			expected: false,
		},

		// Test a second simple case.
		{
			a:        "10.4.0.0/24",
			b:        "10.4.2.0/24",
			mask:     24,
			expected: true,
		},

		// Test a case where the mask is bigger than the space.
		{
			a:        "10.4.0.0/24",
			b:        "10.4.2.0/24",
			mask:     23,
			expected: false,
		},
	}

	for index, test := range tests {
		_, a, _ := net.ParseCIDR(test.a)
		_, b, _ := net.ParseCIDR(test.b)
		mask := net.CIDRMask(test.mask, 32)

		hasSpace := Space(*a, *b, mask)

		if hasSpace != test.expected {
			t.Fatalf(
				"%v: unexpected has space returned.\na: %v, b: %v, mask: %v\nexpected: %v, returned: %v",
				index,
				test.a,
				test.b,
				test.mask,
				test.expected,
				hasSpace,
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
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{"10.4.0.0/26"},
			expectedNetwork: "10.4.0.64/26",
			expectedError:   nil,
		},

		// Test that a network with two existing fragmented subnets,
		// with a mask that does not fall on an octet boundary, returns the correct subnet.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{"10.4.0.0/26", "10.4.0.128/26"},
			expectedNetwork: "10.4.0.64/26",
			expectedError:   nil,
		},

		// Test a setup with multiple, fragmented networks, of different sizes.
		// {
		// 	network: "10.4.0.0/24",
		// 	mask:    29,
		// 	existing: []string{
		// 		"10.4.0.0/26",
		// 		"10.4.0.64/28",
		// 		"10.4.0.80/28",
		// 		"10.4.0.112/28",
		// 		"10.4.0.128/26",
		// 	},
		// 	expectedNetwork: "10.4.0.96/29",
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
		if !ipNetEqual(returnedNetwork, *expected) {
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
