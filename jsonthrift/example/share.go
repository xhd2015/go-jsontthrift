package example

import (
	json ".."
)

var (
	SimpleIntSchema = json.MustNewJsonSchema(`{
    "type": "Int32"
}`)
	PersonSchema = json.MustNewJsonSchema(`{
    "type": "Object",
    "fields": [
        {
            "name": "age",
                "id":1,
            "description": {
                "type": "int32"
            }
        },
        {
            "name": "alias",
                "id":2,
            "description": {
                "type": "string"
            }
        }
    ]
}`)
)
