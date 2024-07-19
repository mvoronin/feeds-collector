package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"feedscollector/internal/models"
	"feedscollector/internal/utils"
	"fmt"
	"github.com/guregu/null"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type initialDataStruct struct {
	channels []models.FeedChannel
}

var testDB *sql.DB
var initialData initialDataStruct

func TestMain(m *testing.M) {
	// Setup database connection
	var err error
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	err = runMigrations(testDB)
	if err != nil {
		panic(err)
	}

	if err := loadInitialData(testDB); err != nil {
		log.Fatalf("could not load initial data: %v", err)
	}

	// Run tests
	os.Exit(m.Run())
}

func loadInitialData(db *sql.DB) error {
	ctx := context.Background()
	queries := models.New(db)

	channelRow1, err := queries.CreateFeedChannel(ctx, models.CreateFeedChannelParams{
		Title:       "Test Channel",
		Description: "A test channel",
		Link:        "http://example.com/rss",
		Host:        "example.com",
	})
	if err != nil {
		return fmt.Errorf("could not create initial feed channel: %w", err)
	}

	// Create initial feed channels
	channelRow2, err := queries.CreateFeedChannel(ctx, models.CreateFeedChannelParams{
		Title:       "Test Channel 2",
		Description: "A test channel 2",
		Link:        "http://example2.com/rss",
		Host:        "example2.com",
	})
	if err != nil {
		return fmt.Errorf("could not create initial feed channel: %w", err)
	}
	initialData.channels = []models.FeedChannel{
		{
			ID:          channelRow1.ID,
			Title:       channelRow1.Title,
			Description: channelRow1.Description,
			Link:        channelRow1.Link,
			Host:        channelRow1.Host,
			Enabled:     true,
		},
		{
			ID:          channelRow2.ID,
			Title:       channelRow2.Title,
			Description: channelRow2.Description,
			Link:        channelRow2.Link,
			Host:        channelRow2.Host,
			Enabled:     true,
		},
	}

	authors := "Author 1, Author 2"
	published := time.Now().UTC().Truncate(time.Second)
	// Create initial feed items
	_, err = queries.CreateFeedItem(ctx, models.CreateFeedItemParams{
		Guid:            null.StringFrom("guid 1"),
		GuidIsPermalink: sql.NullBool{Bool: true, Valid: true},
		Title:           "Item 1",
		Description:     null.StringFrom("Description 1"),
		Link:            "http://example.com/item1",
		Author:          &authors,
		Published:       null.TimeFromPtr(&published),
	})
	if err != nil {
		return fmt.Errorf("could not create initial feed item: %w", err)
	}

	return nil
}

func runMigrations(db *sql.DB) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("could not create sqlite driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"sqlite3", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}

func TestListChannels(t *testing.T) {
	apiInstance := NewAPI(testDB)
	router := mux.NewRouter()
	apiInstance.RegisterRoutes(router)

	req, err := http.NewRequest("GET", "/channels", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var channels []models.FeedChannel
	if err := json.NewDecoder(rr.Body).Decode(&channels); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	var expectedChannel1 *models.FeedChannel = &(initialData.channels[0])
	// if channels[0] != expectedChannel {
	//	 t.Errorf("Expected channel %v, got %v", expectedChannel, channels[0])
	// }
	isEqual, propertyName, err := utils.CompareFeedChannels(&channels[0], expectedChannel1, []string{"ID", "Title", "Description", "Link", "Host", "Enabled"})
	if err != nil {
		t.Errorf("Failed to compare channels: %v", err)
	}
	if !isEqual {
		t.Errorf("Channels are not equal. Property '%v' is different. Expected %v, got %v",
			propertyName, reflect.ValueOf(*expectedChannel1).FieldByName(propertyName), reflect.ValueOf(channels[0]).FieldByName(propertyName))
	}

	var expectedChannel2 *models.FeedChannel = &(initialData.channels[1])
	// if channels[1] != expectedChannel {
	//	t.Errorf("Expected channel %v, got %v", expectedChannel, channels[1])
	//}
	isEqual, propertyName, err = utils.CompareFeedChannels(&channels[1], expectedChannel2, []string{"ID", "Title", "Description", "Link", "Host", "Enabled"})
	if err != nil {
		t.Errorf("Failed to compare channels: %v", err)
	}
	if !isEqual {
		t.Errorf("Channels are not equal. Property '%v' is different. Expected %v, got %v",
			propertyName, reflect.ValueOf(*expectedChannel2).FieldByName(propertyName), reflect.ValueOf(channels[1]).FieldByName(propertyName))
	}
}

func TestAddChannel(t *testing.T) {
	apiInstance := NewAPI(testDB)
	router := mux.NewRouter()
	apiInstance.RegisterRoutes(router)

	newChannel := models.CreateFeedChannelParams{
		Title:       "Test Channel",
		Description: "A test channel",
		Link:        "http://example.com/rss",
		Host:        "example.com",
	}
	body, err := json.Marshal(newChannel)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", "/channels", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestUpdateChannel(t *testing.T) {
	apiInstance := NewAPI(testDB)
	router := mux.NewRouter()
	apiInstance.RegisterRoutes(router)

	// First, add a channel
	newChannel := models.CreateFeedChannelParams{
		Title:       "Test Channel",
		Description: "A test channel",
		Link:        "http://example.com/rss",
		Host:        "example.com",
	}
	_, err := models.New(testDB).CreateFeedChannel(context.Background(), newChannel)
	if err != nil {
		t.Fatalf("Failed to add initial channel: %v", err)
	}

	// Now update it
	updateData := models.UpdateFeedChannelParams{
		ID:          1,
		Title:       "Updated Channel",
		Description: "An updated channel",
		Link:        "http://example.com/rss",
		Host:        "example.com",
	}
	body, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("PUT", "/channels/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestDeleteChannel(t *testing.T) {
	apiInstance := NewAPI(testDB)
	router := mux.NewRouter()
	apiInstance.RegisterRoutes(router)

	// First, add a channel
	newChannel := models.CreateFeedChannelParams{
		Title:       "Test Channel",
		Description: "A test channel",
		Link:        "http://example.com/rss",
		Host:        "example.com",
	}
	_, err := models.New(testDB).CreateFeedChannel(context.Background(), newChannel)
	if err != nil {
		t.Fatalf("Failed to add initial channel: %v", err)
	}

	req, err := http.NewRequest("DELETE", "/channels/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestGetChannelItemList(t *testing.T) {
	apiInstance := NewAPI(testDB)
	router := mux.NewRouter()
	apiInstance.RegisterRoutes(router)

	req, err := http.NewRequest("GET", fmt.Sprintf("/channels/%d/items", initialData.channels[0].ID), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
