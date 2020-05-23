# About
This project implements a "Dynamic" thrift call schema using JSON

# Problem it solves
We all known that thrift prefers static code generation rather than dynamic processing.

Thus to call a server a client needs to get a full list of IDL definitions of the server and calls to `thrift` the CLI to generate code stubs.However this is proved less efficient under developing.

# Demonstration
Take a look into [example](jsonthrift/example/), the   [thrift_server.go](jsonthrift/example/thrift_server.go) start a server listening at `0.0.0.0:8888` to accept RPC call, and it handles two call:
 - `Pow2`: given a number returns its square
 - `AddPersonAge`: given a `person` struct with `age`, increment its `age` and return
  
The [json_client.go](jsonthrift/example/json_client.go) calls to `Pow2` and `AddPersonAge`, where server handles `Pow2` using the same `JsonSchema` as the client does, however the server handles `AddPersonAge` without knowing the client uses a `JsonSchema`,it just uses the thrift generated stubs.

This example shows that a while the server side uses thrift generated struct, it is static, the client is free to choose either thrift or a dynamic json schema.