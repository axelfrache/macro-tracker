import { useState, useEffect, useCallback } from 'react';
import { Typography, Paper, Box, Alert, CircularProgress } from '@mui/material';
import { MacroChart } from '../components/MacroChart';
import { MealPlan } from '../types';
import { getMealPlans } from '../api';
import { useUser } from '../context/UserContext';

export const Home = () => {
  const [mealPlans, setMealPlans] = useState<MealPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { currentUserId, currentUser } = useUser();

  const fetchMealPlans = useCallback(async () => {
    if (!currentUserId) return;
    
    setLoading(true);
    setError(null);
    try {
      const response = await getMealPlans(currentUserId);
      setMealPlans(response.data);
    } catch (error) {
      console.error('Erreur lors de la récupération des données:', error);
      setError('Impossible de charger les données. Veuillez réessayer plus tard.');
    } finally {
      setLoading(false);
    }
  }, [currentUserId]);

  useEffect(() => {
    fetchMealPlans();
    return () => {
    };
  }, [fetchMealPlans]);

  const calculateTotalMacros = (plan: MealPlan) => {
    if (!plan.items || !Array.isArray(plan.items) || plan.items.length === 0) {
      return { proteins: 0, carbs: 0, fats: 0, calories: 0, fiber: 0 };
    }
    
    return plan.items.reduce(
      (acc, item) => ({
        proteins: acc.proteins + (item.proteins || 0),
        carbs: acc.carbs + (item.carbs || 0),
        fats: acc.fats + (item.fats || 0),
        calories: acc.calories + (item.calories || 0),
        fiber: acc.fiber + (item.fiber || 0),
      }),
      { proteins: 0, carbs: 0, fats: 0, calories: 0, fiber: 0 }
    );
  };

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Tableau de bord {currentUser ? `de ${currentUser.name}` : ''}
      </Typography>
      
      {loading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
          <CircularProgress />
        </Box>
      )}
      
      {error && (
        <Alert severity="error" sx={{ my: 2 }}>
          {error}
        </Alert>
      )}
      
      {!loading && !error && mealPlans.length === 0 && (
        <Alert severity="info" sx={{ my: 2 }}>
          Vous n'avez pas encore de journées types. Créez-en une dans la section "Journées Types".
        </Alert>
      )}
      
      {!loading && !error && mealPlans.length > 0 && (
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
                  fiber={macros.fiber}
                />
                <Typography variant="body2" color="text.secondary">
                  Total calories: {macros.calories.toFixed(0)} kcal
                </Typography>
              </Paper>
            );
          })}
        </Box>
      )}
    </Box>
  );
};
