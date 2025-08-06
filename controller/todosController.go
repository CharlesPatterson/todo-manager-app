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

func GetTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	todo, err := model.GetTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func UpdateTodoHandler(c *gin.Context) {
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

	updatedTodo, err := model.UpdateTodo(&todo)
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

	if err := model.CreateTodo(&newTodo); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newTodo)
}

// @Summary Get all todos
// @Description Get all todos without any filtering
// @Success 200 {object} Todo
// @Router /v1/todos [get]
func GetAllTodosHandler(c *gin.Context) {
	todos, err := model.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func DeleteTodoByIdHandler(c *gin.Context) {
	id := c.Param("id")

	err := model.DeleteTodoById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, "")
}
