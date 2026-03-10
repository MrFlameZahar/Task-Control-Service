package redis

import (
	taskDomain "TaskControlService/internal/domain/task"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/go-redis/redis"
)

type TaskCache struct {
	client *goredis.Client
	ttl    time.Duration
}

func NewTaskCache(client *goredis.Client) *TaskCache {
	return &TaskCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (t *TaskCache) GetTasks(key string) ([]taskDomain.Task, bool, error) {
	value, err := t.client.Get(key).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var tasks []taskDomain.Task
	if err := json.Unmarshal([]byte(value), &tasks); err != nil {
		return nil, false, err
	}

	return tasks, true, nil
}

func (t *TaskCache) SetTasks(key string, tasks []taskDomain.Task) error {
	payload, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return t.client.Set(key, payload, t.ttl).Err()
}

func (t *TaskCache) InvalidateTeamTasks(teamID uuid.UUID) error {
	pattern := fmt.Sprintf("tasks:team:%s:*", teamID.String())

	var cursor uint64
	for {
		keys, nextCursor, err := t.client.Scan(cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := t.client.Del(keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}