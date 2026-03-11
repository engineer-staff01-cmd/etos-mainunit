package app_test

import (
	"etos-mainunit/app"
	"math"
	"testing"
)

func TestCalcElectricEnergy(t *testing.T) {
	type arg struct {
		activePower float64
	}

	tests := []struct {
		name string
		args arg
		want float64
	}{
		{
			name: "testcase-1",
			args: arg{
				activePower: 500,
			},
			want: 500.0 / 3600,
		},
		{
			name: "testcase-2",
			args: arg{
				activePower: 10000,
			},
			want: 10000.0 / 3600,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := app.CalcElectricEnergy(tt.args.activePower)
			if !withinTolerance(got, tt.want, 1e-12) {
				t.Logf("got and equal is not equal. got:%f want:%f", got, tt.want)
				t.Fail()
			}
		})
	}
}

func TestCalcOccupancyRate(t *testing.T) {
	type arg struct {
		electricPower         float64
		ratedPowerConsumption float64
	}

	tests := []struct {
		name string
		args arg
		want float64
	}{
		{
			name: "testcase-1",
			args: arg{
				electricPower:         3.0,
				ratedPowerConsumption: 10.0,
			},
			want: 3.0 / 10.0,
		},
		{
			name: "testcase-2",
			args: arg{
				electricPower:         250.5,
				ratedPowerConsumption: 230,
			},
			want: 250.5 / 230,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := app.CalcOccupancyRate(tt.args.electricPower, tt.args.ratedPowerConsumption)
			if !withinTolerance(got, tt.want, 1e-12) {
				t.Logf("got and equal is not equal. got:%f want:%f", got, tt.want)
				t.Fail()
			}
		})
	}
}

func TestCalcCurrentDiscomfortIndex(t *testing.T) {
	type arg struct {
		temperature float64
		humidity    float64
	}

	tests := []struct {
		name string
		args arg
		want float64
	}{
		{
			name: "testcase-1",
			args: arg{
				temperature: 10.5,
				humidity:    0,
			},
			want: 0,
		},
		{
			name: "testcase-2",
			args: arg{
				temperature: 10.5,
				humidity:    50,
			},
			want: 0.81*10.5 + 0.01*50*(0.99*10.5-14.3) + 46.3,
		},
		{
			name: "testcase-3",
			args: arg{
				temperature: -18.5,
				humidity:    35.5,
			},
			want: 0.81*-18.5 + 0.01*35.5*(0.99*-18.5-14.3) + 46.3,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := app.CalcCurrentDiscomfortIndex(tt.args.temperature, tt.args.humidity)
			if !withinTolerance(got, tt.want, 1e-12) {
				t.Logf("got and equal is not equal. got:%f want:%f", got, tt.want)
				t.Fail()
			}
		})
	}
}

// withinTolerance 少数の誤差範囲を許容する
func withinTolerance(a, b, e float64) bool {
	if a == b {
		return true
	}
	d := math.Abs(a - b)
	if b == 0 {
		return d < e
	}
	return (d / math.Abs(b)) < e
}
