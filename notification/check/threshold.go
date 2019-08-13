package check

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/parser"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/notification"
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
	p := parser.ParseSource(t.Query.Text)

	if errs := ast.GetErrors(p); len(errs) != 0 {
		return "", multiError(errs)
	}

	f := t.GenerateFluxAST()
	p.Files = append(p.Files, f)

	return ast.Format(p), nil
}

func (t Threshold) GenerateFluxAST() *ast.File {
	return File(
		"threshold.flux",
		Imports("influxdata/influxdb/alerts"),
		t.generateFluxASTBody(),
	)
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
	fn := Function(FunctionParams("r", "check"), String(t.StatusMessageTemplate))
	return DefineVariable("messageFn", fn)
}

func (t Threshold) generateFluxASTChecksFunction() ast.Statement {
	return ExpressionStatement(Pipe(Identifier("data"), t.generateFluxASTChecksCall()))
}

func (t Threshold) generateFluxASTChecksCall() *ast.CallExpression {
	objectProps := append(([]*ast.Property)(nil), Property("check", Identifier("check")))
	objectProps = append(objectProps, Property("messageFn", Identifier("messageFn")))
	objectProps = append(objectProps, Property("ok", Identifier("ok")))
	objectProps = append(objectProps, Property("info", Identifier("info")))
	objectProps = append(objectProps, Property("warn", Identifier("warn")))
	objectProps = append(objectProps, Property("crit", Identifier("crit")))

	return CallExpression(Member("alerts", "check"), Object(objectProps...))
}

func (t Threshold) generateFluxASTCheckDefinition() ast.Statement {
	tagProperties := []*ast.Property{}
	for _, tag := range t.Tags {
		tagProperties = append(tagProperties, Property(tag.Key, String(tag.Value)))
	}
	tags := Property("tags", Object(tagProperties...))

	checkID := Property("checkID", String(t.ID.String()))

	return DefineVariable("check", Object(checkID, tags))
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
	lvl := strings.ToLower(c.Level.String())

	fnBody := GreaterThan(Member("r", "_value"), Float(*c.LowerBound))
	fn := Function(FunctionParams("r"), fnBody)

	return DefineVariable(lvl, fn)
}

func (c ThresholdConfig) generateFluxASTLesserThresholdFunction() ast.Statement {
	fnBody := LessThan(Member("r", "_value"), Float(*c.UpperBound))
	fn := Function(FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return DefineVariable(lvl, fn)
}

func (c ThresholdConfig) generateFluxASTRangeThresholdFunction() ast.Statement {
	fnBody := And(
		LessThan(Member("r", "_value"), Float(*c.UpperBound)),
		GreaterThan(Member("r", "_value"), Float(*c.LowerBound)),
	)
	fn := Function(FunctionParams("r"), fnBody)

	lvl := strings.ToLower(c.Level.String())

	return DefineVariable(lvl, fn)
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
