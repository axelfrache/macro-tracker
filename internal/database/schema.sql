CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INTEGER,
    weight FLOAT,
    height FLOAT,
    gender VARCHAR(10),
    target_macros JSONB
);

CREATE TYPE meal_type AS ENUM ('breakfast', 'snack1', 'lunch', 'snack2', 'dinner');

CREATE TABLE IF NOT EXISTS meals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    meal_type VARCHAR(50) NOT NULL,
    meal_date TIMESTAMP NOT NULL,
    food_id INTEGER NOT NULL,
    food_name VARCHAR(255) NOT NULL,
    amount FLOAT NOT NULL,
    proteins FLOAT NOT NULL,
    carbs FLOAT NOT NULL,
    fats FLOAT NOT NULL,
    calories FLOAT NOT NULL,
    fiber FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS meal_plans (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS meal_plan_items (
    id SERIAL PRIMARY KEY,
    meal_plan_id INTEGER REFERENCES meal_plans(id),
    meal_type VARCHAR(50) NOT NULL,
    food_id INTEGER NOT NULL,
    food_name VARCHAR(255) NOT NULL,
    amount FLOAT NOT NULL,
    proteins FLOAT NOT NULL,
    carbs FLOAT NOT NULL,
    fats FLOAT NOT NULL,
    calories FLOAT NOT NULL,
    fiber FLOAT NOT NULL
);
