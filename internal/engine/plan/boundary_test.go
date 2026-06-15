package plan

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
	domainproject "github.com/gerdong/loomidbx/internal/domain/project"
	domainschema "github.com/gerdong/loomidbx/internal/domain/schema"
	"github.com/gerdong/loomidbx/internal/engine/lifecycle"
)

func TestDependencyPlanDoesNotMutateProjectTableExecutionOrder(t *testing.T) {
	projectSnapshot := domainproject.Project{ID: 101, Name: "demo"}
	tables := []domainproject.ProjectTable{
		{ID: 302, ProjectID: 101, TableID: 202, ExecutionOrder: 9},
		{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 4},
	}
	originalOrders := map[int64]int{
		302: tables[0].ExecutionOrder,
		301: tables[1].ExecutionOrder,
	}
	foreignKeys := []domainschema.ForeignKey{{ID: 401, TableID: 202, ReferencedTableID: 201}}

	result := PlanDependencies(projectSnapshot, tables, foreignKeys, nil, nil)

	if !result.Passed {
		t.Fatalf("expected dependency plan to pass, got %#v", result.BlockingErrors)
	}
	for _, table := range tables {
		if table.ExecutionOrder != originalOrders[table.ID] {
			t.Fatalf("ProjectTable %d ExecutionOrder mutated to %d, want persisted snapshot value %d", table.ID, table.ExecutionOrder, originalOrders[table.ID])
		}
	}
	assertPlannedTableOrder(t, result.Plan.OrderedTables, []PlannedTable{
		{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1},
		{ProjectTableID: 302, TableID: 202, ExecutionOrder: 2},
	})
}

func TestDependencyPlannerLifecycleSeamOnlyPublishesExecutionPlanArtifact(t *testing.T) {
	planner := NewDependencyPlanner(
		domainproject.Project{ID: 101, Name: "demo"},
		[]domainproject.ProjectTable{{ID: 301, ProjectID: 101, TableID: 201, ExecutionOrder: 7}},
		nil,
		nil,
		nil,
	)
	context := lifecycle.DownstreamContext{Input: &lifecycle.ExecutionInput{ProjectID: 101}}

	result := planner.Plan(context)

	if result.Status != lifecycle.DownstreamStageSucceeded {
		t.Fatalf("Status = %s, want %s: %#v", result.Status, lifecycle.DownstreamStageSucceeded, result.Failure)
	}
	if result.Failure != nil {
		t.Fatalf("Failure = %#v, want nil", result.Failure)
	}
	if context.PlanArtifact != nil || context.GenerationArtifact != nil {
		t.Fatalf("planner mutated lifecycle context artifacts: %#v", context)
	}
	artifact, ok := result.Artifact.(*ExecutionPlan)
	if !ok {
		t.Fatalf("Artifact = %T, want *ExecutionPlan", result.Artifact)
	}
	assertPlannedTableOrder(t, artifact.OrderedTables, []PlannedTable{{ProjectTableID: 301, TableID: 201, ExecutionOrder: 1}})
}

func TestDependencyPlanBoundaryDoesNotExtendLifecycleOrPhase2StateEnums(t *testing.T) {
	lifecycleStates := map[lifecycle.LifecycleState]string{
		lifecycle.LifecycleStateInitialized: "INITIALIZED",
		lifecycle.LifecycleStatePrechecking: "PRECHECKING",
		lifecycle.LifecycleStateReady:       "READY",
		lifecycle.LifecycleStateRunning:     "RUNNING",
		lifecycle.LifecycleStateCancelling:  "CANCELLING",
		lifecycle.LifecycleStateCancelled:   "CANCELLED",
		lifecycle.LifecycleStateFailed:      "FAILED",
		lifecycle.LifecycleStateCompleted:   "COMPLETED",
	}
	if len(lifecycleStates) != 8 {
		t.Fatalf("lifecycle state boundary length = %d, want 8", len(lifecycleStates))
	}
	for state, want := range lifecycleStates {
		if got := state.String(); got != want {
			t.Fatalf("LifecycleState.String() = %q, want %q", got, want)
		}
	}

	taskStatuses := map[domainexecution.ExecutionTaskStatus]string{
		domainexecution.ExecutionTaskStatusRunning:       "RUNNING",
		domainexecution.ExecutionTaskStatusSuccess:       "SUCCESS",
		domainexecution.ExecutionTaskStatusPartialFailed: "PARTIAL_FAILED",
		domainexecution.ExecutionTaskStatusFailed:        "FAILED",
	}
	if len(taskStatuses) != 4 {
		t.Fatalf("execution task status boundary length = %d, want 4", len(taskStatuses))
	}
	for status, want := range taskStatuses {
		if got := status.String(); got != want || !status.IsKnown() {
			t.Fatalf("ExecutionTaskStatus = (%q, known=%t), want (%q, true)", got, status.IsKnown(), want)
		}
	}

	tableStatuses := map[domainexecution.ExecutionTableStatus]string{
		domainexecution.ExecutionTableStatusPending: "PENDING",
		domainexecution.ExecutionTableStatusRunning: "RUNNING",
		domainexecution.ExecutionTableStatusSuccess: "SUCCESS",
		domainexecution.ExecutionTableStatusFailed:  "FAILED",
		domainexecution.ExecutionTableStatusSkipped: "SKIPPED",
	}
	if len(tableStatuses) != 5 {
		t.Fatalf("execution table status boundary length = %d, want 5", len(tableStatuses))
	}
	for status, want := range tableStatuses {
		if got := status.String(); got != want || !status.IsKnown() {
			t.Fatalf("ExecutionTableStatus = (%q, known=%t), want (%q, true)", got, status.IsKnown(), want)
		}
	}
}

