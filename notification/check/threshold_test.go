package check_test

import (
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/notification"
	"github.com/influxdata/influxdb/notification/check"
)

func TestThreshold_FluxAST(t *testing.T) {

	var l float64 = 10
	var u float64 = 40

	threshold := check.Threshold{
		Base: check.Base{
			Name: "moo",
			Tags: []notification.Tag{
				{Key: "aaa", Value: "vaaa"},
				{Key: "bbb", Value: "vbbb"},
			},
			StatusMessageTemplate: "whoa! {check.yeah}",
			Query: influxdb.DashboardQuery{
				Text: `data = from(bucket: "foo") |> range(start: -1d)`,
			},
		},
		Thresholds: []check.ThresholdConfig{
			check.ThresholdConfig{
				Level:      1,
				LowerBound: &l,
			},
			check.ThresholdConfig{
				Level:      2,
				UpperBound: &u,
			},
			check.ThresholdConfig{
				Level:      3,
				LowerBound: &l,
				UpperBound: &u,
			},
		},
	}

	script, _ := threshold.GenerateFlux()
	t.Error(script)
}
