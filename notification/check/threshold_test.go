package check_test

import (
	"strings"
	"testing"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/influxdb/notification"
	"github.com/influxdata/influxdb/notification/check"
)

func GenerateAST(t check.Threshold) *ast.File {
	return &ast.File{
		Name:    "threshold.flux",
		Imports: Imports("influxdata/influxdb/alerts"),
		Body:    GenerateBody(t),
	}
}

func GenerateBody(t check.Threshold) []ast.Statement {
	var statements []ast.Statement
	statements = append(statements, CheckDefinition(t))
	statements = append(statements, ThresholdFunctions(t.Thresholds)...)
	statements = append(statements, MessageFunction(t.StatusMessageTemplate))
	statements = append(statements, ChecksFunction())
	return statements
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

func MessageFunction(m string) ast.Statement {
	fn := Function(FunctionParams("r", "check"), String(m))
	return DefineVariable("messageFn", fn)
}

func GreaterThresholdFunction(c check.ThresholdConfig) ast.Statement {
	lvl := strings.ToLower(c.Level.String())

	fnBody := GreaterThan(Member("r", "_value"), Float(*c.LowerBound))
	fn := Function(FunctionParams("r"), fnBody)

	return DefineVariable(lvl, fn)
}

func LesserThresholdFunction(c check.ThresholdConfig) ast.Statement {
	fnBody := LessThan(Member("r", "_value"), Float(*c.UpperBound))
	fn := Function(FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return DefineVariable(lvl, fn)
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

func ChecksFunction() *ast.ExpressionStatement {
	return ExpressionStatement(Pipe(Identifier("data"), ChecksCall()))
}

func ChecksCall() *ast.CallExpression {
	objectProps := append(([]*ast.Property)(nil), Property("check", Identifier("check")))
	objectProps = append(objectProps, Property("messageFn", Identifier("messageFn")))
	objectProps = append(objectProps, Property("ok", Identifier("ok")))
	objectProps = append(objectProps, Property("info", Identifier("info")))
	objectProps = append(objectProps, Property("warn", Identifier("warn")))
	objectProps = append(objectProps, Property("crit", Identifier("crit")))

	return CallExpression(Member("alerts", "check"), Object(objectProps...))
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

func And(lhs, rhs ast.Expression) ast.Expression {
	return &ast.LogicalExpression{
		Operator: ast.AndOperator,
		Left:     lhs,
		Right:    rhs,
	}
}

func Pipe(base ast.Expression, calls ...*ast.CallExpression) *ast.PipeExpression {
	if len(calls) < 1 {
		panic("must pipe forward to at least one *ast.CallExpression")
	}
	pe := appendPipe(base, calls[0])
	for _, call := range calls[1:] {
		pe = appendPipe(pe, call)
	}

	return pe
}

func appendPipe(base ast.Expression, next *ast.CallExpression) *ast.PipeExpression {
	return &ast.PipeExpression{
		Argument: base,
		Call:     next,
	}
}

func CallExpression(fn ast.Expression, args *ast.ObjectExpression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: fn,
		Arguments: []ast.Expression{
			args,
		},
	}
}

func String(s string) *ast.StringLiteral {
	return &ast.StringLiteral{
		Value: s,
	}
}

func Identifier(i string) *ast.Identifier {
	return &ast.Identifier{Name: i}
}

func ExpressionStatement(e ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{Expression: e}
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

func Imports(pkgs ...string) []*ast.ImportDeclaration {
	var is []*ast.ImportDeclaration
	for _, pkg := range pkgs {
		is = append(is, ImportDeclaration(pkg))
	}
	return is
}

func ImportDeclaration(pkg string) *ast.ImportDeclaration {
	return &ast.ImportDeclaration{
		Path: &ast.StringLiteral{
			Value: "influxdata/influxdb/alerts",
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
}
