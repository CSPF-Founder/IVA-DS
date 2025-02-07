package iputils

import (
	"math"
	"net"
)

func GetIPCountIfRange(cidr string) (int, error) {

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}

	// Get the number of bits in the network mask
	ones, _ := ipnet.Mask.Size()

	// Calculate the number of host bits
	hostBits := 32 - ones

	// Calculate the number of usable IP addresses (excluding network and broadcast addresses)
	totalIPs := int(math.Pow(2, float64(hostBits)))

	return totalIPs, nil
}
