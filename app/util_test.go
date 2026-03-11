package app_test

import (
	"etos-mainunit/app"
	"testing"
)

func TestClamp(t *testing.T) {
	type arg struct {
		min, max, value int64
	}

	tests := []struct {
		name string
		args arg
		want int64
	}{
		{
			name: "lower-than-min",
			args: arg{
				min:   10,
				max:   20,
				value: 5,
			},
			want: 10,
		},
		{
			name: "equal-min",
			args: arg{
				min:   10,
				max:   20,
				value: 10,
			},
			want: 10,
		},
		{
			name: "within",
			args: arg{
				min:   10,
				max:   20,
				value: 15,
			},
			want: 15,
		},
		{
			name: "equal-max",
			args: arg{
				min:   10,
				max:   20,
				value: 20,
			},
			want: 20,
		},
		{
			name: "higher-than-max",
			args: arg{
				min:   10,
				max:   20,
				value: 30,
			},
			want: 20,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := app.Clamp(tt.args.min, tt.args.max, tt.args.value)
			if got != tt.want {
				t.Logf("got and equal is not equal. got:%d want:%d", got, tt.want)
				t.Fail()
			}
		})
	}
}
