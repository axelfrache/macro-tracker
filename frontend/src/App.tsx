import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material';
import { Layout } from './components/Layout';
import { Home } from './pages/Home';
import { MealPlans } from './pages/MealPlans';
import { Profile } from './pages/Profile';
import { Login } from './pages/Login';
import { UserProvider, useUser } from './context/UserContext';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

// Composant qui vérifie si l'utilisateur est connecté
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { currentUserId, loading } = useUser();
  
  if (loading) {
    return <div>Chargement...</div>;
  }
  
  if (!currentUserId) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

// Composant principal de l'application
export const App = () => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <UserProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/" element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }>
              <Route index element={<Home />} />
              <Route path="meal-plans" element={<MealPlans />} />
              <Route path="profile" element={<Profile />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </UserProvider>
    </ThemeProvider>
  );
};
