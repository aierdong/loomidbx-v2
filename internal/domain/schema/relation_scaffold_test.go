package schema

import (
	"reflect"
	"testing"
)

func TestRelationDomainScaffoldExportsStableShapes(t *testing.T) {
	assertJSONTags(t, reflect.TypeOf(ForeignKey{}), map[string]string{
		"ID":                  "id",
		"TableID":             "tableId",
		"FKName":              "fkName",
		"ReferencedTableID":   "referencedTableId",
		"ColumnIDs":           "columnIds",
		"ReferencedColumnIDs": "referencedColumnIds",
		"CreatedAt":           "createdAt",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(ForeignKey{}), []string{"id", "tableId", "fkName", "referencedTableId", "columnIds", "referencedColumnIds", "createdAt"})

	assertJSONTags(t, reflect.TypeOf(TableRelation{}), map[string]string{
		"ID":              "id",
		"RelationType":    "relationType",
		"ParentTableID":   "parentTableId",
		"ChildTableID":    "childTableId",
		"ParentColumnIDs": "parentColumnIds",
		"ChildColumnIDs":  "childColumnIds",
		"MultiplierMin":   "multiplierMin",
		"MultiplierMax":   "multiplierMax",
		"IsLogical":       "isLogical",
		"CreatedAt":       "createdAt",
		"UpdatedAt":       "updatedAt",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(TableRelation{}), []string{"id", "relationType", "parentTableId", "childTableId", "parentColumnIds", "childColumnIds", "multiplierMin", "multiplierMax", "isLogical", "createdAt", "updatedAt"})

	assertJSONTags(t, reflect.TypeOf(RelationMultiplicity{}), map[string]string{
		"Min": "min",
		"Max": "max",
	})
	assertStructJSONFieldSet(t, reflect.TypeOf(RelationMultiplicity{}), []string{"min", "max"})
}

func TestRelationDomainScaffoldDeclaresStableEnumsAndEntryPoints(t *testing.T) {
	if got, want := string(RelationTypeParentChild), "PARENT_CHILD"; got != want {
		t.Fatalf("RelationTypeParentChild = %q, want %q", got, want)
	}
	if got, want := string(RelationTypeJoinTable), "JOIN_TABLE"; got != want {
		t.Fatalf("RelationTypeJoinTable = %q, want %q", got, want)
	}

	var _ []SchemaValidationIssue = ValidateForeignKey(ForeignKey{}, SchemaValidationModeDraft)
	var _ []SchemaValidationIssue = ValidateTableRelation(TableRelation{}, SchemaValidationModeDraft)
	var _ []SchemaValidationIssue = ValidateRelationType(RelationTypeParentChild)
	var _ []SchemaValidationIssue = ValidateRelationMultiplicity(0, 0)

	multiplicity, issues := NewRelationMultiplicity(0, 0)
	if multiplicity.Min != 0 || multiplicity.Max != 0 {
		t.Fatalf("NewRelationMultiplicity scaffold returned %#v, want zero min and max", multiplicity)
	}
	if issues == nil {
		t.Fatalf("NewRelationMultiplicity should return an issue slice, even when empty")
	}

	var _ ForeignKey
	var _ TableRelation
	var _ RelationMultiplicity
	var _ RelationType
}
