package wire

import (
	"encoding/json"
	neturl "net/url"
)

type Request struct {
	Action   string // adverb: get, call, list
	Base     string
	Relative string
	Params   map[string]string
	Args     any
	Cached   WireMessage
}

func (r *Request) Body() (string, []byte, error) {
	if r.Args == nil {
		return "", nil, nil
	}
	content_type := "application/json"

	bytes, err := json.Marshal(r.Args)
	return content_type, bytes, err
}

func (r *Request) Url(base string) string {
	if r.Base != "" {
		base = r.Base
	}
	if r.Relative != "" {
		url, _ := neturl.JoinPath(base, r.Relative)
		return url
	} else {
		return base
	}
}

