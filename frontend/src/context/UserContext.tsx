import React, { createContext, useState, useContext, useEffect, ReactNode } from 'react';
import { getUser } from '../api';
import { User } from '../types';

interface UserContextType {
  currentUserId: number | null;
  currentUser: User | null;
  setCurrentUserId: (id: number | null) => void;
  loading: boolean;
  error: string | null;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export const UserProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [currentUserId, setCurrentUserId] = useState<number | null>(() => {
    const savedUserId = localStorage.getItem('currentUserId');
    return savedUserId ? parseInt(savedUserId, 10) : null;
  });
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (currentUserId) {
      localStorage.setItem('currentUserId', currentUserId.toString());
      
      const fetchUser = async () => {
        setLoading(true);
        setError(null);
        try {
          const response = await getUser(currentUserId);
          setCurrentUser(response.data);
        } catch (err) {
          console.error('Erreur lors de la récupération de l\'utilisateur:', err);
          setError('Impossible de charger les informations de l\'utilisateur');
          setCurrentUser(null);
          setCurrentUserId(null);
          localStorage.removeItem('currentUserId');
        } finally {
          setLoading(false);
        }
      };

      fetchUser();
    } else {
      localStorage.removeItem('currentUserId');
      setCurrentUser(null);
    }
  }, [currentUserId]);

  return (
    <UserContext.Provider value={{ currentUserId, currentUser, setCurrentUserId, loading, error }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = (): UserContextType => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
};
