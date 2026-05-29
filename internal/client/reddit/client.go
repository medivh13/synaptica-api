package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client interface {
	FetchLatestPosts(ctx context.Context, subreddit string, limit int) ([]Post, error)
}

type client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

type listingResponse struct {
	Data listingData `json:"data"`
}

type listingData struct {
	Children []listingChild `json:"children"`
}

type listingChild struct {
	Data postData `json:"data"`
}

type postData struct {
	ID          string  `json:"id"`
	Subreddit   string  `json:"subreddit"`
	Title       string  `json:"title"`
	SelfText    string  `json:"selftext"`
	Author      string  `json:"author"`
	Score       int     `json:"score"`
	UpvoteRatio float64 `json:"upvote_ratio"`
	NumComments int     `json:"num_comments"`
	Permalink   string  `json:"permalink"`
	URL         string  `json:"url"`
	CreatedUTC  float64 `json:"created_utc"`
}

func NewClient(baseURL string, userAgent string, timeout time.Duration) Client {
	return &client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  userAgent,
	}
}

func (c *client) FetchLatestPosts(ctx context.Context, subreddit string, limit int) ([]Post, error) {
	endpoint, err := url.JoinPath(c.baseURL, "r", subreddit, "new.json")
	if err != nil {
		return nil, fmt.Errorf("build reddit request URL: %w", err)
	}

	requestURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse reddit request URL: %w", err)
	}

	query := requestURL.Query()
	query.Set("limit", fmt.Sprintf("%d", limit))
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create reddit request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch reddit posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("reddit returned non-2xx status: %d", resp.StatusCode)
	}

	var listing listingResponse
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("decode reddit response: %w", err)
	}

	if listing.Data.Children == nil {
		return nil, fmt.Errorf("reddit response missing posts")
	}

	posts := make([]Post, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		data := child.Data
		if data.ID == "" {
			continue
		}
		posts = append(posts, Post{
			ID:          data.ID,
			Subreddit:   data.Subreddit,
			Title:       data.Title,
			SelfText:    data.SelfText,
			Author:      data.Author,
			Score:       data.Score,
			UpvoteRatio: data.UpvoteRatio,
			NumComments: data.NumComments,
			Permalink:   "https://www.reddit.com" + data.Permalink,
			URL:         data.URL,
			CreatedUTC:  int64(data.CreatedUTC),
		})
	}

	return posts, nil
}
