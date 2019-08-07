package check_test

import (
	"strings"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/influxdb/notification"
	"github.com/influxdata/influxdb/notification/check"
)

// "github.com/davecgh/go-spew/spew"
// "github.com/influxdata/flux/parser"

func GenerateAST(t check.Threshold) *ast.File {
	return &ast.File{
		Name:    "threshold.flux",
		Imports: []*ast.ImportDeclaration{ImportStatement()},
		Body:    GenerateBody(t),
	}
}

func ImportStatement() *ast.ImportDeclaration {
	return &ast.ImportDeclaration{
		Path: &ast.StringLiteral{
			Value: "influxdata/influxdb/alerts",
		},
	}
}

func GenerateBody(t check.Threshold) []ast.Statement {
	statements := []ast.Statement{}

	statements = append(statements, CheckDefinition(t))

	statements = append(statements, ThresholdFunctions(t.Thresholds)...)

	statements = append(statements, MessageFunction(t.StatusMessageTemplate))

	statements = append(statements, ChecksFunction(t))

	return statements
}

func ThresholdFunctions(cs []check.ThresholdConfig) []ast.Statement {
	thresholdStatements := []ast.Statement{}

	for _, c := range cs {
		if c.UpperBound == nil {
			thresholdStatements = append(thresholdStatements, GreaterThresholdFunction(c))
		} else if c.LowerBound == nil {
			thresholdStatements = append(thresholdStatements, LesserThresholdFunction(c))
		} else {
			thresholdStatements = append(thresholdStatements, RangeThresholdFunction(c))
		}
		//need without range here
	}
	return thresholdStatements
}

func GreaterThresholdFunction(c check.ThresholdConfig) ast.Statement {
	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: strings.ToLower(c.Level.String()),
		},
		Init: &ast.FunctionExpression{
			Params: []*ast.Property{
				&ast.Property{
					Key: &ast.Identifier{Name: "r"},
				},
			},
			Body: &ast.BinaryExpression{
				Operator: ast.GreaterThanOperator,
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "r"},
					Property: &ast.Identifier{Name: "_value"},
				},
				Right: &ast.FloatLiteral{
					Value: *c.LowerBound,
				},
			},
		},
	}
}

func LesserThresholdFunction(c check.ThresholdConfig) ast.Statement {
	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: strings.ToLower(c.Level.String()),
		},
		Init: &ast.FunctionExpression{
			Params: []*ast.Property{
				&ast.Property{
					Key: &ast.Identifier{Name: "r"},
				},
			},
			Body: &ast.BinaryExpression{
				Operator: ast.LessThanOperator,
				Left: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "r"},
					Property: &ast.Identifier{Name: "_value"},
				},
				Right: &ast.FloatLiteral{
					Value: *c.UpperBound,
				},
			},
		},
	}
}

func RangeThresholdFunction(c check.ThresholdConfig) ast.Statement {
	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: strings.ToLower(c.Level.String()),
		},
		Init: &ast.FunctionExpression{
			Params: []*ast.Property{
				&ast.Property{
					Key: &ast.Identifier{Name: "r"},
				},
			},
			Body: &ast.LogicalExpression{
				Operator: ast.AndOperator,
				Left: &ast.BinaryExpression{
					Operator: ast.GreaterThanOperator,
					Left: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "r"},
						Property: &ast.Identifier{Name: "_value"},
					},
					Right: &ast.FloatLiteral{
						Value: *c.LowerBound,
					},
				},
				Right: &ast.BinaryExpression{
					Operator: ast.LessThanOperator,
					Left: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "r"},
						Property: &ast.Identifier{Name: "_value"},
					},
					Right: &ast.FloatLiteral{
						Value: *c.UpperBound,
					},
				},
			},
		},
	}
}

func MessageFunction(m string) ast.Statement {

	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: "messageFn",
		},
		Init: &ast.FunctionExpression{
			Params: []*ast.Property{
				&ast.Property{
					Key: &ast.Identifier{Name: "r"},
				},
				&ast.Property{
					Key: &ast.Identifier{Name: "check"},
				},
			},
			Body: &ast.StringLiteral{Value: m}, // TODO: call string interpolation here.
		},
	}
}

func ChecksFunction(t check.Threshold) *ast.ExpressionStatement {
	var body ast.Expression = &ast.Identifier{Name: "data"}

	body = AppendPipe(body, ChecksCall(t))

	return &ast.ExpressionStatement{Expression: body}
}

func AppendPipe(base ast.Expression, next *ast.CallExpression) *ast.PipeExpression {
	return &ast.PipeExpression{
		Argument: base,
		Call:     next,
	}
}

