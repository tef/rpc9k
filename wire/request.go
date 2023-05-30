package wire

import (
	neturl "net/url"
)

type Request struct {
	Action   string // adverb: get, call, list
	Base     string
	Relative string
	Params   map[string]string
	Args     Envelope
	Cached   WireMessage
}

func (r *Request) Body() (*Blob,  error) {
	if r.Args.IsEmpty() {
		return nil, nil
	}
	
	return r.Args.Blob()
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

