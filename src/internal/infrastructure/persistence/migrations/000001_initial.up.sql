CREATE TABLE IF NOT EXISTS feed_channel (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    link TEXT NOT NULL,
    host TEXT NOT NULL,
    published DATETIME DEFAULT (datetime('now')),
    enabled INTEGER DEFAULT (1),
    created DATETIME DEFAULT (datetime('now')),
    updated DATETIME
);

CREATE TABLE IF NOT EXISTS feed_item (
    id INTEGER PRIMARY KEY,
    guid TEXT,
    guid_is_permalink BOOLEAN DEFAULT FALSE,
    title TEXT NOT NULL,
    description TEXT,
    link TEXT NOT NULL UNIQUE,
    author TEXT,
    published DATETIME,
    read INTEGER DEFAULT (0),
    deleted INTEGER DEFAULT (0),
    created DATETIME DEFAULT (datetime('now')),
    updated DATETIME
);

CREATE TABLE IF NOT EXISTS feed_channel_item (
    channel_id INTEGER,
    item_id INTEGER,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id),
    FOREIGN KEY (item_id) REFERENCES feed_item(id),
    PRIMARY KEY (channel_id, item_id)
);

CREATE TABLE IF NOT EXISTS feed_group (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER,
    FOREIGN KEY (parent_id) REFERENCES feed_group(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS feed_group_channel (
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES feed_group(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, channel_id)
);

CREATE TABLE IF NOT EXISTS feed_channel_log (
    id INTEGER PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    last_update DATETIME DEFAULT (datetime('now')), -- Uses SQLite function for current timestamp
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id)
);

CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS feed_channel_tag (
    channel_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, tag_id)
);