func ChecksCall(t check.Threshold) *ast.CallExpression {
	return &ast.CallExpression{Callee: &ast.MemberExpression{
		Object: &ast.Identifier{Name: "alerts"}, Property: &ast.Identifier{Name: "check"},
	}, Arguments: []ast.Expression{
		&ast.ObjectExpression{
			Properties: []*ast.Property{
				{
					Key: &ast.Identifier{Name: "check"}, Value: &ast.Identifier{Name: "check"},
				},
				{
					Key: &ast.Identifier{Name: "messageFn"}, Value: &ast.Identifier{Name: "messageFn"},
				},
				{
					Key: &ast.Identifier{Name: "ok"}, Value: &ast.Identifier{Name: "ok"},
				},
				{
					Key: &ast.Identifier{Name: "info"}, Value: &ast.Identifier{Name: "info"},
				},
				{
					Key: &ast.Identifier{Name: "warn"}, Value: &ast.Identifier{Name: "warn"},
				},
				{
					Key: &ast.Identifier{Name: "crit"}, Value: &ast.Identifier{Name: "crit"},
				},
			},
		},
	},
	}
}

func ObjectPropertyString(key, value string) *ast.Property {
	return &ast.Property{
		Key:   &ast.Identifier{Name: key},
		Value: &ast.StringLiteral{Value: value},
	}
}

func ObjectPropertyTags(tags []notification.Tag) *ast.Property {

	values := []*ast.Property{}

	for _, t := range tags {
		values = append(values, &ast.Property{Key: &ast.Identifier{Name: t.Key}, Value: &ast.StringLiteral{Value: t.Value}})
	}

	return &ast.Property{
		Key:   &ast.Identifier{Name: "tags"},
		Value: &ast.ObjectExpression{Properties: values},
	}
}

func CheckDefinition(t check.Threshold) ast.Statement {
	properties := []*ast.Property{}

	properties = append(properties, ObjectPropertyString("checkID", t.ID.String()))
	properties = append(properties, ObjectPropertyTags(t.Tags))

	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: "check",
		},
		Init: &ast.ObjectExpression{
			Properties: properties,
		},
	}
}

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
			StatusMessageTemplate: "whoa!",
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

	t.Error(ast.Format(GenerateAST(threshold)))
	// t.Error(spew.Sdump(GreaterThanEqualTo(10)))
	// spewString := `data |> alerts.check(check: check, info:info, crit:crit)`
	// t.Error(spew.Sdump(parser.ParseSource(spewString)))
}

// data = from(bucket: "defbuck") |> range(start: -5m) |> aggregateWindow(period: 1m, fn: count)
//

// Example 1
// {
//  ...
//  "name": "defcheck",
//  "tags": [{"defkey": "defvalue"}]
//  "threshold": [
//   {
//    "type": "greater",
//    "allValues": false,
//    "level": "info",
//    "value": 10.0
//   }
//  }
// }

// import "influxdata/influxdb/alerts"
//
// info = (r) => r._value >= 10.0
// sets = (tables=<-) => tables |> set(key: "checkID", value: "<id of check>") |> set(key: "defkey", value: "defvalue")
// data |> sets() |> alerts.check( name: "defcheck", info: info)

///////////////////////////////////////////////////////////////////////////////////////////////

// Example 2
// {
//  ...
//  "name": "defcheck",
//  "tags": [{"defkey": "defvalue"}]
//  "threshold": [
//   {
//    "type": "range",
//    "allValues": false,
//    "level": "info",
//    "upperBound": 10.0
//    "lowerBound": 1.0
//   }
//  }
// }

// import "influxdata/influxdb/alerts"
//
// info = (r) => r._value <= 10.0 and r._value >= 1.0
// sets = (tables=<-) => tables |> set(key: "checkID", value: "<id of check>") |> set(key: "defkey", value: "defvalue")
// data |> sets() |> alerts.check( name: "defcheck", info: info)
//
// data |> alerts.check(check: check, info:info, crit:crit)

//  { ..., _level: "info", _name: "defcheck" }
//  { ..., _level: "info", _name: "defcheck" }
//  { ..., _level: "info", _name: "defcheck" }
//  { ..., _level: "info", _name: "defcheck" }
//  { ..., _level: "info", _name: "defcheck" }

///////////////////////////////////////////////////////////////////////////////////////////////

// Example 3
// {
//  ...

// check  = {
//   name: "asdfn",
//   id: "chgeckID",
//   tags: ...
// }
//  "name": "defcheck",
//  "tags": [{"defkey": "defvalue"}]
//  "threshold": [
//   {
//    "type": "greater",
//    "allValues": false,
//    "level": "info",
//    "value": 10.0
//   }
//  }
// }

// import "influxdata/influxdb/alerts"
//
// info = (r) => r._value >= 10.0
// sets = (tables=<-) => tables |> set(key: "checkID", value: "<id of check>") |> set(key: "defkey", value: "defvalue")
// data |> sets() |> alerts.check( name: "defcheck", info: info)
