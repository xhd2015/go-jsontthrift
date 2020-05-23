package json

import (
	encodejson "encoding/json"
	"errors"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"strings"
)

// the TStruct Read()
//             Write()

// the underlying is binary, for write, we write JSON,then it is transferred to Binary
//                           for read, we read Binary, then transfer it to JSON

// TTransport

//
//    TStruct
// type JsonStruct struct {
//        json interface{}
//        schema Schema
// }
// JsonStruct.Write(p){
//     for k,v in schema {
//             p.WriteXXXBegin()
//             p.WriteXXX(...)
//             p.WriteXXXEnd()
//     }
// }
//
//

type JsonSchema interface {
	Write(protocol thrift.TProtocol, json interface{}) error
	Read(protocol thrift.TProtocol) (interface{}, error)
	NewJsonTStructForWriteOut(val interface{}) *JsonStruct
	NewJsonTStructForReadIn() *JsonStruct
}

func NewJsonSchema(jsonSchema string) (JsonSchema, error) {
	decoder := encodejson.NewDecoder(strings.NewReader(jsonSchema))

	schemaImp := JsonSchemaImp{}
	err := decoder.Decode(&schemaImp)
	if err != nil {
		return nil, err
	}
	return &schemaImp, nil
}

func MustNewJsonSchema(jsonSchema string) JsonSchema {
	schema, err := NewJsonSchema(jsonSchema)
	if err != nil {
		panic(err)
	}
	return schema
}

type JsonValType int

const (
	// Actually Any is not allowed to express a strongly typed protocol
	// Because the thrift.Protocol API does not provide a way to detect forward value type
	//Any JsonValType = iota
	_ JsonValType = iota
	Int64
	Int32
	Int16
	Double
	Bool
	String
	Object
	Array
	Map
)

func (jsonValType *JsonValType) UnmarshalJSON(s []byte) error {
	var name string
	err := encodejson.Unmarshal(s, &name)
	if err != nil {
		return err
	}
	name = strings.ToLower(name)
	switch name {
	case "int", "int32":
		*jsonValType = Int32
		return nil
	case "int16":
		*jsonValType = Int16
		return nil
	case "int64":
		*jsonValType = Int64
		return nil
	case "double":
		*jsonValType = Double
		return nil
	case "bool":
		*jsonValType = Bool
		return nil
	case "string":
		*jsonValType = String
		return nil
	case "map":
		*jsonValType = Map
		return nil
	case "object":
		*jsonValType = Object
		return nil
	case "array", "list":
		*jsonValType = Array
		return nil
	}
	return errors.New(fmt.Sprintf("unmarshal JsonValType error:%s", string(s)))
}

func (jsonValType JsonValType) TType() thrift.TType {
	switch jsonValType {
	case Int64:
		return thrift.I64
	case Int32:
		return thrift.I32
	case Int16:
		return thrift.I16
	case Double:
		return thrift.DOUBLE
	case String:
		return thrift.STRING
	case Bool:
		return thrift.BOOL
	case Object:
		return thrift.STRUCT
	case Map:
		return thrift.MAP
	case Array:
		return thrift.LIST
	default:
		panic("Unmapped json val type:" + fmt.Sprintf("%#v", jsonValType))
	}
}

type FieldAndType struct {
	Name string `json:"name"`
	// ID is the field id
	ID    int16         `json:"id"`
	Type_ JsonSchemaImp `json:"description"`
}

type JsonSchemaImp struct {
	ObjectType JsonValType `json:"type"`
	// ObjectFields is valid when ObjectType is Object, this fields should be strongly ordered
	ObjectFields []*FieldAndType `json:"fields"`
	// ElementType is valid when ObjectType is Array or is Map
	ElementType *JsonSchemaImp `json:"element"`
}

type JsonStruct struct {
	schema JsonSchema
	val    interface{}
}

func (j *JsonStruct) Write(p thrift.TProtocol) error {
	return j.schema.Write(p, j.val)
}

func (j *JsonStruct) Read(p thrift.TProtocol) error {
	val, err := j.schema.Read(p)
	j.val = val
	return err
}

func (j *JsonStruct) Val() interface{} {
	return j.val
}

func (schema *JsonSchemaImp) NewJsonTStructForWriteOut(val interface{}) *JsonStruct {
	return &JsonStruct{
		schema: schema,
		val:    val,
	}
}

func (schema *JsonSchemaImp) NewJsonTStructForReadIn() *JsonStruct {
	return &JsonStruct{
		schema: schema,
	}
}

func (schema *JsonSchemaImp) TType() thrift.TType {
	return schema.ObjectType.TType()
}

