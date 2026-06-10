package rule

import "github.com/gerdong/loomidbx/internal/domain/schema"

func ruleValidationIssue(path string, code schema.SchemaIssueCode, message string) schema.SchemaValidationIssue {
	return schema.SchemaValidationIssue{
		Path:     path,
		Code:     code,
		Severity: schema.SchemaIssueSeverityError,
		Message:  message,
	}
}
