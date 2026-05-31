package member

import "time"

type ImportResult struct {
	Imported int           `json:"imported"`
	Failed   int           `json:"failed"`
	JobID    string        `json:"job_id,omitempty"`
	Errors   []ImportError `json:"errors,omitempty"`
}

type ImportError struct {
	Row   int      `json:"row"`
	Error string   `json:"error"`
	Data  []string `json:"data"`
}

type ImportJob struct {
	ID        string
	CreatedAt time.Time
	Headers   []string
	ErrorRows []ImportError
}

func normalizeRow(row []string, expectedSize int) []string {
	for len(row) < expectedSize {
		row = append(row, "")
	}

	return row
}
