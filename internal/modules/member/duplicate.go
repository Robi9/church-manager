package member

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const duplicateCandidateLimit = 10

type DuplicateRisk string

const (
	RiskNone   DuplicateRisk = "none"
	RiskLow    DuplicateRisk = "low"
	RiskMedium DuplicateRisk = "medium"
	RiskHigh   DuplicateRisk = "high"
)

type NormalizedMember struct {
	Name          string
	Phone         string
	Address       string
	AddressNumber string
	Neighborhood  string
	City          string
	Congregation  string
}

type DuplicateCandidate struct {
	MemberID      int64         `json:"member_id"`
	Name          string        `json:"name"`
	Phone         string        `json:"phone"`
	Congregation  string        `json:"congregation"`
	Score         int           `json:"score"`
	Risk          DuplicateRisk `json:"risk"`
	MatchedFields []string      `json:"matched_fields"`
}

type DuplicateCheckResult struct {
	HasPossibleDuplicates bool                 `json:"has_possible_duplicates"`
	HighestRisk           DuplicateRisk        `json:"highest_risk"`
	Candidates            []DuplicateCandidate `json:"candidates"`
}

type DuplicateCheckRequest struct {
	Member
	ExcludeMemberID int64 `json:"exclude_member_id,omitempty"`
}

type MemberMutationRequest struct {
	Member
	ForceCreate bool `json:"force_create"`
}

type DuplicateAudit struct {
	MemberID          int64
	CandidateMemberID int64
	ConfirmedBy       int64
	Score             int
	MatchedFields     []string
	Operation         string
}

type CandidateFinder interface {
	FindDuplicateCandidates(normalized NormalizedMember, excludeID int64) ([]Member, error)
}

// DuplicateMatcher is intentionally small so an approximate name matcher can
// be introduced later without changing handlers, import code, or persistence.
type DuplicateMatcher interface {
	Compare(input, candidate Member) DuplicateCandidate
}

type ExactDuplicateMatcher struct{}

type DuplicateChecker struct {
	finder  CandidateFinder
	matcher DuplicateMatcher
	limit   int
}

func NewDuplicateChecker(finder CandidateFinder) *DuplicateChecker {
	return &DuplicateChecker{
		finder:  finder,
		matcher: ExactDuplicateMatcher{},
		limit:   duplicateCandidateLimit,
	}
}

func (c *DuplicateChecker) Check(input Member, excludeID int64) (DuplicateCheckResult, error) {
	candidates, err := c.finder.FindDuplicateCandidates(NormalizeMember(input), excludeID)
	if err != nil {
		return DuplicateCheckResult{}, err
	}

	result := DuplicateCheckResult{
		HighestRisk: RiskNone,
		Candidates:  make([]DuplicateCandidate, 0),
	}

	for _, candidate := range candidates {
		if excludeID != 0 && candidate.ID == excludeID {
			continue
		}
		match := c.matcher.Compare(input, candidate)
		if match.Risk != RiskHigh && match.Risk != RiskMedium {
			continue
		}
		result.Candidates = append(result.Candidates, match)
	}

	sort.SliceStable(result.Candidates, func(i, j int) bool {
		left := result.Candidates[i]
		right := result.Candidates[j]
		if riskRank(left.Risk) != riskRank(right.Risk) {
			return riskRank(left.Risk) > riskRank(right.Risk)
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		return left.MemberID < right.MemberID
	})

	if len(result.Candidates) > c.limit {
		result.Candidates = result.Candidates[:c.limit]
	}
	if len(result.Candidates) > 0 {
		result.HasPossibleDuplicates = true
		result.HighestRisk = result.Candidates[0].Risk
	}

	return result, nil
}

func (ExactDuplicateMatcher) Compare(input, candidate Member) DuplicateCandidate {
	left := NormalizeMember(input)
	right := NormalizeMember(candidate)
	matched := make([]string, 0, 7)
	score := 0

	nameMatched := valuesMatch(left.Name, right.Name)
	phoneMatched := valuesMatch(left.Phone, right.Phone)
	addressMatched := valuesMatch(left.Address, right.Address)
	numberMatched := valuesMatch(left.AddressNumber, right.AddressNumber)

	add := func(field string, matches bool, points int) {
		if matches {
			matched = append(matched, field)
			score += points
		}
	}

	add("phone", phoneMatched, 45)
	add("name", nameMatched, 40)
	add("address", addressMatched, 10)
	add("address_number", numberMatched, 10)
	add("neighborhood", valuesMatch(left.Neighborhood, right.Neighborhood), 5)
	add("city", valuesMatch(left.City, right.City), 5)
	add("congregation", valuesMatch(left.Congregation, right.Congregation), 5)

	risk := RiskLow
	// Name + address + number is an explicit high-risk product rule even though
	// the configured weights for those three fields total 60 points.
	if nameMatched && ((phoneMatched && score >= 70) || (addressMatched && numberMatched)) {
		risk = RiskHigh
	} else if score >= 40 {
		risk = RiskMedium
	}

	return DuplicateCandidate{
		MemberID:      candidate.ID,
		Name:          candidate.Name,
		Phone:         MaskPhone(candidate.Phone),
		Congregation:  candidate.Congregation,
		Score:         score,
		Risk:          risk,
		MatchedFields: matched,
	}
}

func NormalizeMember(member Member) NormalizedMember {
	return NormalizedMember{
		Name:          NormalizeName(member.Name),
		Phone:         NormalizePhone(member.Phone),
		Address:       NormalizeText(member.Address),
		AddressNumber: NormalizeAddressNumber(member.AddressNumber),
		Neighborhood:  NormalizeText(member.Neighborhood),
		City:          NormalizeText(member.City),
		Congregation:  NormalizeText(member.Congregation),
	}
}

func NormalizeName(value string) string {
	return NormalizeText(value)
}

func NormalizeText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Join(strings.Fields(value), " ")

	decomposed := norm.NFD.String(value)
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, decomposed)
}

func NormalizePhone(value string) string {
	var digits strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}

	normalized := digits.String()
	if strings.HasPrefix(normalized, "55") && (len(normalized) == 12 || len(normalized) == 13) {
		normalized = normalized[2:]
	}
	return normalized
}

func NormalizeAddressNumber(value string) string {
	return NormalizeText(value)
}

func MaskPhone(value string) string {
	normalized := NormalizePhone(value)
	if normalized == "" {
		return ""
	}
	if len(normalized) <= 4 {
		return strings.Repeat("*", len(normalized))
	}
	return strings.Repeat("*", len(normalized)-4) + normalized[len(normalized)-4:]
}

func valuesMatch(left, right string) bool {
	return left != "" && right != "" && left == right
}

func riskRank(risk DuplicateRisk) int {
	switch risk {
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	case RiskLow:
		return 1
	default:
		return 0
	}
}
