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
