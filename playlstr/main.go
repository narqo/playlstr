package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/narqo/playlstr/feedly"
)

var (
	userID      = os.Getenv("FEEDLY_USER_ID")
	accessToken = os.Getenv("FEEDLY_ACCESS_TOKEN")
)

const PlaylstrBoard = "Yr Next Playlist"

func getAlbumsFromBoard(ctx context.Context, fc *feedly.Client, board string) {
	tags, err := fc.Tags(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	var tag feedly.Tag
	for _, t := range tags {
		if t.Label == board {
			tag = t
			break
		}
	}

	f := feedly.StreamFilter{
		ID: tag.ID,
	}
	stream, err := fc.StreamContent(ctx, f)
	if err != nil {
		log.Fatalln(err)
	}

	for _, data := range stream.Items {
		parseAlbumFromEntry(ctx, data)
	}
}

type AlbumMeta struct {
	Artist      string
	Title       string
	ReleaseDate time.Time
}

func parseAlbumFromEntry(ctx context.Context, data feedly.Entry) error {
	switch data.Origin.URL {
	case "http://funkysouls.com/":
		parseFunkySoulsEntry(ctx, data)
	default:
		fmt.Errorf("unknown entry's origin: %v (%q)", data.Origin.URL, data.Origin.Title)
	}
	return nil
}

func parseFunkySoulsEntry(ctx context.Context, data feedly.Entry) error {
	parts := strings.SplitN(data.Title, "-", 2)
	if len(parts) < 2 {
		return fmt.Errorf("could not parse album title %q: unsupported format", data.Title)
	}

	artist, title := parts[0], parts[1]
	if epos := strings.LastIndexByte(title, '['); epos != -1 {
		title = title[:epos]
	}

	meta := AlbumMeta{
		Artist: strings.TrimSpace(artist),
		Title:  strings.TrimSpace(title),
	}

	// TODO(varankinv): parse release date

	fmt.Printf("%+v\n", meta)

	return nil
}

func main() {
	if accessToken == "" {
		log.Fatalln("FEEDLY_ACCESS_TOKEN must be set")
	}

	ctx := context.Background()

	t := StaticTokenTransport{accessToken}
	fc := feedly.NewClient(t.Client())

	getAlbumsFromBoard(ctx, fc, PlaylstrBoard)
}

type StaticTokenTransport struct {
	Token string
}

func (t *StaticTokenTransport) Client() *http.Client {
	return &http.Client{
		Transport: t,
	}
}

func (t *StaticTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "OAuth "+t.Token)
	return http.DefaultTransport.RoundTrip(req)
}