func TestDependencyPlanPackageDoesNotReferenceRuntimeStateEnums(t *testing.T) {
	for _, file := range planProductionFiles(t) {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		source := string(content)
		for _, forbidden := range []string{
			"internal/domain/" + "execution",
			"Execution" + "TaskStatus",
			"Execution" + "TableStatus",
			"Lifecycle" + "State",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s references runtime state boundary %q", filepath.Base(file), forbidden)
			}
		}
	}
}

func TestDependencyPlanPackageDoesNotDependOnUIStoreFacadeOrDatabaseDrivers(t *testing.T) {
	for _, file := range planProductionFiles(t) {
		parsed := parsePlanProductionFile(t, file)
		for _, importSpec := range parsed.Imports {
			path := strings.Trim(importSpec.Path.Value, "\"")
			for _, forbidden := range []string{
				"github.com/wailsapp/" + "wails",
				"internal/" + "api",
				"internal/" + "binding",
				"internal/" + "store",
				"internal/" + "facade",
				"internal/" + "database",
				"modernc.org/" + "sqlite",
				"database/" + "sql",
				"gorm.io/" + "gorm",
			} {
				if strings.Contains(path, forbidden) {
					t.Fatalf("%s imports forbidden dependency boundary %q", filepath.Base(file), path)
				}
			}
		}

		source := formattedAST(t, parsed)
		for _, forbidden := range []string{
			"Wails",
			"Vue",
			"Frontend",
			"Store",
			"Facade",
			"DatabaseDriver",
			"DBDriver",
			"DriverName",
			"ProductName",
			"DatabaseProduct",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s references forbidden package boundary identifier %q", filepath.Base(file), forbidden)
			}
		}
	}
}

func TestDependencyPlanPackageDoesNotBranchByDatabaseProduct(t *testing.T) {
	for _, file := range planProductionFiles(t) {
		parsed := parsePlanProductionFile(t, file)
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch statement := node.(type) {
			case *ast.IfStmt:
				assertNoDatabaseProductCondition(t, file, statement.Cond)
			case *ast.SwitchStmt:
				assertNoDatabaseProductCondition(t, file, statement.Tag)
			}
			return true
		})
	}
}

func TestExecutionPlanArtifactShapeStaysTableLevelOnly(t *testing.T) {
	planType := reflect.TypeFor[ExecutionPlan]()
	wantFields := []string{"ProjectID", "OrderedTables", "Edges", "Warnings"}
	if planType.NumField() != len(wantFields) {
		t.Fatalf("ExecutionPlan field count = %d, want %d", planType.NumField(), len(wantFields))
	}
	for index, name := range wantFields {
		field := planType.Field(index)
		if field.Name != name {
			t.Fatalf("ExecutionPlan field[%d] = %s, want %s", index, field.Name, name)
		}
	}

	if planType.Field(1).Type != reflect.TypeFor[[]PlannedTable]() {
		t.Fatalf("OrderedTables type = %s, want []PlannedTable", planType.Field(1).Type)
	}
	if planType.Field(2).Type != reflect.TypeFor[[]DependencyEdge]() {
		t.Fatalf("Edges type = %s, want []DependencyEdge", planType.Field(2).Type)
	}
	if planType.Field(3).Type != reflect.TypeFor[[]PlanIssue]() {
		t.Fatalf("Warnings type = %s, want []PlanIssue", planType.Field(3).Type)
	}

	for _, forbidden := range []string{"RowCountPlan", "GenerationContext", "GeneratorRegistry", "BatchLoop", "WriterAdapter", "Transaction", "WriteResult", "RowsWritten"} {
		if _, ok := planType.FieldByName(forbidden); ok {
			t.Fatalf("ExecutionPlan must not expose future capability field %s", forbidden)
		}
	}
}

func TestDependencyPlanPackageDoesNotImplementFutureGenerationOrWriteCapabilities(t *testing.T) {
	for _, file := range planProductionFiles(t) {
		parsed := parsePlanProductionFile(t, file)
		source := formattedAST(t, parsed)
		for _, forbidden := range []string{
			"RowCountPlan",
			"GenerationContext",
			"GeneratorRegistry",
			"BatchLoop",
			"WriterAdapter",
			"Transaction",
			"WriteResult",
			"RowsWritten",
			"BeginTx",
			"ExecContext",
			"QueryContext",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s implements future capability boundary %q", filepath.Base(file), forbidden)
			}
		}
	}
}

func planProductionFiles(t *testing.T) []string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate current test file")
	}
	dir := filepath.Dir(currentFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read plan package directory: %v", err)
	}
	files := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	return files
}

func parsePlanProductionFile(t *testing.T, file string) *ast.File {
	t.Helper()
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, file, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	return parsed
}

func formattedAST(t *testing.T, parsed *ast.File) string {
	t.Helper()
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), parsed); err != nil {
		t.Fatalf("format AST: %v", err)
	}
	return buffer.String()
}

func assertNoDatabaseProductCondition(t *testing.T, file string, expr ast.Expr) {
	t.Helper()
	if expr == nil {
		return
	}
	var buffer bytes.Buffer
	if err := format.Node(&buffer, token.NewFileSet(), expr); err != nil {
		t.Fatalf("format condition in %s: %v", file, err)
	}
	condition := buffer.String()
	for _, forbidden := range []string{"postgres", "mysql", "sqlite", "sqlserver", "mongodb", "redis", "databaseProduct", "driverName"} {
		if strings.Contains(strings.ToLower(condition), strings.ToLower(forbidden)) {
			t.Fatalf("%s branches on database product in condition %q", filepath.Base(file), condition)
		}
	}
}
