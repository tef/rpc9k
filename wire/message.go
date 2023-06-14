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
	Call(args Variant, base string) *Request
	Scan(args any) error
	Wrap() Variant
	// Empty() bool
	// Blob() Blob
	//
}

type Variant struct {
	Kind string      `json:"Kind"`
	Msg  WireMessage `json:"-"`
}

func (e *Variant) Unwrap(out WireMessage) bool {
	if e == nil && out == nil {
		return true
	} else if e == nil || out == nil {
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

func (e *Variant) Blob() (*Blob, error) {
	fmt.Println("Blobbing", e.Msg)
	if b, ok := e.Msg.(*Blob); ok {
		return b, nil
	}
	if v, ok := e.Msg.(*Value); ok {
		content_type := "application/json"

		bytes, err := json.Marshal(v.Value)
		fmt.Println("Value is", v.Value)
		return &Blob{ContentType: content_type, Value: bytes}, err
	}
	if v, ok := e.Msg.(*JSON); ok {
		content_type := "application/json"
		fmt.Println("Raw JSON Value", v.Value)

		return &Blob{ContentType: content_type, Value: v.Value}, nil
	}

	fmt.Println("Message", e.Msg)
	bytes, err := e.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return &Blob{ContentType: "application/9k+json", Value: bytes}, nil

}

func (e *Variant) UnmarshalJSON(bytes []byte) error {
	var M struct {
		Kind string `json:"Kind"`
	}

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

func (e *Variant) MarshalJSON() ([]byte, error) {
	if e == nil {
		return json.Marshal(Empty{})
	} else {
		return json.Marshal(e.Msg)
	}
}

func (b Variant) Ptr() *Variant {
	return &b
}

func (b Variant) Wrap() Variant {
	return b
}

func (b Variant) IsEmpty() bool {
	return (b.Msg == nil || b.Kind == "Empty")
}

func (b Variant) Routes() []string {
	return b.Msg.Routes()
}

func (b Variant) Fetch(name string, base string) *Request {
	return b.Msg.Fetch(name, base)
}

func (b Variant) Call(args Variant, base string) *Request {
	return b.Msg.Call(args, base)
}

func (b Variant) Scan(args any) error {
	return b.Msg.Scan(args)
}

type MessageBuilder func() WireMessage

var Messages = map[string]MessageBuilder{
	"Module":    func() WireMessage { return &Module{} },
	"Instance":   func() WireMessage { return &Instance{} },
	"Procedure": func() WireMessage { return &Procedure{} },
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

func (b *Header) Call(args Variant, base string) *Request {
	return nil
}

func (b *Header) Scan(args any) error {
	return errors.New("no value")
}

type Module struct {
	Header
	Names  []string            `json:"Names"`
	Urls   map[string]string   `json:"Urls"`
	Embeds map[string]Variant `json:"Embeds"`
}

func (n *Module) Wrap() Variant {
	return Variant{Kind: "Module", Msg: n}
}

func (n *Module) Routes() []string {
	return n.Names
}

func (n *Module) Fetch(name string, base string) *Request {
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

type Instance struct {
	Header
	Params  map[string]string
	Methods []string
	Urls    map[string]string
	Embeds  map[string]Variant
}

func (n *Instance) Wrap() Variant {
	return Variant{Kind: "Instance", Msg: n}
}

func (s *Instance) Routes() []string {
	return s.Methods
}

func (s *Instance) Fetch(name string, base string) *Request {
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
	Result    Variant
}

func (n *Procedure) Wrap() Variant {
	return Variant{Kind: "Procedure", Msg: n}
}

func (p *Procedure) Call(args Variant, base string) *Request {
	// if args is a struct, wrap it into message that
	// gets turned into json and sent as json

	// if args is a list, or a map, create a new map
	// with the named arguments, and send over an
	// arguments message, which has a Kind field

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

type Map struct {
	Header
	Next    string
	Entries map[string]Variant
}

type List struct {
	Header
	Next    string
	Entries []Variant
}

type JSON struct {
	Header
	Value json.RawMessage
}

func (n *JSON) Wrap() Variant {
	return Variant{Kind: "JSON", Msg: n}
}

func (m *JSON) Scan(out any) error {
	return json.Unmarshal(m.Value, out)
}

type Blob struct {
	Header
	ContentType string
	Value       []byte
}

func (n *Blob) Wrap() Variant {
	return Variant{Kind: "Blob", Msg: n}
}
func (e *Blob) Scan(out any) error {
	if e == nil && out == nil {
		return nil
	} else if e != nil || out == nil {
		return errors.New("can't scan from/to nil")
	}

	output := reflect.Indirect(reflect.ValueOf(out))
	input := reflect.Indirect(reflect.ValueOf(e.Value))

	if output.Kind() == input.Kind() {
		return errors.New("can't scan diff types")
	}

	output.Set(input)
	return nil
}

type Value struct {
	Header
	Value any
}

func (n *Value) Wrap() Variant {
	return Variant{Kind: "Value", Msg: n}
}

func (e *Value) Scan(out any) error {
	if e == nil && out == nil {
		return nil
	} else if e != nil || out == nil {
		return errors.New("can't scan from/to nil")
	}

	output := reflect.Indirect(reflect.ValueOf(out))
	input := reflect.Indirect(reflect.ValueOf(e.Value))

	if output.Kind() == input.Kind() {
		return errors.New("can't scan diff types")
	}

	output.Set(input)
	return nil
}

type Empty struct { // HTTP 203
	Header
}

func (n *Empty) Wrap() Variant {
	return Variant{Kind: "Empty", Msg: n}
}

type Redirect struct {
	Header
	Target string
}

func (n *Redirect) Wrap() Variant {
	return Variant{Kind: "Redirect", Msg: n}
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

func (n *Error) Wrap() Variant {
	return Variant{Kind: "Error", Msg: n}
}

func NewError(id string, args ...any) Variant {
	errText := fmt.Sprint(args)
	return Variant{
		Kind: "Error",
		Msg: &Error{
			Header: Header{
				Kind:       "Error",
				ApiVersion: "0",
			},
			Id:   id,
			Text: errText,
		},
	}
}

func NewErr(id string, err error) Variant {
	return Variant{
		Kind: "Error",
		Msg: &Error{
			Header: Header{
				Kind:       "Error",
				ApiVersion: "0",
			},
			Id:   id,
			Text: err.Error(),
		},
	}
}

// ClientError? ServerError
