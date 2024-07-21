package gatherer

import (
	"context"
	"database/sql"
	"errors"
	"feedscollector/internal"
	"feedscollector/internal/infrastructure/config"
	"feedscollector/internal/models"
	"github.com/guregu/null"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

func RunGathererLoop(ctx context.Context, db *sql.DB, config *config.Config) {
	FetchListFeedChannels(ctx, db)

	ticker := time.NewTicker(config.Gatherer.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			FetchListFeedChannels(ctx, db)
		case <-ctx.Done():
			return
		default:
			time.Sleep(10 * time.Second)
		}
	}
}

func FetchListFeedChannels(ctx context.Context, db *sql.DB) {
	queries := models.New(db)

	// NOTE: This table is not going to grow very large, so I'm not using pagination
	feedRows, err := queries.ListFeedChannel(ctx, sql.NullString{String: "1", Valid: true})
	if err != nil {
		internal.ErrorLogger.Fatalf("error listing feeds: %v", err)
	}

	feedDataChannel := make(chan models.ListFeedChannelRow)

	// create 10 goroutines to fetch feeds (function fetchFeed)
	var wgFetcher sync.WaitGroup
	wgFetcher.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wgFetcher.Done()
			for feedInfo := range feedDataChannel {
				err := UpdateFeed(ctx, &feedInfo, db)
				if err != nil {
					internal.ErrorLogger.Printf("Error processing feed: %v", err)
					return
				}
			}
		}()
	}

	for _, feedRow := range feedRows {
		feedDataChannel <- feedRow
	}

	close(feedDataChannel)
	wgFetcher.Wait()
}

func UpdateFeed(ctx context.Context, feedChannelInfo *models.ListFeedChannelRow, db *sql.DB) error {
	// Request the feed
	client := &http.Client{
		Timeout: 4 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedChannelInfo.Link, nil)
	if err != nil {
		internal.ErrorLogger.Printf("Error creating request: %v", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			internal.ErrorLogger.Printf("Request timed out: %v", err)
			// TODO: добавить это в БД лог конкретного канала (feed)
			return err
		}
		internal.ErrorLogger.Printf("Error requesting feed: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			internal.ErrorLogger.Fatalf("Error closing response body: %v", err)
		}
	}(resp.Body)

	// Parse the feed
	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		internal.ErrorLogger.Printf("Error parsing feed: %v", err)
		return err
	}

	// Iterate over feed items and send them to the channel
	for _, itemXML := range feed.Items {
		err := processFeedItem(feedChannelInfo.ID, itemXML, ctx, db)
		if err != nil {
			internal.ErrorLogger.Printf("Error processing feed item \"%v\": %v", itemXML.Title, err)
			return err
		}
		// internal.InfoLogger.Printf("Processed feed item: %v", itemXML.Title)
	}

	queries := models.New(db)
	// Update a feed channel log
	err = queries.CreateFeedChannelLog(ctx, feedChannelInfo.ID)
	if err != nil {
		internal.ErrorLogger.Printf("Error updating feed channel log: %v", err)
		return err
	}

	return nil
}

func processFeedItem(feedChannelID int64, itemXML *gofeed.Item, ctx context.Context, db *sql.DB) error {
	authors := getAuthorsString(itemXML)

	feedItemNew := models.CreateFeedItemParams{
		Guid:        null.StringFrom(itemXML.GUID),
		Title:       itemXML.Title,
		Description: null.StringFrom(itemXML.Description),
		Link:        itemXML.Link,
		Author:      authors,
		Published:   null.TimeFromPtr(itemXML.PublishedParsed),
	}

	queries := models.New(db)

	feedItem, createdFlag, err := getOrCreateFeedItem(ctx, queries, &feedItemNew)
	if err != nil {
		internal.ErrorLogger.Printf("Error getting or creating feed item: %v", err)
		return err
	}
	if !createdFlag { // The feed item is in the database, we need to check if it's changed in the source
		channelsIDsString, err := getChannelsIDs(ctx, queries, feedItem.ID)
		if err != nil {
			internal.ErrorLogger.Printf("Error getting channels IDs: %v", err)
			return err
		}
		isEqualFlag, err := compareFeedItems(feedItem, &feedItemNew, channelsIDsString)
		if err != nil {
			internal.ErrorLogger.Fatalf("Error comparing feed items: %v", err)
		}
		if !isEqualFlag {
			args := models.UpdateFeedItemShortParams{
				Title:       feedItemNew.Title,
				Description: feedItemNew.Description,
				Link:        feedItemNew.Link,
				ID:          feedItem.ID,
			}
			err = queries.UpdateFeedItemShort(ctx, args)
			if err != nil {
				internal.ErrorLogger.Printf("Error updating feed item: %v", err)
				return err
			}
		}
	}

	// Feed item exists, but it may be associated with another channel.
	// Trying to create a new relation (channel to item).
	err = addItemToChannel(ctx, queries, feedChannelID, feedItem.ID)
	if err != nil {
		return err
	}

	return nil
}

func getAuthorsString(itemXML *gofeed.Item) *string {
	// Get list of authors from the feed (itemXML.Authors)
	var authors string
	for i, author := range itemXML.Authors {
		authors += author.Name
		if author.Email != "" {
			authors = authors + " (" + author.Email + ")"
		}
		if i < len(itemXML.Authors)-1 {
			authors += ", "
		}
	}
	return &authors
}

