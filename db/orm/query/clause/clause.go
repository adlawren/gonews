package clause

import (
	"fmt"
	"strings"
)

// Clause represents a SQL query clause
type Clause struct {
	text string
	args []interface{}
}

// Text returns the clause text
func (c *Clause) Text() string {
	return c.text
}

// Args returns the clause arguments
func (c *Clause) Args() []interface{} {
	return c.args
}

// Wrap creates a new Clause consisting of the current Clause surrounded by parentheses
func (c *Clause) Wrap() *Clause {
	return New(fmt.Sprintf("(%s)", c.Text()), c.Args()...)
}

// New creates a Clause from the given arguments
func New(clause string, args ...interface{}) *Clause {
	return &Clause{
		text: clause,
		args: args,
	}
}

// GroupBy returns a group-by clause from the given arguments
func GroupBy(clause string) *Clause {
	return New(fmt.Sprintf("group by %s", clause))
}

// In creates an in clause from the given arguments
func In(args ...interface{}) *Clause {
	if len(args) == 0 {
		return New("in")
	}

	var paramStrings []string
	for idx := 0; idx < len(args); idx++ {
		paramStrings = append(paramStrings, "?")
	}

	return New(fmt.Sprintf("in (%s)", strings.Join(paramStrings, ",")), args...)
}

// InnerJoin returns an inner-join clause from the given arguments
func InnerJoin(clause string) *Clause {
	return New(fmt.Sprintf("inner join %s", clause))
}

// LeftJoin returns a left-join clause from the given arguments
func LeftJoin(clause string) *Clause {
	return New(fmt.Sprintf("left join %s", clause))
}

// Limit creates a limit clause from the given arguments
func Limit(n uint) *Clause {
	return New(fmt.Sprintf("limit %d", n))
}

// OrderBy creates an order-by clause from the given arguments
func OrderBy(clause string) *Clause {
	return New(fmt.Sprintf("order by %s", clause))
}

// Select creates a select clause from the given arguments
func Select(clause string) *Clause {
	return New(fmt.Sprintf("select %s", clause))
}

// Union returns a union clause from the given arguments
func Union(clause string) *Clause {
	return New(fmt.Sprintf("union %s", clause))
}

// Where creates a where clause from the given arguments
func Where(clause string, args ...interface{}) *Clause {
	return New(fmt.Sprintf("where %s", clause), args...)
}
