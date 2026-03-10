CREATE TABLE teams (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_teamscreated_by
        FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_teams_created_by ON teams(created_by);

CREATE TABLE team_members (
    user_id CHAR(36) NOT NULL,
    team_id CHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL,

    PRIMARY KEY (user_id, team_id),

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (team_id) REFERENCES teams(id)
);