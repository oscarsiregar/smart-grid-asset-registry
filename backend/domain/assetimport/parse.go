package assetimport

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"smart-grid-asset-registry/domain/errorcode"
)

const bytesPerMB = 1024 * 1024

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Parse reads a CSV import file of at most maxBytes bytes. It is the only
// place the size limit is enforced. The whole file is refused with a
// FileError (spec FR-003) when it is too large, is not UTF-8, contains any CSV
// syntax error, has a header that repeats or lacks a required column, or has
// no data rows. A record with the wrong number of fields is not a file error:
// it is returned as a row and rejected later as MALFORMED_ROW.
func Parse(r io.Reader, maxBytes int64) (ParsedFile, *FileError) {
	content, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return ParsedFile{}, &FileError{Code: errorcode.ImportFileNotCSV, Args: []any{1}}
	}
	if int64(len(content)) > maxBytes {
		return ParsedFile{}, &FileError{Code: errorcode.ImportFileTooLarge, Args: []any{formatSize(maxBytes)}}
	}

	content = bytes.TrimPrefix(content, utf8BOM)
	if !utf8.Valid(content) {
		return ParsedFile{}, &FileError{Code: errorcode.ImportFileNotCSV, Args: []any{firstInvalidUTF8Line(content)}}
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return ParsedFile{}, &FileError{Code: errorcode.ImportFileEmpty}
	}

	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1 // field counts are checked per row
	reader.LazyQuotes = false

	header, err := reader.Read()
	if err != nil {
		return ParsedFile{}, csvSyntaxError(err)
	}
	columnIndex, fileErr := resolveHeader(header)
	if fileErr != nil {
		return ParsedFile{}, fileErr
	}

	file := ParsedFile{HeaderFieldCount: len(header)}
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return ParsedFile{}, csvSyntaxError(err)
		}

		line, _ := reader.FieldPos(0)
		file.Rows = append(file.Rows, Row{
			Number:     line,
			Values:     rowValues(record, columnIndex),
			FieldCount: len(record),
		})
	}

	if len(file.Rows) == 0 {
		return ParsedFile{}, &FileError{Code: errorcode.ImportFileEmpty}
	}
	return file, nil
}

// resolveHeader maps every required column to its position. Names are
// trimmed and matched case-sensitively; unknown columns are ignored.
func resolveHeader(header []string) (map[Column]int, *FileError) {
	positions := map[string][]int{}
	for index, name := range header {
		trimmed := strings.TrimSpace(name)
		positions[trimmed] = append(positions[trimmed], index)
	}

	var repeated, missing []string
	for _, column := range Columns {
		switch indexes := positions[string(column)]; {
		case len(indexes) > 1:
			repeated = append(repeated, string(column))
		case len(indexes) == 0:
			missing = append(missing, string(column))
		}
	}

	if len(repeated) > 0 {
		return nil, &FileError{Code: errorcode.ImportDuplicateColumns, Args: []any{strings.Join(repeated, ", ")}}
	}
	if len(missing) > 0 {
		return nil, &FileError{Code: errorcode.ImportMissingColumns, Args: []any{strings.Join(missing, ", ")}}
	}

	columnIndex := make(map[Column]int, len(Columns))
	for _, column := range Columns {
		columnIndex[column] = positions[string(column)][0]
	}
	return columnIndex, nil
}

// rowValues picks the required columns out of a record. Columns beyond the
// end of a short record are empty.
func rowValues(record []string, columnIndex map[Column]int) map[Column]string {
	values := make(map[Column]string, len(columnIndex))
	for column, index := range columnIndex {
		if index < len(record) {
			values[column] = record[index]
		} else {
			values[column] = ""
		}
	}
	return values
}

func csvSyntaxError(err error) *FileError {
	line := 1
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) {
		line = parseErr.StartLine
	}
	return &FileError{Code: errorcode.ImportFileNotCSV, Args: []any{line}}
}

func firstInvalidUTF8Line(content []byte) int {
	line := 1
	for len(content) > 0 {
		r, size := utf8.DecodeRune(content)
		if r == utf8.RuneError && size <= 1 {
			return line
		}
		if r == '\n' {
			line++
		}
		content = content[size:]
	}
	return line
}

func formatSize(maxBytes int64) string {
	if maxBytes%bytesPerMB == 0 {
		return fmt.Sprintf("%d MB", maxBytes/bytesPerMB)
	}
	return fmt.Sprintf("%d bytes", maxBytes)
}
