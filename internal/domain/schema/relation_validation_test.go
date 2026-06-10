package schema

import (
	"reflect"
	"testing"
	"time"
)

func TestValidateForeignKeyDraftAndPersistedUseDifferentIDAndAuditRules(t *testing.T) {
	foreignKey := ForeignKey{
		TableID:             41,
		FKName:              "fk_order_customer",
		ReferencedTableID:   42,
		ColumnIDs:           []int64{51},
		ReferencedColumnIDs: []int64{61},
	}

	if issues := ValidateForeignKey(foreignKey, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateForeignKey(draft) = %#v, want no issues for zero primary key and createdAt", issues)
	}

	persistedIssues := ValidateForeignKey(foreignKey, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt"})
	assertIssueCodes(t, persistedIssues, map[string]SchemaIssueCode{
		"id":        SchemaIssueCodeInvalidID,
		"createdAt": SchemaIssueCodeInvalidTime,
	})

	negativeID := foreignKey
	negativeID.ID = -1
	assertIssuePaths(t, ValidateForeignKey(negativeID, SchemaValidationModeDraft), []string{"id"})
}

func TestValidateForeignKeyReturnsFieldLevelIssuesForReferencesNamesAndMappings(t *testing.T) {
	issues := ValidateForeignKey(ForeignKey{
		ID:                  -7,
		TableID:             0,
		FKName:              "bad/name",
		ReferencedTableID:   0,
		ColumnIDs:           []int64{0, -2},
		ReferencedColumnIDs: []int64{0},
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "tableId", "fkName", "referencedTableId", "columnIds[0]", "columnIds[1]", "referencedColumnIds[0]", "referencedColumnIds"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{
		"id":                     SchemaIssueCodeInvalidID,
		"tableId":                SchemaIssueCodeInvalidID,
		"fkName":                 SchemaIssueCodeInvalidName,
		"referencedTableId":      SchemaIssueCodeInvalidID,
		"columnIds[0]":           SchemaIssueCodeInvalidID,
		"columnIds[1]":           SchemaIssueCodeInvalidID,
		"referencedColumnIds[0]": SchemaIssueCodeInvalidID,
		"referencedColumnIds":    SchemaIssueCodeInvalidMapping,
	})
	assertAllIssuesSafeErrors(t, issues)

	emptyMappingIssues := ValidateForeignKey(ForeignKey{
		TableID:           1,
		FKName:            "fk_order_customer",
		ReferencedTableID: 2,
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, emptyMappingIssues, []string{"columnIds", "referencedColumnIds"})
	assertIssueCodes(t, emptyMappingIssues, map[string]SchemaIssueCode{
		"columnIds":           SchemaIssueCodeRequired,
		"referencedColumnIds": SchemaIssueCodeRequired,
	})
}

func TestValidateTableRelationDraftAndPersistedUseDifferentIDAndAuditRules(t *testing.T) {
	relation := TableRelation{
		RelationType:    RelationTypeParentChild,
		ParentTableID:   42,
		ChildTableID:    41,
		ParentColumnIDs: []int64{61},
		ChildColumnIDs:  []int64{51},
		MultiplierMin:   0,
		MultiplierMax:   2,
	}

	if issues := ValidateTableRelation(relation, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateTableRelation(draft) = %#v, want no issues for zero primary key and audit times", issues)
	}

	persistedIssues := ValidateTableRelation(relation, SchemaValidationModePersisted)
	assertIssuePaths(t, persistedIssues, []string{"id", "createdAt", "updatedAt"})
	assertIssueCodes(t, persistedIssues, map[string]SchemaIssueCode{
		"id":        SchemaIssueCodeInvalidID,
		"createdAt": SchemaIssueCodeInvalidTime,
		"updatedAt": SchemaIssueCodeInvalidTime,
	})

	negativeID := relation
	negativeID.ID = -1
	assertIssuePaths(t, ValidateTableRelation(negativeID, SchemaValidationModeDraft), []string{"id"})
}

func TestValidateTableRelationReturnsFieldLevelIssuesForEnumsReferencesRangesAndMappings(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(-time.Minute)

	issues := ValidateTableRelation(TableRelation{
		ID:              -3,
		RelationType:    RelationType("FUTURE_RELATION"),
		ParentTableID:   0,
		ChildTableID:    0,
		ParentColumnIDs: []int64{0, -4},
		ChildColumnIDs:  []int64{0},
		MultiplierMin:   -1,
		MultiplierMax:   -2,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, SchemaValidationModeDraft)

	assertIssuePaths(t, issues, []string{"id", "relationType", "parentTableId", "childTableId", "parentColumnIds[0]", "parentColumnIds[1]", "childColumnIds[0]", "childColumnIds", "multiplierMin", "multiplierMax", "updatedAt"})
	assertIssueCodes(t, issues, map[string]SchemaIssueCode{
		"id":                 SchemaIssueCodeInvalidID,
		"relationType":       SchemaIssueCodeInvalidType,
		"parentTableId":      SchemaIssueCodeInvalidID,
		"childTableId":       SchemaIssueCodeInvalidID,
		"parentColumnIds[0]": SchemaIssueCodeInvalidID,
		"parentColumnIds[1]": SchemaIssueCodeInvalidID,
		"childColumnIds[0]":  SchemaIssueCodeInvalidID,
		"childColumnIds":     SchemaIssueCodeInvalidMapping,
		"multiplierMin":      SchemaIssueCodeInvalidRange,
		"multiplierMax":      SchemaIssueCodeInvalidRange,
		"updatedAt":          SchemaIssueCodeInvalidTime,
	})
	assertAllIssuesSafeErrors(t, issues)

	emptyMappingIssues := ValidateTableRelation(TableRelation{
		RelationType:  RelationTypeJoinTable,
		ParentTableID: 1,
		ChildTableID:  2,
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, emptyMappingIssues, []string{"parentColumnIds", "childColumnIds"})
	assertIssueCodes(t, emptyMappingIssues, map[string]SchemaIssueCode{
		"parentColumnIds": SchemaIssueCodeRequired,
		"childColumnIds":  SchemaIssueCodeRequired,
	})
}

func TestValidateRelationColumnMappingsRejectDuplicateIDsInMemory(t *testing.T) {
	foreignKeyIssues := ValidateForeignKey(ForeignKey{
		TableID:             41,
		FKName:              "fk_order_customer",
		ReferencedTableID:   42,
		ColumnIDs:           []int64{51, 51},
		ReferencedColumnIDs: []int64{61, 61},
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, foreignKeyIssues, []string{"columnIds", "referencedColumnIds"})
	assertIssueCodes(t, foreignKeyIssues, map[string]SchemaIssueCode{
		"columnIds":           SchemaIssueCodeInvalidMapping,
		"referencedColumnIds": SchemaIssueCodeInvalidMapping,
	})

	relationIssues := ValidateTableRelation(TableRelation{
		RelationType:    RelationTypeParentChild,
		ParentTableID:   42,
		ChildTableID:    41,
		ParentColumnIDs: []int64{61, 61},
		ChildColumnIDs:  []int64{51, 51},
		MultiplierMin:   1,
		MultiplierMax:   1,
	}, SchemaValidationModeDraft)
	assertIssuePaths(t, relationIssues, []string{"parentColumnIds", "childColumnIds"})
	assertIssueCodes(t, relationIssues, map[string]SchemaIssueCode{
		"parentColumnIds": SchemaIssueCodeInvalidMapping,
		"childColumnIds":  SchemaIssueCodeInvalidMapping,
	})
	assertAllIssuesSafeErrors(t, append(foreignKeyIssues, relationIssues...))
}

func TestRelationValidationRejectsUnknownMode(t *testing.T) {
	foreignKeyIssues := ValidateForeignKey(ForeignKey{TableID: 1, FKName: "fk_order_customer", ReferencedTableID: 2, ColumnIDs: []int64{3}, ReferencedColumnIDs: []int64{4}}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, foreignKeyIssues, []string{"mode"})

	relationIssues := ValidateTableRelation(TableRelation{RelationType: RelationTypeParentChild, ParentTableID: 1, ChildTableID: 2, ParentColumnIDs: []int64{3}, ChildColumnIDs: []int64{4}}, SchemaValidationMode("runtime"))
	assertIssuePaths(t, relationIssues, []string{"mode"})
}

func TestRelationIndexedIssuePathsSatisfyValidationIssueContract(t *testing.T) {
	issue := SchemaValidationIssue{
		Path:     "columnIds[0]",
		Code:     SchemaIssueCodeInvalidID,
		Severity: SchemaIssueSeverityError,
		Message:  "columnIds[0] must reference a saved column",
	}

	if issues := ValidateIssue(issue); len(issues) != 0 {
		t.Fatalf("ValidateIssue(indexed path) = %#v, want indexed array paths accepted", issues)
	}

	badIndexIssue := issue
	badIndexIssue.Path = "columnIds[]"
	issues := ValidateIssue(badIndexIssue)
	if !reflect.DeepEqual(issuePaths(issues), []string{"path"}) {
		t.Fatalf("ValidateIssue(empty index path) paths = %#v, want path contract issue in %#v", issuePaths(issues), issues)
	}
}

func TestDecodeRelationJSONReportsMissingRequiredFieldsBeforeValidation(t *testing.T) {
	_, foreignKeyIssues := DecodeForeignKeyJSON([]byte(`{"tableId":41,"referencedTableId":42,"columnIds":[51],"referencedColumnIds":[61]}`), SchemaValidationModeDraft)
	assertIssuePaths(t, foreignKeyIssues, []string{"id", "fkName", "createdAt"})
	assertIssueCodes(t, foreignKeyIssues, map[string]SchemaIssueCode{
		"id":        SchemaIssueCodeRequired,
		"fkName":    SchemaIssueCodeRequired,
		"createdAt": SchemaIssueCodeRequired,
	})
	assertAllIssuesSafeErrors(t, foreignKeyIssues)

	_, relationIssues := DecodeTableRelationJSON([]byte(`{"id":0,"relationType":"PARENT_CHILD","parentTableId":42,"childTableId":41,"parentColumnIds":[61],"childColumnIds":[51],"multiplierMax":1,"createdAt":"0001-01-01T00:00:00Z"}`), SchemaValidationModeDraft)
	assertIssuePaths(t, relationIssues, []string{"multiplierMin", "isLogical", "updatedAt"})
	assertIssueCodes(t, relationIssues, map[string]SchemaIssueCode{
		"multiplierMin": SchemaIssueCodeRequired,
		"isLogical":     SchemaIssueCodeRequired,
		"updatedAt":     SchemaIssueCodeRequired,
	})
	assertAllIssuesSafeErrors(t, relationIssues)
}

func TestDecodeTableRelationJSONAcceptsExplicitFalseAndZeroRequiredValues(t *testing.T) {
	_, issues := DecodeTableRelationJSON([]byte(`{"id":0,"relationType":"PARENT_CHILD","parentTableId":42,"childTableId":41,"parentColumnIds":[61],"childColumnIds":[51],"multiplierMin":0,"multiplierMax":0,"isLogical":false,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`), SchemaValidationModeDraft)
	if len(issues) != 0 {
		t.Fatalf("DecodeTableRelationJSON(explicit zero and false required values) issues = %#v, want none", issues)
	}
}

func TestRelationValidationKeepsCrossObjectIntegrityOutOfScope(t *testing.T) {
	foreignKey := ForeignKey{
		TableID:             41,
		FKName:              "fk_self_parent",
		ReferencedTableID:   41,
		ColumnIDs:           []int64{501},
		ReferencedColumnIDs: []int64{601},
	}
	if issues := ValidateForeignKey(foreignKey, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateForeignKey(self reference with plausible scalar ids) issues = %#v, want no table existence, column ownership, uniqueness, or cycle checks", issues)
	}

	relation := TableRelation{
		RelationType:    RelationTypeJoinTable,
		ParentTableID:   41,
		ChildTableID:    41,
		ParentColumnIDs: []int64{601},
		ChildColumnIDs:  []int64{501},
		MultiplierMin:   0,
		MultiplierMax:   0,
		IsLogical:       true,
	}
	if issues := ValidateTableRelation(relation, SchemaValidationModeDraft); len(issues) != 0 {
		t.Fatalf("ValidateTableRelation(self relation with plausible scalar ids) issues = %#v, want no table existence, column ownership, graph cycle, topological order, or join capacity checks", issues)
	}
}

func issuePaths(issues []SchemaValidationIssue) []string {
	paths := make([]string, 0, len(issues))
	for _, issue := range issues {
		paths = append(paths, issue.Path)
	}
	return paths
}
