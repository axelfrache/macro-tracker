import { useState, useEffect } from 'react';
import {
  Typography,
  Paper,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import { MealPlan } from '../types';
import { getMealPlans, createMealPlan } from '../api';

export const MealPlans = () => {
  const [mealPlans, setMealPlans] = useState<MealPlan[]>([]);
  const [open, setOpen] = useState(false);
  const [newPlan, setNewPlan] = useState({ name: '', description: '', items: [] });
  const userId = 1; // À remplacer par l'ID de l'utilisateur connecté

  useEffect(() => {
    fetchMealPlans();
  }, []);

  const fetchMealPlans = async () => {
    try {
      const response = await getMealPlans(userId);
      setMealPlans(response.data);
    } catch (error) {
      console.error('Erreur lors de la récupération des journées types:', error);
    }
  };

  const handleCreatePlan = async () => {
    try {
      await createMealPlan(userId, newPlan);
      setOpen(false);
      setNewPlan({ name: '', description: '', items: [] });
      fetchMealPlans();
    } catch (error) {
      console.error('Erreur lors de la création de la journée type:', error);
    }
  };

  return (
    <div>
      <Typography variant="h4" gutterBottom>
        Journées Types
      </Typography>
      <Button
        variant="contained"
        color="primary"
        onClick={() => setOpen(true)}
        sx={{ mb: 2 }}
      >
        Nouvelle journée type
      </Button>
      <Paper>
        <List>
          {mealPlans.map((plan, index) => (
            <div key={plan.id}>
              {index > 0 && <Divider />}
              <ListItem>
                <ListItemText
                  primary={plan.name}
                  secondary={plan.description}
                />
              </ListItem>
            </div>
          ))}
        </List>
      </Paper>

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>Nouvelle journée type</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Nom"
            fullWidth
            value={newPlan.name}
            onChange={(e) => setNewPlan({ ...newPlan, name: e.target.value })}
          />
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            multiline
            rows={4}
            value={newPlan.description}
            onChange={(e) => setNewPlan({ ...newPlan, description: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Annuler</Button>
          <Button onClick={handleCreatePlan} variant="contained">
            Créer
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};
