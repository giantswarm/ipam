package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
	"sort"

	"github.com/giantswarm/microerror"
)

type IPNets []net.IPNet

func (s IPNets) Len() int {
	return len(s)
}

func (s IPNets) Less(i, j int) bool {
	return IPToDecimal(s[i].IP) < IPToDecimal(s[j].IP)
}

func (s IPNets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

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

// Boundaries takes a network, and a list of networks. A list of the IP addresses,
// as uint32 is returned. The start and end of the first supplied networks are the
// first and last entries, with each networks first and last IPs between.
func Boundaries(network net.IPNet, existing []net.IPNet) ([]uint32, error) {
	// TODO: Remove `Next` from here, we can have an actual function,
	// and drop `Next` entirely.

	boundaries := []uint32{
		IPToDecimal(network.IP),
	}

	for _, existingNetwork := range existing {
		start := IPToDecimal(existingNetwork.IP)
		end := IPToDecimal(Next(existingNetwork).IP)
		end--

		boundaries = append(boundaries, start, end)
	}

	end := IPToDecimal(Next(network).IP)
	end--
	boundaries = append(boundaries, end)

	// Invariant: There should be an even number of boundaries.
	if len(boundaries)%2 != 0 {
		panic("incorrect number of points") // TODO: microerror
	}

	return boundaries, nil
}

// Free takes a network, a mask, and a list of networks.
// An available network, within the first network, is returned.
// fragmented is defined as not having a contigous set of ipnets
func Free(network net.IPNet, mask net.IPMask, existing []net.IPNet) (net.IPNet, error) {
	// TODO: test full

	n := net.IPNet{IP: network.IP, Mask: mask}

	if Size(network.Mask) < Size(mask) {
		return n, microerror.Maskf(
			maskTooBigError, "have: %v, requested: %v", network.Mask, mask,
		)
	}

	sort.Sort(IPNets(existing))

	boundaries, err := Boundaries(network, existing)
	if err != nil {
		return n, microerror.Mask(err)
	}

	type pair struct {
		start uint32
		end   uint32
	}

	pairs := []pair{}
	for i := 0; i < len(boundaries)-1; i = i + 2 {
		pairs = append(pairs, pair{start: boundaries[i], end: boundaries[i+1]})
	}

	for _, pair := range pairs {
		if pair.end-pair.start >= Size(mask) {
			x := pair.start
			n.IP = DecimalToIP(x)
			break
		}
	}

	// Invariant: The IP of the network returned should not be nil.
	if n.IP == nil {
		return n, microerror.Mask(nilIPError)
	}

	// Invariant: The IP of the network returned should be contained
	// within the network supplied.
	if !network.Contains(n.IP) {
		return n, microerror.Maskf(
			ipNotContainedError, "%v is not contained by %v", n.IP, network,
		)
	}

	// Invariant: The mask of the network returned should be equal to
	// the mask supplied as an argument.
	if !bytes.Equal(mask, n.Mask) {
		return n, microerror.Maskf(
			maskIncorrectSizeError, "have: %v, requested: %v", n.Mask, mask,
		)
	}

	return n, nil
}
