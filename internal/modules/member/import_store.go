package member

import (
	"sync"
	"time"
)

var (
	importJobs   = make(map[string]ImportJob)
	importClaims = make(map[string]map[int]bool)
	importJobsMu sync.RWMutex
)

const importJobTTL = 30 * time.Minute
const maxImportJobs = 20

func init() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for now := range ticker.C {
			importJobsMu.Lock()
			cleanupImportJobs(now)
			importJobsMu.Unlock()
		}
	}()
}

// Caller holds importJobsMu.
func cleanupImportJobs(now time.Time) {
	for id, job := range importJobs {
		if !now.Before(job.CreatedAt.Add(importJobTTL)) {
			delete(importJobs, id)
			delete(importClaims, id)
		}
	}
}

func SaveImportJob(job ImportJob) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	cleanupImportJobs(time.Now())
	if len(importJobs) >= maxImportJobs {
		var oldest string
		for id, existing := range importJobs {
			if oldest == "" || existing.CreatedAt.Before(importJobs[oldest].CreatedAt) {
				oldest = id
			}
		}
		delete(importJobs, oldest)
		delete(importClaims, oldest)
	}
	importJobs[job.ID] = job
	delete(importClaims, job.ID)
}

func GetImportJob(id string, ownerID int64) (ImportJob, bool) {
	importJobsMu.RLock()
	defer importJobsMu.RUnlock()

	job, ok := importJobs[id]
	if !ok || job.OwnerID != ownerID || !time.Now().Before(job.CreatedAt.Add(importJobTTL)) {
		return ImportJob{}, false
	}
	return job, true
}

func ClaimImportError(jobID string, row int, ownerID int64) (ImportError, bool) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	job, exists := importJobs[jobID]
	if !exists || job.OwnerID != ownerID || !time.Now().Before(job.CreatedAt.Add(importJobTTL)) {
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
