package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configuration CORS pour permettre les requêtes depuis le frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://frontend:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Routes API
	api := r.Group("")
	{
		// Users
		api.GET("/users/:id", handleGetUser)
		api.POST("/users", handleCreateUser)
		api.PUT("/users/:id", handleUpdateUser)

		// Meal Plans
		api.GET("/users/:id/meal-plans", handleGetMealPlans)
		api.POST("/users/:id/meal-plans", handleCreateMealPlan)
		api.POST("/meal-plans/:planId/items", handleAddMealPlanItem)

		// Food Search
		api.GET("/food/search", handleSearchFood)
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// Handlers temporaires qui renvoient des données de test
func handleGetUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     1,
		"name":   "John Doe",
		"age":    30,
		"weight": 75.5,
		"height": 180,
	})
}

func handleCreateUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":     1,
		"name":   "John Doe",
		"age":    30,
		"weight": 75.5,
		"height": 180,
	})
}

func handleUpdateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     1,
		"name":   "John Doe",
		"age":    30,
		"weight": 75.5,
		"height": 180,
	})
}

func handleGetMealPlans(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{
			"id":          1,
			"user_id":     1,
			"name":        "Plan standard",
			"description": "Plan équilibré pour la journée",
			"items":       []gin.H{},
		},
	})
}

func handleCreateMealPlan(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":          2,
		"user_id":     1,
		"name":        "Nouveau plan",
		"description": "Description du plan",
		"items":       []gin.H{},
	})
}

func handleAddMealPlanItem(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":           1,
		"meal_plan_id": 1,
		"name":         "Item test",
		"quantity":     100,
		"unit":         "g",
	})
}

func handleSearchFood(c *gin.Context) {
	c.JSON(http.StatusOK, []gin.H{
		{
			"id":   1,
			"name": "Pomme",
			"nutrients": gin.H{
				"calories": 52,
				"protein":  0.3,
				"fat":      0.2,
				"carbs":    14,
			},
		},
	})
}
