package capability_test

import (
	"testing"

	"github.com/gerdong/loomidbx/internal/dbx/capability"
)

func TestExampleCapabilitiesExposeNegotiationDifferences(t *testing.T) {
	mysql := capability.MySQLExample()
	postgres := capability.PostgreSQLExample()

	if !mysql.SupportsBatchInsert || !postgres.SupportsBatchInsert {
		t.Fatal("example capabilities must allow batch-insert strategy selection")
	}
	if mysql.SupportsReturning {
		t.Fatal("MySQL example must not claim RETURNING support for first-phase validation")
	}
	if !postgres.SupportsReturning {
		t.Fatal("PostgreSQL example must keep RETURNING capability visible")
	}
	if mysql.MaxIdentifierLength == postgres.MaxIdentifierLength {
		t.Fatal("example capabilities should preserve identifier length differences")
	}
}
