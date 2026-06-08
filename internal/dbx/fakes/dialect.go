package fakes

import (
	"fmt"
	"strings"

	"github.com/gerdong/loomidbx/internal/dbx/core"
	"github.com/gerdong/loomidbx/internal/dbx/dialect"
)

// Dialect is a deterministic test-only SQL primitive renderer.
type Dialect struct {
	// QuoteLeft stores the opening identifier quote character.
	QuoteLeft string

	// QuoteRight stores the closing identifier quote character.
	QuoteRight string

	// PlaceholderPrefix stores the placeholder prefix for numbered placeholders.
	PlaceholderPrefix string

	// Statements stores configured statements returned by BuildInsert when non-empty.
	Statements []dialect.Statement

	// Err stores a configured BuildInsert failure.
	Err error

	// Calls stores the number of BuildInsert invocations.
	Calls int

	// LastRequest stores the last insert request.
	LastRequest dialect.InsertRequest
}

// NewDialect returns a fake dialect with double-quoted identifiers and question placeholders.
func NewDialect() *Dialect {
	return &Dialect{QuoteLeft: "\"", QuoteRight: "\""}
}

// QuoteIdentifier returns a quoted identifier with escaped quote characters.
func (d *Dialect) QuoteIdentifier(name string) string {
	left := d.QuoteLeft
	right := d.QuoteRight
	if left == "" {
		left = "\""
	}
	if right == "" {
		right = left
	}
	return left + strings.ReplaceAll(name, right, right+right) + right
}

// Placeholder returns a question mark or numbered placeholder for a one-based argument index.
func (d *Dialect) Placeholder(index int) string {
	if d.PlaceholderPrefix == "" {
		return "?"
	}
	return fmt.Sprintf("%s%d", d.PlaceholderPrefix, index)
}

// BuildInsert returns configured statements or builds a simple parameterized insert statement.
func (d *Dialect) BuildInsert(req dialect.InsertRequest) ([]dialect.Statement, error) {
	d.Calls++
	d.LastRequest = req
	if d.Err != nil {
		return nil, d.Err
	}
	if len(d.Statements) > 0 {
		return append([]dialect.Statement(nil), d.Statements...), nil
	}
	if req.Table == "" || len(req.Columns) == 0 || len(req.Rows) == 0 {
		return nil, core.UnsupportedDialectOperationError("build insert")
	}

	target := d.QuoteIdentifier(req.Table)
	if req.Schema != "" {
		target = d.QuoteIdentifier(req.Schema) + "." + target
	}
	columnNames := make([]string, len(req.Columns))
	for i, column := range req.Columns {
		columnNames[i] = d.QuoteIdentifier(column.Name)
	}

	args := make([]any, 0, len(req.Columns)*len(req.Rows))
	rowSQL := make([]string, len(req.Rows))
	argIndex := 1
	for rowIndex, row := range req.Rows {
		placeholders := make([]string, len(req.Columns))
		for colIndex, column := range req.Columns {
			placeholders[colIndex] = d.Placeholder(argIndex)
			args = append(args, row[column.Name])
			argIndex++
		}
		rowSQL[rowIndex] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	statement := dialect.Statement{
		SQL:  "INSERT INTO " + target + " (" + strings.Join(columnNames, ", ") + ") VALUES " + strings.Join(rowSQL, ", "),
		Args: args,
	}
	return []dialect.Statement{statement}, nil
}
