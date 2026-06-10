package schema

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRelationModelsSerializeStableIdentityParentReferencesAndCoreFields(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 10, 11, 0, 0, 0, time.UTC)

	foreignKey := ForeignKey{
		ID:                  701,
		TableID:             401,
		FKName:              "fk_order_customer",
		ReferencedTableID:   402,
		ColumnIDs:           []int64{501, 502},
		ReferencedColumnIDs: []int64{601, 602},
		CreatedAt:           createdAt,
	}
	assertJSONRoundTrip(t, "ForeignKey", foreignKey)
	foreignKeyFields := marshalRelationJSONFields(t, foreignKey)
	assertJSONFieldsPresent(t, foreignKeyFields, "id", "tableId", "fkName", "referencedTableId", "columnIds", "referencedColumnIds", "createdAt")
	assertRelationJSONInt64ArrayField(t, foreignKeyFields, "columnIds", []int64{501, 502})
	assertRelationJSONInt64ArrayField(t, foreignKeyFields, "referencedColumnIds", []int64{601, 602})

	tableRelation := TableRelation{
		ID:              801,
		RelationType:    RelationTypeParentChild,
		ParentTableID:   402,
		ChildTableID:    401,
		ParentColumnIDs: []int64{601, 602},
		ChildColumnIDs:  []int64{501, 502},
		MultiplierMin:   0,
		MultiplierMax:   3,
		IsLogical:       false,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	assertJSONRoundTrip(t, "TableRelation", tableRelation)
	tableRelationFields := marshalRelationJSONFields(t, tableRelation)
	assertJSONFieldsPresent(t, tableRelationFields, "id", "relationType", "parentTableId", "childTableId", "parentColumnIds", "childColumnIds", "multiplierMin", "multiplierMax", "isLogical", "createdAt", "updatedAt")
	assertRelationJSONInt64ArrayField(t, tableRelationFields, "parentColumnIds", []int64{601, 602})
	assertRelationJSONInt64ArrayField(t, tableRelationFields, "childColumnIds", []int64{501, 502})
	if got := string(tableRelationFields["isLogical"]); got != "false" {
		t.Fatalf("isLogical JSON = %s, want explicit false", got)
	}
	if got := string(tableRelationFields["relationType"]); got != `"PARENT_CHILD"` {
		t.Fatalf("relationType JSON = %s, want stable enum string", got)
	}
}

func TestRelationModelsUseScalarReferencesAndNoOutOfScopeFields(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(ForeignKey{}), reflect.TypeOf(TableRelation{})} {
		assertNoRelationOutOfScopeFields(t, typ)
	}

	foreignKeyType := reflect.TypeOf(ForeignKey{})
	assertFieldType(t, foreignKeyType, "ID", reflect.TypeOf(int64(0)))
	assertFieldType(t, foreignKeyType, "TableID", reflect.TypeOf(int64(0)))
	assertFieldType(t, foreignKeyType, "ReferencedTableID", reflect.TypeOf(int64(0)))
	assertFieldType(t, foreignKeyType, "ColumnIDs", reflect.TypeOf([]int64{}))
	assertFieldType(t, foreignKeyType, "ReferencedColumnIDs", reflect.TypeOf([]int64{}))
	assertFieldType(t, foreignKeyType, "CreatedAt", reflect.TypeOf(time.Time{}))

	tableRelationType := reflect.TypeOf(TableRelation{})
	assertFieldType(t, tableRelationType, "ID", reflect.TypeOf(int64(0)))
	assertFieldType(t, tableRelationType, "RelationType", reflect.TypeOf(RelationType("")))
	assertFieldType(t, tableRelationType, "ParentTableID", reflect.TypeOf(int64(0)))
	assertFieldType(t, tableRelationType, "ChildTableID", reflect.TypeOf(int64(0)))
	assertFieldType(t, tableRelationType, "ParentColumnIDs", reflect.TypeOf([]int64{}))
	assertFieldType(t, tableRelationType, "ChildColumnIDs", reflect.TypeOf([]int64{}))
	assertFieldType(t, tableRelationType, "MultiplierMin", reflect.TypeOf(int(0)))
	assertFieldType(t, tableRelationType, "MultiplierMax", reflect.TypeOf(int(0)))
	assertFieldType(t, tableRelationType, "IsLogical", reflect.TypeOf(false))
	assertFieldType(t, tableRelationType, "CreatedAt", reflect.TypeOf(time.Time{}))
	assertFieldType(t, tableRelationType, "UpdatedAt", reflect.TypeOf(time.Time{}))
}

func marshalRelationJSONFields(t *testing.T, value any) map[string]json.RawMessage {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal(%T) returned error: %v", value, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("Unmarshal encoded %T into field map returned error: %v", value, err)
	}
	return fields
}

func assertRelationJSONInt64ArrayField(t *testing.T, fields map[string]json.RawMessage, field string, expected []int64) {
	t.Helper()

	raw, ok := fields[field]
	if !ok {
		t.Fatalf("encoded JSON missing array field %q", field)
	}
	var actual []int64
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("field %q should be an int64 array: %v", field, err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("field %q = %#v, want %#v", field, actual, expected)
	}
}

func assertNoRelationOutOfScopeFields(t *testing.T, typ reflect.Type) {
	t.Helper()

	for index := range typ.NumField() {
		field := typ.Field(index)
		fieldName := strings.ToLower(field.Name)
		jsonName := strings.ToLower(strings.Split(field.Tag.Get("json"), ",")[0])
		for _, forbidden := range []string{"service", "api", "ui", "wails", "vue", "execution", "engine", "driver", "sql", "project", "row", "order", "sort", "algorithm"} {
			if strings.Contains(fieldName, forbidden) || strings.Contains(jsonName, forbidden) {
				t.Fatalf("%s.%s exposes out-of-scope field matching %q with json tag %q", typ.Name(), field.Name, forbidden, field.Tag.Get("json"))
			}
		}
	}
}
