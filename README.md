# Youtube Video Summary

A small app that provides a short summary of any YouTube video. The logic is the following:

1. Fetch captions from YouTube video
2. Ask ChatGPT to provide the summary of the video by provided captions

The results are cached (in-memory, expiration duration is configurable). \
HTML page is embedded to the binary file, so the whole app is a single file

The service can be easily deployed to fly.io with `flyctl deploy` command. The `app` variable in `fly.toml` has to be
updated accordingly
There is a rate limiter (by default 3 requests per minute).

## Environment variables

### Mandatory variables

1. `OPENAI_KEY` - the OpenAI key that is used by ChatGPT

### Optional variables

* `MAX_VIDEO_DURATION_MINUTES` - the maximum video length (in minutes). Default value is `20`. It's required to avoid
  sending long videos to ChatGPT, because it might be costly
* `CACHE_EXPIRATION_HOURS` - how long to store data in cache. Default value is `12` hours
* `REQUEST_PER_MINUTE_RATE` - maximum requests per minute per IP Address. Default value is `3`

The app is accessible on https://youtube-summarise.fly.dev/
