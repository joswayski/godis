### Godis

Helpful Discord bot! For now, replaces links in Discord so that video embeds work properly. I plan to add more functionality soon.

<img src="./assets/godis.png" alt="godis" style="width: 200px;" />



#### Link replacements:

- `twitter.com` and `x.com` with `vxtwitter.com`  
- `facebook.com` with `facebed.com`  
- `instagram.com` with `eeinstagram.com`


#### Prerequisites
A `DISCORD_TOKEN` for a bot with the following:

- #### Scopes
    - applications.commands
    - bot
- #### Permissions
    - Manage Messages
    - Manage Webhooks
    - Send Messages


#### Running

Create a `.env` file at the root, and `go run ./cmd/godis`
```bash
{"time":"2026-02-24T12:29:20.323438+03:00","level":"INFO","msg":"Godis is starting..."}
{"time":"2026-02-24T12:29:20.323658+03:00","level":"INFO","msg":"Godis is ready!"}
```
