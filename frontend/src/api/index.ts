import axios, { AxiosResponse } from 'axios';
import { User, MealPlan, MealPlanItem, Food, MacroNutrients } from '../types';

let apiBaseUrl = import.meta.env.VITE_API_URL;

if (typeof window !== 'undefined' && apiBaseUrl.includes('backend')) {
  apiBaseUrl = 'http://localhost:8080';
}

const api = axios.create({
  baseURL: apiBaseUrl,
  timeout: 10000,
});

api.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    return Promise.reject(error);
  }
);

interface CacheItem<T> {
  data: T;
  timestamp: number;
  expiresIn: number;
}

class ApiCache {
  private cache: Record<string, CacheItem<any>> = {};
  
  set<T>(key: string, data: T, expiresIn: number = 60000): void {
    this.cache[key] = {
      data,
      timestamp: Date.now(),
      expiresIn
    };
  }
  
  get<T>(key: string): T | null {
    const item = this.cache[key];
    if (!item) return null;
    
    const isExpired = Date.now() > item.timestamp + item.expiresIn;
    if (isExpired) {
      delete this.cache[key];
      return null;
    }
    
    return item.data;
  }
  
  invalidate(keyPattern: RegExp): void {
    Object.keys(this.cache).forEach(key => {
      if (keyPattern.test(key)) {
        delete this.cache[key];
      }
    });
  }
}

const cache = new ApiCache();

export const getUser = async (id: number): Promise<AxiosResponse<User>> => {
  const cacheKey = `user_${id}`;
  const cachedData = cache.get<AxiosResponse<User>>(cacheKey);
  
  if (cachedData) {
    return cachedData;
  }
  
  const response = await api.get<User>(`/users/${id}`);
  cache.set(cacheKey, response, 300000);
  return response;
};

export const createUser = async (user: Omit<User, 'id'>): Promise<AxiosResponse<User>> => {
  const response = await api.post<User>('/users', user);
  cache.invalidate(/^user_/);
  return response;
};

export const updateUser = async (id: number, user: Partial<User>): Promise<AxiosResponse<User>> => {
  const response = await api.put<User>(`/users/${id}`, user);
  cache.invalidate(new RegExp(`^user_${id}`));
  return response;
};

export const getMealPlans = async (userId: number): Promise<AxiosResponse<MealPlan[]>> => {
  const cacheKey = `meal_plans_${userId}`;
  const cachedData = cache.get<AxiosResponse<MealPlan[]>>(cacheKey);
  
  if (cachedData) {
    return cachedData;
  }
  
  const response = await api.get<MealPlan[]>(`/users/${userId}/meal-plans`);
  cache.set(cacheKey, response, 60000);
  return response;
};

export const createMealPlan = async (userId: number, plan: Omit<MealPlan, 'id' | 'user_id'>): Promise<AxiosResponse<MealPlan>> => {
  const response = await api.post<MealPlan>(`/users/${userId}/meal-plans`, plan);
  cache.invalidate(new RegExp(`^meal_plans_${userId}`));
  return response;
};

export const addMealPlanItem = async (planId: number, item: Omit<MealPlanItem, 'id' | 'meal_plan_id'>): Promise<AxiosResponse<MealPlanItem>> => {
  const response = await api.post<MealPlanItem>(`/meal-plans/${planId}/items`, item);
  cache.invalidate(/^meal_plans_/);
  return response;
};

export const deleteMealPlanItem = async (itemId: number): Promise<AxiosResponse<{message: string}>> => {
  const response = await api.delete<{message: string}>(`/meal-plan-items/${itemId}`);
  cache.invalidate(/^meal_plans_/);
  return response;
};

export const updateMealPlanItemMealType = async (itemId: number, mealType: string): Promise<AxiosResponse<{message: string}>> => {
  const response = await api.put<{message: string}>(`/meal-plan-items/${itemId}/meal-type`, { meal_type: mealType });
  cache.invalidate(/^meal_plans_/);
  return response;
};

let searchTimeout: number | null = null;
export const searchFood = async (query: string): Promise<AxiosResponse<Food[]>> => {
  if (!query.trim()) {
    return {
      data: [],
      status: 200,
      statusText: 'OK',
      headers: {},
      config: {} as any
    } as AxiosResponse<Food[]>;
  }
  
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  
  return new Promise((resolve, reject) => {
    searchTimeout = window.setTimeout(async () => {
      try {
        const response = await api.get<Food[]>(`/food/search?query=${encodeURIComponent(query)}`);
        resolve(response);
      } catch (error) {
        reject(error);
      }
    }, 300);
  });
};

export const getFood = async (foodId: number): Promise<AxiosResponse<Food>> => {
  const cacheKey = `food_${foodId}`;
  const cachedData = cache.get<AxiosResponse<Food>>(cacheKey);
  
  if (cachedData) {
    return cachedData;
  }
  
  const response = await api.get<Food>(`/food/${foodId}`);
  cache.set(cacheKey, response, 3600000);
  return response;
};

