package prototype_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aoldershaw/prototype-sdk-go"
	"github.com/stretchr/testify/require"
)

type SimpleObject struct {
	Foo string `json:"foo" prototype:"required"`
	Bar int    `json:"bar"`
}

type SimpleParams struct {
	Baz string `json:"baz" prototype:"required"`
}

type CustomUnmarshal struct {
	PowerOfTen int `json:"power_of_ten"`
}

func (c *CustomUnmarshal) UnmarshalJSON(data []byte) error {
	type target CustomUnmarshal
	var dst target
	if err := json.Unmarshal(data, &dst); err != nil {
		return err
	}
	if !isPowerOfTen(dst.PowerOfTen) {
		return fmt.Errorf("%d is not a power of ten!", dst.PowerOfTen)
	}
	return nil
}

func isPowerOfTen(num int) bool {
	if num == 1 {
		return true
	}
	return num%10 == 0 && isPowerOfTen(num/10)
}

func noop(_ interface{}) []prototype.MessageResponse { return nil }

func TestPrototypeInfo(t *testing.T) {
	for _, tt := range []struct {
		desc         string
		prototype    prototype.Prototype
		object       map[string]interface{}
		expectedMsgs []string
	}{
		{
			desc: "simple object",
			prototype: prototype.New(
				prototype.WithObject(SimpleObject{},
					prototype.WithMessage("msg1", noop),
					prototype.WithMessage("msg2", noop),
				)),
			object: map[string]interface{}{
				"foo": "blah",
			},
			expectedMsgs: []string{"msg1", "msg2"},
		},
		{
			desc: "simple object and params",
			prototype: prototype.New(
				prototype.WithObject(SimpleObject{},
					prototype.WithMessage("msg1", noop),
					prototype.WithMessage("msg2", func(_ SimpleObject, _ SimpleParams) []prototype.MessageResponse {
						return nil
					}),
				)),
			object: map[string]interface{}{
				"foo": "abc",
				"baz": "def",
			},
			// no msg1 because there's an extra unused key in the object
			expectedMsgs: []string{"msg2"},
		},
		{
			desc: "missing required field",
			prototype: prototype.New(
				prototype.WithObject(SimpleObject{},
					prototype.WithMessage("msg1", noop),
					prototype.WithMessage("msg2", noop),
				)),
			object: map[string]interface{}{
				"bar": 123,
			},
			expectedMsgs: []string{},
		},
		{
			desc: "missing required field in params",
			prototype: prototype.New(
				prototype.WithObject(SimpleObject{},
					prototype.WithMessage("msg1", noop),
					prototype.WithMessage("msg2", func(_ SimpleObject, _ SimpleParams) []prototype.MessageResponse {
						return nil
					}),
				)),
			object: map[string]interface{}{
				"foo": "foo",
			},
			expectedMsgs: []string{"msg1"},
		},
		{
			desc: "unmarshal error",
			prototype: prototype.New(
				prototype.WithObject(SimpleObject{},
					prototype.WithMessage("msg1", noop),
				)),
			object: map[string]interface{}{
				"foo": 123, // should be a string
			},
			expectedMsgs: []string{},
		},
		{
			desc: "custom unmarshal success",
			prototype: prototype.New(
				prototype.WithObject(CustomUnmarshal{},
					prototype.WithMessage("msg1", noop),
				)),
			object: map[string]interface{}{
				"power_of_ten": 100,
			},
			expectedMsgs: []string{"msg1"},
		},
		{
			desc: "custom unmarshal error",
			prototype: prototype.New(
				prototype.WithObject(CustomUnmarshal{},
					prototype.WithMessage("msg1", noop),
				)),
			object: map[string]interface{}{
				"power_of_ten": 123,
			},
			expectedMsgs: []string{},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			response, err := tt.prototype.Info(prototype.InfoRequest{Object: tt.object})
			require.NoError(t, err)
			require.ElementsMatch(t, tt.expectedMsgs, response.Messages)
		})
	}
}

func TestPrototypeRun(t *testing.T) {
	expectedObject := SimpleObject{Foo: "foo", Bar: 123}
	expectedParams := SimpleParams{Baz: "baz"}
	expectedResponse := []prototype.MessageResponse{
		{Object: map[string]interface{}{"hello": "world"}},
	}

	proto := prototype.New(
		prototype.WithObject(SimpleObject{},
			prototype.WithMessage("msg", func(object SimpleObject, params SimpleParams) []prototype.MessageResponse {
				require.Equal(t, expectedObject, object)
				require.Equal(t, expectedParams, params)

				return expectedResponse
			}),
		))
	response, err := proto.Run("msg", prototype.MessageRequest{
		Object: map[string]interface{}{
			"foo": expectedObject.Foo,
			"bar": expectedObject.Bar,
			"baz": expectedParams.Baz,
		},
	})
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}
