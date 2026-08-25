package member

import (
	"sync"
)

var (
	importJobs   = make(map[string]ImportJob)
	importClaims = make(map[string]map[int]bool)
	importJobsMu sync.RWMutex
)

func SaveImportJob(job ImportJob) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	importJobs[job.ID] = job
	delete(importClaims, job.ID)
}

func GetImportJob(id string) (ImportJob, bool) {
	importJobsMu.RLock()
	defer importJobsMu.RUnlock()

	job, ok := importJobs[id]
	return job, ok
}

func ClaimImportError(jobID string, row int) (ImportError, bool) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	job, exists := importJobs[jobID]
	if !exists {
		return ImportError{}, false
	}
	if importClaims[jobID] == nil {
		importClaims[jobID] = make(map[int]bool)
	}
	if importClaims[jobID][row] {
		return ImportError{}, false
	}
	for _, importError := range job.ErrorRows {
		if importError.Row == row {
			importClaims[jobID][row] = true
			return importError, true
		}
	}
	return ImportError{}, false
}

func CompleteImportError(jobID string, row int, remove bool) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	delete(importClaims[jobID], row)
	if !remove {
		return
	}
	job, exists := importJobs[jobID]
	if !exists {
		return
	}
	remaining := make([]ImportError, 0, len(job.ErrorRows))
	for _, importError := range job.ErrorRows {
		if importError.Row != row {
			remaining = append(remaining, importError)
		}
	}
	job.ErrorRows = remaining
	importJobs[jobID] = job
}
