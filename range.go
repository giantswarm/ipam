package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"net"
	"sort"

	"github.com/giantswarm/microerror"
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

// Range takes an IPNet, and returns the start and end IPs.
func Range(network net.IPNet) (uint32, uint32) {
	start := IPToDecimal(network.IP)
	end := start + Size(network.Mask) - 1

	return start, end
}

// Boundaries takes a network, and a list of networks. A list of the IP addresses,
// as uint32 is returned. The start and end of the first supplied networks are the
// first and last entries, with each networks first and last IPs between.
func Boundaries(network net.IPNet, existing []net.IPNet) ([]uint32, error) {
	sort.Sort(IPNets(existing))

	rangeStart, rangeEnd := Range(network)

	boundaries := []uint32{rangeStart}

	for _, existingNetwork := range existing {
		start, end := Range(existingNetwork)
		boundaries = append(boundaries, start, end)
	}

	boundaries = append(boundaries, rangeEnd)

	// Invariant: The number of boundaries must be even.
	if len(boundaries)%2 != 0 {
		return nil, microerror.Maskf(
			incorrectNumberOfBoundariesError, "%v ipnets, %v boundaries", len(existing), len(boundaries),
		)
	}

	return boundaries, nil
}

// FreeRanges converts a list of boundaries into a list of free ranges.
// e.g: [168034304, 168099839] to [{168034304, 168099839}].
// These free ranges are then used to find a free space for a new network.
func FreeRanges(boundaries []uint32) ([]freeRange, error) {
	// Invariant: The number of boundaries must be even.
	if len(boundaries)%2 != 0 {
		return nil, microerror.Maskf(
			incorrectNumberOfBoundariesError, "%v boundaries", len(boundaries),
		)
	}

	freeRanges := []freeRange{}

	for i := 0; i < len(boundaries)-1; i = i + 2 {
		freeRanges = append(
			freeRanges,
			freeRange{start: boundaries[i], end: boundaries[i+1]},
		)
	}

	// Invariant: Number of free ranges is half the number of boundaries.
	if len(freeRanges) != len(boundaries)/2 {
		return nil, microerror.Maskf(
			incorrectNumberOfFreeRangesError, "%v boundaries, %v free ranges", len(boundaries), len(freeRanges),
		)
	}

	return freeRanges, nil
}

// Space takes a list of free ranges, and returns the first IP that has space.
func Space(freeRanges []freeRange, mask net.IPMask) (uint32, error) {
	for i := 0; i < len(freeRanges); i++ {
		free := freeRanges[i]

		if free.end-free.start+1 >= Size(mask) {
			start := free.start

			// In the case that we are not at the start of the overall range,
			// we need to increment to the start of the next range.
			if i != 0 {
				start++
			}

			return start, nil
		}
	}

	return 0, microerror.Maskf(spaceExhaustedError, "tried to fit: %v", mask)
}

// Free takes a network, a mask, and a list of networks.
// An available network, within the first network, is returned.
func Free(network net.IPNet, mask net.IPMask, existing []net.IPNet) (net.IPNet, error) {
	if Size(network.Mask) < Size(mask) {
		return net.IPNet{}, microerror.Maskf(
			maskTooBigError, "have: %v, requested: %v", network.Mask, mask,
		)
	}

	sort.Sort(IPNets(existing))

	// Get the start and end of each network currently in use.
	boundaries, err := Boundaries(network, existing)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	// Determine the amount of free space between each network in use,
	// as well as betweeen the start and end of the total space.
	freeRanges, err := FreeRanges(boundaries)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	// Attempt to find a free space, of the required size.
	free, err := Space(freeRanges, mask)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	freeIP := DecimalToIP(free)
	freeNetwork := net.IPNet{IP: freeIP, Mask: mask}

	// Invariant: The IP of the network returned should not be nil.
	if freeNetwork.IP == nil {
		return net.IPNet{}, microerror.Mask(nilIPError)
	}

	// Invariant: The IP of the network returned should be contained
	// within the network supplied.
	if !network.Contains(freeNetwork.IP) {
		return net.IPNet{}, microerror.Maskf(
			ipNotContainedError, "%v is not contained by %v", freeNetwork.IP, network,
		)
	}

	// Invariant: The mask of the network returned should be equal to
	// the mask supplied as an argument.
	if !bytes.Equal(mask, freeNetwork.Mask) {
		return net.IPNet{}, microerror.Maskf(
			maskIncorrectSizeError, "have: %v, requested: %v", freeNetwork.Mask, mask,
		)
	}

	return freeNetwork, nil
}