func (schema *JsonSchemaImp) Write(protocol thrift.TProtocol, json interface{}) error {
	switch schema.ObjectType {
	case Int64:
		if json == nil {
			return protocol.WriteI64(0)
		}
		if val, ok := json.(int64); ok {
			return protocol.WriteI64(val)
		}
		return errors.New("data type not int64")
	case Int32:
		if json == nil {
			return protocol.WriteI32(0)
		}
		if val, ok := json.(int32); ok {
			return protocol.WriteI32(val)
		}

		if val, ok := json.(int); ok {
			return protocol.WriteI32(int32(val))
		}
		return errors.New("data type not int32")
	case Int16:
		if json == nil {
			return protocol.WriteI16(0)
		}
		if val, ok := json.(int16); ok {
			return protocol.WriteI16(val)
		}
		return errors.New("data type not int16")
	case Double:
		if json == nil {
			return protocol.WriteDouble(0)
		}
		if val, ok := json.(float64); ok {
			return protocol.WriteDouble(val)
		}
		return errors.New("data type not double")
	case String:
		if json == nil {
			return protocol.WriteString("")
		}
		if val, ok := json.(string); ok {
			return protocol.WriteString(val)
		}
		return errors.New("data type not string")
	case Bool:
		if json == nil {
			return protocol.WriteBool(false)
		}
		if val, ok := json.(bool); ok {
			return protocol.WriteBool(val)
		}
		return errors.New("data type not bool")
	case Object:
		err := protocol.WriteStructBegin("JsonObject")
		if err != nil {
			return err
		}
		// we write all keys with empty values
		var mapData map[string]interface{}
		var ok bool
		if json != nil {
			if mapData, ok = json.(map[string]interface{}); !ok {
				return errors.New("data should be map<string,...>")
			}
		}
		// so it must be a map

		// should we skip non recognized fields?
		for _, fieldType := range schema.ObjectFields {
			// id starts from 1
			err := protocol.WriteFieldBegin(fieldType.Name, fieldType.Type_.TType(), fieldType.ID)
			if err != nil {
				return err
			}

			err = fieldType.Type_.Write(protocol, mapData[fieldType.Name])
			if err != nil {
				return err
			}

			err = protocol.WriteFieldEnd()
			if err != nil {
				return err
			}
		}

		// must write a Stop
		err = protocol.WriteFieldStop()
		if err != nil {
			return err
		}

		return protocol.WriteStructEnd()
	case Array:

		// we write all keys with empty values
		var listData []interface{}
		var ok bool
		if json != nil {
			if listData, ok = json.([]interface{}); !ok {
				return errors.New("data should be list")
			}
		}

		size := len(listData)
		err := protocol.WriteListBegin(schema.ElementType.TType(), size)
		if err != nil {
			return err
		}

		for i := 0; i < size; i++ {
			err := schema.ElementType.Write(protocol, listData[i])
			if err != nil {
				return err
			}
		}
		return protocol.WriteListEnd()
	case Map:
		// we write all keys with empty values
		var mapData map[string]interface{}
		var ok bool
		if json != nil {
			if mapData, ok = json.(map[string]interface{}); !ok {
				return errors.New("data should be map<string,...>")
			}
		}
		// so it must be a map
		err := protocol.WriteMapBegin(thrift.STRING, schema.ElementType.TType(), len(mapData))
		if err != nil {
			return err
		}

		for k, v := range mapData {
			err := protocol.WriteString(k)
			if err != nil {
				return err
			}
			err = schema.ElementType.Write(protocol, v)
			if err != nil {
				return err
			}
		}
		return protocol.WriteMapEnd()
	default:
		return errors.New(fmt.Sprintf("Unknown json schema type to write:%#v", schema.ObjectType))
	}
}

func (schema *JsonSchemaImp) Read(protocol thrift.TProtocol) (interface{}, error) {
	switch schema.ObjectType {
	case Int64:
		return protocol.ReadI64()
	case Int32:
		return protocol.ReadI32()
	case Int16:
		return protocol.ReadI16()
	case Double:
		return protocol.ReadDouble()
	case String:
		return protocol.ReadString()
	case Bool:
		return protocol.ReadBool()
	case Object:
		_, err := protocol.ReadStructBegin()
		if err != nil {
			return nil, err
		}
		data := make(map[string]interface{}, len(schema.ObjectFields))

		// should we skip non recognized fields?
		for _, fieldType := range schema.ObjectFields {
			_, _, id, err := protocol.ReadFieldBegin()

			if err != nil {
				return data, err
			}

			if fieldType.ID != id {
				return data, errors.New(fmt.Sprintf("field %v expect id = %v, actaul = %v", fieldType.Name, fieldType.ID, id))
			}

			value, err := fieldType.Type_.Read(protocol)
			if err != nil {
				return data, err
			}
			data[fieldType.Name] = value

			err = protocol.ReadFieldEnd()
			if err != nil {
				return data, err
			}
		}

		err = protocol.ReadStructEnd()
		if err != nil {
			return data, err
		}
		return data, nil
	case Array:
		_, size, err := protocol.ReadListBegin()
		if err != nil {
			return nil, err
		}
		var listData = make([]interface{}, size, size)
		for i := 0; i < size; i++ {
			elem, err := schema.ElementType.Read(protocol)
			if err != nil {
				return listData, err
			}
			listData = append(listData, elem)
		}
		err = protocol.ReadListEnd()
		if err != nil {
			return listData, err
		}
		return listData, nil
	case Map:
		_, _, size, err := protocol.ReadMapBegin()
		if err != nil {
			return nil, err
		}
		data := make(map[string]interface{}, size)

		for i := 0; i < size; i++ {
			key, err := protocol.ReadString()
			if err != nil {
				return data, err
			}
			value, err := schema.ElementType.Read(protocol)
			if err != nil {
				return data, err
			}
			data[key] = value
		}

		err = protocol.ReadMapEnd()
		if err != nil {
			return data, err
		}
		return data, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown json schema type to read:%#v", schema.ObjectType))
	}
}
