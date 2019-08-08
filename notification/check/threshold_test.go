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
	var statements []ast.Statement
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
	lvl := strings.ToLower(c.Level.String())

	fnBody := GreaterThan(Member("r", "_value"), Float(*c.LowerBound))
	fn := Function(FunctionParams("r"), fnBody)

	return DefineVariable(lvl, fn)
}

func GreaterThan(lhs, rhs ast.Expression) ast.Expression {
	return &ast.BinaryExpression{
		Operator: ast.GreaterThanOperator,
		Left:     lhs,
		Right:    rhs,
	}
}

func LessThan(lhs, rhs ast.Expression) ast.Expression {
	return &ast.BinaryExpression{
		Operator: ast.LessThanOperator,
		Left:     lhs,
		Right:    rhs,
	}
}

func Member(p, c string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: p},
		Property: &ast.Identifier{Name: c},
	}
}

func LesserThresholdFunction(c check.ThresholdConfig) ast.Statement {
	fnBody := LessThan(Member("r", "_value"), Float(*c.UpperBound))
	fn := Function(FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return DefineVariable(lvl, fn)
}

func And(lhs, rhs ast.Expression) ast.Expression {
	return &ast.LogicalExpression{
		Operator: ast.AndOperator,
		Left:     lhs,
		Right:    rhs,
	}
}

func RangeThresholdFunction(c check.ThresholdConfig) ast.Statement {
	fnBody := And(
		LessThan(Member("r", "_value"), Float(*c.UpperBound)),
		GreaterThan(Member("r", "_value"), Float(*c.LowerBound)),
	)
	fn := Function(FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return DefineVariable(lvl, fn)
}

func MessageFunction(m string) ast.Statement {
	fn := Function(FunctionParams("r", "check"), String(m))
	return DefineVariable("messageFn", fn)
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
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "alerts"},
			Property: &ast.Identifier{Name: "check"},
		},
		Arguments: []ast.Expression{
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

func String(s string) *ast.StringLiteral {
	return &ast.StringLiteral{
		Value: s,
	}
}

func Float(f float64) *ast.FloatLiteral {
	return &ast.FloatLiteral{
		Value: f,
	}
}

func Function(params []*ast.Property, b ast.Expression) *ast.FunctionExpression {
	return &ast.FunctionExpression{
		Params: params,
		Body:   b,
	}
}

func FunctionParams(args ...string) []*ast.Property {
	var params []*ast.Property
	for _, arg := range args {
		params = append(params, &ast.Property{Key: &ast.Identifier{Name: arg}})
	}
	return params
}

func CheckDefinition(t check.Threshold) ast.Statement {
	tagProperties := []*ast.Property{}
	for _, tag := range t.Tags {
		tagProperties = append(tagProperties, Property(tag.Key, String(tag.Value)))
	}
	tags := Property("tags", Object(tagProperties...))
	checkID := Property("checkID", String(t.ID.String()))

	return DefineVariable("check", Object(checkID, tags))
}

func DefineVariable(id string, e ast.Expression) *ast.VariableAssignment {
	return &ast.VariableAssignment{
		ID: &ast.Identifier{
			Name: id,
		},
		Init: e,
	}
}

func Property(key string, e ast.Expression) *ast.Property {
	return &ast.Property{
		Key: &ast.Identifier{
			Name: key,
		},
		Value: e,
	}
}

func Object(ps ...*ast.Property) *ast.ObjectExpression {
	return &ast.ObjectExpression{
		Properties: ps,
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
			StatusMessageTemplate: "whoa! {check.yeah}",
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
