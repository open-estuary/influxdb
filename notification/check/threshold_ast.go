package check

import (
	"github.com/influxdata/flux/ast"
)

func File(name string, imports []*ast.ImportDeclaration, body []ast.Statement) *ast.File {
	return &ast.File{
		Name:    name,
		Imports: imports,
		Body:    body,
	}
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
