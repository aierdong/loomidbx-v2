package project

// RelationValueSource identifies how a ProjectTableRelation obtains relationship values.
type RelationValueSource string

const (
	// RelationValueSourceFromExecution means relationship values come only from records generated in the current execution.
	RelationValueSourceFromExecution RelationValueSource = "FROM_EXECUTION"

	// RelationValueSourceFromDBQuery means relationship values come from existing database records selected by relSourceSql.
	RelationValueSourceFromDBQuery RelationValueSource = "FROM_DB_QUERY"

	// RelationValueSourceMerged means relationship values merge current execution records with existing records selected by relSourceSql.
	RelationValueSourceMerged RelationValueSource = "MERGED"
)

// IsKnown reports whether source is one of the stable Project relation value source enum strings.
func (source RelationValueSource) IsKnown() bool {
	switch source {
	case RelationValueSourceFromExecution, RelationValueSourceFromDBQuery, RelationValueSourceMerged:
		return true
	default:
		return false
	}
}
