CREATE TABLE tasks (
    id CHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    team_id CHAR(36) NOT NULL,
    assignee_id CHAR(36) NULL,
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_tasks_team
        FOREIGN KEY (team_id) REFERENCES teams(id),

    CONSTRAINT fk_tasks_assignee
        FOREIGN KEY (assignee_id) REFERENCES users(id),

    CONSTRAINT fk_tasks_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_tasks_team_id ON tasks(team_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_team_status_assignee ON tasks(team_id, status, assignee_id);

CREATE TABLE task_history (
    id CHAR(36) PRIMARY KEY,
    task_id CHAR(36) NOT NULL,
    field VARCHAR(64) NOT NULL,
    old_value TEXT NULL,
    new_value TEXT NULL,
    changed_by CHAR(36) NOT NULL,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_task_history_task
        FOREIGN KEY (task_id) REFERENCES tasks(id),

    CONSTRAINT fk_task_history_changed_by
        FOREIGN KEY (changed_by) REFERENCES users(id)
);

CREATE INDEX idx_task_history_task_id ON task_history(task_id);
CREATE INDEX idx_task_history_changed_at ON task_history(changed_at);

CREATE TABLE task_comments (
    id CHAR(36) PRIMARY KEY,
    task_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_task_comments_task
        FOREIGN KEY (task_id) REFERENCES tasks(id),

    CONSTRAINT fk_task_comments_user
        FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);