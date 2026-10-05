CREATE TABLE deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_node INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    build_node INTEGER REFERENCES nodes(id) ON DELETE SET NULL,

    action TEXT NOT NULL,
    config_file TEXT NOT NULL DEFAULT '',
    command TEXT NOT NULL,

    is_upgrade BOOLEAN NOT NULL DEFAULT FALSE,
    is_rollback BOOLEAN NOT NULL DEFAULT FALSE,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,

    exit_code INTEGER NOT NULL DEFAULT -1,
    error TEXT NOT NULL DEFAULT '',

    executed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deployments_target_node ON deployments(target_node);
CREATE INDEX idx_deployments_executed_at ON deployments(executed_at);
