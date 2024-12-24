# protoc-gen-nexus

A Protobuf plugin for generating Nexus code.

**⚠️ EXPERIMENTAL: Generated code structure is subject to change as feedback is collected. ⚠️**

Supported languages:

- Golang
- Java (TBD)

## Installation

### From GitHub releases (recommended)

1. Download an archive from the [latest release](https://github.com/bergundy/protoc-gen-nexus/releases/latest).
2. Extract and add to your system's path.

### Using go install

```
go install github.com/bergundy/protoc-gen-nexus/cmd/protoc-gen-nexus@latest
```

## Usage

### Create a proto file

> NOTE: the directory structure here determines the directory structure of the generated files.

`example/v1/service.proto`

```protobuf
syntax="proto3";

package example.v1;

option go_package = "github.com/bergundy/greet-nexus-example/gen/example/v1;example";

message GreetInput {
  string name = 1;
}

message GreetOutput {
  string greeting = 1;
}

service Greeting {
  rpc Greet(GreetInput) returns (GreetOutput) {
  }
}
```

### Create `buf` config files

> NOTE: Alternatively you may use protoc directly.

`buf.yaml`

```yaml
version: v2
modules:
  - path: .
deps:
  - buf.build/bergundy/protoc-gen-nexus
lint:
  use:
    - BASIC
  except:
    - FIELD_NOT_REQUIRED
    - PACKAGE_NO_IMPORT_CYCLE
breaking:
  use:
    - FILE
  except:
    - EXTENSION_NO_DELETE
    - FIELD_SAME_DEFAULT
```

`buf.gen.yaml`

```yaml
version: v2
clean: true
managed:
  enabled: true
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen
    opt:
      - paths=source_relative
  - local: protoc-gen-nexus
    out: gen
    strategy: all
    opt:
      - paths=source_relative
      - lang=go
```

### Generate code 

```
buf generate
```

### Create a handler and serve on localhost:7243

`server/main.go`

```go
package main

import (
        "context"
        "log"
        "net"
        "net/http"

        "github.com/bergundy/greet-nexus-example/gen/example/v1"
        "github.com/bergundy/greet-nexus-example/gen/example/v1/examplenexus"
        "github.com/nexus-rpc/sdk-go/contrib/nexusproto"
        "github.com/nexus-rpc/sdk-go/nexus"
)

type handler struct {
        examplenexus.UnimplementedGreetingNexusServiceHandler
}

func (h *handler) Greet(name string) nexus.Operation[*example.GreetInput, *example.GreetOutput] {
        return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.GreetInput, options nexus.StartOperationOptions) (*example.GreetOutput, error) {
                return &example.GreetOutput{
                        Greeting: "Hello, " + input.Name,
                }, nil
        })
}

func main() {
        service, err := examplenexus.NewGreetingNexusService(&handler{})
        if err != nil {
                log.Fatal(err)
        }
	        if err != nil {
                log.Fatal(err)
        }
        registry := nexus.NewServiceRegistry()
        if err := registry.Register(service); err != nil {
                log.Fatal(err)
        }
        rh, err := registry.NewHandler()
        if err != nil {
                log.Fatal(err)
        }
        h := nexus.NewHTTPHandler(nexus.HandlerOptions{
                Handler:    rh,
                Serializer: nexusproto.Serializer(nexusproto.SerializerModePreferJSON),
        })

        listener, err := net.Listen("tcp", "localhost:7243")
        if err != nil {
                log.Fatal(err)
        }
        defer listener.Close()
        if err = http.Serve(listener, h); err != nil {
                log.Fatal(err)
        }
}
```

`go run ./server`

### Execute an operation with the generated client

`client/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bergundy/greet-nexus-example/gen/example/v1"
	"github.com/bergundy/greet-nexus-example/gen/example/v1/examplenexus"
	"github.com/nexus-rpc/sdk-go/nexus"
)

func main() {
	ctx := context.Background()
	c, err := examplenexus.NewGreetingNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: "http://localhost:7243",
	})
	if err != nil {
		log.Fatal(err)
	}
	output, err := c.Greet(ctx, &example.GreetInput{Name: "World"}, nexus.ExecuteOperationOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got sync greeting:", output.Greeting)
	result, err := c.GreetAsync(ctx, &example.GreetInput{Name: "World"}, nexus.StartOperationOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got async greeting:", result.Successful.Greeting)
}
```

`go run ./client`

## Options

### Service

#### (nexus.v1.service).name

`string`

Defines the Nexus Service name. Defaults to the proto Service full name.

**Example:**

```protobuf
syntax = "proto3";

package example.v1;

import "nexus/v1/options.proto";

service ExampleService {
  option (nexus.v1.service).name = "example.v1.Example";
}
```

### Method

#### (nexus.v1.operation).name

`string`

Defines the Nexus Operation name. Defaults to the proto Method name.

**Example:**

```protobuf
syntax = "proto3";

package example.v1;

import "nexus/v1/options.proto";

service ExampleService {
  rpc Foo(FooInput) returns (FooResponse) {
	option (nexus.v1.operation).name = "foo";
  }
}
```

## Contributing

### Prerequisites

- Go >=1.23
- [Buf](https://buf.build/docs/installation/)

### Generate the proto extension's Go code

```
rm -rf ./gen && buf generate
```

### Build the plugin

```
go build ./cmd/...
```

### Generate example code from protos

```
(cd example && PATH=${PWD}/..:${PATH} buf generate)
```

### Run sanity tests

```
go test ./...
```

### Lint code

[Install](https://golangci-lint.run/welcome/install/) the latest version of `golangci-lint` and run:

```
golangci-lint run ./...
```
