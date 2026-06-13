package lifecycle

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	domainexecution "github.com/gerdong/loomidbx/internal/domain/execution"
)

func TestLifecyclePackageDoesNotDependOnUIBindingOrDatabaseLayers(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller must locate lifecycle package directory")
	}
	packageDir := filepath.Dir(currentFile)
	files, err := filepath.Glob(filepath.Join(packageDir, "*.go"))
	if err != nil {
		t.Fatalf("glob lifecycle package files: %v", err)
	}
	for _, filePath := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports from %s: %v", filePath, err)
		}
		for _, importSpec := range parsed.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			for _, disallowed := range []string{
				"wails",
				"frontend",
				"binding",
				"database/sql",
				"gorm.io/",
				"github.com/mattn/go-sqlite3",
				"modernc.org/sqlite",
				"github.com/go-sql-driver/mysql",
				"github.com/lib/pq",
				"github.com/jackc/pgx",
				"github.com/jmoiron/sqlx",
			} {
				if strings.Contains(strings.ToLower(importPath), disallowed) {
					t.Fatalf("lifecycle package import %q from %s crosses boundary %q", importPath, filepath.Base(filePath), disallowed)
				}
			}
		}
	}
}

func TestLifecycleDoesNotExtendPhase2ExecutionTaskStatusEnum(t *testing.T) {
	knownPhase2Statuses := []domainexecution.ExecutionTaskStatus{
		domainexecution.ExecutionTaskStatusRunning,
		domainexecution.ExecutionTaskStatusSuccess,
		domainexecution.ExecutionTaskStatusPartialFailed,
		domainexecution.ExecutionTaskStatusFailed,
	}
	for _, status := range knownPhase2Statuses {
		if !status.IsKnown() {
			t.Fatalf("Phase 2 execution task status %s should remain known", status)
		}
	}

	for _, lifecycleOnlyStatus := range []domainexecution.ExecutionTaskStatus{
		domainexecution.ExecutionTaskStatus(LifecycleStateInitialized.String()),
		domainexecution.ExecutionTaskStatus(LifecycleStatePrechecking.String()),
		domainexecution.ExecutionTaskStatus(LifecycleStateReady.String()),
		domainexecution.ExecutionTaskStatus(LifecycleStateCancelling.String()),
		domainexecution.ExecutionTaskStatus(LifecycleStateCancelled.String()),
		domainexecution.ExecutionTaskStatus(LifecycleStateCompleted.String()),
	} {
		if lifecycleOnlyStatus.IsKnown() {
			t.Fatalf("Phase 2 execution task status enum must not absorb lifecycle-only state %s", lifecycleOnlyStatus)
		}
	}
}

func TestLifecycleProductionCodeDoesNotDeclareFutureExecutionCapabilities(t *testing.T) {
	for _, filePath := range lifecycleProductionFiles(t) {
		parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, 0)
		if err != nil {
			t.Fatalf("parse lifecycle production file %s: %v", filePath, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			for _, disallowed := range []string{
				"DependencyGraph",
				"TopologicalSort",
				"RowCountPlan",
				"GeneratorRegistry",
				"BatchLoop",
				"ResultAggregator",
				"WriterAdapter",
				"WriteRows",
			} {
				if identifier.Name == disallowed {
					t.Fatalf("lifecycle production code declares future capability %s in %s", disallowed, filepath.Base(filePath))
				}
			}
			return true
		})
	}
}

func TestDownstreamSeamsExposeOnlyStageResultAndSafeErrorBoundary(t *testing.T) {
	resultType := reflect.TypeFor[DownstreamStageResult]()
	assertStructFields(t, resultType, []string{"Status", "Failure", "Artifact"})
	if resultType.Field(1).Type != reflect.TypeFor[*LifecycleError]() {
		t.Fatalf("DownstreamStageResult.Failure = %s, want *LifecycleError", resultType.Field(1).Type)
	}
	if resultType.Field(2).Type.Kind() != reflect.Interface {
		t.Fatalf("DownstreamStageResult.Artifact kind = %s, want opaque interface", resultType.Field(2).Type.Kind())
	}

	for _, seam := range []struct {
		name       string
		typ        reflect.Type
		methodName string
	}{
		{name: "PlannerPort", typ: reflect.TypeFor[PlannerPort](), methodName: "Plan"},
		{name: "GenerationPort", typ: reflect.TypeFor[GenerationPort](), methodName: "Generate"},
		{name: "ResultPort", typ: reflect.TypeFor[ResultPort](), methodName: "Summarize"},
	} {
		method, ok := seam.typ.MethodByName(seam.methodName)
		if !ok {
			t.Fatalf("%s missing method %s", seam.name, seam.methodName)
		}
		if method.Type.NumIn() != 1 || method.Type.In(0) != reflect.TypeFor[DownstreamContext]() {
			t.Fatalf("%s.%s input signature = %s, want DownstreamContext only", seam.name, seam.methodName, method.Type)
		}
		if method.Type.NumOut() != 1 || method.Type.Out(0) != reflect.TypeFor[DownstreamStageResult]() {
			t.Fatalf("%s.%s output signature = %s, want DownstreamStageResult only", seam.name, seam.methodName, method.Type)
		}
	}
}

func lifecycleProductionFiles(t *testing.T) []string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller must locate lifecycle package directory")
	}
	packageDir := filepath.Dir(currentFile)
	files, err := filepath.Glob(filepath.Join(packageDir, "*.go"))
	if err != nil {
		t.Fatalf("glob lifecycle package files: %v", err)
	}
	productionFiles := make([]string, 0, len(files))
	for _, filePath := range files {
		if !strings.HasSuffix(filePath, "_test.go") {
			productionFiles = append(productionFiles, filePath)
		}
	}
	return productionFiles
}
