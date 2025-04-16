package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/frachea/macro-tracker/internal/database"
	"github.com/frachea/macro-tracker/internal/fdc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var db *database.DB
var fdcClient *fdc.Client

func main() {
	dbHost := getEnv("DB_HOST", "db")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "macro_tracker")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = database.NewDB(connStr)
	if err != nil {
		log.Fatalf("Erreur de connexion à la base de données: %v\n", err)
	}

	fdcApiKey := getEnv("FDC_API_KEY", "DEMO_KEY")
	fdcClient = fdc.NewClient(fdcApiKey)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Origines spécifiques pour le développement
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Routes API
	api := r.Group("")
	{
		api.GET("/users", handleGetUsers)
		api.GET("/users/:id", handleGetUser)
		api.POST("/users", handleCreateUser)
		api.PUT("/users/:id", handleUpdateUser)

		api.GET("/users/:id/meal-plans", handleGetMealPlans)
		api.POST("/users/:id/meal-plans", handleCreateMealPlan)
		api.POST("/meal-plans/:planId/items", handleAddMealPlanItem)
		api.PUT("/meal-plan-items/:itemId", handleUpdateMealPlanItem)
		api.PUT("/meal-plan-items/:itemId/meal-type", handleUpdateMealPlanItem)
		api.DELETE("/meal-plan-items/:itemId", handleDeleteMealPlanItem)

		api.GET("/food/search", handleSearchFood)
		api.GET("/food/:id", handleGetFood)
	}

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func handleGetUsers(c *gin.Context) {
	users, err := db.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des utilisateurs"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func handleGetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	user, err := db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func handleCreateUser(c *gin.Context) {
	var user database.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.AddUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func handleUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	var userUpdate database.User
	if err := c.ShouldBindJSON(&userUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	if userUpdate.Name != "" {
		user.Name = userUpdate.Name
	}
	if userUpdate.Age != 0 {
		user.Age = userUpdate.Age
	}
	if userUpdate.Weight != 0 {
		user.Weight = userUpdate.Weight
	}
	if userUpdate.Height != 0 {
		user.Height = userUpdate.Height
	}
	if userUpdate.Gender != "" {
		user.Gender = userUpdate.Gender
	}
	if len(userUpdate.TargetMacros) > 0 {
		user.TargetMacros = userUpdate.TargetMacros
	}

	err = db.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de l'utilisateur: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func handleGetMealPlans(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	plans, err := db.GetMealPlans(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, 0, len(plans))
	for _, plan := range plans {
		items, err := db.GetMealPlanItems(plan.ID)
		if err != nil {
			log.Printf("Erreur lors de la récupération des items pour le plan %d: %v", plan.ID, err)
			items = []database.MealPlanItem{}
		}

		planMap := map[string]interface{}{
			"id":          plan.ID,
			"user_id":     plan.UserID,
			"name":        plan.Name,
			"description": plan.Description,
			"items":       items,
		}
		result = append(result, planMap)
	}

	c.JSON(http.StatusOK, result)
}

func handleCreateMealPlan(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	var planData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&planData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan := &database.MealPlan{
		UserID:      userID,
		Name:        planData.Name,
		Description: planData.Description,
	}

	err = db.CreateMealPlan(plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"id":          plan.ID,
		"user_id":     plan.UserID,
		"name":        plan.Name,
		"description": plan.Description,
		"items":       []database.MealPlanItem{},
	}

	c.JSON(http.StatusCreated, response)
}

func handleAddMealPlanItem(c *gin.Context) {
	planIDStr := c.Param("planId")
	planID, err := strconv.Atoi(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de plan invalide"})
		return
	}

	type ItemRequest struct {
		MealType string  `json:"meal_type"`
		FoodID   int     `json:"food_id"`
		FoodName string  `json:"food_name"`
		Amount   float64 `json:"amount"`
		Proteins float64 `json:"proteins"`
		Carbs    float64 `json:"carbs"`
		Fats     float64 `json:"fats"`
		Calories float64 `json:"calories"`
		Fiber    float64 `json:"fiber"`
	}

	var itemReq ItemRequest
	if err := c.ShouldBindJSON(&itemReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Type de repas reçu: %s\n", itemReq.MealType)

	item := database.MealPlanItem{
		MealPlanID: planID,
		MealType:   database.MealType(itemReq.MealType), // Conversion explicite
		FoodID:     itemReq.FoodID,
		FoodName:   itemReq.FoodName,
		Amount:     itemReq.Amount,
		Proteins:   itemReq.Proteins,
		Carbs:      itemReq.Carbs,
		Fats:       itemReq.Fats,
		Calories:   itemReq.Calories,
		Fiber:      itemReq.Fiber,
	}

	err = db.AddMealPlanItem(&item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func handleUpdateMealPlanItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'élément invalide"})
		return
	}

	type ItemRequest struct {
		MealType string `json:"meal_type"`
	}

	var itemReq ItemRequest
	if err := c.ShouldBindJSON(&itemReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Mise à jour du type de repas: %s pour l'élément %d\n", itemReq.MealType, itemID)

	err = db.UpdateMealPlanItemMealType(itemID, database.MealType(itemReq.MealType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Type de repas mis à jour avec succès"})
}

func handleDeleteMealPlanItem(c *gin.Context) {
	itemIDStr := c.Param("itemId")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'élément invalide"})
		return
	}

	err = db.DeleteMealPlanItem(itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Élément non trouvé"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Élément supprimé avec succès"})
}

func handleSearchFood(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requête de recherche manquante"})
		return
	}

	result, err := fdcClient.SearchFoods(query)
	if err != nil {
		log.Printf("Erreur lors de la recherche FDC pour '%s': %v", query, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(result.Foods) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	// Vérifier que les résultats ont bien des nutriments
	processedFoods := make([]map[string]interface{}, 0, len(result.Foods))
	for i, food := range result.Foods {
		// Limiter à 10 résultats pour des performances optimales
		if i >= 10 {
			break
		}
		
		// Ajouter des informations nutritionnelles si absentes
		detailedFood := &food
		if len(food.Nutrients) == 0 {
			tmpFood, err := fdcClient.GetFood(food.FdcID)
			if err == nil && tmpFood != nil {
				detailedFood = tmpFood
			}
		}
		
		// Calculer les macros pour chaque aliment
		proteins, carbs, fats, calories, fiber := detailedFood.GetMacros()
		
		// Vérifier si les valeurs sont valides
		validNutrients := proteins > 0 || carbs > 0 || fats > 0 || calories > 0
		
		// Ajouter l'information des macros dans les log pour le débogage
		log.Printf("Aliment: %s, Protéines: %.2f, Glucides: %.2f, Lipides: %.2f, Calories: %.2f, Fibres: %.2f, Valide: %v",
			detailedFood.Description, proteins, carbs, fats, calories, fiber, validNutrients)
		
		// Créer un objet avec les informations nécessaires
		processedFood := map[string]interface{}{
			"fdcId":       detailedFood.FdcID,
			"description": detailedFood.Description,
			"dataType":    detailedFood.DataType,
			"nutrients":   detailedFood.Nutrients,
			"macros": map[string]float64{
				"proteins": proteins,
				"carbs":    carbs,
				"fats":     fats,
				"calories": calories,
				"fiber":    fiber,
			},
		}
		
		// N'ajouter que les aliments avec des valeurs nutritionnelles valides
		if validNutrients {
			processedFoods = append(processedFoods, processedFood)
		}
	}

	c.JSON(http.StatusOK, processedFoods)
}

func handleGetFood(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID aliment invalide"})
		return
	}

	food, err := fdcClient.GetFood(id)
	if err != nil {
		log.Printf("Erreur lors de la récupération de l'aliment %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculer les macros
	proteins, carbs, fats, calories, fiber := food.GetMacros()
	
	// Ajouter les macros aux détails de l'aliment pour simplifier l'utilisation côté client
	response := map[string]interface{}{
		"fdcId":       food.FdcID,
		"description": food.Description,
		"nutrients":   food.Nutrients,
		"macros": map[string]float64{
			"proteins": proteins,
			"carbs":    carbs,
			"fats":     fats, 
			"calories": calories,
			"fiber":    fiber,
		},
	}
	
	log.Printf("Détail aliment: %s, Protéines: %.2f, Glucides: %.2f, Lipides: %.2f, Calories: %.2f, Fibres: %.2f",
		food.Description, proteins, carbs, fats, calories, fiber)

	c.JSON(http.StatusOK, response)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
