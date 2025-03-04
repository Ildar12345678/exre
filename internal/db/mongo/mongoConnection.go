package mongo

import (
	"context"
	"database/sql"
	"time"
	"expenses2/internal/types"
	// "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type MongoDB struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewMongoDB(path string) (*MongoDB, error) {
	// path = uri|dbName|cooName
	params := strings.Split(path, "|")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(params[0]))
	if err != nil {
		return nil, err
	}
	coll := client.Database(params[1]).Collection(params[2])
	return &MongoDB{client: client, coll: coll}, nil
}

func (db *MongoDB) Initialize() error {
	// Create indexes for efficient querying
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"date", 1}}},     // Index on date for range queries
		{Keys: bson.D{{"name", 1}}},     // Index on name for searches
		{Keys: bson.D{{"category", 1}}}, // Index on category for grouping
	}
	_, err := db.coll.Indexes().CreateMany(context.Background(), indexes)
	return err
}

func (db *MongoDB) GetExpenses(dateLow, dateHigh time.Time) ([]*types.ExpenseShow, error) {
	ctx := context.Background()
	filter := bson.M{
		"date": bson.M{
			"$gte": dateLow,
			"$lte": dateHigh,
		},
	}
	opts := options.Find().SetSort(bson.D{{"date", 1}})
	cursor, err := db.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var dest []*types.ExpenseShow
	for cursor.Next(ctx) {
		var doc struct {
			Date     time.Time `bson:"date"`
			Count    float32   `bson:"count"`
			Price    float32   `bson:"price"`
			Name     string    `bson:"name"`
			Category string    `bson:"category"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		expense := &types.ExpenseShow{
			Date:     doc.Date.Format("2006-01-02"), // Convert to YYYY-MM-DD
			Price:    doc.Count * doc.Price,         // Round to int64 as in SQLite
			Name:     doc.Name,
			Category: doc.Category,
		}
		dest = append(dest, expense)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return nil, types.ErrNoRecord
	}
	return dest, nil
}

func (db *MongoDB) GetStatistics(dateLow, dateHigh time.Time) ([]*types.Statistics, error) {
	ctx := context.Background()
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{
			"date": bson.M{"$gte": dateLow, "$lte": dateHigh},
		}}},
		bson.D{{"$group", bson.M{
			"_id":          "$category",
			"sum_category": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$count", "$price"}}},
		}}},
		bson.D{{"$sort", bson.M{"_id": -1}}},
		bson.D{{"$project", bson.M{
			"_id":          0,
			"category":     "$_id",
			"sum_category": 1,
		}}},
	}
	cursor, err := db.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var dest []*types.Statistics
	for cursor.Next(ctx) {
		var stat types.Statistics
		if err := cursor.Decode(&stat); err != nil {
			return nil, err
		}
		dest = append(dest, &stat)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return nil, types.ErrNoRecord
	}
	return dest, nil
}

func (db *MongoDB) SearchExpense(name string) ([]*types.ExpenseSearch, error) {
	ctx := context.Background()
	filter := bson.M{
		"name": bson.M{"$regex": name, "$options": "i"}, // Case-insensitive regex
	}
	opts := options.Find().SetSort(bson.D{{"date", 1}})
	cursor, err := db.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var dest []*types.ExpenseSearch
	for cursor.Next(ctx) {
		var doc struct {
			Name  string    `bson:"name"`
			Price string    `bson:"price"`
			Date  time.Time `bson:"date"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		expense := &types.ExpenseSearch{
			Name:  doc.Name,
			Price: doc.Price,
			Date:  doc.Date.Format("2006-01-02"),
		}
		dest = append(dest, expense)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return nil, types.ErrNoRecord
	}
	return dest, nil
}

func (db *MongoDB) GetCities() ([]string, error) {
	return nil, nil
}

func (db *MongoDB) GetCategories() ([]string, error) {
	return nil, nil
}

func (db *MongoDB) GetExpensesNames(ids ...any) ([]string, error) {
	ctx := context.Background()
	filter := bson.M{}
	if len(ids) > 0 {
		filter["_id"] = bson.M{"$in": ids} // Assumes ids are ObjectIDs or convertible
	}
	opts := options.Find().SetProjection(bson.M{"name": 1, "_id": 0})
	cursor, err := db.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var dest []string
	for cursor.Next(ctx) {
		var doc struct {
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		dest = append(dest, doc.Name)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if len(dest) == 0 {
		return nil, types.ErrNoRecord
	}
	return dest, nil
}

func (db *MongoDB) AddExpense(expense *types.ExpenseAdd) error {
	ctx := context.Background()
	doc := bson.M{
		"date":     expense.Date,
		"name":     expense.Name,
		"category": expense.Category,
		"city":     expense.City,
		"online":   expense.Online,
		"count":    expense.Count,
		"price":    expense.Price,
	}
	res, err := db.coll.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("expense insert error: %v", err)
	}
	// MongoDB returns an ObjectID, not an int64. Convert or adjust return type if needed.
	if _, ok := res.InsertedID.(primitive.ObjectID); ok {
		return nil // Placeholder; adjust as needed
	}
	return nil
}

func (db *MongoDB) Close() error {
	return db.client.Disconnect(context.Background())
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
