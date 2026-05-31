package member

import (
	"sync"
)

var (
	importJobs   = make(map[string]ImportJob)
	importJobsMu sync.RWMutex
)

func SaveImportJob(job ImportJob) {
	importJobsMu.Lock()
	defer importJobsMu.Unlock()

	importJobs[job.ID] = job
}

func GetImportJob(id string) (ImportJob, bool) {
	importJobsMu.RLock()
	defer importJobsMu.RUnlock()

	job, ok := importJobs[id]
	return job, ok
}
