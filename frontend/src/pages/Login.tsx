import { useState, useEffect } from 'react';
import { 
  Box, 
  Typography, 
  TextField, 
  Button, 
  Paper, 
  Container,
  CircularProgress,
  Alert,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemButton
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useUser } from '../context/UserContext';
import { createUser } from '../api';
import axios from 'axios';

export const Login = () => {
  const { setCurrentUserId, currentUserId } = useUser();
  const [userId, setUserId] = useState<string>('');
  const [newUserName, setNewUserName] = useState<string>('');
  const [newUserAge, setNewUserAge] = useState<string>('');
  const [newUserWeight, setNewUserWeight] = useState<string>('');
  const [newUserHeight, setNewUserHeight] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [existingUsers, setExistingUsers] = useState<{id: number, name: string}[]>([]);
  const [loadingUsers, setLoadingUsers] = useState<boolean>(false);
  
  const navigate = useNavigate();

  useEffect(() => {
    if (currentUserId) {
      navigate('/');
    }
  }, [currentUserId, navigate]);

  useEffect(() => {
    const fetchUsers = async () => {
      setLoadingUsers(true);
      try {
        setExistingUsers([
          { id: 1, name: "Utilisateur 1" },
          { id: 2, name: "Nouveau" }
        ]);
      } catch (err) {
        console.error("Erreur lors du chargement des utilisateurs:", err);
      } finally {
        setLoadingUsers(false);
      }
    };

    fetchUsers();
  }, [success]);

  const handleLogin = () => {
    const id = parseInt(userId, 10);
    if (isNaN(id) || id <= 0) {
      setError('Veuillez entrer un ID utilisateur valide');
      return;
    }

    setCurrentUserId(id);
    navigate('/');
  };

  const handleCreateUser = async () => {
    if (!newUserName || !newUserAge || !newUserWeight || !newUserHeight) {
      setError('Veuillez remplir tous les champs');
      return;
    }

    const age = parseInt(newUserAge, 10);
    const weight = parseFloat(newUserWeight);
    const height = parseFloat(newUserHeight);

    if (isNaN(age) || isNaN(weight) || isNaN(height)) {
      setError('Veuillez entrer des valeurs numériques valides');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const response = await createUser({
        name: newUserName,
        age,
        weight,
        height,
        target_macros: {}
      });

      setSuccess(`Compte créé avec succès! Votre ID est: ${response.data.id}`);
      setNewUserName('');
      setNewUserAge('');
      setNewUserWeight('');
      setNewUserHeight('');
      
      // Connecter automatiquement avec le nouvel utilisateur
      setCurrentUserId(response.data.id);
      navigate('/');
    } catch (err) {
      console.error('Erreur lors de la création du compte:', err);
      if (axios.isAxiosError(err) && err.response) {
        setError(`Erreur: ${err.response.data.error || err.message}`);
      } else {
        setError('Une erreur est survenue lors de la création du compte');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleSelectUser = (id: number) => {
    setCurrentUserId(id);
    navigate('/');
  };

  return (
    <Container maxWidth="sm">
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom align="center">
          Macro-Tracker
        </Typography>
        <Typography variant="h6" gutterBottom align="center">
          Suivi de vos macronutriments
        </Typography>

        <Paper elevation={3} sx={{ p: 3, mt: 4 }}>
          <Typography variant="h5" gutterBottom>
            Utilisateurs existants
          </Typography>
          
          {loadingUsers ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', my: 2 }}>
              <CircularProgress size={24} />
            </Box>
          ) : (
            <List>
              {existingUsers.map((user) => (
                <ListItem key={user.id} disablePadding>
                  <ListItemButton onClick={() => handleSelectUser(user.id)}>
                    <ListItemText 
                      primary={user.name} 
                      secondary={`ID: ${user.id}`} 
                    />
                  </ListItemButton>
                </ListItem>
              ))}
            </List>
          )}

          <Divider sx={{ my: 3 }} />

          <Typography variant="h5" gutterBottom>
            Se connecter avec un ID
          </Typography>
          
          <TextField
            label="ID Utilisateur"
            variant="outlined"
            fullWidth
            margin="normal"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            type="number"
          />
          
          <Button 
            variant="contained" 
            color="primary" 
            fullWidth 
            sx={{ mt: 2 }}
            onClick={handleLogin}
            disabled={loading}
          >
            Se connecter
          </Button>

          <Divider sx={{ my: 3 }} />

          <Typography variant="h5" gutterBottom>
            Créer un nouveau compte
          </Typography>
          
          <TextField
            label="Nom"
            variant="outlined"
            fullWidth
            margin="normal"
            value={newUserName}
            onChange={(e) => setNewUserName(e.target.value)}
          />
          
          <TextField
            label="Âge"
            variant="outlined"
            fullWidth
            margin="normal"
            value={newUserAge}
            onChange={(e) => setNewUserAge(e.target.value)}
            type="number"
          />
          
          <TextField
            label="Poids (kg)"
            variant="outlined"
            fullWidth
            margin="normal"
            value={newUserWeight}
            onChange={(e) => setNewUserWeight(e.target.value)}
            type="number"
            inputProps={{ step: 0.1 }}
          />
          
          <TextField
            label="Taille (cm)"
            variant="outlined"
            fullWidth
            margin="normal"
            value={newUserHeight}
            onChange={(e) => setNewUserHeight(e.target.value)}
            type="number"
            inputProps={{ step: 0.1 }}
          />
          
          <Button 
            variant="contained" 
            color="secondary" 
            fullWidth 
            sx={{ mt: 2 }}
            onClick={handleCreateUser}
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : 'Créer un compte'}
          </Button>
          
          {error && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {error}
            </Alert>
          )}
          
          {success && (
            <Alert severity="success" sx={{ mt: 2 }}>
              {success}
            </Alert>
          )}
        </Paper>
      </Box>
    </Container>
  );
};
