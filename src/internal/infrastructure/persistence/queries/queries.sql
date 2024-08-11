
-- Feed Channel Queries

-- name: ListFeedChannel :many
SELECT id, link, host, last_update
FROM (
    SELECT
        fc.id,
        fc.link,
        fc.host,
        MAX(fl.last_update) AS last_update
    FROM feed_channel AS fc
    LEFT JOIN feed_channel_log fl ON fc.id = fl.channel_id
    WHERE fc.enabled = 1
    GROUP BY fc.id, fc.link, fc.host
) AS LatestUpdates
WHERE datetime('now', '-' || @minutes || ' minutes') > COALESCE(last_update, '1970-01-01')
ORDER BY host, last_update DESC;

-- name: ListAllFeedChannel :many
SELECT id, title, description, link, host, published, enabled
FROM feed_channel
ORDER BY title;

-- name: GetFeedChannelsIDs :many
SELECT channel_id
FROM feed_channel_item
WHERE item_id = ?;

-- name: GetFeedChannel :one
SELECT fc.id, fc.title, fc.description, fc.link, fc.host, fc.published
FROM feed_channel AS fc
WHERE fc.id = @id
LIMIT 1;

-- name: CreateFeedChannel :one
INSERT INTO feed_channel (title, description, link, host)
VALUES (@title, @description, @link, @host)
RETURNING id, title, description, link, host, published;

-- name: UpdateFeedChannel :exec
UPDATE feed_channel
SET title = @title, description = @description, link = @link, host = @host
WHERE id = @id;

-- name: UpdateFeedChannelFTitle :exec
UPDATE feed_channel
SET title = @title
WHERE id = @id;

-- name: UpdateFeedChannelFDescription :exec
UPDATE feed_channel
SET description = @description
WHERE id = @id;

-- name: UpdateFeedChannelFTitleAndDescription :exec
UPDATE feed_channel
SET title = @title, description = @description
WHERE id = @id;

-- name: UpdateFeedChannelFLink :exec
UPDATE feed_channel
SET link = @link
WHERE id = @id;

-- name: DeleteFeedChannel :exec
DELETE FROM feed_channel
WHERE id = ?;

-- name: CreateFeedChannelLog :exec
INSERT INTO feed_channel_log (channel_id, last_update)
VALUES (?, datetime('now'));

-- name: GetLastChannelUpdateDate :one
SELECT last_update
FROM feed_channel_log
WHERE channel_id = @id
ORDER BY last_update DESC
LIMIT 1;

-- Feed Data Queries

-- name: ListFeedItem :many
SELECT fi.id, fi.guid, fi.guid_is_permalink, fi.title, fi.description, fi.link, fi.author, fi.published
FROM feed_item AS fi
LEFT JOIN feed_channel_item AS fci ON fi.id = fci.item_id
WHERE fci.channel_id = @id
ORDER BY published DESC;

-- name: GetFeedItemByGuid :one
SELECT fi.id, fi.guid, fi.guid_is_permalink, fi.title, fi.description, fi.link, fi.author, fi.published
FROM feed_item AS fi
WHERE fi.guid = @guid
LIMIT 1;

-- name: GetFeedItemByLink :one
SELECT fi.id, fi.guid, fi.guid_is_permalink, fi.title, fi.description, fi.link, fi.author, fi.published
FROM feed_item AS fi
WHERE fi.link = @link AND (fi.guid IS NULL OR fi.guid = '')
LIMIT 1;

-- name: CreateFeedItem :one
INSERT INTO feed_item (guid, guid_is_permalink, title, description, link, author, published)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id, read, created, updated;

-- name: CreateFeedChannelItem :exec
INSERT INTO feed_channel_item (channel_id, item_id)
VALUES (?, ?)
ON CONFLICT DO NOTHING;

-- name: UpdateFeedItem :exec
UPDATE feed_item
SET guid = ?, guid_is_permalink = ?, title = ?, description = ?, link = ?, author = ?, published = ?
WHERE id = ?;

-- name: UpdateFeedItemShort :exec
UPDATE feed_item
SET title = ?, description = ?, link = ?
WHERE id = ?;

-- name: UpdateFeedItemFDeleted :exec
UPDATE feed_item
SET deleted = 1
WHERE id = ?;

-- name: RemoveFeedItemFromChannel :exec
DELETE FROM feed_channel_item
WHERE channel_id = @channel_id AND item_id = @item_id;

-- name: DeleteFeedItem :exec
DELETE FROM feed_item
WHERE id = ?;


-- Feed Group Queries

-- name: ListGroup :many
SELECT id, name
FROM feed_group
ORDER BY @order || ' ' || @direction;

-- name: CreateGroup :exec
INSERT INTO feed_group (name, parent_id)
VALUES (@name, @parent_id);

-- name: UpdateGroup :exec
UPDATE feed_group
SET name = @name
WHERE id = @id;

-- name: DeleteGroup :exec
DELETE FROM feed_group
WHERE id = @id;

-- name: AddChannelToGroup :exec
INSERT INTO feed_group_channel (group_id, channel_id)
VALUES (@group_id, @channel_id);

-- name: RemoveChannelFromGroup :exec
DELETE FROM feed_group_channel
WHERE group_id = @group_id AND channel_id = @channel_id;
