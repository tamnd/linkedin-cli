package linkedin

import (
	"context"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	return postsFromGraph(doc, pageURL), nil
}

// FetchProfilePosts returns the recent posts carried as DiscussionForumPosting
// nodes in a member profile page's JSON-LD graph.
func (c *Client) FetchProfilePosts(ctx context.Context, cache *Cache, cfg Config, in string) ([]Post, error) {
	slug := NormalizeProfileSlug(in)
	pageURL := ProfileURL(slug)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	return postsFromGraph(doc, pageURL), nil
}

// FetchProfileArticles returns the long-form articles carried as Article nodes
// in a member profile page's JSON-LD graph.
func (c *Client) FetchProfileArticles(ctx context.Context, cache *Cache, cfg Config, in string) ([]Article, error) {
	slug := NormalizeProfileSlug(in)
	pageURL := ProfileURL(slug)
	doc, err := c.CachingFetchHTML(ctx, cache, cfg, pageURL)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var out []Article
	for _, raw := range nodesOfType(ldNodes(doc), "Article") {
		if art, ok := articleFromNode(raw, now); ok {
			out = append(out, art)
		}
	}
	return out, nil
}

// postsFromGraph turns the DiscussionForumPosting nodes in a page's JSON-LD
// graph into Post records, falling back to the page URL when a node omits its own.
func postsFromGraph(doc *goquery.Document, pageURL string) []Post {
	now := time.Now()
	var out []Post
	for _, raw := range nodesOfType(ldNodes(doc), "DiscussionForumPosting") {
		post, ok := postFromNode(raw, now)
		if !ok {
			continue
		}
		if post.URL == "" {
			post.URL = pageURL
		}
		out = append(out, post)
	}
	return out
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
