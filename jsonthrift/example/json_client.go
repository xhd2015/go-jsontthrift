package example

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
)

func Run() {
	var err error

	//transport := thrift.NewTMemoryBuffer()

	transport, err := thrift.NewTSocket("localhost:8888")
	if err != nil {
		fmt.Printf("cannot connect to socket:%+v", err)
		return
	}
	err = transport.Open()
	if err != nil {
		fmt.Printf("cannot open transport:%+v", err)
		return
	}

	protocol := thrift.NewTBinaryProtocolTransport(transport)
	client := thrift.NewTStandardClient(protocol, protocol)

	ctx := context.Background()

	req := SimpleIntSchema.NewJsonTStructForWriteOut(45)
	resp := SimpleIntSchema.NewJsonTStructForReadIn()

	err = client.Call(ctx, "Pow2", req, resp)

	if err != nil {
		fmt.Printf("call Pow2 error:%+v\n", err)
		return
	}
	fmt.Printf("resp = %+v\n", resp.Val())

	// AddPersonAge
	personReq := PersonSchema.NewJsonTStructForWriteOut(map[string]interface{}{
		"age":   12,
		"alias": "Duckes",
	})
	personResp := PersonSchema.NewJsonTStructForReadIn()
	err = client.Call(ctx, "AddPersonAge", personReq, personResp)

	if err != nil {
		fmt.Printf("call AddPersonAge error:%+v\n", err)
		return
	}
	fmt.Printf("person age added:%+v", personResp)
}
