package main

import (
	"bytes"
	"net"
	"reflect"
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
		{
			ip:              "255.255.255.255",
			expectedDecimal: 4294967295,
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

// TestRanges tests the Boundaries function.
func TestBoundaries(t *testing.T) {
	tests := []struct {
		network            string
		existing           []string
		expectedBoundaries []uint32
	}{
		// Test an empty set of networks returns the IPs for the
		// start and end of the overall network.
		{
			network:  "10.4.0.0/16",
			existing: []string{},
			expectedBoundaries: []uint32{
				168034304,
				168099839,
			},
		},

		// Test one net at the start of the overall network returns the
		// IPs for the start and end of the overall network,
		// with the smaller network in the middle.
		{
			network:  "10.4.0.0/16",
			existing: []string{"10.4.0.0/24"},
			expectedBoundaries: []uint32{
				168034304,
				168034304,
				168034559,
				168099839,
			},
		},

		// Test a fragmented network, with one subnet.
		{
			network:  "10.4.0.0/16",
			existing: []string{"10.4.1.0/24"},
			expectedBoundaries: []uint32{
				168034304,
				168034560,
				168034815,
				168099839,
			},
		},

		// Test two, different sized, fragment subnets.
		// {
		// 	network:  "10.4.0.0/8",
		// 	existing: []string{"10.4.1.0/25", "10.4.9.0/30"},
		// 	expectedBoundaries: []uint32{
		// 		168034304,
		// 		168034560,
		// 		168034687,
		// 		168036608,
		// 		168036611,
		// 		184549375,
		// 	},
		// },
	}

	for index, test := range tests {
		_, network, _ := net.ParseCIDR(test.network)

		existing := []net.IPNet{}
		for _, e := range test.existing {
			_, n, _ := net.ParseCIDR(e)
			existing = append(existing, *n)
		}

		// TODO: test errors
		returnedBoundaries, _ := Boundaries(*network, existing)

		if !reflect.DeepEqual(returnedBoundaries, test.expectedBoundaries) {
			t.Fatalf(
				"%v: unexpected boundaries returned.\nexpected: %v\nreturned: %v\n",
				index,
				test.expectedBoundaries,
				returnedBoundaries,
			)
		}
	}
}

// TestFree tests the Free function.
func TestFree(t *testing.T) {
	tests := []struct {
		network              string
		mask                 int
		existing             []string
		expectedNetwork      string
		expectedErrorHandler func(error) bool
	}{
		// Test that a network with no existing subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{},
			expectedNetwork: "10.4.0.0/24",
		},

		// Test that a network with one existing subnet returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.0.0/24"},
			expectedNetwork: "10.4.1.0/24",
		},

		// Test that a network with two existing (non-fragmented) subnets returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.0.0/24", "10.4.1.0/24"},
			expectedNetwork: "10.4.2.0/24",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// and can fit one network before, returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.1.0/24"},
			expectedNetwork: "10.4.0.0/24",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// and can fit one network before, returns the correct subnet,
		// given a smaller mask.
		{
			network:         "10.4.0.0/16",
			mask:            25,
			existing:        []string{"10.4.1.0/24"},
			expectedNetwork: "10.4.0.0/25",
		},

		// Test that a network with an existing subnet, that is fragmented,
		// but can't fit the requested network size before, returns the correct subnet.
		{
			network:         "10.4.0.0/16",
			mask:            23,
			existing:        []string{"10.4.1.0/24"}, // 10.4.1.0 - 10.4.1.255
			expectedNetwork: "10.4.2.0/23",           // 10.4.2.0 - 10.4.3.255
		},

		// Test that a network with no existing subnets returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{},
			expectedNetwork: "10.4.0.0/26",
		},

		// Test that a network with one existing subnet returns the correct subnet,
		// for a mask that does not fall on an octet boundary.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{"10.4.0.0/26"},
			expectedNetwork: "10.4.0.64/26",
		},

		// Test that a network with two existing fragmented subnets,
		// with a mask that does not fall on an octet boundary, returns the correct subnet.
		{
			network:         "10.4.0.0/24",
			mask:            26,
			existing:        []string{"10.4.0.0/26", "10.4.0.128/26"},
			expectedNetwork: "10.4.0.64/26",
		},

		// Test a setup with multiple, fragmented networks, of different sizes.
		{
			network: "10.4.0.0/24",
			mask:    29,
			existing: []string{
				"10.4.0.0/26",
				"10.4.0.64/28",
				"10.4.0.80/28",
				"10.4.0.112/28",
				"10.4.0.128/26",
			},
			expectedNetwork: "10.4.0.96/29",
		},

		// Test where a network the same size as the main network is requested.
		{
			network:         "10.4.0.0/16",
			mask:            16,
			existing:        []string{},
			expectedNetwork: "10.4.0.0/16",
		},

		// Test a setup where a network larger than the main network is requested.
		{
			network:              "10.4.0.0/16",
			mask:                 15,
			existing:             []string{},
			expectedErrorHandler: IsMaskTooBig,
		},

		// Test where the existing networks are not ordered.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.1.0/24", "10.4.0.0/24"},
			expectedNetwork: "10.4.2.0/24",
		},

		// Test where the existing networks are fragmented, and not ordered.
		{
			network:         "10.4.0.0/16",
			mask:            24,
			existing:        []string{"10.4.2.0/24", "10.4.0.0/24"},
			expectedNetwork: "10.4.1.0/24",
		},

		// Test where the range is full.
		// {
		// 	network:              "10.4.0.0/16",
		// 	mask:                 17,
		// 	existing:             []string{"10.4.0.0/17", "10.4.128.0/17"},
		// 	expectedErrorHandler: nil,
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

		if err != nil {
			if test.expectedErrorHandler == nil {
				t.Fatalf("%v: unexpected error returned.\nreturned: %v", index, err)
			} else {
				if !test.expectedErrorHandler(err) {
					t.Fatalf("%v: incorrect error returned.\nreturned: %v", index, err)
				}
			}
		} else {
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
}
