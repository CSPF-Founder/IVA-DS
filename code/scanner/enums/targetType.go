package enums

type TargetType string

const (
	TargetTypeIP      TargetType = "ip"
	TargetTypeIPRange TargetType = "ip_range"
	TargetTypeURL     TargetType = "url"
	TargetTypeInvalid TargetType = ""
)
