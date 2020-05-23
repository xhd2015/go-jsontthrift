package example

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"time"
)

// give val, return val*val
// default processor serves 2 methods:
//    Pow2: given a number,return its square
//    AddPersonAge: given a person with alias and age, increment its age,and return
type DefaultProcessor struct {
}

func (p *DefaultProcessor) Process(ctx context.Context, in, out thrift.TProtocol) (bool, thrift.TException) {
	name, typeId, seqid, err := in.ReadMessageBegin()
	if err != nil {
		return false, nil
	}
	fmt.Printf("name=%v,typeId=%#v,seqid=%v\n", name, typeId, seqid)

	if name == "Pow2" {
		jsonStruct := SimpleIntSchema.NewJsonTStructForReadIn()

		err = jsonStruct.Read(in)
		if err != nil {
			fmt.Printf("read json error:%+v", err)
			return false, nil
		}

		fmt.Printf("jsonVal = %+v\n", jsonStruct.Val())

		val := jsonStruct.Val().(int32)

		result := SimpleIntSchema.NewJsonTStructForWriteOut(val * val)
		// name and seqid must match
		out.WriteMessageBegin(name, thrift.REPLY, seqid)
		result.Write(out)
		out.WriteMessageEnd()
	} else if name == "AddPersonAge" {
		person := NewPerson()

		err = person.Read(in)
		if err != nil {
			fmt.Printf("error read person:%+v", err)
			return false, nil
		}
		fmt.Printf("person read = %+v", person)

		person.Age++
		out.WriteMessageBegin(name, thrift.REPLY, seqid)
		err = person.Write(out)
		if err != nil {
			fmt.Printf("err write person:%+v", err)
			return false, nil
		}
		out.WriteMessageEnd()

	}
	return true, nil
}

func (p *DefaultProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
	fmt.Printf("ProcessorMap called")
	return nil
}

func (p *DefaultProcessor) AddToProcessorMap(string, thrift.TProcessorFunction) {
	fmt.Printf("Add to map")
}

func RunAndServe() {
	var processor = thrift.NewTMultiplexedProcessor()

	processor.RegisterDefault(&DefaultProcessor{})

	serverTransport, err := thrift.NewTServerSocketTimeout(":8888", 10*time.Second)
	if err != nil {
		fmt.Printf("Unable to create server socket:%+v", err)
		return
	}

	server := thrift.NewTSimpleServer2(processor, serverTransport)
	server.Serve()
}
