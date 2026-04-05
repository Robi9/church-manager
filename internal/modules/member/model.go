package member

import "time"

type Member struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status Status `json:"status"`

	MemberSince *time.Time `json:"member_since,omitempty"`
	Baptized    bool       `json:"baptized"`
	BaptismDate *time.Time `json:"baptism_date,omitempty"`

	ChurchRole         string        `json:"church_role"`
	MaritalStatus      MaritalStatus `json:"marital_status"`
	OriginDenomination string        `json:"origin_denomination"`

	MembershipCourseCompleted   bool       `json:"membership_course_completed"`
	MembershipCourseCompletedAt *time.Time `json:"membership_course_completed_at,omitempty"`

	Contacted   bool       `json:"contacted"`
	ContactedAt *time.Time `json:"contacted_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy int64     `json:"created_by"`
}

type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

type MaritalStatus string

const (
	Single   MaritalStatus = "single"
	Married  MaritalStatus = "married"
	Divorced MaritalStatus = "divorced"
	Widowed  MaritalStatus = "widowed"
)
