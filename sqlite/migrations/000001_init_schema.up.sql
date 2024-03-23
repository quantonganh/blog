CREATE TABLE IF NOT EXISTS activities (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    country    TEXT NOT NULL,
    browser    TEXT NOT NULL,
    os         TEXT NOT NULL,
    referer    TEXT NOT NULL,
    url        TEXT NOT NULL,
    time       TIMESTAMP
);
