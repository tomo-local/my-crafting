# URL Shortener with Analytics Tracking

Building a URL shortener with click analytics to understand ID generation, redirects, and event tracking.

## Implementations

- [Go](./go)

## Goals

- Short ID generation (Base62 encoding)
- Redirect with 301 / 302
- Click tracking: timestamp, user-agent, referrer, geo (IP-based)
- Analytics aggregation (clicks per day, top referrers)
