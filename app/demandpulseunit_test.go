package app_test

import (
	"etos-mainunit/app"
	"testing"
)

func TestNewDemandPulseUnit(t *testing.T) {
	d := app.NewDemandPulseUnit()

	if d == nil {
		t.Errorf("failed to NewDemandPulseUnit")
	}
}
