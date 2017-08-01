package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
)

// IPToDecimal converts a net.IP to a uint32.
func IPToDecimal(ip net.IP) uint32 {
	t := ip
	if len(ip) == 16 {
		t = ip[12:16]
	}

	return binary.BigEndian.Uint32(t)
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
	return IPToDecimal(a.IP)+Size(a.Mask)+Size(mask) <= IPToDecimal(b.IP)
}

// Free takes a network, a mask, and a list of networks.
// An available network, within the first network, is returned.
// fragmented is defined as not having a contigous set of ipnets
func Free(network net.IPNet, mask net.IPMask, existing []net.IPNet) (net.IPNet, error) {
	// TODO: check mask larger than network.
	// TODO: test existing not ordered
	// TODO: test full

	numExisting := len(existing)

	// Define the initial network for the search as the original network,
	// with the mask we want.
	n := net.IPNet{IP: network.IP, Mask: mask}

	// If there is only one existing network,
	if numExisting == 1 {
		// Check that the existing network does not match the initial network,
		if network.IP.Equal(existing[0].IP) {
			// if it does, use the next one.
			n = Next(n)
		}
	}

	// If there is more than one existing network,
	if numExisting > 1 {
		// Loop over each network.
		for i := 0; i < numExisting; i++ {
			// Advance to the next available network,
			// taking care to advance to the end of the current network
			// being checked, instead of just incrementing the network.
			n = net.IPNet{IP: Next(existing[i]).IP, Mask: mask}

			// If we have one more network ahead of the search.
			if i < numExisting-1 {
				// Check if we can fit n between the two networks.
				if Space(existing[i], existing[i+1], mask) {
					// And quit if we can.
					break
				}
			}
		}
	}

	if !bytes.Equal(mask, n.Mask) {
		panic("mask incorrect size") // TODO: microerror
	}

	return n, nil
}
