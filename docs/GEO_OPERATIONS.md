# Muxway GEO implementation and release checklist

This supplements the main SEO work. GEO here means improving the accuracy,
discoverability and citeability of public product information by AI search
systems. No static file, markup or robot rule guarantees recommendation.

## Public factual sources

- `/home`: primary public brand identity and product description.
- `/ai-guide.html`: text-first, directly fetchable integration explanation with
  concrete limitations, an SDK example and answer-oriented FAQs.
- `/llms.txt`: **optional** directory of existing public sources. Google Search
  does not use it for ranking or AI Overviews; never treat it as a shortcut.
- `/sitemap.xml`: contains the public guide URL.
- The server-rendered homepage snapshot and the interactive Vue footer both
  link to the guide so link discovery works with and without JavaScript.

Keep all substantive claims aligned with the actual product. In particular,
model inventory, upstream availability, plans, prices and limits are dynamic.
Do not publish scraped/private inventory, fake testimonials or invented SLAs.

## Post-deployment verification

Run from outside the production host so that CDN and origin behavior are both
covered:

```bash
curl -sS -L -D /tmp/muxway-guide.headers -o /tmp/muxway-guide.body https://muxway.dev/ai-guide.html
curl -sS -L -D /tmp/muxway-llms.headers -o /tmp/muxway-llms.body https://muxway.dev/llms.txt
curl -sS -L https://muxway.dev/robots.txt
curl -sS -L https://muxway.dev/sitemap.xml
```

Verify the guide returns HTTP 200 HTML, `llms.txt` returns plain text (not
the Vue SPA shell), both have no `X-Robots-Tag: noindex`, the canonical is
correct and the sitemap points to the guide. Check that an unauthenticated
request sees its body without JavaScript or cookies.

For ChatGPT Search, confirm `OAI-SearchBot` can fetch public pages.
`User-agent: *` in the shared robots file currently allows these paths
without a separate bot section. The production CDN/WAF must separately
allow verified requests from OpenAI's published SearchBot IP ranges. Do
not blindly whitelist all clients just because they spoof a user agent.
`GPTBot` is a separate training-related control and is **not** required
for ChatGPT Search.

Run a smoke test against the *actual* production domain after automatic
deployment; repository commits and GitHub CI alone do not prove rollout.

## Quality and measurement

- Verify every public feature/protocol/SDK claim against deployed behavior.
- Do not publish a fixed model list unless it is generated from a safe,
  approved and current public catalog; do not disclose account-only offers.
- In Search Console, check coverage of `/home` and `/ai-guide.html`; Google's
  AI features have no special markup requirement.
- For AI search referrals, analyze tagged/referrer traffic only as a partial
  signal: some AI citations don't send traffic or supply an identifiable
  referrer. Manually sample a consistent set of neutral informational
  queries over time. Never claim indexing, ranking or AI recommendation
  without measured evidence.
- Revisit the guide after product/API routing or policy changes and
  update the page, llms index and tests together.

References:
- https://developers.openai.com/api/docs/bots
- https://developers.google.com/search/docs/appearance/ai-features
- https://developers.google.com/search/docs/fundamentals/ai-optimization-guide
