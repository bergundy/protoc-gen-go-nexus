package example

import (
	"context"
	"testing"

	"github.com/bergundy/protoc-gen-go-nexus/example/gen/example/v1"
	examplenexus "github.com/bergundy/protoc-gen-go-nexus/example/gen/example/v1/examplenexus"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
)

type twoWayHandler struct {
	examplenexus.UnimplementedTwoWayNexusHandler
}

// Example implements examplenexus.TwoWayNexusHandler.
func (t *twoWayHandler) Example(name string) nexus.Operation[*example.ExampleInput, *example.ExampleOutput] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.ExampleInput, options nexus.StartOperationOptions) (*example.ExampleOutput, error) {
		return &example.ExampleOutput{
			Foo: input.Foo,
		}, nil
	})
}

var _ examplenexus.TwoWayNexusHandler = &twoWayHandler{}

func TestTwoWay(t *testing.T) {
	_, err := examplenexus.NewTwoWayNexusService(&twoWayHandler{})
	require.NoError(t, err)
}

type oneWayHandler struct {
	examplenexus.UnimplementedOneWayNexusHandler
}

// NoInput implements examplenexus.OneWayNexusHandler.
func (o *oneWayHandler) NoInput(name string) nexus.Operation[nexus.NoValue, *example.ExampleOutput] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, _ nexus.NoValue, options nexus.StartOperationOptions) (*example.ExampleOutput, error) {
		return &example.ExampleOutput{Foo: "bar"}, nil
	})
}

// NoOutput implements examplenexus.OneWayNexusHandler.
func (o *oneWayHandler) NoOutput(name string) nexus.Operation[*example.ExampleInput, nexus.NoValue] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.ExampleInput, options nexus.StartOperationOptions) (nexus.NoValue, error) {
		if input.Foo != "bar" {
			return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "input.Foo must be bar")
		}
		return nil, nil
	})
}

var _ examplenexus.OneWayNexusHandler = &oneWayHandler{}

func TestOneWay(t *testing.T) {
	_, err := examplenexus.NewOneWayNexusService(&oneWayHandler{})
	require.NoError(t, err)
}

type includeMeWithExcludedMethodsHandler struct {
	examplenexus.UnimplementedIncludeWithExcludedMethodsNexusHandler
}

func (h *includeMeWithExcludedMethodsHandler) Example(name string) nexus.Operation[*example.ExampleInput, *example.ExampleOutput] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.ExampleInput, options nexus.StartOperationOptions) (*example.ExampleOutput, error) {
		return &example.ExampleOutput{
			Foo: input.Foo,
		}, nil
	})
}

var _ examplenexus.IncludeWithExcludedMethodsNexusHandler = &includeMeWithExcludedMethodsHandler{}
