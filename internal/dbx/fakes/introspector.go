package fakes

import (
	"context"

	"github.com/gerdong/loomidbx/internal/dbx/introspect"
	"github.com/gerdong/loomidbx/internal/dbx/schema"
)

// Introspector is a deterministic test-only schema introspector.
type Introspector struct {
	// Database stores the schema snapshot returned by Introspect.
	Database *schema.Database

	// Err stores a configured failure returned by Introspect.
	Err error

	// Calls stores the number of Introspect invocations.
	Calls int

	// LastOptions stores the last introspection options.
	LastOptions introspect.Options
}

// NewIntrospector returns a fake introspector for a deterministic schema snapshot.
func NewIntrospector(database *schema.Database) *Introspector {
	return &Introspector{Database: database}
}

// Introspect returns the configured schema snapshot or failure without querying a network connection.
func (i *Introspector) Introspect(ctx context.Context, conn introspect.Connection, opts introspect.Options) (*schema.Database, error) {
	_ = ctx
	_ = conn
	i.Calls++
	i.LastOptions = opts
	if i.Err != nil {
		return nil, i.Err
	}
	return cloneDatabase(i.Database), nil
}

// SampleDatabase returns deterministic fake schema data for tests.
func SampleDatabase() *schema.Database {
	return &schema.Database{
		Name:    "fake_db",
		Catalog: "fake_catalog",
		Raw:     schema.RawMetadata{"source": "fake"},
		Namespaces: []schema.Namespace{{
			Name: "public",
			Raw:  schema.RawMetadata{"source": "fake_namespace"},
			Tables: []schema.Table{{
				Name: "users",
				Columns: []schema.Column{
					{Name: "id", NativeType: "integer", LogicalType: schema.LogicalType{Kind: schema.KindInteger, NativeType: "integer"}, Ordinal: 1, Primary: true, Identity: true, Raw: schema.RawMetadata{"ordinal": 1}},
					{Name: "email", NativeType: "varchar(255)", LogicalType: schema.LogicalType{Kind: schema.KindString, NativeType: "varchar(255)"}, Nullable: false, Ordinal: 2, Unique: true, Raw: schema.RawMetadata{"length": 255}},
				},
				PrimaryKey:        &schema.PrimaryKey{Name: "users_pkey", Columns: []string{"id"}, Raw: schema.RawMetadata{"source": "fake_pk"}},
				UniqueConstraints: []schema.UniqueConstraint{{Name: "users_email_key", Columns: []string{"email"}, Raw: schema.RawMetadata{"source": "fake_unique"}}},
				Indexes:           []schema.Index{{Name: "users_email_idx", Columns: []string{"email"}, Unique: true, Raw: schema.RawMetadata{"source": "fake_index"}}},
				Raw:               schema.RawMetadata{"source": "fake_table"},
			}},
			Views: []schema.View{{Name: "active_users", Definition: "select id, email from users", Raw: schema.RawMetadata{"source": "fake_view"}}},
		}},
	}
}

func cloneDatabase(database *schema.Database) *schema.Database {
	if database == nil {
		return nil
	}
	clone := *database
	clone.Raw = cloneRaw(database.Raw)
	clone.Namespaces = make([]schema.Namespace, len(database.Namespaces))
	for i := range database.Namespaces {
		clone.Namespaces[i] = cloneNamespace(database.Namespaces[i])
	}
	return &clone
}

func cloneNamespace(namespace schema.Namespace) schema.Namespace {
	clone := namespace
	clone.Raw = cloneRaw(namespace.Raw)
	clone.Tables = make([]schema.Table, len(namespace.Tables))
	for i := range namespace.Tables {
		clone.Tables[i] = cloneTable(namespace.Tables[i])
	}
	clone.Views = make([]schema.View, len(namespace.Views))
	for i := range namespace.Views {
		clone.Views[i] = cloneView(namespace.Views[i])
	}
	return clone
}

func cloneTable(table schema.Table) schema.Table {
	clone := table
	clone.Raw = cloneRaw(table.Raw)
	clone.Columns = make([]schema.Column, len(table.Columns))
	for i := range table.Columns {
		clone.Columns[i] = cloneColumn(table.Columns[i])
	}
	if table.PrimaryKey != nil {
		pk := *table.PrimaryKey
		pk.Columns = append([]string(nil), table.PrimaryKey.Columns...)
		pk.Raw = cloneRaw(table.PrimaryKey.Raw)
		clone.PrimaryKey = &pk
	}
	clone.ForeignKeys = append([]schema.ForeignKey(nil), table.ForeignKeys...)
	clone.UniqueConstraints = append([]schema.UniqueConstraint(nil), table.UniqueConstraints...)
	clone.CheckConstraints = append([]schema.CheckConstraint(nil), table.CheckConstraints...)
	clone.Indexes = append([]schema.Index(nil), table.Indexes...)
	return clone
}

func cloneView(view schema.View) schema.View {
	clone := view
	clone.Raw = cloneRaw(view.Raw)
	clone.Columns = make([]schema.Column, len(view.Columns))
	for i := range view.Columns {
		clone.Columns[i] = cloneColumn(view.Columns[i])
	}
	return clone
}

func cloneColumn(column schema.Column) schema.Column {
	clone := column
	clone.Raw = cloneRaw(column.Raw)
	return clone
}

func cloneRaw(raw schema.RawMetadata) schema.RawMetadata {
	if raw == nil {
		return nil
	}
	clone := make(schema.RawMetadata, len(raw))
	for key, value := range raw {
		clone[key] = value
	}
	return clone
}
