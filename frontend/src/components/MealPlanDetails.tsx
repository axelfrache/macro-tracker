import { useState } from 'react';
import {
  Typography,
  Paper,
  Box,
  Divider,
  List,
  ListItem,
  ListItemText,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
  IconButton,
  Tooltip,
  Menu,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { MealPlan, MealPlanItem } from '../types';
import { addMealPlanItem, searchFood, deleteMealPlanItem, updateMealPlanItemMealType } from '../api';

interface MealPlanDetailsProps {
  mealPlan: MealPlan;
  onUpdate: () => void;
}

export const MealPlanDetails = ({ mealPlan, onUpdate }: MealPlanDetailsProps) => {
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);
  const [selectedFood, setSelectedFood] = useState<any | null>(null);
  const [mealType, setMealType] = useState('breakfast');
  const [amount, setAmount] = useState('100');
  const [loading, setLoading] = useState(false);
  const [searchLoading, setSearchLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deletingItemId, setDeletingItemId] = useState<number | null>(null);
  const [editingItemId, setEditingItemId] = useState<number | null>(null);
  const [editAnchorEl, setEditAnchorEl] = useState<null | HTMLElement>(null);
  const [updatingItemId, setUpdatingItemId] = useState<number | null>(null);

  const mealTypeOptions = [
    { value: 'breakfast', label: 'Petit-déjeuner' },
    { value: 'snack1', label: 'Collation 1' },
    { value: 'lunch', label: 'Déjeuner' },
    { value: 'snack2', label: 'Collation 2' },
    { value: 'dinner', label: 'Dîner' },
  ];

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    
    setSearchLoading(true);
    setError(null);
    try {
      const response = await searchFood(searchQuery);
      // Le backend renvoie directement un tableau d'aliments, pas un objet avec une propriété foods
      if (response.data && Array.isArray(response.data)) {
        setSearchResults(response.data);
        console.log('Résultats de recherche:', response.data);
        
        if (response.data.length === 0) {
          setError('Aucun aliment trouvé pour cette recherche. Essayez avec d\'autres termes.');
        }
      } else if (response.data && typeof response.data === 'object' && response.data.foods) {
        // Au cas où le format de réponse changerait
        setSearchResults(response.data.foods);
        console.log('Résultats de recherche (format alternatif):', response.data.foods);
        
        if (response.data.foods.length === 0) {
          setError('Aucun aliment trouvé pour cette recherche. Essayez avec d\'autres termes.');
        }
      } else {
        console.error('Format de réponse inattendu:', response.data);
        setSearchResults([]);
        setError('Format de réponse inattendu. Veuillez réessayer.');
      }
    } catch (error: any) {
      console.error('Erreur lors de la recherche:', error);
      // Afficher un message d'erreur plus détaillé si disponible
      const errorMessage = error.response?.data?.error || error.message || 'Erreur inconnue';
      setError(`Impossible de rechercher des aliments: ${errorMessage}. Veuillez réessayer.`);
    } finally {
      setSearchLoading(false);
    }
  };

  const handleSelectFood = (food: any) => {
    setSelectedFood(food);
  };

  const handleAddItem = async () => {
    if (!selectedFood) {
      setError('Veuillez sélectionner un aliment');
      return;
    }

    const amountValue = parseFloat(amount);
    if (isNaN(amountValue) || amountValue <= 0) {
      setError('Veuillez entrer une quantité valide');
      return;
    }

    setLoading(true);
    setError(null);
    
    try {
      // Calculer les macros en fonction de la quantité
      const { fdcId, description, foodNutrients } = selectedFood;
      
      // Extraire les nutriments
      const proteins = foodNutrients.find((n: any) => n.nutrientName === 'Protein')?.value || 0;
      const carbs = foodNutrients.find((n: any) => n.nutrientName === 'Carbohydrate, by difference')?.value || 0;
      const fats = foodNutrients.find((n: any) => n.nutrientName === 'Total lipid (fat)')?.value || 0;
      const calories = foodNutrients.find((n: any) => n.nutrientName === 'Energy')?.value || 0;
      const fiber = foodNutrients.find((n: any) => n.nutrientName === 'Fiber, total dietary')?.value || 0;
      
      // Calculer les macros pour la quantité donnée (les valeurs sont pour 100g)
      const multiplier = amountValue / 100.0;
      
      const newItem: Omit<MealPlanItem, 'id' | 'meal_plan_id'> = {
        meal_type: mealType,
        food_id: fdcId,
        food_name: description,
        amount: amountValue,
        proteins: proteins * multiplier,
        carbs: carbs * multiplier,
        fats: fats * multiplier,
        calories: calories * multiplier,
        fiber: fiber * multiplier,
      };
      
      await addMealPlanItem(mealPlan.id, newItem);
      setOpen(false);
      resetForm();
      onUpdate();
    } catch (error) {
      console.error('Erreur lors de l\'ajout de l\'aliment:', error);
      setError('Impossible d\'ajouter l\'aliment. Veuillez réessayer.');
    } finally {
      setLoading(false);
    }
  };
  
  // Fonction pour supprimer un élément du meal plan
  // Fonction pour ouvrir le menu d'édition du type de repas
  const handleOpenEditMenu = (event: React.MouseEvent<HTMLElement>, itemId: number) => {
    setEditAnchorEl(event.currentTarget);
    setEditingItemId(itemId);
  };

  // Fonction pour fermer le menu d'édition
  const handleCloseEditMenu = () => {
    setEditAnchorEl(null);
    setEditingItemId(null);
  };

  // Fonction pour mettre à jour le type de repas d'un élément
  const handleUpdateMealType = async (itemId: number, newMealType: string) => {
    try {
      setUpdatingItemId(itemId);
      await updateMealPlanItemMealType(itemId, newMealType);
      handleCloseEditMenu();
      onUpdate();
    } catch (error) {
      console.error('Erreur lors de la mise à jour du type de repas:', error);
      setError('Impossible de mettre à jour le type de repas. Veuillez réessayer.');
    } finally {
      setUpdatingItemId(null);
    }
  };

  const handleDeleteItem = async (itemId: number) => {
    try {
      setDeletingItemId(itemId); // Indiquer quel élément est en cours de suppression
      await deleteMealPlanItem(itemId);
      onUpdate(); // Rafraîchir les données après la suppression
    } catch (error) {
      console.error('Erreur lors de la suppression de l\'élément:', error);
      setError('Impossible de supprimer l\'élément. Veuillez réessayer.');
    } finally {
      setDeletingItemId(null);
    }
  };

  const resetForm = () => {
    setSearchQuery('');
    setSearchResults([]);
    setSelectedFood(null);
    setMealType('breakfast');
    setAmount('100');
    setError(null);
  };

  // Grouper les items par type de repas et les trier selon l'ordre des types de repas
  const itemsByMealType: Record<string, MealPlanItem[]> = {};
  
  // Initialiser les types de repas dans l'ordre
  mealTypeOptions.forEach(option => {
    itemsByMealType[option.value] = [];
  });
  
  // Ajouter les items à leurs types de repas respectifs
  mealPlan.items?.forEach(item => {
    if (itemsByMealType[item.meal_type]) {
      itemsByMealType[item.meal_type].push(item);
    } else {
      // Au cas où il y aurait un type de repas non reconnu
      itemsByMealType[item.meal_type] = [item];
    }
  });

  // Calculer les macros totaux
  const totalMacros = mealPlan.items?.reduce(
    (acc, item) => {
      return {
        proteins: acc.proteins + item.proteins,
        carbs: acc.carbs + item.carbs,
        fats: acc.fats + item.fats,
        calories: acc.calories + item.calories,
      };
    },
    { proteins: 0, carbs: 0, fats: 0, calories: 0 }
  ) || { proteins: 0, carbs: 0, fats: 0, calories: 0 };

  return (
    <Paper sx={{ p: 3, mt: 2, maxWidth: '100%', overflowX: 'auto' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">{mealPlan.name}</Typography>
        <Button 
          variant="contained" 
          color="primary" 
          onClick={() => setOpen(true)}
        >
          Ajouter un aliment
        </Button>
      </Box>
      
      {mealPlan.description && (
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          {mealPlan.description}
        </Typography>
      )}
      
      <Box sx={{ mb: 4 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>Macros totaux</Typography>
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', md: 'repeat(4, 1fr)' }, gap: 2 }}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 2, 
              bgcolor: 'rgba(25, 118, 210, 0.08)', 
              border: '1px solid rgba(25, 118, 210, 0.2)',
              borderRadius: 2
            }}
          >
            <Typography variant="h6" sx={{ color: 'primary.main', fontWeight: 'bold' }}>
              {totalMacros.proteins.toFixed(1)}g
            </Typography>
            <Typography variant="body2" sx={{ color: 'primary.main' }}>
              Protéines
            </Typography>
          </Paper>
          
          <Paper 
            elevation={0} 
            sx={{ 
              p: 2, 
              bgcolor: 'rgba(46, 125, 50, 0.08)', 
              border: '1px solid rgba(46, 125, 50, 0.2)',
              borderRadius: 2
            }}
          >
            <Typography variant="h6" sx={{ color: 'success.main', fontWeight: 'bold' }}>
              {totalMacros.carbs.toFixed(1)}g
            </Typography>
            <Typography variant="body2" sx={{ color: 'success.main' }}>
              Glucides
            </Typography>
          </Paper>
          
          <Paper 
            elevation={0} 
            sx={{ 
              p: 2, 
              bgcolor: 'rgba(237, 108, 2, 0.08)', 
              border: '1px solid rgba(237, 108, 2, 0.2)',
              borderRadius: 2
            }}
          >
            <Typography variant="h6" sx={{ color: 'warning.main', fontWeight: 'bold' }}>
              {totalMacros.fats.toFixed(1)}g
            </Typography>
            <Typography variant="body2" sx={{ color: 'warning.main' }}>
              Lipides
            </Typography>
          </Paper>
          
          <Paper 
            elevation={0} 
            sx={{ 
              p: 2, 
              bgcolor: 'rgba(211, 47, 47, 0.08)', 
              border: '1px solid rgba(211, 47, 47, 0.2)',
              borderRadius: 2
            }}
          >
            <Typography variant="h6" sx={{ color: 'error.main', fontWeight: 'bold' }}>
              {totalMacros.calories.toFixed(0)} kcal
            </Typography>
            <Typography variant="body2" sx={{ color: 'error.main' }}>
              Calories
            </Typography>
          </Paper>
        </Box>
      </Box>
      
      <Divider sx={{ mb: 3 }} />
      
      {Object.entries(itemsByMealType).filter(([_, items]) => items.length > 0).length === 0 ? (
        <Alert severity="info">
          Aucun aliment n'a été ajouté à cette journée type. Utilisez le bouton "Ajouter un aliment" pour commencer.
        </Alert>
      ) : (
        Object.entries(itemsByMealType).map(([type, items]) => (
          items.length > 0 && (
            <Box key={type} sx={{ mb: 3 }}>
              <Typography variant="h6" sx={{ backgroundColor: '#f5f5f5', p: 1, borderRadius: 1 }} gutterBottom>
                {mealTypeOptions.find(opt => opt.value === type)?.label || type}
              </Typography>
              {/* Utiliser un div au lieu d'un Paper pour contenir la liste */}
              <Box sx={{ mb: 2, border: '1px solid #e0e0e0', borderRadius: 1, overflow: 'hidden' }}>
                {items.map((item, index) => (
                  <Box 
                    key={item.id} 
                    sx={{
                      p: 2,
                      borderBottom: index < items.length - 1 ? '1px solid #e0e0e0' : 'none',
                      '&:hover': { bgcolor: 'rgba(0, 0, 0, 0.02)' },
                      position: 'relative'
                    }}
                  >
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                      <Typography variant="subtitle1" sx={{ fontWeight: 'medium' }}>
                        {item.food_name} ({item.amount}g)
                      </Typography>
                      <Box>
                        <Tooltip title="Modifier le type de repas">
                          <IconButton 
                            size="small" 
                            color="primary" 
                            onClick={(e) => handleOpenEditMenu(e, item.id)}
                            disabled={updatingItemId === item.id}
                            sx={{ ml: 1 }}
                          >
                            {updatingItemId === item.id ? <CircularProgress size={20} /> : <EditIcon />}
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Supprimer">
                          <IconButton 
                            size="small" 
                            color="error" 
                            onClick={() => handleDeleteItem(item.id)}
                            disabled={deletingItemId === item.id}
                            sx={{ ml: 1 }}
                          >
                            {deletingItemId === item.id ? <CircularProgress size={20} /> : <DeleteIcon />}
                          </IconButton>
                        </Tooltip>
                      </Box>
                    </Box>
                    
                    <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr 1fr', sm: 'repeat(4, 1fr)' }, gap: 1.5 }}>
                      <Box sx={{ p: 1, bgcolor: 'rgba(25, 118, 210, 0.08)', borderRadius: 1, border: '1px solid rgba(25, 118, 210, 0.2)' }}>
                        <Typography variant="body2" sx={{ fontWeight: 'medium', color: 'primary.main' }}>
                          Protéines: {item.proteins.toFixed(1)}g
                        </Typography>
                      </Box>
                      <Box sx={{ p: 1, bgcolor: 'rgba(46, 125, 50, 0.08)', borderRadius: 1, border: '1px solid rgba(46, 125, 50, 0.2)' }}>
                        <Typography variant="body2" sx={{ fontWeight: 'medium', color: 'success.main' }}>
                          Glucides: {item.carbs.toFixed(1)}g
                        </Typography>
                      </Box>
                      <Box sx={{ p: 1, bgcolor: 'rgba(237, 108, 2, 0.08)', borderRadius: 1, border: '1px solid rgba(237, 108, 2, 0.2)' }}>
                        <Typography variant="body2" sx={{ fontWeight: 'medium', color: 'warning.main' }}>
                          Lipides: {item.fats.toFixed(1)}g
                        </Typography>
                      </Box>
                      <Box sx={{ p: 1, bgcolor: 'rgba(211, 47, 47, 0.08)', borderRadius: 1, border: '1px solid rgba(211, 47, 47, 0.2)' }}>
                        <Typography variant="body2" sx={{ fontWeight: 'medium', color: 'error.main' }}>
                          Calories: {item.calories.toFixed(0)} kcal
                        </Typography>
                      </Box>
                    </Box>
                  </Box>
                ))}
              </Box>
            </Box>
          )
        ))
      )}
      
      <Dialog 
        open={open} 
        onClose={() => !loading && setOpen(false)}
        fullWidth
        maxWidth="md"
        PaperProps={{
          sx: { maxHeight: '80vh', overflowY: 'auto' }
        }}
      >
        <DialogTitle>Ajouter un aliment</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mt: 2, mb: 2 }}>
              {error}
            </Alert>
          )}
          
          <Box sx={{ display: 'flex', mt: 2, mb: 2 }}>
            <TextField
              label="Rechercher un aliment"
              fullWidth
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              sx={{ mr: 1 }}
            />
            <Button 
              variant="contained" 
              onClick={handleSearch}
              disabled={searchLoading || !searchQuery.trim()}
            >
              {searchLoading ? <CircularProgress size={24} /> : 'Rechercher'}
            </Button>
          </Box>
          
          {searchLoading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
              <CircularProgress />
            </Box>
          )}
          
          {!searchLoading && searchResults.length > 0 && (
            <Box sx={{ mb: 3, border: '1px solid #e0e0e0', borderRadius: 1, mt: 2 }}>
              <Typography variant="subtitle1" sx={{ p: 1, bgcolor: '#f5f5f5', borderBottom: '1px solid #e0e0e0' }}>
                Résultats de recherche
              </Typography>
              <Box sx={{ maxHeight: '200px', overflow: 'auto' }}>
                <List disablePadding>
                  {searchResults.map((food) => (
                    <ListItem 
                      key={food.fdcId} 
                      onClick={() => handleSelectFood(food)}
                      divider
                      sx={{ 
                        cursor: 'pointer',
                        backgroundColor: selectedFood?.fdcId === food.fdcId ? 'rgba(0, 0, 0, 0.04)' : 'inherit',
                        '&:hover': {
                          backgroundColor: 'rgba(0, 0, 0, 0.08)'
                        }
                      }}
                    >
                      <ListItemText primary={food.description} />
                    </ListItem>
                  ))}
                </List>
              </Box>
            </Box>
          )}
          
          {selectedFood && (
            <Box sx={{ mt: 3, mb: 3, pb: 2, borderBottom: '1px solid #eee' }}>
              <Typography variant="subtitle1" sx={{ mb: 2, fontWeight: 'bold' }}>
                Aliment sélectionné: {selectedFood.description}
              </Typography>
              
              <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                <Box sx={{ flex: 1 }}>
                  <FormControl fullWidth variant="outlined">
                    <InputLabel id="meal-type-label">Type de repas</InputLabel>
                    <Select
                      labelId="meal-type-label"
                      id="meal-type-select"
                      value={mealType}
                      onChange={(e) => setMealType(e.target.value)}
                      label="Type de repas"
                    >
                      {mealTypeOptions.map((option) => (
                        <MenuItem key={option.value} value={option.value}>
                          {option.label}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                </Box>
                <Box sx={{ flex: 1 }}>
                  <TextField
                    label="Quantité (g)"
                    type="number"
                    fullWidth
                    variant="outlined"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    InputProps={{ inputProps: { min: 1 } }}
                  />
                </Box>
              </Box>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)} disabled={loading}>Annuler</Button>
          <Button 
            onClick={handleAddItem} 
            variant="contained" 
            disabled={loading || !selectedFood}
          >
            {loading ? <CircularProgress size={24} /> : 'Ajouter'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Menu pour modifier le type de repas */}
      <Menu
        anchorEl={editAnchorEl}
        open={Boolean(editAnchorEl)}
        onClose={handleCloseEditMenu}
      >
        {mealTypeOptions.map((option) => (
          <MenuItem 
            key={option.value} 
            onClick={() => editingItemId && handleUpdateMealType(editingItemId, option.value)}
            disabled={updatingItemId !== null}
          >
            {option.label}
          </MenuItem>
        ))}
      </Menu>
    </Paper>
  );
};
