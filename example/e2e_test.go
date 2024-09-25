package example

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bergundy/protoc-gen-nexus/example/gen/example/v1"
	examplenexus "github.com/bergundy/protoc-gen-nexus/example/gen/example/v1/examplenexus"
	"github.com/nexus-rpc/sdk-go/contrib/nexusproto"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
)

type twoWayHandler struct {
	examplenexus.UnimplementedTwoWayNexusServiceHandler
}

// Example implements examplenexus.TwoWayNexusServiceHandler.
func (t *twoWayHandler) Example(name string) nexus.Operation[*example.ExampleInput, *example.ExampleOutput] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.ExampleInput, options nexus.StartOperationOptions) (*example.ExampleOutput, error) {
		return &example.ExampleOutput{
			Foo: input.Foo,
		}, nil
	})
}

var _ examplenexus.TwoWayNexusServiceHandler = &twoWayHandler{}

func TestTwoWay(t *testing.T) {
	svc, err := examplenexus.NewTwoWayNexusService(&twoWayHandler{})
	require.NoError(t, err)
	ctx, baseURL := setup(t, svc)
	c, err := examplenexus.NewTwoWayNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	output, err := c.ExecuteExample(ctx, &example.ExampleInput{Foo: "bar"}, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", output.Foo)
	result, err := c.StartExample(ctx, &example.ExampleInput{Foo: "bar"}, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", result.Successful.Foo)
}

type oneWayHandler struct {
	examplenexus.UnimplementedOneWayNexusServiceHandler
}

// NoInput implements examplenexus.OneWayNexusServiceHandler.
func (o *oneWayHandler) NoInput(name string) nexus.Operation[nexus.NoValue, *example.ExampleOutput] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, _ nexus.NoValue, options nexus.StartOperationOptions) (*example.ExampleOutput, error) {
		return &example.ExampleOutput{Foo: "bar"}, nil
	})
}

// NoOutput implements examplenexus.OneWayNexusServiceHandler.
func (o *oneWayHandler) NoOutput(name string) nexus.Operation[*example.ExampleInput, nexus.NoValue] {
	return nexus.NewSyncOperation(name, func(ctx context.Context, input *example.ExampleInput, options nexus.StartOperationOptions) (nexus.NoValue, error) {
		if input.Foo != "bar" {
			return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "input.Foo must be bar")
		}
		return nil, nil
	})
}

var _ examplenexus.OneWayNexusServiceHandler = &oneWayHandler{}

func TestOneWay(t *testing.T) {
	svc, err := examplenexus.NewOneWayNexusService(&oneWayHandler{})
	require.NoError(t, err)
	ctx, baseURL := setup(t, svc)
	c, err := examplenexus.NewOneWayNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	err = c.ExecuteNoOutput(ctx, &example.ExampleInput{Foo: "bar"}, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	noOutResult, err := c.StartNoOutput(ctx, &example.ExampleInput{Foo: "bar"}, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Nil(t, noOutResult.Pending)

	output, err := c.ExecuteNoInput(ctx, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", output.Foo)
	noInResult, err := c.StartNoInput(ctx, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", noInResult.Successful.Foo)
}

func setup(t *testing.T, service *nexus.Service) (ctx context.Context, baseURL string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	reg := nexus.NewServiceRegistry()
	require.NoError(t, reg.Register(service))
	handler, err := reg.NewHandler()
	require.NoError(t, err)

	httpHandler := nexus.NewHTTPHandler(nexus.HandlerOptions{
		Handler:    handler,
		Serializer: nexusproto.Serializer(nexusproto.SerializerModePreferJSON),
	})

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })

	go func() {
		// Ignore for test purposes
		_ = http.Serve(listener, httpHandler)
	}()

	return ctx, fmt.Sprintf("http://%s/", listener.Addr().String())
}
