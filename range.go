package main

import (
	"encoding/binary"
	"math"
	"net"
)

// IPToDecimal converts a net.IP to a uint32.
func IPToDecimal(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

// DecimalToIP converts a uint32 to a net.IP.
func DecimalToIP(ip uint32) net.IP {
	t := make(net.IP, 4)
	binary.BigEndian.PutUint32(t, ip)

	return t
}

// Size takes a mask, and returns the number of addresses.
func Size(mask net.IPMask) uint32 {
	ones, _ := mask.Size()
	size := uint32(math.Pow(2, float64(32-ones)))

	return size
}

// Next takes an IPNet, and returns the IPNet that is contigous.
// e.g: given 10.4.0.0/24, the 'next' network is 10.4.1.0/24.
// In the rare edge case that the given network is at the end of the
// IPv4 address space (e.g: 255.255.255.0/24), the 'next' network
// will be at the start of the IPv4 address space (e.g: 0.0.0.0/24).
func Next(network net.IPNet) net.IPNet {
	next := DecimalToIP(
		IPToDecimal(network.IP) + Size(network.Mask),
	)

	nextNetwork := net.IPNet{
		IP:   next,
		Mask: network.Mask,
	}

	return nextNetwork
}

// Spaces takes two IPNets, and a mask. If a network of the size given by the
// mask would fit between the two supplied networks, true is returned,
// false otherwise.
// e.g: 10.4.0.0/24, 10.4.2.0/24, /24 has space.
func Space(a, b net.IPNet, mask net.IPMask) bool {
	aIPDecimal := binary.BigEndian.Uint32(a.IP)
	bIPDeciaml := binary.BigEndian.Uint32(b.IP)

	size := Size(mask)

	hasSpace := aIPDecimal+size < bIPDeciaml

	return hasSpace
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
		if !n.IP.Equal(existingNetwork.IP) {
			// If they do not, this range is free
			return n, nil
		}

		// We then increment the network IP to the start of the next range
		n = Next(n)
	}

	return n, nil
}
