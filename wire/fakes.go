package wire

import (
	"encoding/json"
	"fmt"
	"errors"
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

