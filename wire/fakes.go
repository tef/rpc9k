package wire

import (
	"encoding/json"
	"errors"
	"fmt"
)

var Root = (&Module{
	Header: Header{
		Kind:       "Module",
		ApiVersion: "0",
	},
	Names:  []string{"Example"},
	Urls:   map[string]string{},
	Embeds: map[string]Variant{
		//	"Example": Variant{Msg: Example},
	},
}).Variant()

var Example = (&Instance{
	Header: Header{
		Kind:       "Instance",
		ApiVersion: "0",
	},
	Methods: []string{"rpc"},
	Urls:    map[string]string{},
	Embeds:  map[string]Variant{
		//	"rpc": Variant{Msg: rpc},
	},
}).Variant()

var rpc = (&Procedure{
	Header: Header{
		Kind:       "Procedure",
		ApiVersion: "0",
	},
	Arguments: []string{"x", "y"},
}).Variant()

func FakeServer(Action string, url string, payload *Blob) (*Variant, error) {
	fmt.Println("serving", Action, url)

	if Action == "get" {
		if url == "/" {
			return &Root, nil
		} else if url == "/Example" {
			redirect := &Variant{
				Kind: "Redirect",
				Msg: &Redirect{
					Header: Header{
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
			err := json.Unmarshal(payload.Value, &output)
			if err != nil {
				return nil, err
			}
			fmt.Println("Got", payload.Value, output)

			reply, err := json.Marshal(output)

			if err != nil {
				fmt.Println("RT", err)
				return nil, err
			}

			return &Variant{Kind: "JSON", Msg: &JSON{Value: reply}}, nil
		}
	}

	return nil, errors.New("no")

}
