CREATE TABLE deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    FOREIGN KEY target_host REFERENCES nodes(id),
    FOREIGN KEY build_host REFERENCES nodes(id),

    command TEXT NOT NULL,

    is_upgrade BOOLEAN,
    executed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
