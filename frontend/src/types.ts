export interface User {
  id: number;
  name: string;
  age: number;
  weight: number;
  height: number;
  gender: string;
  target_macros: Record<string, number>;
}

export interface MacroNutrients {
  proteins: number;
  carbs: number;
  fats: number;
  calories: number;
  fiber: number;
}

export interface Food {
  fdcId: number;
  description: string;
  dataType?: string;
  nutrients?: any[];
  macros: MacroNutrients;
}

export interface MealPlanItem {
  id: number;
  meal_plan_id: number;
  meal_type: string;
  food_id: number;
  food_name: string;
  amount: number;
  proteins: number;
  carbs: number;
  fats: number;
  calories: number;
  fiber: number;
}

export interface MealPlan {
  id: number;
  user_id: number;
  name: string;
  description: string;
  items: MealPlanItem[];
} 