package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/chhoumann/blogaggregator/internal/database"
	"github.com/mmcdole/gofeed"
)

type Scraper struct {
	DB     *database.Queries
	Logger *slog.Logger
}

func NewScraper(db *database.Queries, logger *slog.Logger) *Scraper {
	return &Scraper{
		DB:     db,
		Logger: logger,
	}
}

func (s *Scraper) fetchFeed(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch feed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *Scraper) processFeeds(ctx context.Context, feeds []database.Feed) {
	for _, feed := range feeds {
		go func(feed database.Feed) {
			content, err := s.fetchFeed(feed.Url)
			if err != nil {
				s.Logger.Error("error fetching feed", "error", err, "feed", feed.Url)
				return
			}

			fp := gofeed.NewParser()
			parsedFeed, err := fp.ParseString(content)
			if err != nil {
				s.Logger.Error("error parsing feed", "error", err, "feed", feed.Url)
				return
			}

			s.Logger.Info("fetched feed", "feed", feed.Url, "title", parsedFeed.Title)

			for _, item := range parsedFeed.Items {
				s.Logger.Info("post found",
					"feed", feed.Url,
					"title", item.Title,
					"published", item.Published,
					"author", item.Authors,
					"link", item.Link,
				)

				publishedAt := sql.NullTime{}
				if item.PublishedParsed != nil {
					publishedAt = sql.NullTime{Time: *item.PublishedParsed, Valid: true}
				} else if item.Published != "" {
					parsedTime, err := time.Parse(time.RFC1123, item.Published)
					if err == nil {
						publishedAt = sql.NullTime{Time: parsedTime, Valid: true}
					}
				}

				post := database.CreatePostParams{
					Title:       item.Title,
					Url:         item.Link,
					Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
					PublishedAt: publishedAt,
					FeedID:      feed.ID,
				}

				_, err := s.DB.CreatePost(ctx, post)
				if err != nil {
					if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
						s.Logger.Info("post already exists", "feed", feed.Url, "post", item.Link)
					} else {
						s.Logger.Error("error saving post to database", "error", err, "feed", feed.Url, "post", item.Link)
					}
					continue
				}
			}

			// Mark the feed as fetched
			err = s.DB.MarkFeedFetched(ctx, feed.ID)
			if err != nil {
				s.Logger.Error("error marking feed as fetched", "error", err)
			}
		}(feed)
	}
}

func (s *Scraper) Run(ctx context.Context, interval time.Duration, fetchLimit int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			feeds, err := s.DB.GetNextFeedsToFetch(ctx, int32(fetchLimit))
			if err != nil {
				s.Logger.Error("error fetching feeds from db", "error", err)
				continue
			}

			s.processFeeds(ctx, feeds)
		case <-ctx.Done():
			s.Logger.Info("feed scraper stopped")
			return
		}
	}
}
