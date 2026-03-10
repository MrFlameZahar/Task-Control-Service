package analytics

type TeamStats struct {
	TeamID            string
	TeamName          string
	MembersCount      int 
	DoneTasksLast7Days int 
}

type TopCreator struct {
	TeamID      string
	UserID      string
	TasksCount  int
	Rank        int
}

type IntegrityIssue struct {
	TaskID      string
	Title       string
	TeamID      string
	AssigneeID  *string
}