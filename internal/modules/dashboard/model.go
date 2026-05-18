package dashboard

type Stats struct {
	TotalMembers    int `json:"total_members"`
	ActiveMembers   int `json:"active_members"`
	InactiveMembers int `json:"inactive_members"`
	NewThisMonth    int `json:"new_this_month"`
}
