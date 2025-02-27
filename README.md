# tgtag: Save Tags from Telegram Exported HTML

## Save data
1. Export Telegram channel history.
2. Configure app `etc/config.yml` (copy from `etc/config.yml.example`).
3. Move html-files to `%system.data_path%/you_channel/*.html`
4. `docker compose up`
5. `go run ./cmd/save/main.go` (go 1.23.4)
6. `docker compose down`.
7. Check `mongodb://localhost:27017`, database: `tgtag`, collection: `messages` (`%mongo.uri%`, `%mongo.database%`, `%mongo.collection_messages%`).
