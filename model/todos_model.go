package model

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection

func init() {
	var ctx = context.TODO()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load .env file: %e", err)
	}

	mongoURI := os.Getenv("DB_URI")
	databaseName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_COLLECTION_NAME")

	credential := options.Credential{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(databaseName).Collection(collectionName)
}

type TodoDocInput struct {
	Text      string `json:"text" bson:"text"`
	Completed bool   `json:"completed" bson:"completed"`
}

type Todo struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	Text      string             `json:"text" bson:"text"`
	Completed bool               `json:"completed" bson:"completed"`
}

func CreateTodo(ctx context.Context, todo *Todo) error {
	_, err := Collection.InsertOne(ctx, todo)
	return err
}

func GetAll(ctx context.Context) ([]*Todo, error) {
	filter := bson.D{{}}
	return FilterTodos(ctx, filter)
}

func GetTodoById(ctx context.Context, id string) (*Todo, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectId}
	t := &Todo{}
	err = Collection.FindOne(ctx, filter).Decode(t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func UpdateTodo(ctx context.Context, todo *Todo, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{primitive.E{
		Key: "_id", Value: objectId,
	}}
	t := &Todo{}
	err = Collection.FindOne(ctx, filter).Decode(t)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"completed":  todo.Completed,
			"text":       todo.Text,
			"updated_at": time.Now(),
		},
	}

	_, err = Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func FilterTodos(ctx context.Context, filter interface{}) ([]*Todo, error) {
	var todos []*Todo

	cur, err := Collection.Find(ctx, filter)
	if err != nil {
		return todos, err
	}

	for cur.Next(ctx) {
		var t Todo
		err := cur.Decode(&t)
		if err != nil {
			return todos, err
		}

		todos = append(todos, &t)
	}

	if err := cur.Err(); err != nil {
		return todos, err
	}

	err = cur.Close(ctx)
	if err != nil {
		return todos, err
	}

	if len(todos) == 0 {
		return todos, mongo.ErrNoDocuments
	}

	return todos, nil
}

func CompleteTodo(ctx context.Context, text string) error {
	filter := bson.D{primitive.E{Key: "text", Value: text}}

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "completed", Value: true},
	}}}

	t := &Todo{}
	return Collection.FindOneAndUpdate(ctx, filter, update).Decode(t)
}

func GetPending(ctx context.Context) ([]*Todo, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: false},
	}

	return FilterTodos(ctx, filter)
}

func GetFinished(ctx context.Context) ([]*Todo, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: true},
	}

	return FilterTodos(ctx, filter)
}

func DeleteTodoById(ctx context.Context, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}

	res, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no todos were deleted")
	}

	return nil
}

func DeleteTodo(ctx context.Context, text string) error {
	filter := bson.D{
		primitive.E{Key: "text", Value: text},
	}

	res, err := Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no todos were deleted")
	}

	return nil
}

func PrintTodos(todos []*Todo) {
	for i, v := range todos {
		if v.Completed {
			color.Green("%d: %s\n", i+1, v.Text)
		} else {
			color.Yellow("%d: %s\n", i+1, v.Text)
		}
	}
}
