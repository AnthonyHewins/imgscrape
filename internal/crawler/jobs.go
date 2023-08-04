package crawler

import (
	"fmt"
	"net/url"
)

func (a *Crawler) AddURLString(urls ...string) error {
	for i, v := range urls {
		uri, err := url.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid URL at index %d: %s; %w", i, v, err)
		}

		a.urls = append(a.urls, uri)
	}

	return nil
}
