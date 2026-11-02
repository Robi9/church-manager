package member

import (
	"testing"
	"time"
)

func TestImportOwnershipAndExpiration(t *testing.T) {
	id := "private-import-test"
	SaveImportJob(ImportJob{ID: id, OwnerID: 9, ErrorRows: []ImportError{{Row: 2}}})
	if _, ok := GetImportJob(id, 8); ok {
		t.Fatal("another user read an import")
	}
	if _, ok := ClaimImportError(id, 2, 8); ok {
		t.Fatal("another user claimed a row")
	}
	if _, ok := GetImportJob(id, 9); !ok {
		t.Fatal("owner cannot read import")
	}
	if _, ok := ClaimImportError(id, 2, 9); !ok {
		t.Fatal("owner cannot claim row")
	}
	CompleteImportError(id, 2, false)
	importJobsMu.Lock()
	job := importJobs[id]
	job.CreatedAt = time.Now().Add(-importJobTTL)
	importJobs[id] = job
	importJobsMu.Unlock()
	if _, ok := GetImportJob(id, 9); ok {
		t.Fatal("expired import still accessible")
	}
	if _, ok := ClaimImportError(id, 2, 9); ok {
		t.Fatal("expired import can be modified")
	}
	importJobsMu.Lock()
	cleanupImportJobs(time.Now())
	_, exists := importJobs[id]
	_, claimsExist := importClaims[id]
	importJobsMu.Unlock()
	if exists || claimsExist {
		t.Fatal("expired personal data was not removed")
	}
}

func TestCSVExportNeutralizesFormulas(t *testing.T) {
	input := []string{"=HYPERLINK(\"https://example.org\")", " +123", "@SUM(A1)", "\tformula", "Maria"}
	output := safeCSVRow(input)
	for i := 0; i < 4; i++ {
		if output[i] != "'"+input[i] {
			t.Fatalf("unsafe cell: %q", output[i])
		}
	}
	if output[4] != "Maria" || input[0][0] != '=' {
		t.Fatal("normal or source data changed")
	}
}
