package asset_import_usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func validateInput(lines ...string) usecase.ValidateAssetImportInput {
	return usecase.ValidateAssetImportInput{
		FileName: "grid.csv",
		File:     csvFile(lines...),
		Mode:     valueobject.ImportModePartial,
	}
}

func TestValidateAssetImport(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")

	tests := []struct {
		name    string
		input   usecase.ValidateAssetImportInput
		setup   func(d deps)
		check   func(t *testing.T, output usecase.ValidateAssetImportOutput)
		wantErr error
	}{
		{
			name: "success: reports accepted, rejected and warning rows against the store",
			input: validateInput(
				"TX-1,SS-1,TRANSFORMER,Tx 1,IN_SERVICE,22,1000,,,SER-1,",
				"TX-2,,TRANSFORMER,Tx 2,IN_SERVICE,22,1000,,,,",
			),
			setup: func(d deps) {
				serial := "SER-1"
				d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, nil, []string{"SS-1", "TX-1", "TX-2"}).
					Return([]*entity.Asset{{AssetID: "SS-1", AssetType: valueobject.AssetTypeSubstation}}, nil)
				d.assetRepo.EXPECT().FindBySerialNumbers(mock.Anything, nil, []string{"SER-1"}).
					Return([]*entity.Asset{{AssetID: "TX-OLD", SerialNumber: &serial}}, nil)
			},
			check: func(t *testing.T, output usecase.ValidateAssetImportOutput) {
				report := output.Report
				assert.Equal(t, 2, report.Counts.Total)
				assert.Equal(t, 1, report.Counts.ToImport)
				assert.Equal(t, 1, report.Counts.Rejected)
				assert.Equal(t, 1, report.Counts.WarningRows)
				require.Len(t, report.Rejected, 1)
				assert.Equal(t, errorcode.ParentRequired, report.Rejected[0].Errors[0].Code)
				assert.Equal(t, errorcode.DuplicateSerialNumber, report.Warnings[0].Issue.Code)
			},
		},
		{
			name:    "invalid request shape: 400, nothing is read",
			input:   usecase.ValidateAssetImportInput{FileName: "grid.csv", Mode: valueobject.ImportModePartial},
			setup:   func(deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "file is required"),
		},
		{
			name:    "file-level problem: 422, nothing is read",
			input:   usecase.ValidateAssetImportInput{FileName: "grid.csv", Mode: valueobject.ImportModeStrict, File: csvFile()},
			setup:   func(deps) {},
			wantErr: apperr.UnprocessableEntity(errorcode.ImportFileEmpty),
		},
		{
			name:  "repository error is returned",
			input: validateInput("SS-1,,SUBSTATION,Sub,,,,,,,"),
			setup: func(d deps) {
				d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, nil, []string{"SS-1"}).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().ValidateAssetImport(ctx, tt.input)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, output)
		})
	}
}

func TestValidateAssetImport_Logging(t *testing.T) {
	ctx := context.Background()

	newLoggingUsecase := func(t *testing.T) (assetImportUsecase, deps, *bytes.Buffer) {
		d := newDeps(t)
		var logs bytes.Buffer
		u := d.usecase()
		u.log = logrus.New()
		u.log.SetOutput(&logs)
		u.log.SetFormatter(&logrus.JSONFormatter{})
		return u, d, &logs
	}

	t.Run("success logs counts, never row contents", func(t *testing.T) {
		u, d, logs := newLoggingUsecase(t)
		d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, nil, mock.Anything).Return(nil, nil)
		d.assetRepo.EXPECT().FindBySerialNumbers(mock.Anything, nil, mock.Anything).Return(nil, nil)

		_, err := u.ValidateAssetImport(ctx, validateInput("SS-1,,SUBSTATION,Secret Substation Name,,,,,,SERIAL-XYZ,"))

		require.NoError(t, err)
		assert.Contains(t, logs.String(), `"msg":"success"`)
		assert.Contains(t, logs.String(), `"fileName":"grid.csv"`)
		assert.Contains(t, logs.String(), `"toImport":1`)
		assert.NotContains(t, logs.String(), "Secret Substation Name")
		assert.NotContains(t, logs.String(), "SERIAL-XYZ")
	})

	t.Run("a technical failure is logged as an error", func(t *testing.T) {
		u, d, logs := newLoggingUsecase(t)
		d.assetRepo.EXPECT().FindByAssetIDs(mock.Anything, nil, mock.Anything).Return(nil, errors.New("db down"))

		_, err := u.ValidateAssetImport(ctx, validateInput("SS-1,,SUBSTATION,Sub,,,,,,,"))

		require.Error(t, err)
		assert.Contains(t, logs.String(), `"level":"error"`)
		assert.Contains(t, logs.String(), "failed to load stored assets")
	})

	t.Run("a file-level 422 is not logged as an error", func(t *testing.T) {
		u, _, logs := newLoggingUsecase(t)

		_, err := u.ValidateAssetImport(ctx, validateInput())

		require.Error(t, err)
		assert.NotContains(t, logs.String(), `"level":"error"`)
	})
}
