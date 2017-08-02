package feedly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const DefaultURL = "https://cloud.feedly.com"

const userAgent = "go-feedly-client/X"

type Client struct {
	client  *http.Client
	baseURL *url.URL
}

func NewClient(c *http.Client) *Client {
	baseURL, _ := url.ParseRequestURI(DefaultURL)
	if c == nil {
		c = http.DefaultClient
	}
	return &Client{
		client:  c,
		baseURL: baseURL,
	}
}

type Feed struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	Website     string   `json:"website"`
	Velocity    float32  `json:"velocity"`
	Label       string   `json:"label,omitempty"`
	Topics      []string `json:"topics"`
	Curated     bool     `json:"curated"`
	Language    string   `json:"language"`
}

func (c *Client) Feed(ctx context.Context, id string) (resp Feed, err error) {
	if id == "" {
		return resp, errors.New("no id to fetch")
	}

	u := c.baseURL
	u.Path = "/v3/feeds/" + id

	err = c.Fetch(ctx, u.String(), &resp)
	return resp, err
}

type Tag struct {
	ID          string `json:"id"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) Tags(ctx context.Context) (resp []Tag, err error) {
	u := c.baseURL
	u.Path = "/v3/tags"

	err = c.Fetch(ctx, u.String(), &resp)
	return resp, err
}

type EntryOrigin struct {
	StreamID string `json:"streamId"`
	Title    string `json:"title"`
	URL      string `json:"htmlURL"`
}

type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Meta struct {
	Type string `json:"type"`
	Href string `json:"href"`
}

type Image struct {
	URL string `json:"url"`
}

type Text struct {
	Direction string `json:"direction"`
	Content   string `json:"content"`
}

type Entry struct {
	ID             string      `json:"id"`
	SID            string      `json:"sid"`
	Title          string      `json:"title"`
	Summary        Text        `json:"summary"`
	Engagement     int         `json:"engagement"`
	EngagementRate float32     `json:"engagementRate"`
	Tags           []Tag       `json:"tag"`
	Author         string      `json:"author"`
	Unread         bool        `json:"unread"`
	OriginID       string      `json:"originId"`
	Origin         EntryOrigin `json:"origin"`
	Published      int64       `json:"published"`
	Updated        int64       `json:"updated"`
	Crawled        int64       `json:"crawled"`
	Recrawled      int64       `json:"recrawled"`
	Categories     []Category  `json:"categories"`
	Canonical      Meta        `json:"canonical"`
	Thumbnail      []Image     `json:"thumbnail"`
	Fingerprint    string      `json:"fingerprint"`
	Keywords       []string    `json:"keywords"`
}

func (c *Client) Entries(ctx context.Context, ids []string) (resp []Entry, err error) {
	u := c.baseURL
	u.Path = "/v3/entries/.mget"

	body := &bytes.Buffer{}
	if err = json.NewEncoder(body).Encode(ids); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	err = c.doRequest(ctx, req, &resp)
	return resp, err
}

func (c *Client) Entry(ctx context.Context, id string) (resp Entry, err error) {
	if id == "" {
		return resp, errors.New("no id to fetch")
	}

	u := c.baseURL
	u.Path = "/v3/entries/" + id

	err = c.Fetch(ctx, u.String(), &resp)
	return resp, err
}

const (
	NewestFirst = "newest"
	OldestFirst = "oldest"
)

type StreamFilter struct {
	ID           string
	Count        uint
	Ranked       string
	UnreadOnly   bool
	NewerThan    time.Time
	Continuation string
}

func encodeStreamFilter(f StreamFilter) string {
	vals := url.Values{}
	vals.Set("streamId", f.ID)

	if f.Ranked != "" {
		vals.Set("ranked", f.Ranked)
	}
	if f.UnreadOnly {
		vals.Set("unreadOnly", "1")
	}
	if !f.NewerThan.IsZero() {
		t := strconv.FormatInt(f.NewerThan.Unix(), 10)
		vals.Set("newerThan", t)
	}
	if f.Continuation != "" {
		vals.Set("continuation", f.Continuation)
	}
	return vals.Encode()
}

func (c *Client) StreamListEntries(ctx context.Context, f StreamFilter) (ids []string, cont string, err error) {
	if f.ID == "" {
		return ids, cont, errors.New("no stream id")
	}

	u := c.baseURL
	u.Path = "/v3/streams/ids"
	u.RawQuery = encodeStreamFilter(f)

	resp := struct {
		IDs          []string `json:"ids"`
		Continuation string   `json:"continuation"`
	}{}

	err = c.Fetch(ctx, u.String(), &resp)

	return resp.IDs, resp.Continuation, err
}

type StreamContent struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Updated      int64   `json:"updated"`
	Items        []Entry `json:"items"`
	Alternate    []Meta  `json:"alternate"`
	Direction    string  `json:"direction"`
	Continuation string  `json:"continuation"`
}

func (c *Client) StreamContent(ctx context.Context, f StreamFilter) (resp StreamContent, err error) {
	if f.ID == "" {
		return resp, errors.New("no stream id")
	}

	u := c.baseURL
	u.Path = "/v3/streams/contents"
	u.RawQuery = encodeStreamFilter(f)

	err = c.Fetch(ctx, u.String(), &resp)
	return resp, err
}

func (c *Client) Fetch(ctx context.Context, urlStr string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	return c.doRequest(ctx, req, v)
}

func (c *Client) doRequest(ctx context.Context, req *http.Request, v interface{}) error {
	req.Header.Set("User-Agent", userAgent)
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching %s: %v", req.URL.String(), resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
