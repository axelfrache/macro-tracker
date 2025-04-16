import { useState, useEffect } from 'react';
import {
  Typography,
  Paper,
  TextField,
  Button,
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem
} from '@mui/material';
import { useNavigate } from 'react-router-dom';

import { User } from '../types';
import { updateUser } from '../api';
import { useUser } from '../context/UserContext';

export const Profile = () => {
  const { currentUser, currentUserId, setCurrentUserId } = useUser();
  const [user, setUser] = useState<User | null>(null);
  const [editing, setEditing] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (currentUser) {
      setUser(currentUser);
    }
  }, [currentUser]);

  const handleLogout = () => {
    setCurrentUserId(null);
    navigate('/login', { state: { autoLogout: true } });
  };

  const handleSave = async () => {
    if (!user || !currentUserId) return;

    try {
      await updateUser(currentUserId, user);
      setEditing(false);
    } catch (error) {
      console.error('Erreur lors de la mise à jour du profil:', error);
    }
  };

  if (!user) {
    return <Typography>Chargement...</Typography>;
  }

  return (
    <div>
      <Typography variant="h4" gutterBottom>
        Profil
      </Typography>
      <Paper sx={{ p: 3 }}>
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3 }}>
          <Box>
            <TextField
              fullWidth
              label="Nom"
              value={user.name}
              disabled={!editing}
              onChange={(e) => setUser({ ...user, name: e.target.value })}
            />
          </Box>
          <Box>
            <TextField
              fullWidth
              label="Âge"
              type="number"
              value={user.age}
              disabled={!editing}
              onChange={(e) => setUser({ ...user, age: parseInt(e.target.value) })}
            />
          </Box>
          <Box>
            <TextField
              fullWidth
              label="Poids (kg)"
              type="number"
              value={user.weight}
              disabled={!editing}
              onChange={(e) => setUser({ ...user, weight: parseFloat(e.target.value) })}
            />
          </Box>
          <Box>
            <TextField
              fullWidth
              label="Taille (cm)"
              type="number"
              value={user.height}
              disabled={!editing}
              onChange={(e) => setUser({ ...user, height: parseFloat(e.target.value) })}
            />
          </Box>
          <Box>
            <FormControl fullWidth>
              <InputLabel id="gender-label">Genre</InputLabel>
              <Select
                labelId="gender-label"
                id="gender"
                value={user.gender}
                label="Genre"
                disabled={!editing}
                onChange={(e) => setUser({ ...user, gender: e.target.value as string })}
              >
                <MenuItem value="homme">Homme</MenuItem>
                <MenuItem value="femme">Femme</MenuItem>
              </Select>
            </FormControl>
          </Box>
        </Box>
        <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end' }}>
          {editing ? (
            <>
              <Button onClick={() => setEditing(false)} sx={{ mr: 1 }}>
                Annuler
              </Button>
              <Button variant="contained" onClick={handleSave}>
                Enregistrer
              </Button>
            </>
          ) : (
            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button variant="contained" onClick={() => setEditing(true)}>
                Modifier
              </Button>
              <Button 
                variant="outlined" 
                color="secondary" 
                onClick={handleLogout}
              >
                Déconnexion
              </Button>
            </Box>
          )}
        </Box>
      </Paper>
    </div>
  );
};
