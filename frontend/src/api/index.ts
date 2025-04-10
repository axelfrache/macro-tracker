import axios, { AxiosResponse } from 'axios';
import { User, MealPlan, MealPlanItem } from '../types';

let apiBaseUrl = import.meta.env.VITE_API_URL;

if (typeof window !== 'undefined' && apiBaseUrl.includes('backend')) {
  apiBaseUrl = 'http://localhost:8081';
}

console.log('Using API URL:', apiBaseUrl);

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
    console.error('API Error:', error.message);
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
export const searchFood = async (query: string): Promise<AxiosResponse<any>> => {
  if (!query.trim()) {
    return Promise.resolve({ data: [] } as AxiosResponse<any>);
  }
  
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  
  return new Promise((resolve, reject) => {
    searchTimeout = window.setTimeout(async () => {
      try {
        const response = await api.get(`/food/search?query=${encodeURIComponent(query)}`);
        resolve(response);
      } catch (error) {
        reject(error);
      }
    }, 300);
  });
};
