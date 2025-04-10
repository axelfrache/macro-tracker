package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/frachea/macro-tracker/internal/database"
	"github.com/frachea/macro-tracker/internal/fdc"
)

var db *database.DB
var fdcClient *fdc.Client

func main() {
	// Initialiser la base de données
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5434")
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

	// Initialiser le client FDC API
	// Utiliser une clé API de développement par défaut si aucune n'est fournie
	// Pour la production, il faudra configurer une vraie clé API via les variables d'environnement
	fdcApiKey := getEnv("FDC_API_KEY", "DEMO_KEY")
	fdcClient = fdc.NewClient(fdcApiKey)

	r := gin.Default()

	// Configuration CORS pour permettre les requêtes depuis le frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Accepter toutes les origines en développement
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // Doit être false quand AllowOrigins contient *
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
		api.PUT("/meal-plan-items/:itemId", handleUpdateMealPlanItem)
		api.PUT("/meal-plan-items/:itemId/meal-type", handleUpdateMealPlanItem)
		api.DELETE("/meal-plan-items/:itemId", handleDeleteMealPlanItem)

		// Food Search
		api.GET("/food/search", handleSearchFood)
	}

	log.Println("Starting server on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
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

	// Récupérer l'utilisateur existant
	user, err := db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Mettre à jour les champs modifiés
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

	// Mettre à jour l'utilisateur dans la base de données
	// Note: Vous devrez ajouter une méthode UpdateUser à votre package database
	// Pour l'instant, nous renvoyons simplement l'utilisateur mis à jour
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

	// Pour chaque plan, récupérer les items associés
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

	// Renvoyer le plan créé avec un tableau d'items vide
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

	// Utiliser une structure intermédiaire pour éviter les problèmes de désérialisation
	type ItemRequest struct {
		MealType  string  `json:"meal_type"`
		FoodID    int     `json:"food_id"`
		FoodName  string  `json:"food_name"`
		Amount    float64 `json:"amount"`
		Proteins  float64 `json:"proteins"`
		Carbs     float64 `json:"carbs"`
		Fats      float64 `json:"fats"`
		Calories  float64 `json:"calories"`
		Fiber     float64 `json:"fiber"`
	}

	var itemReq ItemRequest
	if err := c.ShouldBindJSON(&itemReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log pour déboguer
	fmt.Printf("Type de repas reçu: %s\n", itemReq.MealType)

	// Créer l'objet MealPlanItem avec les bonnes valeurs
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

	// Utiliser une structure intermédiaire pour éviter les problèmes de désérialisation
	type ItemRequest struct {
		MealType string `json:"meal_type"`
	}

	var itemReq ItemRequest
	if err := c.ShouldBindJSON(&itemReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log pour déboguer
	fmt.Printf("Mise à jour du type de repas: %s pour l'élément %d\n", itemReq.MealType, itemID)

	// Mettre à jour le type de repas dans la base de données
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.Foods)
}

// Fonction utilitaire pour récupérer les variables d'environnement avec une valeur par défaut
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
