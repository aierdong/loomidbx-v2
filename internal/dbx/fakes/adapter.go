package fakes

import (
	"context"
	"sync"

	"github.com/gerdong/loomidbx/internal/dbx/capability"
	"github.com/gerdong/loomidbx/internal/dbx/core"
	"github.com/gerdong/loomidbx/internal/dbx/dialect"
	"github.com/gerdong/loomidbx/internal/dbx/introspect"
	"github.com/gerdong/loomidbx/internal/dbx/typex"
)

// Adapter is a configurable test-only DBX adapter.
type Adapter struct {
	mu sync.Mutex

	// DBType stores the fake database type returned by Type.
	DBType core.DBType

	// Name stores the fake display name returned by DisplayName.
	Name string

	// Caps stores the fake capability model.
	Caps capability.Capabilities

	// ConnectionResult stores the result returned by TestConnection.
	ConnectionResult core.ConnectionTestResult

	// DialectValue stores the fake dialect returned by Dialect.
	DialectValue dialect.Dialect

	// IntrospectorValue stores the fake introspector returned by Introspector.
	IntrospectorValue introspect.Introspector

	// MapperValue stores the fake mapper returned by TypeMapper.
	MapperValue typex.Mapper

	// LastConnectionConfig stores the last connection config passed to TestConnection.
	LastConnectionConfig core.ConnectionConfig

	// TestConnectionCalls stores the number of TestConnection invocations.
	TestConnectionCalls int

	// CapabilitiesCalls stores the number of Capabilities invocations.
	CapabilitiesCalls int

	// DialectCalls stores the number of Dialect invocations.
	DialectCalls int

	// IntrospectorCalls stores the number of Introspector invocations.
	IntrospectorCalls int

	// TypeMapperCalls stores the number of TypeMapper invocations.
	TypeMapperCalls int
}

// NewAdapter returns a fake adapter with deterministic default subcomponents.
func NewAdapter(dbType core.DBType) *Adapter {
	return &Adapter{
		DBType:            dbType,
		Name:              "Fake " + string(dbType),
		ConnectionResult:  core.ConnectionTestResult{OK: true, Message: "ok"},
		DialectValue:      NewDialect(),
		IntrospectorValue: NewIntrospector(SampleDatabase()),
		MapperValue:       NewMapper(),
	}
}

// Type returns the stable fake database type.
func (a *Adapter) Type() core.DBType {
	return a.DBType
}

// DisplayName returns non-sensitive fake adapter metadata.
func (a *Adapter) DisplayName() string {
	if a.Name == "" {
		return "Fake " + string(a.DBType)
	}
	return a.Name
}

// Capabilities returns the configured fake capabilities and records the call.
func (a *Adapter) Capabilities() capability.Capabilities {
	a.mu.Lock()
	a.CapabilitiesCalls++
	a.mu.Unlock()
	return a.Caps
}

// TestConnection returns the configured fake result and records the call.
func (a *Adapter) TestConnection(ctx context.Context, cfg core.ConnectionConfig) core.ConnectionTestResult {
	_ = ctx
	a.mu.Lock()
	a.TestConnectionCalls++
	a.LastConnectionConfig = cfg
	a.mu.Unlock()
	return a.ConnectionResult
}

// Dialect returns the configured fake dialect and records the call.
func (a *Adapter) Dialect() dialect.Dialect {
	a.mu.Lock()
	a.DialectCalls++
	a.mu.Unlock()
	return a.DialectValue
}

// Introspector returns the configured fake introspector and records the call.
func (a *Adapter) Introspector() introspect.Introspector {
	a.mu.Lock()
	a.IntrospectorCalls++
	a.mu.Unlock()
	return a.IntrospectorValue
}

// TypeMapper returns the configured fake mapper and records the call.
func (a *Adapter) TypeMapper() typex.Mapper {
	a.mu.Lock()
	a.TypeMapperCalls++
	a.mu.Unlock()
	return a.MapperValue
}
