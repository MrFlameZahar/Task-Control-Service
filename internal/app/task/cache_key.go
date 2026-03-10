package task

import (
	taskDomain "TaskControlService/internal/domain/task"
	"fmt"
)

func buildTasksCacheKey(filter taskDomain.ListFilter) string {
	status := "all"
	if filter.Status != nil {
		status = string(*filter.Status)
	}

	assignee := "all"
	if filter.AssigneeID != nil {
		assignee = filter.AssigneeID.String()
	}

	return fmt.Sprintf(
		"tasks:team:%s:status:%s:assignee:%s:limit:%d:offset:%d",
		filter.TeamID.String(),
		status,
		assignee,
		filter.Limit,
		filter.Offset,
	)
}