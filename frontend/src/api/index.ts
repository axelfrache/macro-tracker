import axios from 'axios';
import { User, MealPlan, MealPlanItem } from '../types';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

export const getUser = (id: number) => api.get<User>(`/users/${id}`);
export const createUser = (user: Omit<User, 'id'>) => api.post<User>('/users', user);
export const updateUser = (id: number, user: Partial<User>) => api.put<User>(`/users/${id}`, user);

export const getMealPlans = (userId: number) => api.get<MealPlan[]>(`/users/${userId}/meal-plans`);
export const createMealPlan = (userId: number, plan: Omit<MealPlan, 'id' | 'user_id'>) => 
  api.post<MealPlan>(`/users/${userId}/meal-plans`, plan);
export const addMealPlanItem = (planId: number, item: Omit<MealPlanItem, 'id' | 'meal_plan_id'>) =>
  api.post<MealPlanItem>(`/meal-plans/${planId}/items`, item);

export const searchFood = (query: string) => api.get(`/food/search?query=${encodeURIComponent(query)}`);
