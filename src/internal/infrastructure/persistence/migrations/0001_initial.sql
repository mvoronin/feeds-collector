-- +goose up
CREATE TABLE feed_channel (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    link TEXT NOT NULL,
    host TEXT NOT NULL,
    published DATETIME NOT NULL DEFAULT (datetime('now')),
    enabled INTEGER NOT NULL DEFAULT (1),
    created DATETIME NOT NULL DEFAULT (datetime('now')),
    updated DATETIME
);

CREATE TABLE feed_item (
    id INTEGER PRIMARY KEY,
    guid TEXT,
    guid_is_permalink BOOLEAN DEFAULT FALSE,
    title TEXT NOT NULL,
    description TEXT,
    link TEXT NOT NULL,
    author TEXT,
    published DATETIME,
    read INTEGER NOT NULL DEFAULT (0),
    deleted INTEGER NOT NULL DEFAULT (0),
    created DATETIME NOT NULL DEFAULT (datetime('now')),
    updated DATETIME
);

CREATE TABLE feed_channel_item (
    channel_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id),
    FOREIGN KEY (item_id) REFERENCES feed_item(id),
    PRIMARY KEY (channel_id, item_id)
);

CREATE TABLE feed_group (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER,
    FOREIGN KEY (parent_id) REFERENCES feed_group(id) ON DELETE CASCADE
);

CREATE TABLE feed_group_channel (
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES feed_group(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, channel_id)
);

CREATE TABLE feed_channel_log (
    id INTEGER PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    last_update DATETIME DEFAULT (datetime('now')), -- Uses SQLite function for current timestamp
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id)
);

CREATE TABLE tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE feed_channel_tag (
    channel_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES feed_channel(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, tag_id)
);

-- +goose down
DROP TABLE IF EXISTS feed_channel;
DROP TABLE IF EXISTS feed_item;
DROP TABLE IF EXISTS feed_channel_item;
DROP TABLE IF EXISTS feed_group;
DROP TABLE IF EXISTS feed_group_channel;
DROP TABLE IF EXISTS feed_channel_log;
DROP TABLE IF EXISTS tag;
DROP TABLE IF EXISTS feed_channel_tag;
