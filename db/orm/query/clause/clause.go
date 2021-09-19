package clause

// Clause represents an SQL clause in a SQL query
type Clause struct {
	str  string
	args []interface{}
}

func (c *Clause) Text() string {
	return c.str
}

func (c *Clause) Args() []interface{} {
	return c.args
}

// New creates a Clause from the given arguments
func New(clause string, args ...interface{}) *Clause {
	return &Clause{
		str:  clause,
		args: args,
	}
}
