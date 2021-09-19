package clause

// Clause represents an SQL clause in a SQL query
type Clause struct {
	text string
	args []interface{}
}

func (c *Clause) Text() string {
	return c.text
}

func (c *Clause) Args() []interface{} {
	return c.args
}

// New creates a Clause from the given arguments
func New(clause string, args ...interface{}) *Clause {
	return &Clause{
		text: clause,
		args: args,
	}
}
