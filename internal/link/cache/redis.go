package cache

import (
	"context"
	"github.com/ilam072/shortener/pkg/errutils"
	"github.com/wb-go/wbf/redis"
	"time"
)

type LinkCache struct {
	client *redis.Client
}

func New(client *redis.Client) *LinkCache {
	return &LinkCache{client: client}
}

func (c *LinkCache) SetURL(ctx context.Context, alias string, url string) error {
	if err := c.client.SetWithExpiration(ctx, alias, url, 24*time.Hour); err != nil {
		return errutils.Wrap("failed to cache status", err)
	}
	return nil
}

func (c *LinkCache) GetURL(ctx context.Context, alias string) (string, error) {
	url, err := c.client.Get(ctx, alias)
	if err != nil {
		return "", errutils.Wrap("failed to get status from redis", err)
	}
	return url, nil
}
