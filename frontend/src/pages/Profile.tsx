import { useState, useEffect } from 'react';
import {
  Typography,
  Paper,
  TextField,
  Button,
  Box,
} from '@mui/material';

import { User } from '../types';
import { getUser, updateUser } from '../api';



export const Profile = () => {
  const [user, setUser] = useState<User | null>(null);
  const [editing, setEditing] = useState(false);
  const userId = 1; // À remplacer par l'ID de l'utilisateur connecté

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const response = await getUser(userId);
        setUser(response.data);
      } catch (error) {
        console.error('Erreur lors de la récupération du profil:', error);
      }
    };

    fetchUser();
  }, [userId]);

  const handleSave = async () => {
    if (!user) return;

    try {
      await updateUser(userId, user);
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
            <Button variant="contained" onClick={() => setEditing(true)}>
              Modifier
            </Button>
          )}
        </Box>
      </Paper>
    </div>
  );
};