export const calculateMacros = (food: Food, amount: number): MacroNutrients => {
  const getNutrientValue = (nutrientIds: number[]): number => {
    if (!food.nutrients || !Array.isArray(food.nutrients)) {
      return 0;
    }
    
    for (const nutrient of food.nutrients) {
      if (nutrient.nutrient) {
        for (const id of nutrientIds) {
          if (nutrient.nutrient.id === id) {
            if (nutrient.amount !== undefined && nutrient.amount > 0) {
              return nutrient.amount;
            }
            
            if (nutrient.value !== undefined && nutrient.value > 0) {
              return nutrient.value;
            }
            
            if (nutrient.nutrient.amount !== undefined && nutrient.nutrient.amount > 0) {
              return nutrient.nutrient.amount;
            }
            
            if (nutrient.nutrient.value !== undefined && nutrient.nutrient.value > 0) {
              return nutrient.nutrient.value;
            }
            
            try {
              const { name } = nutrient.nutrient;
              
              if (name && name.toLowerCase().includes('protein') && (id === 1003 || id === 203)) {
                const amountStr = String(nutrient.amount || 0);
                const amountValue = parseFloat(amountStr);
                return isNaN(amountValue) ? 0 : amountValue;
              }
              
              if (name && name.toLowerCase().includes('carbohydrate') && (id === 1005 || id === 205)) {
                const amountStr = String(nutrient.amount || 0);
                const amountValue = parseFloat(amountStr);
                return isNaN(amountValue) ? 0 : amountValue;
              }
              
              if (name && name.toLowerCase().includes('fat') && (id === 1004 || id === 204)) {
                const amountStr = String(nutrient.amount || 0);
                const amountValue = parseFloat(amountStr);
                return isNaN(amountValue) ? 0 : amountValue;
              }
              
              if (name && name.toLowerCase().includes('energy') && (id === 1008 || id === 208)) {
                const amountStr = String(nutrient.amount || 0);
                const amountValue = parseFloat(amountStr);
                return isNaN(amountValue) ? 0 : amountValue;
              }
              
              if (name && name.toLowerCase().includes('fiber') && (id === 1079 || id === 291)) {
                const amountStr = String(nutrient.amount || 0);
                const amountValue = parseFloat(amountStr);
                return isNaN(amountValue) ? 0 : amountValue;
              }
            } catch (error) {
            }
          }
        }
      }
      
      for (const id of nutrientIds) {
        if (nutrient.id === id || nutrient.nutrientId === id) {
          if (nutrient.amount !== undefined && nutrient.amount > 0) {
            return nutrient.amount;
          }
          
          if (nutrient.value !== undefined && nutrient.value > 0) {
            return nutrient.value;
          }
        }
      }
    }
    
    for (const nutrient of food.nutrients) {
      try {
        const name = nutrient.name || (nutrient.nutrient && nutrient.nutrient.name);
        if (name) {
          for (const id of nutrientIds) {
            if ((id === 1003 || id === 203) && name.toLowerCase().includes('protein')) {
              const amountValue = nutrient.amount || nutrient.value || 0;
              return amountValue;
            }
            
            if ((id === 1005 || id === 205) && name.toLowerCase().includes('carbohydrate')) {
              const amountValue = nutrient.amount || nutrient.value || 0;
              return amountValue;
            }
            
            if ((id === 1004 || id === 204) && name.toLowerCase().includes('fat')) {
              const amountValue = nutrient.amount || nutrient.value || 0;
              return amountValue;
            }
            
            if ((id === 1008 || id === 208) && (name.toLowerCase().includes('energy') || name.toLowerCase().includes('calor'))) {
              const amountValue = nutrient.amount || nutrient.value || 0;
              return amountValue;
            }
            
            if ((id === 1079 || id === 291) && name.toLowerCase().includes('fiber')) {
              const amountValue = nutrient.amount || nutrient.value || 0;
              return amountValue;
            }
          }
        }
      } catch (error) {
      }
    }
    
    return 0;
  };
  
  const proteinIds = [1003, 203];
  const carbIds = [1005, 205];
  const fatIds = [1004, 204];
  const calorieIds = [1008, 208];
  const fiberIds = [1079, 291];
  
  const proteins = getNutrientValue(proteinIds);
  const carbs = getNutrientValue(carbIds);
  const fats = getNutrientValue(fatIds);
  const calories = getNutrientValue(calorieIds);
  const fiber = getNutrientValue(fiberIds);
  
  const factor = amount / 100;
  
  const result = {
    proteins: proteins * factor,
    carbs: carbs * factor,
    fats: fats * factor,
    calories: calories * factor,
    fiber: fiber * factor
  };
  
  return result;
};
