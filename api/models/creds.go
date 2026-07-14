// Package models defines the persistence models for kuang's credential store.
package models

// Credential is the database model for agent credentials.
type Credential struct {
	ID    string `db:"id" constraints:"primarykey"`
	Agent string `db:"agent" constraints:"notnull"`
	Key   string `db:"key" constraints:"notnull"`
	Value string `db:"value" constraints:"notnull"`
}
