package database

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type User struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	Age          int             `json:"age"`
	Weight       float64         `json:"weight"`
	Height       float64         `json:"height"`
	TargetMacros json.RawMessage `json:"target_macros"`
}

type MealType string

const (
	Breakfast MealType = "breakfast"
	Snack1    MealType = "snack1"
	Lunch     MealType = "lunch"
	Snack2    MealType = "snack2"
	Dinner    MealType = "dinner"
)

type Meal struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MealType  string    `json:"meal_type"`
	MealDate  time.Time `json:"meal_date"`
	FoodID    int       `json:"food_id"`
	FoodName  string    `json:"food_name"`
	Amount    float64   `json:"amount"`
	Proteins  float64   `json:"proteins"`
	Carbs     float64   `json:"carbs"`
	Fats      float64   `json:"fats"`
	Calories  float64   `json:"calories"`
	Fiber     float64   `json:"fiber"`
}

type MealPlan struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MealPlanItem struct {
	ID         int      `json:"id"`
	MealPlanID int      `json:"meal_plan_id"`
	MealType   MealType `json:"meal_type"`
	FoodID     int      `json:"food_id"`
	FoodName   string   `json:"food_name"`
	Amount     float64  `json:"amount"`
	Proteins   float64  `json:"proteins"`
	Carbs      float64  `json:"carbs"`
	Fats       float64  `json:"fats"`
	Calories   float64  `json:"calories"`
	Fiber      float64  `json:"fiber"`
}

func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) AddUser(user *User) error {
	query := `
		INSERT INTO users (name, age, weight, height, target_macros)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	
	return db.QueryRow(query, user.Name, user.Age, user.Weight, user.Height, user.TargetMacros).Scan(&user.ID)
}

func (db *DB) GetUser(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, name, age, weight, height, target_macros FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Age, &user.Weight, &user.Height, &user.TargetMacros)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetUsers() ([]User, error) {
	query := `SELECT id, name, age, weight, height, target_macros FROM users ORDER BY id`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Weight, &user.Height, &user.TargetMacros)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (db *DB) AddMeal(meal *Meal) error {
	query := `
		INSERT INTO meals (user_id, meal_type, meal_date, food_id, food_name, amount, proteins, carbs, fats, calories, fiber)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	
	return db.QueryRow(
		query,
		meal.UserID,
		meal.MealType,
		meal.MealDate,
		meal.FoodID,
		meal.FoodName,
		meal.Amount,
		meal.Proteins,
		meal.Carbs,
		meal.Fats,
		meal.Calories,
		meal.Fiber,
	).Scan(&meal.ID)
}

func (db *DB) GetDailyMeals(userID int, date time.Time) ([]Meal, error) {
	rows, err := db.DB.Query(`
		SELECT id, user_id, meal_type, date, food_id, food_name, amount, proteins, carbs, fats, calories, fiber
		FROM meals
		WHERE user_id = $1 AND DATE(date) = DATE($2)
		ORDER BY date ASC
	`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meals []Meal
	for rows.Next() {
		var meal Meal
		err := rows.Scan(
			&meal.ID, &meal.UserID, &meal.MealType, &meal.MealDate,
			&meal.FoodID, &meal.FoodName, &meal.Amount,
			&meal.Proteins, &meal.Carbs, &meal.Fats, &meal.Calories, &meal.Fiber,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, meal)
	}

	return meals, nil
}

func (db *DB) GetDailyTotals(userID int, date time.Time) (proteins, carbs, fats, calories, fiber float64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(proteins), 0) as total_proteins,
			COALESCE(SUM(carbs), 0) as total_carbs,
			COALESCE(SUM(fats), 0) as total_fats,
			COALESCE(SUM(calories), 0) as total_calories,
			COALESCE(SUM(fiber), 0) as total_fiber
		FROM meals
		WHERE user_id = $1 AND meal_date = $2`

	err = db.QueryRow(query, userID, date).Scan(&proteins, &carbs, &fats, &calories, &fiber)
	return
}

func (db *DB) CreateMealPlan(plan *MealPlan) error {
	query := `
		INSERT INTO meal_plans (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id`

	return db.QueryRow(query, plan.UserID, plan.Name, plan.Description).Scan(&plan.ID)
}

func (db *DB) GetMealPlans(userID int) ([]MealPlan, error) {
	rows, err := db.Query(`
		SELECT id, user_id, name, description
		FROM meal_plans
		WHERE user_id = $1
		ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []MealPlan
	for rows.Next() {
		var plan MealPlan
		err := rows.Scan(&plan.ID, &plan.UserID, &plan.Name, &plan.Description)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (db *DB) GetMealPlanItems(planID int) ([]MealPlanItem, error) {
	rows, err := db.Query(`
		SELECT id, meal_plan_id, meal_type, food_id, food_name, amount, proteins, carbs, fats, calories, fiber
		FROM meal_plan_items
		WHERE meal_plan_id = $1
		ORDER BY id
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MealPlanItem
	for rows.Next() {
		var item MealPlanItem
		err := rows.Scan(
			&item.ID,
			&item.MealPlanID,
			&item.MealType,
			&item.FoodID,
			&item.FoodName,
			&item.Amount,
			&item.Proteins,
			&item.Carbs,
			&item.Fats,
			&item.Calories,
			&item.Fiber,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (db *DB) AddMealPlanItem(item *MealPlanItem) error {
	query := `
		INSERT INTO meal_plan_items (meal_plan_id, meal_type, food_id, food_name, amount, proteins, carbs, fats, calories, fiber)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	return db.QueryRow(
		query,
		item.MealPlanID,
		item.MealType,
		item.FoodID,
		item.FoodName,
		item.Amount,
		item.Proteins,
		item.Carbs,
		item.Fats,
		item.Calories,
		item.Fiber,
	).Scan(&item.ID)
}

func (db *DB) UpdateMealPlanItemMealType(itemID int, mealType MealType) error {
	query := `UPDATE meal_plan_items SET meal_type = $1 WHERE id = $2`
	
	result, err := db.Exec(query, mealType, itemID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

func (db *DB) DeleteMealPlanItem(itemID int) error {
	query := `DELETE FROM meal_plan_items WHERE id = $1`
	
	result, err := db.Exec(query, itemID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}
