// Package plan builds engine dependency plans from Project and Schema snapshots.
package plan

import (
	"strconv"

	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
)

// PlanInput stores the minimum Project and Schema snapshots needed by dependency planning.
type PlanInput struct {
	// ProjectID stores the owning Project identity for the plan boundary.
	ProjectID int64

	// Tables stores validated table-level node inputs.
	Tables []PlanTableInput

	// ForeignKeys stores physical foreign key snapshots available for edge mapping.
	ForeignKeys []domainschema.ForeignKey

	// TableRelations stores Schema-level logical relation snapshots available for edge mapping.
	TableRelations []domainschema.TableRelation

	// ProjectRelations stores Project-level relation configuration snapshots available for edge mapping.
	ProjectRelations []domainproject.ProjectTableRelation
}

// PlanTableInput stores the validated node input for one ProjectTable.
type PlanTableInput struct {
	// ProjectTableID stores the ProjectTable identity used as the graph node identity.
	ProjectTableID int64

	// ProjectID stores the owning Project reference copied from the ProjectTable snapshot.
	ProjectID int64

	// TableID stores the Schema table identity mapped to this ProjectTable node.
	TableID int64

	// ExistingExecutionOrder stores the persisted snapshot order used only as a stable sort key.
	ExistingExecutionOrder int
}

// BuildPlanInput validates ProjectTable snapshots and maps them into dependency planner input.
func BuildPlanInput(projectSnapshot domainproject.Project, tables []domainproject.ProjectTable, foreignKeys []domainschema.ForeignKey, tableRelations []domainschema.TableRelation, projectRelations []domainproject.ProjectTableRelation) (*PlanInput, PlanPrecheckResult) {
	result := PlanPrecheckResult{Passed: true}
	planTables := make([]PlanTableInput, 0, len(tables))
	tableIDs := make(map[int64]struct{}, len(tables))

	for index, table := range tables {
		prefix := "tables[" + strconv.Itoa(index) + "]"
		valid := true
		if table.ID <= 0 {
			result.addBlockingIssue(prefix+".id", PlanErrorCodeRequired, PlanStageInputMapping, "project table identity is required for dependency planning")
			valid = false
		}
		if table.ProjectID <= 0 || (projectSnapshot.ID > 0 && table.ProjectID != projectSnapshot.ID) {
			result.addBlockingIssue(prefix+".projectId", PlanErrorCodeInvalidReference, PlanStageInputMapping, "project table must reference the submitted project")
			valid = false
		}
		if table.TableID <= 0 {
			result.addBlockingIssue(prefix+".tableId", PlanErrorCodeInvalidReference, PlanStageInputMapping, "schema table reference is required for dependency planning")
			valid = false
		}
		if table.TableID > 0 {
			if _, exists := tableIDs[table.TableID]; exists {
				result.addBlockingIssue(prefix+".tableId", PlanErrorCodeDuplicateNode, PlanStageInputMapping, "schema table reference must map to a single project table node")
				valid = false
			} else {
				tableIDs[table.TableID] = struct{}{}
			}
		}
		if !valid {
			continue
		}
		planTables = append(planTables, PlanTableInput{
			ProjectTableID:         table.ID,
			ProjectID:              table.ProjectID,
			TableID:                table.TableID,
			ExistingExecutionOrder: table.ExecutionOrder,
		})
	}

	if len(result.BlockingErrors) > 0 {
		result.Passed = false
		return nil, result
	}

	return &PlanInput{
		ProjectID:        projectSnapshot.ID,
		Tables:           planTables,
		ForeignKeys:      append([]domainschema.ForeignKey(nil), foreignKeys...),
		TableRelations:   append([]domainschema.TableRelation(nil), tableRelations...),
		ProjectRelations: append([]domainproject.ProjectTableRelation(nil), projectRelations...),
	}, result
}
