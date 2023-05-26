package wire

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"reflect"
	"time"
)

var Root = (&Namespace{
	CommonMessage: CommonMessage{
		Kind:       "Namespace",
		ApiVersion: "0",
	},
	Names: []string{"Example"},
	Urls:  map[string]string{},
	Embeds: map[string]Envelope{
	//	"Example": Envelope{Msg: Example},
	},
}).Wrap()

var Example = (&Service{
	CommonMessage: CommonMessage{
		Kind:       "Service",
		ApiVersion: "0",
	},
	Methods: []string{"rpc"},
	Urls:    map[string]string{},
	Embeds: map[string]Envelope{
	//	"rpc": Envelope{Msg: rpc},
	},
}).Wrap()

var rpc = (&Procedure{
	CommonMessage: CommonMessage{
		Kind:       "Service",
		ApiVersion: "0",
	},
	Arguments: []string{"x", "y"},
}).Wrap()

func FakeServer(Action string, url string, content_type string, buf []byte) (*Envelope, error) {
	fmt.Println("serving", Action, url)

	if Action == "get" {
		if url == "/" {
			return &Root, nil
		} else if url == "/Example" {
			redirect := &Envelope{
				Kind:"Redirect", 
				Msg: &Redirect {
					CommonMessage: CommonMessage{
						Kind:       "Redirect",
						ApiVersion: "0",
					},
					Target: "/Example/",
				},
			}
			return redirect, nil
		} else if url == "/Example/" {
			return &Example, nil
		} else if url == "/Example/rpc" {
			return &rpc, nil
		}
	}

	if Action == "post" {
		if url == "/Example/rpc" {
			var output any
			err := json.Unmarshal(buf, &output)
			if err != nil {
				return nil, err
			}
			fmt.Println("Got", output)

			reply, err := json.Marshal(output)

			if err != nil {
				fmt.Println("RT", err)
				return nil, err
			}

			return &Envelope{Kind:"JSON", Msg:&JSON{Value: reply}}, nil
		}
	}

	return nil, errors.New("no")

}

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

// Can't use WireMessage as a struct member when said struct
// gets converted to and from json, encoder doesn't know
// how to turn JSON into a given interface, and we can't
// hook a method onto the WireMessage interface type either.

// ... but we can override a struct's behaviour, and so
// we have a container struct that contains one field,
// a message, and the container knows how to encode or
// decode to json

type WireMessage interface {
	Routes() []string
	Fetch(name string, base string) *Request
	Call(args any, base string) *Request
	Scan(args any) error
	Wrap() Envelope
}

type Envelope struct {
	Kind string
	Msg WireMessage
}

func (e *Envelope) Unwrap(out WireMessage) bool {
	output := reflect.Indirect(reflect.ValueOf(out))
	input := reflect.Indirect(reflect.ValueOf(e.Msg))
	if output.Kind() == input.Kind() {
		return false
	}

	output.Set(input)
	return true
}

func (e *Envelope) UnmarshalJSON(bytes []byte) error {
	var M CommonMessage

	err := json.Unmarshal(bytes, &M)
	if err != nil {
		return err
	}

	builder, ok := Messages[M.Kind]
	if !ok {
		return errors.New("Unknown message: " + M.Kind)
	}
	e.Kind = M.Kind
	e.Msg = builder()
	return json.Unmarshal(bytes, e.Msg)
}

func (e Envelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Msg)
}

func (b Envelope) Wrap() Envelope {
	return b
}

func (b Envelope) Routes() []string {
	return b.Msg.Routes()
}

func (b Envelope) Fetch(name string, base string) *Request {
	return b.Msg.Fetch(name, base)
}

func (b Envelope) Call(args any, base string) *Request {
	return b.Msg.Call(args, base)
}

func (b Envelope) Scan(args any) error {
	return b.Msg.Scan(args)
}

type MessageBuilder func() WireMessage

