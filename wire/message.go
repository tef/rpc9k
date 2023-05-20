package wire

import (
	"encoding/json"
	"time"
)

var Root = &Namespace{
	CommonMessage: CommonMessage{
		Kind:       "Namespace",
		ApiVersion: "0",
	},
	Names: []string{"Example"},
	Urls:  map[string]string{},
	Embeds: map[string]Envelope{
		"Example": Envelope{M: Example},
	},
}

var Example = &Namespace{
	CommonMessage: CommonMessage{
		Kind:       "Namespace",
		ApiVersion: "0",
	},
	Names:  []string{"rpc"},
	Urls:   map[string]string{},
	Embeds: map[string]Envelope{},
}

type Request struct {
	Action string // adverb: get, call, list
	Path   string
	State  any
	Args   any
	Cached Message
}

type Message interface {
	Routes() []string
	Fetch(name string) *Request

	Call(args any) *Request
}

type Envelope struct {
	M Message
}

func (e *Envelope) UnmarshalJSON(bytes []byte) error {
	e.M = &Namespace{}
	return json.Unmarshal(bytes, e.M)
}

func (e Envelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.M)
}

type Metadata struct {
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	Version   int       `json:"Version"`
}

type CommonMessage struct {
	Kind       string   `json:"Kind"`
	ApiVersion string   `json:"ApiVersion"`
	Metadata   Metadata `json:"Metadata"`
}

func (b *CommonMessage) Routes() []string {
	return []string{}
}

func (b *CommonMessage) Fetch(name string) *Request {
	return nil
}

func (b *CommonMessage) Call(args any) *Request {
	return nil
}

type Error struct {
	CommonMessage
	Id   string
	Text string
}

type Namespace struct {
	CommonMessage
	Names  []string            `json:"Names"`
	Urls   map[string]string   `json:"Urls"`
	Embeds map[string]Envelope `json:"Embeds"`
}

func (n *Namespace) Routes() []string {
	return n.Names
}

func (n *Namespace) Fetch(name string) *Request {
	request := &Request{
		Action: "get",
	}
	url, ok := n.Urls[name]
	if ok {
		request.Path = url
	} else {
		request.Path = name + "/"
	}

	message, ok := n.Embeds[name]

	if ok {
		request.Cached = message.M
	}

	return request
}

type Service struct {
	CommonMessage
	State   map[string]string
	Methods []string
	Urls    map[string]string
	Embeds  map[string]Envelope
}

type Procedure struct {
	CommonMessage
	State     map[string]string
	Arguments []string
	Result    Message
}

type JSON struct {
	CommonMessage
	Value *json.RawMessage
}
