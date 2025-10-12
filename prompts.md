# SocialQueue - Actionable Development Prompts

## 🚀 Immediate Action Items (Do These Now!)

### Prompt 1: Fix Database Schema
```
Task: Generate and run the migration to fix the users table schema mismatch.

Requirements:
1. Add username, first_name, last_name columns to users table
2. Make username unique and not null
3. Update SQLC queries to match new schema
4. Test with existing seed data

Generate:
- Migration SQL file
- Updated SQLC queries
- Test to verify migration works
```

### Prompt 2: Complete User Use Cases
```
Task: Implement the remaining user use cases in the application layer.

Create these use cases:
1. UpdateUser - update profile information
2. GetUser - retrieve user by ID
3. DeleteUser - soft delete user
4. VerifyEmail - verify email with token
5. ResetPassword - reset password with token

For each use case, include:
- Input/Output DTOs
- Validation logic
- Error handling
- Domain service calls
- Event publishing

File paths should be: backend/internal/application/user/{usecase}.go
```

### Prompt 3: Implement User Repository (PostgreSQL)
```
Task: Create PostgreSQL implementation of the user.Repository interface using SQLC.

Requirements:
1. Implement all methods from domain/user/repository.go interface
2. Use SQLC generated queries
3. Handle transactions properly
4. Map between domain entities and database models
5. Include proper error handling

File path: backend/internal/infrastructure/persistence/user_repository.go
Include: Unit tests with testify
```

## 📝 Week 1 Development Prompts

### Day 1-2: Team Management
```
Task: Implement team use cases and repository.

Create:
1. Application use cases:
   - CreateTeam
   - InviteMember  
   - UpdateTeamSettings
   - RemoveMember
   - UpdateMemberRole

2. Repository implementation:
   - PostgreSQL team repository using SQLC

3. HTTP handlers:
   - Team creation endpoint
   - Member invitation endpoint
   - Settings update endpoint

Include proper authorization checks (only admins can invite/remove).
```

### Day 3-4: Post Scheduling System
```
Task: Build the complete post scheduling system.

Implement:
1. Post use cases:
   - CreatePost (with media upload)
   - SchedulePost (with timezone handling)
   - UpdatePost
   - DeletePost  
   - ApprovePost
   - GetPostQueue

2. Scheduler service:
   - Queue management
   - Priority handling
   - Rate limit checking

3. API endpoints:
   - POST /api/posts
   - GET /api/posts/queue
   - PATCH /api/posts/:id
   - DELETE /api/posts/:id
```

### Day 5: Social Account Management
```
Task: Implement social account connection flow.

Build:
1. OAuth flow handlers:
   - Initiate OAuth redirect
   - Handle OAuth callback
   - Store encrypted tokens

2. Use cases:
   - ConnectSocialAccount
   - DisconnectSocialAccount
   - RefreshSocialTokens
   - ListConnectedAccounts

3. Token encryption:
   - AES encryption for tokens
   - Secure storage in database

File paths: 
- backend/internal/application/social/*.go
- backend/internal/adapters/oauth/*.go
```

## 🎨 Frontend Development Prompts

### Frontend Setup & Auth
```
Task: Set up Next.js frontend with authentication.

Create:
1. Authentication pages:
   - Login page with form validation
   - Signup page with password requirements
   - Email verification page
   - Password reset flow

2. Auth context/hooks:
   - useAuth hook with React Query
   - Protected route wrapper
   - Token refresh logic

3. Components:
   - Navigation with user menu
   - Loading states
   - Error boundaries

Use shadcn/ui components (Button, Input, Card, Form).
Use react-hook-form for forms and zod for validation.
```

### Dashboard Implementation
```
Task: Build the main dashboard with post composer.

Create:
1. Dashboard layout:
   - Sidebar navigation
   - Top header with user info
   - Main content area

2. Post composer:
   - Rich text editor (TipTap or Slate)
   - Image/video upload with preview
   - Platform selector (Twitter, LinkedIn, etc.)
   - Schedule picker with timezone
   - Character counter per platform

3. Post queue view:
   - Drag-and-drop reordering
   - Calendar view
   - List view with filters
   - Quick actions (edit, delete, reschedule)

Use Tailwind CSS for styling and shadcn/ui components.
```

## 🔧 Infrastructure Prompts

### Redis Cache Implementation
```
Task: Implement Redis caching layer.

Create:
1. Redis cache service implementing CacheService interface
2. Caching strategies for:
   - User sessions (1 hour TTL)
   - Team data (10 minutes TTL)
   - Post queue (5 minutes TTL)
   - Analytics data (30 minutes TTL)

3. Cache invalidation on:
   - User updates
   - Team changes
   - New posts scheduled

File path: backend/internal/infrastructure/services/redis_cache.go
```

### Worker System
```
Task: Build the background worker for post publishing.

Implement:
1. Worker main loop:
   - Poll for due posts every minute
   - Process posts in priority order
   - Respect rate limits per platform

2. Publishing logic:
   - Acquire lock on post
   - Call social adapter
   - Handle success/failure
   - Update post status
   - Record analytics

3. Retry mechanism:
   - Exponential backoff
   - Max 3 retries
   - Dead letter queue

File path: backend/cmd/worker/main.go
```

## 🧪 Testing Prompts

### Integration Test Suite
```
Task: Create comprehensive integration tests.

Write tests for:
1. Complete user journey:
   - Signup → Verify email → Login
   - Create team → Invite member
   - Connect social account
   - Create and schedule post
   - View analytics

2. API endpoint tests:
   - Authentication required
   - Input validation
   - Error responses
   - Success cases

3. Database tests:
   - Repository methods
   - Transactions
   - Concurrent access

Use testcontainers for Postgres and Redis.
```

## 📊 Analytics Implementation
```
Task: Build analytics collection and visualization.

Create:
1. Analytics collector:
   - Fetch metrics from platforms
   - Store in time-series format
   - Calculate engagement rates

2. Aggregation service:
   - Daily/weekly/monthly rollups
   - Per-post metrics
   - Per-platform metrics

3. Frontend charts:
   - Line chart for impressions over time
   - Bar chart for engagement by platform
   - Pie chart for audience demographics
   - Table for top performing posts

Use Recharts for visualization.
```

## 🚀 Production Deployment
```
Task: Prepare for production deployment.

Set up:
1. GitHub Actions CI/CD:
   - Run tests on PR
   - Build Docker images
   - Push to registry
   - Deploy to staging

2. Kubernetes manifests:
   - Deployment for API
   - Deployment for workers
   - Service definitions
   - Ingress with TLS

3. Monitoring:
   - Prometheus metrics
   - Grafana dashboards
   - Sentry error tracking
   - Uptime monitoring

4. Security:
   - Environment secrets
   - Database SSL
   - API rate limiting
   - CORS configuration
```

## 💡 Tips for Using These Prompts

1. **Always specify file paths** in your prompts
2. **Include test requirements** for each component
3. **Ask for error handling** explicitly
4. **Request documentation** inline with code
5. **Specify which patterns** to follow (Clean Architecture)

## Example Enhanced Prompt Format:
```
Task: [Specific task]
Context: [Current state of the project]
Requirements:
- [Requirement 1]
- [Requirement 2]
Generate:
- [File 1 with path]
- [File 2 with path]
- [Tests]
Include: Error handling, logging, and inline documentation
```