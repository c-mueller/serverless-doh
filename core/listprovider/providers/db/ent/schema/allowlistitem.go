package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// AllowlistItem holds the schema definition for the AllowlistItem entity.
type AllowlistItem struct {
	ent.Schema
}

func (AllowlistItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("qname").Unique(),
	}
}

// Fields of the AllowlistItem.
func (AllowlistItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("qname").NotEmpty(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the AllowlistItem.
func (AllowlistItem) Edges() []ent.Edge {
	return nil
}
