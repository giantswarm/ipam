package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
)

// IPNetEqual returns true if both IPNets given are equal.
func IPNetEqual(a, b net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

// Next takes an IPNet, and returns the IPNet that is contigous.
// e.g: given 10.4.0.0/24, the 'next' network is 10.4.1.0/24.
// In the rare edge case that the given network is at the end of the
// IPv4 address space (e.g: 255.255.255.0/24), the 'next' network
// will be at the start of the IPv4 address space (e.g: 0.0.0.0/24).
func Next(network net.IPNet) net.IPNet {
	// Calculate size of network.
	ones, _ := network.Mask.Size()
	addresses := uint32(math.Pow(2, float64(32-ones)))

	// Convert IP to decimal.
	ipDecimal := binary.BigEndian.Uint32(network.IP)

	// Add size of network to ip.
	startOfNextRangeDecimal := uint32(ipDecimal) + addresses

	// Convert decimal back to byte slice.
	startOfNextRangeIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(startOfNextRangeIP, startOfNextRangeDecimal)

	// Create new network with next IP, and original mask.
	nextNetwork := net.IPNet{
		IP:   startOfNextRangeIP,
		Mask: network.Mask,
	}

	return nextNetwork
}

// Free takes a network, a mask, and a list of networks.
// An available network, within the first network, is returned.
// fragmented is defined as not having a contigous set of ipnets
func Free(network net.IPNet, mask net.IPMask, existing []net.IPNet) (net.IPNet, error) {
	// TODO: check network is not nil.
	// TODO: check mask is not nil.
	// TODO: check existing is not nil.
	// TODO: check mask larger than network.

	// Do we assume existing is ordered? How do we order?

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