var Messages = map[string]MessageBuilder{
	"Namespace": func() WireMessage { return &Namespace{} },
	"Service":   func() WireMessage { return &Service{} },
	"Procedure": func() WireMessage { return &Namespace{} },
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

func (b CommonMessage) Routes() []string {
	return []string{}
}

func (b CommonMessage) Fetch(name string, base string) *Request {
	return nil
}

func (b CommonMessage) Call(args any, base string) *Request {
	return nil
}

func (b CommonMessage) Scan(args any) error {
	return errors.New("no value")
}

type Namespace struct {
	CommonMessage
	Names  []string            `json:"Names"`
	Urls   map[string]string   `json:"Urls"`
	Embeds map[string]Envelope `json:"Embeds"`
}
func (n *Namespace) Wrap() Envelope {
	return Envelope{Kind: "Namespace", Msg: n}
}


func (n *Namespace) Routes() []string {
	return n.Names
}

func (n *Namespace) Fetch(name string, base string) *Request {
	request := &Request{
		Action: "get",
		Base:   base,
	}
	url, ok := n.Urls[name]
	if ok {
		request.Relative = url
	} else {
		request.Relative = name 
	}

	message, ok := n.Embeds[name]

	if ok {
		request.Cached = message.Msg
	}

	return request
}

type Service struct {
	CommonMessage
	Params  map[string]string
	Methods []string
	Urls    map[string]string
	Embeds  map[string]Envelope
}

func (n *Service) Wrap() Envelope {
	return Envelope{Kind: "Service", Msg: n}
}

func (s *Service) Routes() []string {
	return s.Methods
}

func (s *Service) Fetch(name string, base string) *Request {
	request := &Request{
		Action: "get",
		Base:   base,
		Params: s.Params,
	}
	url, ok := s.Urls[name]
	if ok {
		request.Relative = url
	} else {
		request.Relative = name
	}

	message, ok := s.Embeds[name]

	if ok {
		request.Cached = message.Msg
	}

	return request
}

type Procedure struct {
	CommonMessage
	Params    map[string]string
	Arguments []string
	Result    Envelope
}
func (n *Procedure) Wrap() Envelope {
	return Envelope{Kind: "Procedure", Msg: n}
}


func (p *Procedure) Call(args any, base string) *Request {
	request := &Request{
		Action:   "post",
		Base:     base,
		Relative: "",
		Params:   p.Params,
		Args:     args,
		Cached:   p.Result.Msg,
	}
	return request
}

type JSON struct {
	CommonMessage
	Value json.RawMessage
}
func (n *JSON) Wrap() Envelope {
	return Envelope{Kind: "JSON", Msg: n}
}


func (m *JSON) Scan(out any) error {
	return json.Unmarshal(m.Value, out)
}

type Blob struct {
	CommonMessage
	ContentType string
	Value       []byte
}
func (n *Blob) Wrap() Envelope {
	return Envelope{Kind: "Blob", Msg: n}
}


type Value struct {
	CommonMessage
	Value any
}

func (n *Value) Wrap() Envelope {
	return Envelope{Kind: "Value", Msg: n}
}

func (e *Value) Scan(out any) error {
	output := reflect.Indirect(reflect.ValueOf(out))
	input := reflect.Indirect(reflect.ValueOf(e.Value))
	if output.Kind() == input.Kind() {
		return errors.New("Can't unwrap")
	}

	output.Set(input)
	return nil
}

type Empty struct { // HTTP 203
	CommonMessage
}
func (n *Empty) Wrap() Envelope {
	return Envelope{Kind: "Empty", Msg: n}
}


type Redirect struct {
	CommonMessage
	Target string
}
func (n *Redirect) Wrap() Envelope {
	return Envelope{Kind: "Redirect", Msg: n}
}


func (r *Redirect) Url(base string) string {
	if r.Target[0] == '/' {
		return r.Target
	}
	url, _ := neturl.JoinPath(base, r.Target)
	return url
}


type Error struct {
	CommonMessage
	Id   string
	Text string
}
func (n *Error) Wrap() Envelope {
	return Envelope{Kind: "Error", Msg: n}
}


// ClientError? ServerError