func getOrCreateFeedItem(ctx context.Context, queries *models.Queries, feedItemNew *models.CreateFeedItemParams) (item *models.FeedItem, created bool, err error) {
	feedItem, err := getItemFromDB(ctx, queries, feedItemNew)
	if err != nil {
		internal.ErrorLogger.Printf("Error checking if feed item exists: %v", err)
		return nil, false, err
	}
	if feedItem != nil {
		return feedItem, false, nil // feed item exists
	}

	// Feed item doesn't exist. So, create it
	itemCreated, err := queries.CreateFeedItem(ctx, *feedItemNew) // TODO
	if err != nil {
		internal.ErrorLogger.Printf("Error creating feed item: %v", err)
		return nil, false, err
	}
	item = &models.FeedItem{
		ID:          itemCreated.ID,
		Title:       feedItemNew.Title,
		Description: feedItemNew.Description,
		Link:        feedItemNew.Link,
		Author:      feedItemNew.Author,
		Guid:        feedItemNew.Guid,
		Published:   feedItemNew.Published,
		Read:        itemCreated.Read,
		Deleted:     false,
		Created:     itemCreated.Created,
		Updated:     itemCreated.Updated,
	}
	return item, true, nil
}

func getItemFromDB(ctx context.Context, queries *models.Queries, feedItemNew *models.CreateFeedItemParams) (*models.FeedItem, error) {
	var item models.FeedItem

	// Checking if a feed with this guid already exists
	// TODO: почему я не могу проверить itemByGuid == nil?
	itemByGuid, err := queries.GetFeedItemByGuid(ctx, feedItemNew.Guid)
	if err != nil {
		// If there is no feed item with the passed guid,
		// then we need to check if a feed with passed link already exists
		if !errors.Is(err, sql.ErrNoRows) {
			internal.ErrorLogger.Printf("Error checking if feed exists: %v", err)
			return nil, err
		}
	} else {
		item = models.FeedItem{
			ID:          itemByGuid.ID,
			Guid:        itemByGuid.Guid,
			Title:       itemByGuid.Title,
			Description: itemByGuid.Description,
			Link:        itemByGuid.Link,
			Author:      itemByGuid.Author,
			Published:   itemByGuid.Published,
		}
		return &item, nil
	}

	// Checking if a feed item with the same link already exists
	itemByLink, err := queries.GetFeedItemByLink(ctx, feedItemNew.Link)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		item = models.FeedItem{
			ID:          itemByLink.ID,
			Guid:        itemByLink.Guid,
			Title:       itemByLink.Title,
			Description: itemByLink.Description,
			Link:        itemByLink.Link,
			Author:      itemByLink.Author,
			Published:   itemByLink.Published,
		}
		return &item, nil
	}

	return nil, nil
}

func compareFeedItems(feedItem *models.FeedItem, feedItemNew *models.CreateFeedItemParams, channelsIDsString *string) (isEqual bool, err error) {
	if feedItemNew.Title != feedItem.Title {
		internal.InfoLogger.Printf("Feed item #%v (channels: %s) title has changed."+
			"\r\nDATABASE: '%s'\r\nNEW: '%s'",
			feedItem.ID, *channelsIDsString, feedItem.Title, feedItemNew.Title)
		return false, nil
	}
	var descriptionNew, description string
	descriptionNew = feedItemNew.Description.String
	description = feedItem.Description.String
	if description != descriptionNew {
		internal.InfoLogger.Printf("Feed item #%v (channels: %v) description has changed."+
			"\r\nDATABASE: '%s'\r\nNEW: '%s'",
			feedItem.ID, *channelsIDsString, description, descriptionNew)
		return false, nil
	}
	if feedItemNew.Link != feedItem.Link {
		internal.InfoLogger.Printf("Feed item #%v (channels: %v) link has changed."+
			"\r\nDATABASE: '%s'\r\nNEW: '%s'",
			feedItem.ID, *channelsIDsString, feedItem.Link, feedItemNew.Link)
		return false, nil
	}

	return true, nil
}

func getChannelsIDs(ctx context.Context, queries *models.Queries, itemID int64) (*string, error) {
	channelsIDs, err := queries.GetFeedChannelsIDs(ctx, itemID)
	if err != nil {
		internal.ErrorLogger.Printf("Error getting feed channels IDs: %v", err)
		return nil, err
	}
	channelsIDsStrings := make([]string, len(channelsIDs))
	for i, id := range channelsIDs {
		channelsIDsStrings[i] = strconv.FormatInt(id, 10)
	}
	s := strings.Join(channelsIDsStrings, ",")
	return &s, nil
}

func addItemToChannel(ctx context.Context, queries *models.Queries, feedChannelID int64, feedItemID int64) error {
	args := models.CreateFeedChannelItemParams{
		ChannelID: feedChannelID,
		ItemID:    feedItemID,
	}
	err := queries.CreateFeedChannelItem(ctx, args)
	if err != nil {
		internal.ErrorLogger.Printf("Error creating feed channel item: %v", err)
		return err
	}
	return nil
}
