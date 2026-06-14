package linkedin

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// FetchProfile fetches and parses a public member profile. The input may be a
// slug, an /in/<slug> path, or a full URL.
func (c *Client) FetchProfile(ctx context.Context, cache *Cache, cfg Config, in string) (*Profile, error) {
	slug := NormalizeProfileSlug(in)
	pageURL := ProfileURL(slug)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	return ParseProfile(doc, slug, pageURL)
}

// FetchCompany fetches and parses a public company page.
func (c *Client) FetchCompany(ctx context.Context, cache *Cache, cfg Config, in string) (*Company, error) {
	slug := NormalizeCompanySlug(in)
	pageURL := CompanyURL(slug)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	return ParseCompany(doc, slug, pageURL)
}

// FetchCompanyPosts returns the recent company posts carried as
// DiscussionForumPosting nodes in the company page's JSON-LD graph.
func (c *Client) FetchCompanyPosts(ctx context.Context, cache *Cache, cfg Config, in string) ([]Post, error) {
	slug := NormalizeCompanySlug(in)
	pageURL := CompanyURL(slug)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var out []Post
	for _, raw := range nodesOfType(ldNodes(doc), "DiscussionForumPosting", "Article") {
		var p postLD
		if json.Unmarshal(raw, &p) != nil {
			continue
		}
		a := parseAuthor(p.Author)
		post := Post{
			URL:       canonicalURL(p.URL),
			Author:    strings.TrimSpace(a.Name),
			AuthorURL: a.URL,
			Title:     cleanText(p.Headline),
			Text:      cleanText(p.Text),
			Published: p.DatePublished,
			ImageURL:  ldString(p.Image),
			Likes:     interactionCount(p.InteractionStatistic, "LikeAction"),
			Comments:  interactionCount(p.InteractionStatistic, "CommentAction"),
			FetchedAt: now,
		}
		if post.URL == "" {
			post.URL = pageURL
		}
		out = append(out, post)
	}
	return out, nil
}

// FetchPost fetches a public post or article, best effort.
func (c *Client) FetchPost(ctx context.Context, cache *Cache, cfg Config, in string) (*Post, error) {
	pageURL := canonicalURL(in)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	return ParsePost(doc, pageURL)
}
