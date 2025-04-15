# Macro-Tracker

Macro-Tracker est une application de suivi nutritionnel complète qui permet de suivre vos apports en macronutriments au quotidien. Elle se compose d'une interface en ligne de commande (CLI) et d'une interface web moderne.

## Prérequis

- Docker et Docker Compose
- Go 1.23 ou supérieur
- Node.js 20 ou supérieur
- Une clé API FoodData Central (FDC) - [Obtenir une clé ici](https://fdc.nal.usda.gov/api-key-signup.html)

## Installation

1. Clonez le dépôt :
```bash
git clone https://github.com/axelfrache/macro-tracker.git
cd macro-tracker
```

2. Lancez l'application avec Docker Compose :
```bash
docker-compose up -d
```

L'application sera disponible aux adresses suivantes :
- Frontend : http://localhost:5173
- Backend API : http://localhost:8080

## Lancement du CLI

Vous pouvez lancer le CLI de deux façons :

### 1. Via Docker (recommandé)

Après l'installation initiale ou après chaque modification du code, reconstruisez les conteneurs :
```bash
docker-compose down
docker-compose build
docker-compose up -d
```

Puis lancez le CLI :
```bash
docker exec -it macro-tracker-backend-1 ./cli
```

### 2. En local (pour le développement)

Depuis la racine du projet :
```bash
# Compilation du CLI
go build -o cli ./cmd/cli

# Lancement du CLI
./cli
```

**Note** : Assurez-vous que la base de données PostgreSQL est en cours d'exécution avant de lancer le CLI.

## Utilisation du CLI

Le CLI offre une interface en ligne de commande pour gérer vos repas et suivre vos macros.

### Première utilisation

Lors du premier lancement, vous devrez créer un compte en fournissant les informations suivantes :
- Nom
- Âge
- Poids (en kg)
- Taille (en cm)

Un ID utilisateur vous sera attribué, conservez-le pour vos prochaines connexions.

### Commandes disponibles

1. **Connexion/Création de compte**
```bash
# Au lancement, entrez :
- 0 pour créer un nouveau compte
- Votre ID utilisateur pour vous connecter
```

2. **Recherche d'aliments** :
```bash
search <nom de l'aliment>
```
Exemple : `search pomme`
- Affiche une liste d'aliments correspondant à votre recherche
- Chaque résultat inclut un ID (fdcId) à utiliser pour l'ajout d'un aliment

3. **Ajout d'un aliment consommé** :
```bash
add <fdcId> <quantité en grammes> <type de repas>
```
Types de repas disponibles :
- petit-dejeuner
- dejeuner
- diner
- collation

Exemple : `add 173944 100 dejeuner`

4. **Bilan nutritionnel** :
```bash
report
```
Affiche pour la journée en cours :
- Liste des repas consommés
- Total des calories
- Total des protéines
- Total des glucides
- Total des lipides
- Total des fibres

5. **Gestion des journées types** :
```bash
plan
```
Permet de :
- Créer une nouvelle journée type
- Consulter les journées types existantes
- Ajouter des repas à une journée type

6. **Quitter l'application** :
```bash
exit
```

### Exemple d'utilisation typique

1. Lancer le CLI et se connecter
2. Rechercher un aliment : `search apple`
3. Ajouter l'aliment au petit-déjeuner : `add 174988 125 petit-dejeuner`
4. Vérifier son bilan : `report`
5. Créer une journée type : `plan`
6. Quitter : `exit`

## Interface Web

L'interface web offre une expérience utilisateur moderne et intuitive avec les fonctionnalités suivantes :

### Pages principales

1. **Tableau de bord** (`/`)
   - Vue d'ensemble des macros du jour
   - Graphiques de progression
   - Ajout rapide d'aliments

2. **Journées types** (`/meal-plans`)
   - Création et gestion des journées types
   - Planification des repas
   - Modèles de repas réutilisables

3. **Profil** (`/profile`)
   - Gestion des informations utilisateur
   - Objectifs de macronutriments
   - Préférences personnelles

## Développement

### Structure du projet

```
macro-tracker/
├── cmd/
│   ├── cli/         # Application en ligne de commande
│   └── server/      # Serveur API
├── config/          # Configuration de l'application
├── frontend/        # Application React
├── internal/
│   ├── database/    # Couche d'accès aux données
│   └── fdc/        # Client API FoodData Central
└── docker-compose.yml
```

### Lancer en mode développement

1. Backend :
```bash
go run cmd/server/main.go
```

2. Frontend :
```bash
cd frontend
npm install
npm run dev
```