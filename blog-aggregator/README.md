# Blog Aggregator Project

## Database Migrations

### Create a new migration
```bash
goose -dir sql/schema -s <migration_name> sql
```
Example:
```bash
goose -dir sql/schema -s posts sql
```

### Apply migrations
```bash
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://<username>:<password>@<host>:<port>/<database>" goose up -dir sql/schema -s
```
Replace `<username>`, `<password>`, `<host>`, `<port>`, and `<database>` with your actual database credentials.

## Generate SQL Code
To generate Go code from SQL:
```bash
sqlc generate
```