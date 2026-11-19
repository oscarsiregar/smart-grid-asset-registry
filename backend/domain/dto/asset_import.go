package dto

// The CSV import HTTP contract (specs/001-csv-import/contracts/imports-api.md).
// Requests are multipart/form-data (parts "file" and "mode") and are read in
// the handler; only the response is a DTO here.

// ImportResultResponse is the result of validating or confirming a file.
type ImportResultResponse struct {
	Outcome      string                      `json:"outcome"`
	IsCommitted  bool                        `json:"isCommitted"`
	Mode         string                      `json:"mode"`
	FileName     string                      `json:"fileName"`
	Counts       ImportCountsResponse        `json:"counts"`
	RejectedRows []ImportRejectedRowResponse `json:"rejectedRows"`
	Warnings     []ImportWarningResponse     `json:"warnings"`
	Import       *ImportRecordResponse       `json:"import"`
}

// ImportCountsResponse: Total = Imported + Unchanged + Rejected. Imported is
// the number of new valid rows; they were saved only when the outcome is
// COMMITTED.
type ImportCountsResponse struct {
	Total       int `json:"total"`
	Imported    int `json:"imported"`
	Unchanged   int `json:"unchanged"`
	Rejected    int `json:"rejected"`
	WarningRows int `json:"warningRows"`
}

type ImportRejectedRowResponse struct {
	Row     int                   `json:"row"`
	AssetID *string               `json:"assetId"`
	Errors  []ImportIssueResponse `json:"errors"`
}

type ImportIssueResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportWarningResponse struct {
	Row     int    `json:"row"`
	AssetID string `json:"assetId"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ImportRecordResponse identifies the stored import record (null for a
// validate-only result).
type ImportRecordResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}
