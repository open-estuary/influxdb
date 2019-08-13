package check

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/notification"
	"github.com/influxdata/influxdb/notification/flux"
)

var _ influxdb.Check = &Threshold{}

// Threshold is the threshold check.
type Threshold struct {
	Base
	Thresholds []ThresholdConfig `json:"thresholds"`
}

// Type returns the type of the check.
func (c Threshold) Type() string {
	return "threshold"
}

// Valid returns error if something is invalid.
func (c Threshold) Valid() error {
	if err := c.Base.Valid(); err != nil {
		return err
	}
	for _, cc := range c.Thresholds {
		if err := cc.Valid(); err != nil {
			return err
		}
	}
	return nil
}

func multiError(errs []error) error {
	var b strings.Builder

	for _, err := range errs {
		b.WriteString(err.Error() + "\n")
	}

	return fmt.Errorf(b.String())
}

// GenerateFlux returns a flux script for the threshold provided. If there
// are any errors in the flux that the user provided the function will return
// an error for each error found when the script is parsed.
func (t Threshold) GenerateFlux() (string, error) {
	p, err := t.GenerateFluxAST()
	if err != nil {
		return "", err
	}

	return ast.Format(p), nil
}

// GenerateFlux returns a flux AST for the threshold provided. If there
// are any errors in the flux that the user provided the function will return
// an error for each error found when the script is parsed.
func (t Threshold) GenerateFluxAST() (*ast.Package, error) {
	p := parser.ParseSource(t.Query.Text)

	if errs := ast.GetErrors(p); len(errs) != 0 {
		return nil, multiError(errs)
	}

	f := flux.File(
		"threshold.flux",
		flux.Imports("influxdata/influxdb/alerts"),
		t.generateFluxASTBody(),
	)
	p.Files = append(p.Files, f)

	return p, nil
}

func (t Threshold) generateFluxASTBody() []ast.Statement {
	var statements []ast.Statement
	statements = append(statements, t.generateFluxASTCheckDefinition())
	statements = append(statements, t.generateFluxASTThresholdFunctions()...)
	statements = append(statements, t.generateFluxASTMessageFunction())
	statements = append(statements, t.generateFluxASTChecksFunction())
	return statements
}

func (t Threshold) generateFluxASTMessageFunction() ast.Statement {
	fn := flux.Function(flux.FunctionParams("r", "check"), flux.String(t.StatusMessageTemplate))
	return flux.DefineVariable("messageFn", fn)
}

func (t Threshold) generateFluxASTChecksFunction() ast.Statement {
	return flux.ExpressionStatement(flux.Pipe(flux.Identifier("data"), t.generateFluxASTChecksCall()))
}

func (t Threshold) generateFluxASTChecksCall() *ast.CallExpression {
	objectProps := append(([]*ast.Property)(nil), flux.Property("check", flux.Identifier("check")))
	objectProps = append(objectProps, flux.Property("messageFn", flux.Identifier("messageFn")))
	objectProps = append(objectProps, flux.Property("ok", flux.Identifier("ok")))
	objectProps = append(objectProps, flux.Property("info", flux.Identifier("info")))
	objectProps = append(objectProps, flux.Property("warn", flux.Identifier("warn")))
	objectProps = append(objectProps, flux.Property("crit", flux.Identifier("crit")))

	return flux.Call(flux.Member("alerts", "check"), flux.Object(objectProps...))
}

func (t Threshold) generateFluxASTCheckDefinition() ast.Statement {
	tagProperties := []*ast.Property{}
	for _, tag := range t.Tags {
		tagProperties = append(tagProperties, flux.Property(tag.Key, flux.String(tag.Value)))
	}
	tags := flux.Property("tags", flux.Object(tagProperties...))

	checkID := flux.Property("checkID", flux.String(t.ID.String()))

	return flux.DefineVariable("check", flux.Object(checkID, tags))
}

func (t Threshold) generateFluxASTThresholdFunctions() []ast.Statement {
	thresholdStatements := []ast.Statement{}

	for _, c := range t.Thresholds {
		if c.UpperBound == nil {
			thresholdStatements = append(thresholdStatements, c.generateFluxASTGreaterThresholdFunction())
		} else if c.LowerBound == nil {
			thresholdStatements = append(thresholdStatements, c.generateFluxASTLesserThresholdFunction())
		} else {
			thresholdStatements = append(thresholdStatements, c.generateFluxASTRangeThresholdFunction())
		}
		//need without range here
	}
	return thresholdStatements
}

func (c ThresholdConfig) generateFluxASTGreaterThresholdFunction() ast.Statement {
	fnBody := flux.GreaterThan(flux.Member("r", "_value"), flux.Float(*c.LowerBound))
	fn := flux.Function(flux.FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return flux.DefineVariable(lvl, fn)
}

func (c ThresholdConfig) generateFluxASTLesserThresholdFunction() ast.Statement {
	fnBody := flux.LessThan(flux.Member("r", "_value"), flux.Float(*c.UpperBound))
	fn := flux.Function(flux.FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return flux.DefineVariable(lvl, fn)
}

func (c ThresholdConfig) generateFluxASTRangeThresholdFunction() ast.Statement {
	fnBody := flux.And(
		flux.LessThan(flux.Member("r", "_value"), flux.Float(*c.UpperBound)),
		flux.GreaterThan(flux.Member("r", "_value"), flux.Float(*c.LowerBound)),
	)
	fn := flux.Function(flux.FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return flux.DefineVariable(lvl, fn)
}

type thresholdAlias Threshold

// MarshalJSON implement json.Marshaler interface.
func (c Threshold) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			thresholdAlias
			Type string `json:"type"`
		}{
			thresholdAlias: thresholdAlias(c),
			Type:           c.Type(),
		})
}

// ThresholdConfig is the base of all threshold config.
type ThresholdConfig struct {
	// If true, only alert if all values meet threshold.
	AllValues  bool                    `json:"allValues"`
	Level      notification.CheckLevel `json:"level"`
	LowerBound *float64                `json:"lowerBound,omitempty"`
	UpperBound *float64                `json:"upperBound,omitempty"`
}

// Valid returns error if something is invalid.
func (c ThresholdConfig) Valid() error {
	if c.LowerBound == nil && c.UpperBound == nil {
		return &influxdb.Error{
			Code: influxdb.EInvalid,
			Msg:  "threshold must have at least one lowerBound or upperBound value",
		}
	}
	return nil
}
