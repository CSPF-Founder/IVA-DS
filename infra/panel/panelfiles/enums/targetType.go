package enums

import (
	"net"
	"net/url"

	"github.com/CSPF-Founder/iva/panel/utils/iputils"
)

type TargetType string

const (
	TargetTypeIP      TargetType = "ip"
	TargetTypeIPRange TargetType = "ip_range"
	TargetTypeURL     TargetType = "url"
	TargetTypeInvalid TargetType = ""
)

func ParseTargetType(targetAddress string) TargetType {
	if ip := net.ParseIP(targetAddress); ip != nil {
		return TargetTypeIP
	} else if parsedURL, err := url.ParseRequestURI(targetAddress); err == nil {
		if parsedURL.Scheme != "" && parsedURL.Host != "" {
		return TargetTypeURL
		}
	} else {
		ipCount, err := iputils.ConvertIPRangeToIPSize(targetAddress)
		if err != nil {
			return TargetTypeInvalid
		}
		if ipCount == nil || ipCount.Int64() > 256 {
			return TargetTypeInvalid
		} else {
			return TargetTypeIPRange
		}
	}
	return TargetTypeInvalid
}
