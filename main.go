package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CharlesPatterson/todos-app/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var collection *mongo.Collection
var ctx = context.TODO()

type Todo struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Text      string             `bson:"text"`
	Completed bool               `bson:"completed"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load .env file: %e", err)
	}

	mongoURI := os.Getenv("DB_URI")
	databaseName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("DB_COLLECTION_NAME")

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(databaseName).Collection(collectionName)
}

func createTodo(todo *Todo) error {
	_, err := collection.InsertOne(ctx, todo)
	return err
}

func getAll() ([]*Todo, error) {
	filter := bson.D{{}}
	return filterTodos(filter)
}

func getTodoById(id string) (*Todo, error) {
	filter := bson.D{primitive.E{
		Key: "_id", Value: id,
	}}
	t := &Todo{}
	err := collection.FindOne(ctx, filter).Decode(t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func filterTodos(filter interface{}) ([]*Todo, error) {
	var todos []*Todo

	cur, err := collection.Find(ctx, filter)
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

	cur.Close(ctx)

	if len(todos) == 0 {
		return todos, mongo.ErrNoDocuments
	}

	return todos, nil
}

func completeTodo(text string) error {
	filter := bson.D{primitive.E{Key: "text", Value: text}}

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "completed", Value: true},
	}}}

	t := &Todo{}
	return collection.FindOneAndUpdate(ctx, filter, update).Decode(t)
}

func getPending() ([]*Todo, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: false},
	}

	return filterTodos(filter)
}

func getFinished() ([]*Todo, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: true},
	}

	return filterTodos(filter)
}

func deleteTodoById(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}

	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no todos were deleted")
	}

	return nil
}

func deleteTodo(text string) error {
	filter := bson.D{
		primitive.E{Key: "text", Value: text},
	}

	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("no todos were deleted")
	}

	return nil
}

func printTodos(todos []*Todo) {
	for i, v := range todos {
		if v.Completed {
			color.Green("%d: %s\n", i+1, v.Text)
		} else {
			color.Yellow("%d: %s\n", i+1, v.Text)
		}
	}
}

func getTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	todo, err := getTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func getAllTodosHandler(c *gin.Context) {
	todos, err := getAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func deleteTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	err := deleteTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, "")
}

func main() {
	app := &cli.App{
		Name:  "Todos App",
		Usage: "A simple CLI program to manage your todos",
		Action: func(c *cli.Context) error {
			todos, err := getPending()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
					return nil
				}
				return err
			}

			printTodos(todos)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a todo to the list",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						return errors.New("cannot add an empty todo")
					}

					todo := &Todo{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Text:      str,
						Completed: false,
					}

					return createTodo(todo)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "list all todos",
				Action: func(c *cli.Context) error {
					todos, err := getAll()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
							return nil
						}

						return err
					}
					printTodos(todos)
					return nil
				},
			},
			{
				Name:    "done",
				Aliases: []string{"d"},
				Usage:   "complete a todo on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					return completeTodo(text)
				},
			},
			{
				Name:    "finished",
				Aliases: []string{"f"},
				Usage:   "list completed todos",
				Action: func(c *cli.Context) error {
					todos, err := getFinished()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("Nothing to see here.\nRun `add 'todo'` to add a todo")
							return nil
						}
						return err
					}

					printTodos(todos)
					return nil
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"rm"},
				Usage:   "deletes a todo on the list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					err := deleteTodo(text)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "starts a server to interact with mongodb",
				Action: func(c *cli.Context) error {
					r := gin.Default()
					r.Use(middleware.TimeoutMiddleware())
					r.Use(gin.Logger())
					r.Use(gin.Recovery())
					r.SetTrustedProxies(nil)
					version := "/v1"
					v1 := r.Group(version)
					{
						v1.GET("/todos", getAllTodosHandler)
						v1.GET("/todos/:id", getTodoByIdHandler)
						v1.DELETE("/todos/:id", deleteTodoByIdHandler)
					}
					port := os.Getenv("PORT")
					r.Run(port)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
