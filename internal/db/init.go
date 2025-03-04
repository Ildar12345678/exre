package db

import (
	"expenses2/internal/app"
	"expenses2/internal/db/sqlite"
	"expenses2/internal/db/mongo"
	"errors"
)

func NewDB(dbName, path string) (app.DB, error) {
	switch dbName {
	case "sqlite":
		return sqlite.NewSQLiteDB(path)
	case "mongodb":
		return mongo.NewMongoDB(path)
	default:
		return nil, errors.New("incorrect db choice")
	}
}
