package crawler

import (
	"context"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/sourcegraph/conc/pool"
)

func (a *Crawler) Run(ctx context.Context) ([][]string, error) {
	p := pool.NewWithResults[[]string]().WithContext(ctx)

	for i, v := range a.urls {
		uri := v.String()
		l := a.logger.With("worker index", i, "url", uri)

		p.Go(func(ctx context.Context) ([]string, error) {
			l.DebugCtx(ctx, "spawning worker")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
			if err != nil {
				l.ErrorCtx(ctx, "failed creating request object", "err", err)
				return nil, err
			}

			l.DebugCtx(ctx, "performing HTTP GET")
			resp, err := a.httpClient.Do(req)
			if err != nil {
				l.ErrorCtx(ctx, "failed fetching page", "err", err)
				return nil, err
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				l.ErrorCtx(ctx, "failed parsing document", "err", err)
				return nil, err
			}

			images := []string{}
			doc.Find("*").Each(func(_ int, item *goquery.Selection) {
				link, exists := item.Find("img").Attr("src")
				if !exists || link == "" {
					return
				}

				l.DebugCtx(ctx, "found link", "link", link)
				images = append(images, link)
			})

			return images, nil
		})
	}

	return p.Wait()
}
