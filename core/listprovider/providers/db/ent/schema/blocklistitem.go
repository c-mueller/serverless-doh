package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// BlocklistItem holds the schema definition for the BlocklistItem entity.
type BlocklistItem struct {
	ent.Schema
}

func (BlocklistItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("qname").Unique(),
	}
}

// Fields of the BlocklistItem.
func (BlocklistItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("qname").NotEmpty(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the BlocklistItem.
func (BlocklistItem) Edges() []ent.Edge {
	return nil
}
