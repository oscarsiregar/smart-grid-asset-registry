package asset_import

import (
	"mime/multipart"
	"time"

	"smart-grid-asset-registry/domain/assetimport"
	"smart-grid-asset-registry/domain/dto"
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/gofiber/fiber/v2"
)

const (
	formFieldFile = "file"
	formFieldMode = "mode"
)

// upload is a parsed multipart import request. The caller closes file.
type upload struct {
	fileName string
	file     multipart.File
	mode     valueobject.ImportMode
}

// readUpload reads the "file" and "mode" parts of a multipart/form-data
// request (Template Exception 12: the file is read in memory, never stored).
// A missing file part or an unknown mode is 400 INVALID_REQUEST; an empty
// mode means PARTIAL.
func readUpload(ctx *fiber.Ctx) (upload, error) {
	mode := valueobject.ImportModePartial
	if raw := ctx.FormValue(formFieldMode); raw != "" {
		parsed, err := valueobject.ParseImportMode(raw)
		if err != nil {
			return upload{}, apperr.BadRequest(errorcode.InvalidRequest, "mode must be PARTIAL or STRICT")
		}
		mode = parsed
	}

	header, err := ctx.FormFile(formFieldFile)
	if err != nil {
		return upload{}, apperr.BadRequest(errorcode.InvalidRequest, "file is required")
	}
	file, err := header.Open()
	if err != nil {
		return upload{}, err
	}

	return upload{fileName: header.Filename, file: file, mode: mode}, nil
}

// toImportResultResponse maps a report to the HTTP contract. Every issue is
// rendered with its English catalog message; lists are never null.
func toImportResultResponse(
	outcome valueobject.ImportOutcome,
	mode valueobject.ImportMode,
	fileName string,
	report assetimport.Report,
	record *entity.AssetImport,
) dto.ImportResultResponse {
	response := dto.ImportResultResponse{
		Outcome:     outcome.String(),
		IsCommitted: outcome == valueobject.ImportOutcomeCommitted,
		Mode:        mode.String(),
		FileName:    fileName,
		Counts: dto.ImportCountsResponse{
			Total:       report.Counts.Total,
			Imported:    report.Counts.ToImport,
			Unchanged:   report.Counts.Unchanged,
			Rejected:    report.Counts.Rejected,
			WarningRows: report.Counts.WarningRows,
		},
		RejectedRows: make([]dto.ImportRejectedRowResponse, 0, len(report.Rejected)),
		Warnings:     make([]dto.ImportWarningResponse, 0, len(report.Warnings)),
	}

	for _, rejected := range report.Rejected {
		issues := make([]dto.ImportIssueResponse, 0, len(rejected.Errors))
		for _, issue := range rejected.Errors {
			issues = append(issues, dto.ImportIssueResponse{Code: issue.Code, Message: apperr.Message(issue.Code, issue.Args...)})
		}
		response.RejectedRows = append(response.RejectedRows, dto.ImportRejectedRowResponse{
			Row:     rejected.Row,
			AssetID: rejected.AssetID,
			Errors:  issues,
		})
	}

	for _, warning := range report.Warnings {
		response.Warnings = append(response.Warnings, dto.ImportWarningResponse{
			Row:     warning.Row,
			AssetID: warning.AssetID,
			Code:    warning.Issue.Code,
			Message: apperr.Message(warning.Issue.Code, warning.Issue.Args...),
		})
	}

	if record != nil {
		response.Import = &dto.ImportRecordResponse{
			ID:        record.ID.String(),
			CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return response
}
