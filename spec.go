package ipam

import (
	"net"
)

// ipRange defines a pair of IPs, over a range.
type ipRange struct {
	start net.IP
	end   net.IP
}

// ipNets is a helper type for sorting net.IPNets.
type ipNets []net.IPNet

func (s ipNets) Len() int {
	return len(s)
}

// Tuntion used to order nets, IP is checked first then Mask in case IP is the same
func (s ipNets) Less(i, j int) bool {
	if ipToDecimal(s[i].IP) == ipToDecimal(s[j].IP) {
		return size(s[i].Mask) > size(s[j].Mask)
	} else {
		return ipToDecimal(s[i].IP) < ipToDecimal(s[j].IP)
	}
}

func (s ipNets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
