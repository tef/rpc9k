package wire

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"reflect"
	"time"
)

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
	// Empty() bool
	// Blob() Blob
	// 
}

type Envelope struct {
	Kind string `json:"Kind"`
	Msg WireMessage `json:"-"`
}

func (e *Envelope) Unwrap(out WireMessage) bool {
	if e == nil && out == nil {
		return true
	} else if e== nil || out == nil {
		return false
	}

	output := reflect.Indirect(reflect.ValueOf(out))
	input := reflect.Indirect(reflect.ValueOf(e.Msg))
	if output.Kind() == input.Kind() {
		return false
	}

	output.Set(input)
	return true
}

func (e *Envelope) UnmarshalJSON(bytes []byte) error {
	var M struct {Kind string `json:"Kind"`}

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

func (e *Envelope) MarshalJSON() ([]byte, error) {
	if e == nil {
		return json.Marshal(Empty{})
	} else { 
		return json.Marshal(e.Msg)
	}
}

func (b Envelope) Ptr() *Envelope {
	return &b
}

func (b Envelope) Wrap() Envelope {
	return b
}

func (b Envelope) IsEmpty() bool {
	return  (b.Msg == nil || b.Kind == "Empty")
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

type Header struct {
	Kind       string   `json:"Kind"`
	ApiVersion string   `json:"ApiVersion"`
	Metadata   Metadata `json:"Metadata"`
}

func (b *Header) Routes() []string {
	return []string{}
}

func (b *Header) Fetch(name string, base string) *Request {
	return nil
}

func (b *Header) Call(args any, base string) *Request {
	return nil
}

func (b *Header) Scan(args any) error {
	return errors.New("no value")
}

type Namespace struct {
	Header
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
	Header
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
	Header
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
	Header
	Value json.RawMessage
}
func (n *JSON) Wrap() Envelope {
	return Envelope{Kind: "JSON", Msg: n}
}


func (m *JSON) Scan(out any) error {
	return json.Unmarshal(m.Value, out)
}

type Blob struct {
	Header
	ContentType string
	Value       []byte
}
func (n *Blob) Wrap() Envelope {
	return Envelope{Kind: "Blob", Msg: n}
}


type Value struct {
	Header
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
	Header
}
func (n *Empty) Wrap() Envelope {
	return Envelope{Kind: "Empty", Msg: n}
}


type Redirect struct {
	Header
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
	Header
	Id   string
	Text string
}
func (n *Error) Wrap() Envelope {
	return Envelope{Kind: "Error", Msg: n}
}


func NewError(id string, args ...any) Envelope {
	errText := fmt.Sprint(args)
	return Envelope{
		Kind: "Error",
		Msg: &Error{
			Header: Header{
				Kind: "Error",
				ApiVersion: "0",
			},
			Id: id, 
			Text: errText,
		},
	}
}


func NewErr(id string, err error) Envelope {
	return Envelope{
		Kind: "Error",
		Msg: &Error{
			Header: Header{
				Kind: "Error",
				ApiVersion: "0",
			},
			Id: id, 
			Text: err.Error(),
		},
	}
}
// ClientError? ServerError
