package app_test

import (
	"etos-mainunit/app"
	"reflect"
	"testing"
)

func TestGenerateSystemAlert(t *testing.T) {
	const testTime = 1000
	type arg struct {
		code      uint32
		argValues []interface{}
	}

	tests := []struct {
		name string
		args arg
		want app.SystemAlert
	}{
		{
			name: "testcase-1",
			args: arg{
				code:      app.ERCD_SUCCESS,
				argValues: []interface{}{"v0.99"},
			},
			want: app.SystemAlert{
				Time:      testTime,
				ErrorCode: "E0000",
				Message:   "connect successfully to cloud. firmware: v0.99",
			},
		},
		{
			name: "testcase-2",
			args: arg{
				code:      app.ERCD_CONTROL_CONDITION_ERROR,
				argValues: nil,
			},
			want: app.SystemAlert{
				Time:      testTime,
				ErrorCode: "E0100",
				Message:   "Control release time is smaller than the required operation time",
			},
		},
		{
			name: "testcase-3",
			args: arg{
				code:      1234,
				argValues: nil,
			},
			want: app.SystemAlert{
				Time:      testTime,
				ErrorCode: "E1234",
				Message:   app.UnknownErrorMessage,
			},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := app.GenerateSystemAlert(testTime, tt.args.code, tt.args.argValues...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("got and equal is not equal. got:%v want:%v", got, tt.want)
				t.Fail()
			}
		})
	}
}
