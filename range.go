package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// IPNetEqual returns true if both IPNets given are equal.
func IPNetEqual(a, b net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

// Next takes an IPNet, and returns an IPNet that is contigous.
// e.g: given 10.4.0.0/24, the 'next' network is 10.4.1.0/24
func Next(network net.IPNet) (net.IPNet, error) {
	fmt.Printf("ip byte: %#v\n", network.IP)
	fmt.Printf("mask byte: %#v\n", network.Mask)

	ipInt := binary.BigEndian.Uint32(network.IP)
	maskInt := binary.BigEndian.Uint32(network.Mask)
	fmt.Printf("ip uint32: %v\n", ipInt)
	fmt.Printf("mask uint32: %v\n", maskInt)

	fmt.Printf("range: %v\n", ipInt^maskInt)

	return network, nil
}

// Free takes a network, a mask, and a list of networks.
// An available network, within the first network, is returned.
// fragmented is defined as not having a contigous set of ipnets
func Free(network net.IPNet, mask net.IPMask, existing []net.IPNet) (net.IPNet, error) {
	// TODO: check network is not nil.
	// TODO: check mask is not nil.
	// TODO: check existing is not nil.
	// TODO: check mask larger than network.

	// Do we assume existing is ordered?

	// Every IPNet we return will have the supplied mask, so this can be ignored.
	// We start the network IP at the network IP.
	n := net.IPNet{IP: network.IP, Mask: mask}

	// For every existing network
	for _, existingNetwork := range existing {
		// we check if the networks match (this assumes we only have networks of the same range)
		if !IPNetEqual(net.IPNet{IP: network.IP, Mask: mask}, existingNetwork) {
			// If they do not, this range is free
			return n, nil
		}

		// We then increment the network IP to the start of the next range
		// this currently is only correct for /24
		network.IP[2]++
	}

	// TODO: Assert that returned IPNet has mask matching 'mask'

	return n, nil
}
