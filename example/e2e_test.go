package example

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/bergundy/protoc-gen-go-nexus/example/gen/example/v1"
	examplenexus "github.com/bergundy/protoc-gen-go-nexus/example/gen/example/v1/examplenexus"
	"github.com/nexus-rpc/sdk-go/contrib/nexusproto"
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
	svc, err := examplenexus.NewTwoWayNexusService(&twoWayHandler{})
	require.NoError(t, err)
	ctx, baseURL := setup(t, svc)
	c, err := examplenexus.NewTwoWayNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	output, err := c.Example(ctx, &example.ExampleInput{Foo: "bar"}, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", output.Foo)
	result, err := c.ExampleAsync(ctx, &example.ExampleInput{Foo: "bar"}, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", result.Successful.Foo)
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
	svc, err := examplenexus.NewOneWayNexusService(&oneWayHandler{})
	require.NoError(t, err)
	ctx, baseURL := setup(t, svc)
	c, err := examplenexus.NewOneWayNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	err = c.NoOutput(ctx, &example.ExampleInput{Foo: "bar"}, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	noOutResult, err := c.NoOutputAsync(ctx, &example.ExampleInput{Foo: "bar"}, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Nil(t, noOutResult.Pending)

	output, err := c.NoInput(ctx, nexus.ExecuteOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", output.Foo)
	noInResult, err := c.NoInputAsync(ctx, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "bar", noInResult.Successful.Foo)
}

type multiStageHandler struct {
	examplenexus.UnimplementedMultiStageNexusHandler
}

type multiStageExecuteHandler struct {
	nexus.UnimplementedOperation[*example.ExecuteRequest, *example.ExecuteResponse]
	name string
}

func (h *multiStageExecuteHandler) Name() string {
	return h.name
}

func (h *multiStageExecuteHandler) Start(ctx context.Context, req *example.ExecuteRequest, _ nexus.StartOperationOptions) (nexus.HandlerStartOperationResult[*example.ExecuteResponse], error) {
	return examplenexus.NewMultiStageExecuteOperationHandlerResultAsync(req.Id, &example.ExecuteStartResult{
		RunId: "run-id",
	}, nil), nil
}

func (h *multiStageExecuteHandler) GetResult(ctx context.Context, id string, _ nexus.GetOperationResultOptions) (*example.ExecuteResponse, error) {
	if id != "my-id" {
		return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "expected ID to be 'my-id', got %q", id)
	}
	return &example.ExecuteResponse{Ok: true}, nil
}

// NoOutput implements examplenexus.OneWayNexusHandler.
func (o *multiStageHandler) Execute(name string) nexus.Operation[*example.ExecuteRequest, *example.ExecuteResponse] {
	return &multiStageExecuteHandler{name: name}
}

var _ examplenexus.OneWayNexusHandler = &oneWayHandler{}

func TestMultiStage(t *testing.T) {
	svc, err := examplenexus.NewMultiStageNexusService(&multiStageHandler{})
	require.NoError(t, err)
	ctx, baseURL := setup(t, svc)
	c, err := examplenexus.NewMultiStageNexusHTTPClient(nexus.HTTPClientOptions{
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	res, err := c.ExecuteAsync(ctx, &example.ExecuteRequest{Id: "my-id"}, nexus.StartOperationOptions{})
	require.NoError(t, err)
	require.Equal(t, "run-id", res.StartResult.RunId)
	outcome, err := res.Pending.GetResult(ctx, nexus.GetOperationResultOptions{})
	require.NoError(t, err)
	require.True(t, outcome.Ok)
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
