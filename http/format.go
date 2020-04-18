package http

type RequestType int

const (
	JSON RequestType = iota
	XML
)

func (s RequestType) String() string {
	return toString[s]
}

func GetRType(rtype string) RequestType {
	return toRequestType[rtype]
}

var toString = map[RequestType]string{
	JSON:   "json",
	XML:  "xml",
}

var toRequestType = map[string]RequestType{
	"json":   JSON,
	"xml":  XML,
}

