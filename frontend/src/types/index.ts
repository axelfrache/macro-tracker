export interface User {
  id: number;
  name: string;
  age: number;
  weight: number;
  height: number;
  gender: string;
  target_macros: {
    proteins?: number;
    carbs?: number;
    fats?: number;
    calories?: number;
    fiber?: number;
  };
}

export interface MealPlan {
  id: number;
  user_id: number;
  name: string;
  description: string;
  items: MealPlanItem[];
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

export interface Food {
  fdcId: number;
  description: string;
  dataType: string;
  nutrients: Nutrient[];
}

export interface Nutrient {
  id?: number;
  nutrientId?: number; // Pour la compatibilité avec différentes structures
  name?: string;
  nutrientName?: string;
  amount?: number;
  value?: number;
  unitName?: string;
  nutrient?: {
    id: number;
    number?: string;
    name: string;
    unitName: string;
    rank?: number;
    amount?: number;
    value?: number;
  };
}

export interface MacroNutrients {
  proteins: number;
  carbs: number;
  fats: number;
  calories: number;
  fiber: number;
}
