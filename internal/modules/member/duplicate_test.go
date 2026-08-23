package member

import (
	"reflect"
	"testing"
)

type candidateFinderStub struct {
	candidates []Member
	excludeID  int64
}

func (f *candidateFinderStub) FindDuplicateCandidates(_ NormalizedMember, excludeID int64) ([]Member, error) {
	f.excludeID = excludeID
	return f.candidates, nil
}

func TestNormalizeNameRemovesAccentsAndCollapsesSpaces(t *testing.T) {
	got := NormalizeName("  MARIA  Dá SILVA ")
	if got != "maria da silva" {
		t.Fatalf("NormalizeName() = %q, want %q", got, "maria da silva")
	}
}

func TestNormalizePhoneTreatsBrazilCountryCodeConsistently(t *testing.T) {
	withCountryCode := NormalizePhone("+55 (88) 9 9999-9999")
	withoutCountryCode := NormalizePhone("(88) 9 9999-9999")
	if withCountryCode != "88999999999" || withCountryCode != withoutCountryCode {
		t.Fatalf("normalized phones differ: %q and %q", withCountryCode, withoutCountryCode)
	}
}

func TestEmptyValuesDoNotScore(t *testing.T) {
	match := (ExactDuplicateMatcher{}).Compare(Member{}, Member{})
	if match.Score != 0 || len(match.MatchedFields) != 0 || match.Risk != RiskLow {
		t.Fatalf("empty values generated a match: %+v", match)
	}
}

func TestDuplicateRiskRules(t *testing.T) {
	tests := []struct {
		name      string
		input     Member
		candidate Member
		wantRisk  DuplicateRisk
	}{
		{
			name:      "same name and phone is high",
			input:     Member{Name: "Maria da Silva", Phone: "88999999999"},
			candidate: Member{Name: "MARIA DÁ SILVA", Phone: "+55 88 99999-9999"},
			wantRisk:  RiskHigh,
		},
		{
			name:      "same name address and number is high",
			input:     Member{Name: "Maria", Address: "Rua José", AddressNumber: "120"},
			candidate: Member{Name: "maria", Address: "RUA JOSE", AddressNumber: " 120 "},
			wantRisk:  RiskHigh,
		},
		{
			name:      "phone only is medium",
			input:     Member{Name: "Maria", Phone: "88999999999"},
			candidate: Member{Name: "Ana", Phone: "+55 88 99999-9999"},
			wantRisk:  RiskMedium,
		},
		{
			name:      "name only is medium",
			input:     Member{Name: "Maria"},
			candidate: Member{Name: "maria"},
			wantRisk:  RiskMedium,
		},
		{
			name: "phone and address with different names is not high",
			input: Member{
				Name: "Maria", Phone: "88999999999", Address: "Rua A", AddressNumber: "1",
			},
			candidate: Member{
				Name: "Ana", Phone: "88999999999", Address: "Rua A", AddressNumber: "1",
			},
			wantRisk: RiskMedium,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := (ExactDuplicateMatcher{}).Compare(test.input, test.candidate)
			if got.Risk != test.wantRisk {
				t.Fatalf("risk = %q, want %q (score %d, fields %v)", got.Risk, test.wantRisk, got.Score, got.MatchedFields)
			}
		})
	}
}

func TestCandidatesAreSortedByRiskThenDescendingScore(t *testing.T) {
	input := Member{
		Name: "Maria", Phone: "88999999999", Address: "Rua A", AddressNumber: "10",
		Neighborhood: "Centro", City: "Quixadá", Congregation: "Sede",
	}
	finder := &candidateFinderStub{candidates: []Member{
		{ID: 1, Name: "Ana", Phone: "88999999999"},
		{ID: 2, Name: "Maria", Phone: "88999999999"},
		{ID: 3, Name: "Maria", Phone: "88999999999", Address: "Rua A", AddressNumber: "10", Neighborhood: "Centro", City: "Quixadá", Congregation: "Sede"},
	}}

	result, err := NewDuplicateChecker(finder).Check(input, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := []int64{result.Candidates[0].MemberID, result.Candidates[1].MemberID, result.Candidates[2].MemberID}
	want := []int64{3, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("candidate order = %v, want %v", got, want)
	}
}

func TestDuplicateCheckIgnoresCurrentMemberOnEdit(t *testing.T) {
	finder := &candidateFinderStub{candidates: []Member{{ID: 7, Name: "Maria", Phone: "88999999999"}}}
	result, err := NewDuplicateChecker(finder).Check(Member{Name: "Maria", Phone: "88999999999"}, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.HasPossibleDuplicates || len(result.Candidates) != 0 {
		t.Fatalf("current member was not ignored: %+v", result)
	}
	if finder.excludeID != 7 {
		t.Fatalf("exclude id = %d, want 7", finder.excludeID)
	}
}
