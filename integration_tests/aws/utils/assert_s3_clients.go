package utils

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

// AssertAwsClients takes an aws operation name, an aws input, and a function to extract the response value from the aws response.
// This method will invoke the aws operation on both the aws client, the sidekick client, and the failover client and then compare the response values.
func AssertAwsClients[I any](
	t *testing.T,
	ctx context.Context,
	awsOp string,
	awsInput I,
	getRespValue func(t *testing.T, v reflect.Value) reflect.Value,
) {
	testCases := []struct {
		name string
		s3c  *s3.Client
	}{
		{name: "aws", s3c: AwsS3c},
		{name: "sidekick", s3c: SidekickS3c},
		// {name: "failover", s3c: AwsS3c},
	}
	responses := make([]reflect.Value, len(testCases))
	for i, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			resp := invoke(t, tt.s3c, awsOp, ctx, awsInput)
			responses[i] = getRespValue(t, resp)
		})
	}

	t.Run("ResponsesEqual", func(t *testing.T) {
		require.Len(t, responses, len(testCases))
		expected := responses[0]
		for _, v := range responses {
			require.Equal(t, expected.Interface(), v.Interface())
		}
	})
}

// invoke invokes a method on any with the given name and args.
func invoke(t *testing.T, any interface{}, name string, args ...interface{}) reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	values := reflect.ValueOf(any).MethodByName(name).Call(inputs)
	resp := values[0]
	err := values[1]
	require.Nil(t, err.Interface())

	return resp
}
