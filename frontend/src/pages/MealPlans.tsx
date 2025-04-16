import { useState, useEffect, useCallback } from 'react';
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
  ListItemSecondaryAction,
  IconButton,
  Divider,
  CircularProgress,
  Alert,
  Box,
  Collapse,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { MealPlan } from '../types';
import { getMealPlans, createMealPlan } from '../api';
import { MealPlanDetails } from '../components/MealPlanDetails';
import { useUser } from '../context/UserContext';

export const MealPlans = () => {
  const [mealPlans, setMealPlans] = useState<MealPlan[]>([]);
  const [open, setOpen] = useState(false);
  const [newPlan, setNewPlan] = useState({ name: '', description: '', items: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [expandedPlanId, setExpandedPlanId] = useState<number | null>(null);
  const { currentUserId } = useUser();

  const fetchMealPlans = useCallback(async () => {
    if (!currentUserId) return;
    
    setLoading(true);
    setError(null);
    try {
      const response = await getMealPlans(currentUserId);
      setMealPlans(response.data);
    } catch (error) {
      console.error('Erreur lors de la récupération des journées types:', error);
      setError('Impossible de charger les journées types. Veuillez réessayer plus tard.');
    } finally {
      setLoading(false);
    }
  }, [currentUserId]);

  useEffect(() => {
    fetchMealPlans();
    return () => {
    };
  }, [fetchMealPlans]);

  const handleCreatePlan = async () => {
    if (!newPlan.name.trim()) {
      setCreateError('Le nom de la journée type est obligatoire');
      return;
    }
    
    setCreating(true);
    setCreateError(null);
    try {
      if (!currentUserId) {
        setCreateError('Utilisateur non connecté');
        setCreating(false);
        return;
      }
      await createMealPlan(currentUserId, newPlan);
      setOpen(false);
      setNewPlan({ name: '', description: '', items: [] });
      await fetchMealPlans();
    } catch (error) {
      console.error('Erreur lors de la création de la journée type:', error);
      setCreateError('Erreur lors de la création de la journée type. Veuillez réessayer.');
    } finally {
      setCreating(false);
    }
  };

  const toggleExpand = (planId: number) => {
    setExpandedPlanId(expandedPlanId === planId ? null : planId);
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
        disabled={loading}
      >
        Nouvelle journée type
      </Button>
      
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
          Vous n'avez pas encore de journées types. Créez-en une avec le bouton ci-dessus.
        </Alert>
      )}
      
      {!loading && !error && mealPlans.length > 0 && (
        <Paper>
          <List>
            {mealPlans.map((plan, index) => (
              <div key={plan.id}>
                {index > 0 && <Divider />}
                <ListItem onClick={() => toggleExpand(plan.id)} sx={{ cursor: 'pointer' }}>
                  <ListItemText
                    primary={plan.name}
                    secondary={plan.description}
                  />
                  <ListItemSecondaryAction>
                    <IconButton edge="end" onClick={(e) => {
                      e.stopPropagation();
                      toggleExpand(plan.id);
                    }}>
                      {expandedPlanId === plan.id ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                    </IconButton>
                  </ListItemSecondaryAction>
                </ListItem>
                <Collapse in={expandedPlanId === plan.id} timeout="auto" unmountOnExit>
                  <MealPlanDetails mealPlan={plan} onUpdate={fetchMealPlans} />
                </Collapse>
              </div>
            ))}
          </List>
        </Paper>
      )}

      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        disablePortal
        keepMounted
        disableEnforceFocus
        disableAutoFocus
      >
        <DialogTitle>Nouvelle journée type</DialogTitle>
        <DialogContent>
          {createError && (
            <Alert severity="error" sx={{ mt: 2, mb: 2 }}>
              {createError}
            </Alert>
          )}
          <TextField
            autoFocus
            margin="dense"
            label="Nom"
            fullWidth
            value={newPlan.name}
            onChange={(e) => setNewPlan({ ...newPlan, name: e.target.value })}
            error={createError?.includes('nom')}
            required
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
          <Button onClick={() => setOpen(false)} disabled={creating}>Annuler</Button>
          <Button 
            onClick={handleCreatePlan} 
            variant="contained" 
            disabled={creating || !newPlan.name.trim()}
          >
            {creating ? <CircularProgress size={24} /> : 'Créer'}
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};
