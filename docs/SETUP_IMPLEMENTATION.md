# Burndler Setup Flow Implementation

## Overview

This document describes the implementation of the initial setup flow for the Burndler system. When users first install and access the system, they can initialize it through a secure and systematic configuration process.

## Implemented Features

### Backend Components

#### 1. Database Model (`internal/models/setup.go`)
- `Setup` model: Tracks system setup state
- Stores setup completion status, admin email, company name, and system configuration
- Automatically records setup completion timestamp

#### 2. Setup Service (`internal/services/setup.go`)
- `SetupService`: Handles setup-related business logic
- Key features:
  - Check setup status (`CheckSetupStatus`)
  - Generate and validate setup tokens (`ValidateSetupToken`)
  - Create initial admin user (`CreateInitialAdmin`)
  - Complete setup process (`CompleteSetup`)

#### 3. Setup Handler (`internal/handlers/setup.go`)
- Provides REST API endpoints:
  - `GET /api/v1/setup/status` - Check setup status
  - `POST /api/v1/setup/init` - Initialize setup
  - `POST /api/v1/setup/admin` - Create admin account
  - `POST /api/v1/setup/complete` - Complete setup

#### 4. Setup Middleware (`internal/middleware/setup.go`)
- `SetupGuard`: Access control based on setup completion status
- `SetupCompleteGuard`: Blocks access to setup pages after completion

#### 5. Database Migration Enhancement
- Improved migration command handling (`cmd/api/main.go`)
- Auto-exit after running `go run cmd/api/main.go migrate`

### Frontend Components

#### 1. Setup Service (`src/services/setup.ts`)
- Service class for communicating with Setup APIs
- Handles API calls for setup status checks, admin creation, and setup completion

#### 2. Setup Hook (`src/hooks/useSetup.tsx`)
- Provides `SetupProvider` and `useSetup` hook
- Manages setup state and global state sharing through Context API

#### 3. Setup Page Components
- **SetupWizard** (`src/pages/SetupWizard.tsx`): Main setup wizard
- **SetupStatus** (`src/components/setup/SetupStatus.tsx`): System status verification
- **AdminSetup** (`src/components/setup/AdminSetup.tsx`): Admin account creation
- **SystemConfig** (`src/components/setup/SystemConfig.tsx`): System configuration
- **SetupComplete** (`src/components/setup/SetupComplete.tsx`): Setup completion

#### 4. Routing Integration (`src/App.tsx`)
- Provides `SetupProvider` to the entire application
- Protects all routes with `SetupGuard` component
- Adds `/setup` route

#### 5. Types (`src/types/setup.ts`)
- Defines TypeScript interfaces for setup-related functionality

## API Endpoints

### Setup Status
```http
GET /api/v1/setup/status

Response:
{
  "is_completed": false,
  "requires_setup": true,
  "admin_exists": false,
  "setup_token": "5c2d659ff669bef367a8269fa231e1e291a1d5165271fe14ca8821f96bf70a5b"
}
```

### Create Admin
```http
POST /api/v1/setup/admin
Content-Type: application/json

{
  "name": "Admin User",
  "email": "admin@burndler.com",
  "password": "password123"
}

Response:
{
  "user": {
    "id": 1,
    "email": "admin@burndler.com",
    "name": "Admin User",
    "role": "Admin",
    "active": true,
    "created_at": "2025-09-15T23:31:22.95462+09:00",
    "updated_at": "2025-09-15T23:31:22.95462+09:00"
  }
}
```

### Complete Setup
```http
POST /api/v1/setup/complete
Content-Type: application/json

{
  "company_name": "Burndler Inc",
  "system_settings": {
    "default_namespace": "burndler",
    "max_concurrent_builds": "3",
    "storage_retention_days": "30",
    "auto_cleanup_enabled": "true"
  }
}

Response:
{
  "message": "Setup completed successfully",
  "completed": true
}
```

## Setup Flow

### 1. Initial Access
- User first accesses the application
- `SetupGuard` checks setup status
- Redirects to `/setup` page if setup is required

### 2. System Status Check
- Verifies database connection status
- Checks for existing admin accounts
- Confirms setup completion status

### 3. Admin Account Creation
- Initial admin account creation form
- Input name, email, and password
- Password confirmation and validation

### 4. System Configuration
- Company name setup
- System default settings:
  - Default namespace
  - Maximum concurrent builds
  - Storage retention period
  - Auto-cleanup enablement
  - Notification email (optional)

### 5. Setup Completion
- Setup completion confirmation
- Redirect to dashboard or login page

## Security Considerations

### 1. Setup Tokens
- Generates 32-byte random tokens
- Valid only during setup process
- Automatically invalidated after setup completion

### 2. Access Control
- Before setup completion: Only setup pages accessible
- After setup completion: Regular authentication flow applies
- Prevents re-setup after admin account creation

### 3. Password Security
- Minimum 8 characters required
- bcrypt hashing applied
- Frontend password confirmation validation

## Test Results

### Integration Tests Passed
1. ✅ Database migration executed successfully
2. ✅ Setup Status API responding correctly
3. ✅ Admin creation API working properly
4. ✅ Setup completion API working properly
5. ✅ Login API working properly
6. ✅ Frontend TypeScript compilation successful
7. ✅ Frontend build successful

### API Response Validation
- Before setup: `requires_setup: true`, `admin_exists: false`
- After admin creation: Admin information properly returned
- After setup completion: `is_completed: true`, `requires_setup: false`

## Usage

### Development Environment Setup
1. Start database: `docker-compose -f compose/dev.compose.yaml up -d postgres`
2. Run migrations: `go run cmd/api/main.go migrate`
3. Start backend server: `go run cmd/api/main.go`
4. Start frontend server: `npm run dev`
5. Access `http://localhost:3000` in browser

### Setup Process
1. Automatically redirects to `/setup` page on first access
2. Check system status and click "Continue Setup"
3. Enter admin account information and create
4. Input company name and system settings
5. Navigate to login page after setup completion

## File Structure

```
backend/
├── cmd/api/main.go                     # Migration command handling
├── internal/
│   ├── models/setup.go                 # Setup model
│   ├── services/setup.go               # Setup service
│   ├── handlers/setup.go               # Setup API handlers
│   ├── middleware/setup.go             # Setup middleware
│   └── server/server.go                # Routing configuration

frontend/
├── src/
│   ├── services/setup.ts               # Setup API service
│   ├── hooks/useSetup.tsx              # Setup Context/Hook
│   ├── types/setup.ts                  # Setup type definitions
│   ├── pages/SetupWizard.tsx           # Main setup page
│   ├── components/
│   │   ├── SetupGuard.tsx              # Setup guard component
│   │   └── setup/
│   │       ├── SetupStatus.tsx         # Status check component
│   │       ├── AdminSetup.tsx          # Admin creation component
│   │       ├── SystemConfig.tsx        # System configuration component
│   │       └── SetupComplete.tsx       # Completion component
│   └── App.tsx                         # Routing configuration
```

This implementation provides a complete initial setup flow for the Burndler system, allowing users to initialize the system in a secure and intuitive manner.