package main

import (
	"net"
)

type freeRange struct {
	start uint32
	end   uint32
}

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
