# Greenlight
Movie API.


## Dir overview
- `bin` for compiled app binaries, ready for production
- `cmd/api` for app-specific code, including code for running the server, read/write for HTTP reqs, and managing auth.
- `internal` for internal packages - various, ancillary code that (could be) used across multiple packages
- `migrations` for SQL migrations
- `remote` for config files & setup scripts for the prod server

