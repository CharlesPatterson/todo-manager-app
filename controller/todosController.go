package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/CharlesPatterson/todos-app/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetRootRedirectHandler(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
}

// @Summary Get a TODO by ID
// @ID get-todo-by-id
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} model.Todo
// @Router /todos/{id} [get]
func GetTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	todo, err := model.GetTodoById(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// @Summary Update a TODO by ID
// @ID update-todo-by-id
// @Produce json
// @Param id path string true "model.Todo ID"
// @Param data body model.Todo true "model.Todo data"
// @Success 200 {object} model.Todo
// @Router /todos/{id} [put]
func UpdateTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	var todo model.Todo
	if err := c.BindJSON(&todo); err != nil {

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		}
		return
	}

	updatedTodo, err := model.UpdateTodo(c, &todo, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, updatedTodo)
}

type ErrorMsg struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "lte":
		return "Should be less than " + fe.Param()
	case "gte":
		return "Should be greater than " + fe.Param()
	}
	return "Unknown error"
}

// @Summary Create a todo
// @ID create-todo
// @Produce json
// @Param data body model.Todo true "model.Todo data"
// @Success 200 {object} model.Todo
// @Router /todos [post]
func CreateTodoHandler(c *gin.Context) {
	var newTodo model.Todo

	if err := c.BindJSON(&newTodo); err != nil {

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make([]ErrorMsg, len(ve))
			for i, fe := range ve {
				out[i] = ErrorMsg{fe.Field(), getErrorMsg(fe)}
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out})
		}
		return
	}

	newTodo.CreatedAt = time.Now()
	newTodo.UpdatedAt = time.Now()
	newTodo.ID = primitive.NewObjectID()

	if err := model.CreateTodo(c, &newTodo); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newTodo)
}

// @Summary Get all todos
// @Description Get all todos without any filtering
// @Success 200 {object} model.Todo
// @Router /todos [get]
func GetAllTodosHandler(c *gin.Context) {
	todos, err := model.GetAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

// @Summary Delete a todo
// @ID delete-todo-by-id
// @Produce json
// @Param id path string true "model.Todo ID"
// @Success 200 {object} model.Todo
// @Router /todos [delete]
func DeleteTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	err := model.DeleteTodoById(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, "")
}
