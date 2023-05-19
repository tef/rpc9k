package wire

import (
	"time"
)

var Root = &Namespace{ 
        CommonMessage: CommonMessage{ 
                kind: "Namespace", 
                ApiVersion: "0", 
        }, 
        routes: []string{"Example",}, 
        urls: map[string]string{}, 
        embeds: map[string]Message{
		"Example": Example,
	},
} 

var Example = &Namespace{
        CommonMessage: CommonMessage{ 
                kind: "Namespace", 
                ApiVersion: "0", 
        }, 
        routes: []string{"rpc",},
        urls: map[string]string{}, 
        embeds: map[string]Message{},
}


type Message interface {
	Kind() string
	Routes() []string
	Fetch(name string) (*Request)
	Call(args any) (*Request)
}

type Request struct {
	Action string // adverb: get, call, list
	Path   string
	State  any
	Args   any
	Cached Message
}

func (r *Request) UrlFrom(base string) string {
	return base+","+r.Path
}

type Metadata struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

type CommonMessage struct {
	kind       string
	ApiVersion string
	Metadata   Metadata
}

func (b *CommonMessage) Kind() string {
	return b.kind
}

func (b *CommonMessage) Routes() []string {
	return []string{}
}

func (b *CommonMessage) Fetch(name string) (*Request) {
	return nil
}

func (b *CommonMessage) Call(args any) *Request {
	return nil
}

type Namespace struct {
	CommonMessage
	routes []string
	urls   map[string]string
	embeds map[string]Message
}

func (n *Namespace) Kind() string {
	return "Namespace"
}

func (n *Namespace) Routes() []string {
	return n.routes
}

func (n *Namespace) Fetch(name string) (*Request) {
	request := &Request{
		Action:"get",
	}
	url, ok := n.urls[name]
	if ok {
		request.Path = url
	} else {
		request.Path = name
	}

	message, ok := n.embeds[name]

	if ok {
		request.Cached = message
	}

	return request
}

type Service struct {
	CommonMessage
}

func (*Service) Kind() string {
	return "Service"
}

type Procedure struct {
	CommonMessage
}

func (*Procedure) Kind() string {
	return "Procedure"
}

var root = &Namespace{
	CommonMessage: CommonMessage{
		kind: "Namespace",
		ApiVersion: "0",
	},
	routes: []string{"Example",},
	urls: map[string]string{},
	embeds: map[string]Message{},
}


