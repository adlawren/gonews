package client

import (
	"database/sql"
	"gonews/db/orm/query"

	"github.com/pkg/errors"
)

type Client interface {
	All(interface{}) error
	Find(interface{}, ...*query.Clause) error
	FindAll(interface{}, ...*query.Clause) error
	Save(interface{}) error
	// Delete(interface{}) error // TODO
}

func New(db *sql.DB) Client {
	return &client{
		db: db,
	}
}

type client struct {
	db *sql.DB
}

func (c *client) All(ptr interface{}) error {
	// All just calls FindAll without using any clauses, but "All(ptr)" is faster to type than "FindAll(ptr)"
	return c.FindAll(ptr)
}

func (c *client) Find(ptr interface{}, clauses ...*query.Clause) error {
	query, err := query.SelectOne(ptr, clauses...)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(c.db), "failed to execute query")
}

func (c *client) FindAll(ptr interface{}, clauses ...*query.Clause) error {
	query, err := query.Select(ptr, clauses...)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(c.db), "failed to execute query")
}

func (c *client) Save(ptr interface{}) error {
	query, err := query.Upsert(ptr)
	if err != nil {
		return errors.Wrap(err, "failed to create query")
	}

	return errors.Wrap(query.Exec(c.db), "failed to execute query")
}
