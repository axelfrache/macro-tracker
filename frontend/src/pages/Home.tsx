import { useState, useEffect } from 'react';
import { Typography, Paper, Box } from '@mui/material';
import { MacroChart } from '../components/MacroChart';
import { MealPlan } from '../types';
import { getMealPlans } from '../api';

export const Home = () => {
  const [mealPlans, setMealPlans] = useState<MealPlan[]>([]);
  const userId = 1; // À remplacer par l'ID de l'utilisateur connecté

  useEffect(() => {
    const fetchMealPlans = async () => {
      try {
        const response = await getMealPlans(userId);
        setMealPlans(response.data);
      } catch (error) {
        console.error('Erreur lors de la récupération des journées types:', error);
      }
    };

    fetchMealPlans();
  }, [userId]);

  const calculateTotalMacros = (plan: MealPlan) => {
    return plan.items.reduce(
      (acc, item) => ({
        proteins: acc.proteins + item.proteins,
        carbs: acc.carbs + item.carbs,
        fats: acc.fats + item.fats,
        calories: acc.calories + item.calories,
      }),
      { proteins: 0, carbs: 0, fats: 0, calories: 0 }
    );
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Tableau de bord
      </Typography>
      <Box sx={{ display: 'grid', gap: 3, gridTemplateColumns: { xs: '1fr', md: 'repeat(2, 1fr)' } }}>
        {mealPlans.map((plan) => {
          const macros = calculateTotalMacros(plan);
          return (
            <Paper key={plan.id} sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                {plan.name}
              </Typography>
              <MacroChart
                proteins={macros.proteins}
                carbs={macros.carbs}
                fats={macros.fats}
              />
              <Typography variant="body2" color="text.secondary">
                Total calories: {macros.calories.toFixed(0)} kcal
              </Typography>
            </Paper>
          );
        })}
      </Box>
    </Box>
  );
};
