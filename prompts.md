# SocialQueue - UPDATED Complete Module Development Prompts

## 🎯 How to Use These Prompts

**Each prompt is:**
- ✅ Based on YOUR actual codebase
- ✅ Gives COMPLETE, working implementations
- ✅ Phase-ordered (do in sequence)
- ✅ Ready to copy-paste into Claude

**Don't modify the prompts.** Use them as-is for best results.

---

## 📦 PHASE 1: Complete Team Member Management

### Prompt: Team Member Management - Complete Implementation

```
I need to COMPLETE my team module by adding the 3 missing member management use cases.

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Following Clean Architecture

WHAT'S ALREADY WORKING:
- Team CRUD (CreateTeam, GetTeam, UpdateTeam, DeleteTeam, ListTeams) ✅
- Team repository at internal/infrastructure/persistence/team_repository.go ✅
- Member repository at internal/infrastructure/persistence/team_member_repository.go ✅
- Team handlers at internal/handlers/team_handler.go ✅
- Domain entities at internal/domain/team/ ✅
- All wired in container.go and router.go ✅

WHAT I NEED - 3 USE CASES:

1. backend/internal/application/team/invite_member.go
   ```
   type InviteMemberInput struct {
       TeamID    uuid.UUID
       Email     string
       Role      team.MemberRole  // use domain types
       InviterID uuid.UUID
   }
   
   type InviteMemberOutput struct {
       Member *MemberDTO
   }
   ```
   
   **Business Logic:**
   - Check inviter is admin or owner (use memberRepo.FindMember)
   - Validate email format
   - Check if user already invited/member
   - Check team seat limits (use team.CanAddMember())
   - Create pending member record
   - Send invitation email (use emailService.SendInvitationEmail)
   - Log action
   
   **Authorization:** Only admins/owners can invite
   **Output:** MemberDTO with status "pending"

2. backend/internal/application/team/remove_member.go
   ```
   type RemoveMemberInput struct {
       TeamID    uuid.UUID
       UserID    uuid.UUID  // member to remove
       RemoverID uuid.UUID  // who is removing
   }
   ```
   
   **Business Logic:**
   - Check remover is admin or owner
   - Cannot remove the team owner
   - Check if member is last admin (prevent)
   - Soft delete member record (memberRepo.RemoveMember)
   - Log action
   
   **Authorization:** Only admins/owners can remove members
   **Output:** void (error on failure)

3. backend/internal/application/team/update_member_role.go
   ```
   type UpdateMemberRoleInput struct {
       TeamID    uuid.UUID
       UserID    uuid.UUID
       NewRole   team.MemberRole
       UpdaterID uuid.UUID
   }
   
   type UpdateMemberRoleOutput struct {
       Member *MemberDTO
   }
   ```
   
   **Business Logic:**
   - Check updater is owner (only owners change roles)
   - Validate new role (owner, admin, editor, viewer)
   - Check if demoting last admin (prevent)
   - Update member role
   - Log action
   
   **Authorization:** Only owner can update roles
   **Output:** Updated MemberDTO

ALSO UPDATE THESE FILES:

4. backend/internal/handlers/team_handler.go
   Add 3 new handler methods:
   ```go
   func (h *TeamHandler) InviteMember(w http.ResponseWriter, r *http.Request)
   func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request)
   func (h *TeamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request)
   ```
   
   Routes:
   - POST /api/v2/teams/:id/members
   - DELETE /api/v2/teams/:id/members/:userId
   - PATCH /api/v2/teams/:id/members/:userId/role

5. backend/cmd/api/container.go
   Add to Container struct:
   ```go
   InviteMemberUC      *appTeam.InviteMemberUseCase
   RemoveMemberUC      *appTeam.RemoveMemberUseCase
   UpdateMemberRoleUC  *appTeam.UpdateMemberRoleUseCase
   ```
   
   Initialize in initializeUseCases():
   ```go
   c.InviteMemberUC = appTeam.NewInviteMemberUseCase(
       teamRepo, memberRepo, userRepo, c.EmailService, c.Logger)
   // ... same for other 2
   ```
   
   Pass to handler in initializeHandlers():
   ```go
   c.TeamHandler = handlers.NewTeamHandler(
       c.CreateTeamUC,
       c.GetTeamUC,
       c.UpdateTeamUC,
       c.DeleteTeamUC,
       c.ListTeamsUC,
       c.InviteMemberUC,      // NEW
       c.RemoveMemberUC,      // NEW
       c.UpdateMemberRoleUC,  // NEW
   )
   ```

6. backend/cmd/api/router.go
   Under Team routes section, add:
   ```go
   r.Route("/teams", func(r chi.Router) {
       // ... existing routes ...
       
       // Member management
       r.Post("/{id}/members", container.TeamHandler.InviteMember)
       r.Delete("/{id}/members/{userId}", container.TeamHandler.RemoveMember)
       r.Patch("/{id}/members/{userId}/role", container.TeamHandler.UpdateMemberRole)
   })
   ```

REQUIREMENTS:
- Follow patterns from create_team.go, update_team.go, delete_team.go
- Use existing domain types (team.MemberRole, etc.)
- Use MapMemberToDTO helper function (create if doesn't exist)
- Include proper error handling
- Add logging for all operations
- Use transactions where needed
- Authorization checks before business logic
- Production-ready code

DELIVERABLES:
- 3 use case files (invite, remove, update role)
- Updated team_handler.go (add 3 methods)
- Updated container.go (wire dependencies)
- Updated router.go (register routes)
- Complete, runnable, testable code
- No TODOs or placeholders

Generate all files with full implementations.
```

---

## 📦 PHASE 2: Complete Post Scheduling Module

### Prompt: Post Module - Complete Implementation

```
I need a COMPLETE post scheduling module for SocialQueue (Buffer clone).

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Following Clean Architecture
- Reference my User and Team modules for patterns

WHAT ALREADY EXISTS:
- ✅ Domain entities at internal/domain/post/
- ✅ Repository interface defined
- ✅ SQLC queries at sql/posts.sql
- ✅ Database schema (scheduled_posts, post_attachments, posts tables)
- ✅ Team and User modules working

WHAT TO CREATE:

1. backend/internal/infrastructure/persistence/post_repository.go
   Implement domain/post/repository.go interface:
   - Create, Update, Delete, FindByID
   - FindByTeamID, FindByUserID
   - FindDuePosts, FindScheduled, FindPublished
   - CountByTeamID, GetTeamPostStats
   - Handle post attachments (media URLs array)
   - Map between domain entities and SQLC models

2. backend/internal/application/post/ (7 use cases):

   a) create_draft.go
      ```
      type CreateDraftInput struct {
          TeamID      uuid.UUID
          AuthorID    uuid.UUID
          Content     string
          Platforms   []post.Platform  // twitter, linkedin, facebook
          Attachments []string         // media URLs
      }
      
      type CreateDraftOutput struct {
          Post *PostDTO
      }
      ```
      - Validate author is team member
      - Validate content not empty
      - Validate platforms are supported
      - Create draft post (status: draft)
      - Save attachments
      - Output: PostDTO
   
   b) schedule_post.go
      ```
      type SchedulePostInput struct {
          PostID      uuid.UUID
          UserID      uuid.UUID
          ScheduledAt time.Time
          Timezone    string
      }
      ```
      - Check authorization (author or admin)
      - Validate future time
      - Check rate limits (team posting limits)
      - Update post with schedule
      - Output: PostDTO with schedule
   
   c) update_post.go
      ```
      type UpdatePostInput struct {
          PostID      uuid.UUID
          UserID      uuid.UUID
          Content     *string
          Platforms   []post.Platform
          Attachments []string
      }
      ```
      - Check authorization
      - Validate not published yet
      - Update fields
      - Output: Updated PostDTO
   
   d) delete_post.go
      ```
      type DeletePostInput struct {
          PostID uuid.UUID
          UserID uuid.UUID
      }
      ```
      - Check authorization
      - Cancel if scheduled
      - Soft delete
   
   e) get_post.go
      ```
      type GetPostInput struct {
          PostID uuid.UUID
          UserID uuid.UUID
      }
      ```
      - Check user is team member
      - Return full post details
   
   f) list_posts.go
      ```
      type ListPostsInput struct {
          TeamID uuid.UUID
          UserID uuid.UUID
          Status *post.Status  // optional filter
          Offset int
          Limit  int
      }
      ```
      - Check user is team member
      - Return paginated list
   
   g) publish_now.go
      ```
      type PublishNowInput struct {
          PostID uuid.UUID
          UserID uuid.UUID
      }
      ```
      - Check authorization
      - Validate ready to publish
      - Mark as queued (for worker)
      - Output: PostDTO

3. backend/internal/handlers/post_handler.go
   ```go
   type PostHandler struct {
       createDraftUC   *post.CreateDraftUseCase
       schedulePostUC  *post.SchedulePostUseCase
       updatePostUC    *post.UpdatePostUseCase
       deletePostUC    *post.DeletePostUseCase
       getPostUC       *post.GetPostUseCase
       listPostsUC     *post.ListPostsUseCase
       publishNowUC    *post.PublishNowUseCase
   }
   ```
   
   Routes:
   - POST /api/v2/posts - CreateDraft
   - GET /api/v2/posts/:id - GetPost
   - PUT /api/v2/posts/:id - UpdatePost
   - DELETE /api/v2/posts/:id - DeletePost
   - POST /api/v2/posts/:id/schedule - SchedulePost
   - POST /api/v2/posts/:id/publish - PublishNow
   - GET /api/v2/teams/:teamId/posts - ListPosts

4. backend/cmd/api/container.go
   Add:
   - PostRepository
   - All 7 use cases
   - PostHandler
   Wire dependencies

5. backend/cmd/api/router.go
   Register all post routes under /api/v2

REQUIREMENTS:
- Use domain/post entities (Post, Status, Platform, Priority)
- Handle timezones correctly (store as UTC)
- Support multiple platforms per post
- Character count validation per platform (Twitter: 280, LinkedIn: 3000)
- Media URL validation
- Schedule validation (only future dates)
- Proper authorization checks
- Comprehensive error handling
- Logging for all operations
- Production-ready

DELIVERABLES:
- post_repository.go (complete implementation)
- 7 use case files
- post_handler.go
- Updated container.go and router.go
- Complete DTOs with json tags
- No placeholders or TODOs

Generate all files with full implementations.
```

---

## 📦 PHASE 3: Complete Social OAuth Module

### Prompt: Social OAuth - Complete Implementation

```
I need a COMPLETE social media OAuth and publishing module for SocialQueue.

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Need OAuth for: Twitter (X), LinkedIn, Facebook
- Following Clean Architecture

WHAT ALREADY EXISTS:
- ✅ Domain entities at internal/domain/social/
- ✅ Repository interface defined
- ✅ Database schema (social_accounts, social_tokens)
- ✅ Some adapter skeletons at internal/adapters/social/

WHAT TO CREATE:

1. backend/internal/adapters/social/adapter.go
   Define the adapter interface:
   ```go
   type Adapter interface {
       // OAuth
       GetAuthURL(state string, scopes []string) string
       ExchangeCode(ctx context.Context, code string) (*Token, error)
       RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
       
       // Publishing
       PublishPost(ctx context.Context, token *Token, content *PostContent) (*PublishResult, error)
       
       // Analytics
       GetPostAnalytics(ctx context.Context, token *Token, postID string) (*Analytics, error)
       
       // Validation
       ValidateToken(ctx context.Context, token *Token) (bool, error)
   }
   ```

2. Platform adapters (3 complete implementations):

   a) backend/internal/adapters/social/twitter/
      - client.go: OAuth 2.0 PKCE flow
      - publisher.go: Tweet creation, media upload
      - Use Twitter API v2
      - Handle rate limits
   
   b) backend/internal/adapters/social/linkedin/
      - client.go: LinkedIn OAuth 2.0
      - publisher.go: Post creation (text + images)
      - Use LinkedIn Marketing API
      - Handle rate limits
   
   c) backend/internal/adapters/social/facebook/
      - client.go: Facebook OAuth
      - publisher.go: Page post creation
      - Use Graph API v18
      - Handle rate limits

3. backend/internal/infrastructure/persistence/social_repository.go
   Implement domain/social/repository.go:
   - CRUD for social accounts
   - Store encrypted OAuth tokens
   - Token refresh logic
   - Map domain entities to DB models

4. backend/internal/infrastructure/services/encryption.go
   Token encryption service:
   - AES-256-GCM encryption
   - Encrypt(plaintext string) (string, error)
   - Decrypt(ciphertext string) (string, error)
   - Use environment variable for key

5. backend/internal/application/social/ (6 use cases):

   a) connect_account.go
      - OAuth code exchange
      - Store encrypted tokens
      - Output: SocialAccountDTO
   
   b) disconnect_account.go
      - Revoke tokens
      - Soft delete account
   
   c) refresh_tokens.go
      - Refresh before expiry
      - Update storage
   
   d) list_accounts.go
      - Get team's connected accounts
      - Show health status
   
   e) publish_post.go
      - Get post content
      - Get social account + tokens
      - Call platform adapter
      - Store platform post ID
      - Output: PublishResult
   
   f) get_analytics.go
      - Fetch from platform
      - Cache results
      - Output: Analytics

6. backend/internal/handlers/social_handler.go
   Routes:
   - GET /api/v2/social/auth/:platform - Get OAuth URL
   - GET /api/v2/social/auth/:platform/callback - OAuth callback
   - POST /api/v2/social/accounts - Connect account
   - GET /api/v2/teams/:teamId/social/accounts - List accounts
   - DELETE /api/v2/social/accounts/:id - Disconnect
   - POST /api/v2/social/accounts/:id/refresh - Refresh tokens

7. backend/cmd/api/container.go & router.go
   Wire all dependencies

8. Environment variables (.env):
   ```
   TWITTER_CLIENT_ID=
   TWITTER_CLIENT_SECRET=
   LINKEDIN_CLIENT_ID=
   LINKEDIN_CLIENT_SECRET=
   FACEBOOK_APP_ID=
   FACEBOOK_APP_SECRET=
   ENCRYPTION_KEY=  # 32-byte key for AES-256
   ```

REQUIREMENTS:
- Secure token storage (encrypted at rest)
- Token refresh before expiry (auto-refresh)
- Rate limiting per platform
- Retry logic with exponential backoff
- Error handling for API failures
- Platform-specific character limits
- Media upload support
- No hardcoded API keys (env vars only)
- Production-ready

DELIVERABLES:
- Complete adapter implementations (Twitter, LinkedIn, Facebook)
- social_repository.go
- encryption.go service
- 6 use case files
- social_handler.go
- Updated container.go and router.go
- Environment variable template
- No placeholders

Generate all files with full, production-ready implementations.
```

---

## 📦 PHASE 4: Worker System

### Prompt: Worker System - Complete Implementation

```
I need a COMPLETE background worker system for processing scheduled posts in SocialQueue.

PROJECT CONTEXT:
- Backend: Go
- Module path: github.com/techappsUT/social-queue
- Need to process scheduled posts automatically
- Use Redis for queues and distributed locks

WHAT TO CREATE:

1. backend/internal/infrastructure/services/redis_cache.go
   Replace in-memory cache with real Redis:
   - Implement common.CacheService interface
   - Connection pooling (go-redis/redis)
   - Methods: Get, Set, Delete, Exists
   - Distributed locking: Lock(key, ttl), Unlock(key)
   - TTL support

2. backend/internal/infrastructure/services/worker_queue.go
   Job queue service:
   - Enqueue(ctx, jobType, payload) error
   - Dequeue(ctx, jobType) (*Job, error)
   - MarkComplete(ctx, jobID) error
   - MarkFailed(ctx, jobID, reason) error
   - Retry logic (exponential backoff)
   - Dead letter queue for permanent failures
   - Use Redis lists/streams

3. backend/cmd/worker/main.go
   Worker binary entry point:
   ```go
   func main() {
       // Load config
       // Initialize database connection
       // Initialize Redis
       // Initialize repositories
       // Initialize social adapters
       // Start job processors (goroutines)
       // Graceful shutdown (SIGTERM/SIGINT)
       // Wait for all jobs to finish
   }
   ```

4. backend/cmd/worker/jobs/publish_post.go
   Post publishing job processor:
   ```go
   type PublishPostProcessor struct {
       postRepo    post.Repository
       socialRepo  social.Repository
       adapters    map[social.Platform]social.Adapter
       queue       services.WorkerQueue
       logger      common.Logger
   }
   
   func (p *PublishPostProcessor) Run(ctx context.Context) {
       // Poll every 30 seconds
       for {
           // Query due posts (scheduled_at <= now, status=scheduled)
           posts := p.postRepo.FindDuePosts(ctx, time.Now())
           
           for _, post := range posts {
               // Acquire distributed lock
               if !p.queue.Lock(post.ID().String(), 5*time.Minute) {
                   continue // Skip if locked
               }
               
               // Get social accounts for post
               accounts := p.socialRepo.FindByIDs(ctx, post.PlatformAccountIDs())
               
               // Publish to each platform
               for _, account := range accounts {
                   adapter := p.adapters[account.Platform()]
                   result, err := adapter.PublishPost(ctx, account.Token(), post.Content())
                   
                   if err != nil {
                       // Retry up to 3 times
                       p.queue.MarkFailed(ctx, post.ID().String(), err.Error())
                       continue
                   }
                   
                   // Store platform post ID
                   post.MarkPublished(result.PlatformPostID)
               }
               
               // Update post status
               p.postRepo.Update(ctx, post)
               
               // Release lock
               p.queue.Unlock(post.ID().String())
           }
           
           time.Sleep(30 * time.Second)
       }
   }
   ```

5. backend/cmd/worker/jobs/fetch_analytics.go
   Analytics fetching job:
   - Query published posts (24hrs+ old, no recent analytics)
   - Fetch metrics from each platform
   - Store in database
   - Run every 6 hours

6. backend/cmd/worker/jobs/cleanup.go
   Maintenance jobs:
   - Delete old draft posts (30+ days)
   - Archive old analytics (1 year+)
   - Clean up dead letter queue
   - Run daily at 2 AM

7. Dockerfile.worker
   ```dockerfile
   FROM golang:1.21-alpine AS builder
   WORKDIR /app
   COPY go.* ./
   RUN go mod download
   COPY . .
   RUN go build -o worker ./cmd/worker
   
   FROM alpine:latest
   RUN apk --no-cache add ca-certificates
   COPY --from=builder /app/worker /worker
   CMD ["/worker"]
   ```

8. Update docker-compose.yml
   Add worker service:
   ```yaml
   worker:
     build:
       context: ./backend
       dockerfile: Dockerfile.worker
     depends_on:
       - postgres
       - redis
     environment:
       - DATABASE_URL=${DATABASE_URL}
       - REDIS_URL=redis://redis:6379
       - TWITTER_CLIENT_ID=${TWITTER_CLIENT_ID}
       - LINKEDIN_CLIENT_ID=${LINKEDIN_CLIENT_ID}
       - FACEBOOK_APP_ID=${FACEBOOK_APP_ID}
     restart: unless-stopped
   ```

REQUIREMENTS:
- Idempotent job processing (handle duplicates)
- Distributed locks (prevent duplicate processing)
- At-least-once delivery semantics
- Graceful shutdown (finish current jobs)
- Health check endpoint (for monitoring)
- Structured logging (JSON logs)
- Error tracking (capture stack traces)
- Metrics (jobs processed, failures, latency)
- No race conditions
- Production-ready

DELIVERABLES:
- redis_cache.go (real Redis implementation)
- worker_queue.go (job queue service)
- worker/main.go (entry point)
- 3 job processors (publish, analytics, cleanup)
- Dockerfile.worker
- Updated docker-compose.yml
- Environment variable template
- No placeholders

Generate all files with full implementations.
```

---

## 📦 PHASE 5: Frontend

### Prompt: Frontend - Complete Implementation

```
I need a COMPLETE Next.js 15 frontend for SocialQueue (Buffer clone).

PROJECT CONTEXT:
- Framework: Next.js 15 with App Router
- TypeScript strict mode
- Styling: Tailwind CSS + shadcn/ui
- Forms: react-hook-form + zod
- State: React Query
- Backend API: http://localhost:8000/api/v2

WHAT ALREADY EXISTS:
- ✅ Next.js 15 setup
- ✅ Tailwind + shadcn/ui configured
- ✅ Theme provider
- ✅ Some auth component skeletons

WHAT TO CREATE:

1. Authentication Pages (src/app/(auth)/)

   a) login/page.tsx
      - Email + password form
      - Form validation (zod schema)
      - Error handling
      - "Remember me" checkbox
      - Link to signup
      - Redirect to /dashboard after login
   
   b) signup/page.tsx
      - Fields: email, username, password, firstName, lastName
      - Password strength indicator
      - Form validation
      - Call POST /api/v2/auth/signup
      - Auto-login after signup
      - Redirect to /dashboard
   
   c) verify-email/page.tsx
      - Show "Check your email" message
      - Resend verification button
      - Auto-refresh status

2. Dashboard Layout (src/app/(dashboard)/)

   a) layout.tsx
      - Sidebar navigation (Dashboard, Compose, Queue, Accounts, Analytics)
      - Top header (team selector, user menu)
      - Protected route wrapper (AuthGuard)
      - Responsive (mobile drawer)
   
   b) dashboard/page.tsx
      - Overview cards (posts scheduled, published, impressions)
      - Recent posts table
      - Quick actions (Compose Post, Connect Account)
      - Loading skeletons

3. Post Composer (src/app/(dashboard)/compose/page.tsx)
   - Rich text editor (Tiptap or textarea with formatting)
   - Character counter per platform
   - Platform selector (checkboxes: Twitter, LinkedIn, Facebook)
   - Media upload (drag & drop, image preview)
   - Schedule picker:
     * Date picker (react-day-picker)
     * Time picker
     * Timezone selector
   - Save as draft button
   - Schedule button
   - Publish now button

4. Post Queue (src/app/(dashboard)/queue/page.tsx)
   - Calendar view (react-big-calendar)
   - List view toggle
   - Filters (status, platform, date range)
   - Post cards with:
     * Content preview
     * Scheduled time
     * Platform icons
     * Quick actions (edit, delete, reschedule)
   - Drag-and-drop reordering
   - Loading states

5. Social Accounts (src/app/(dashboard)/accounts/page.tsx)
   - Connected accounts grid
   - Account cards showing:
     * Platform logo
     * Account name/handle
     * Connection status
     * Health indicator
     * Disconnect button
   - Connect new account buttons
   - OAuth flow handling (popup or redirect)
   - Loading states

6. Analytics (src/app/(dashboard)/analytics/page.tsx)
   - Date range selector
   - Summary cards (total posts, impressions, engagement)
   - Charts (recharts):
     * Line chart: impressions over time
     * Bar chart: engagement by platform
     * Pie chart: post distribution
   - Top performing posts table
   - Export CSV button

7. Shared Components (src/components/)

   - ui/ (from shadcn/ui):
     * button, input, card, form, dialog, dropdown, etc.
   
   - layout/
     * sidebar.tsx - Navigation sidebar
     * header.tsx - Top bar with user menu
     * footer.tsx - Optional footer
   
   - auth/
     * login-form.tsx - Form component
     * signup-form.tsx - Form component
     * AuthGuard.tsx - Route protection HOC
   
   - posts/
     * post-card.tsx - Post display card
     * post-composer.tsx - Reusable composer
     * schedule-picker.tsx - Date/time picker
     * post-calendar.tsx - Calendar view
   
   - social/
     * account-card.tsx - Social account card
     * oauth-button.tsx - OAuth connect button
     * platform-icon.tsx - Platform logos

8. API Integration (src/lib/)

   a) api.ts - API client
      ```typescript
      import axios from 'axios';
      
      const api = axios.create({
        baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v2',
      });
      
      // Request interceptor (add auth token)
      api.interceptors.request.use((config) => {
        const token = localStorage.getItem('accessToken');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      });
      
      // Response interceptor (handle 401)
      api.interceptors.response.use(
        (response) => response,
        async (error) => {
          if (error.response?.status === 401) {
            // Logout user
            localStorage.removeItem('accessToken');
            window.location.href = '/login';
          }
          return Promise.reject(error);
        }
      );
      
      export default api;
      ```
   
   b) auth.ts - Auth utilities
      - login(email, password)
      - signup(data)
      - logout()
      - refreshToken()
      - getUser()

9. React Query Hooks (src/hooks/)

   a) use-auth.ts
      - useLogin()
      - useSignup()
      - useLogout()
      - useUser()
   
   b) use-posts.ts
      - useCreatePost()
      - useUpdatePost()
      - useDeletePost()
      - usePosts(teamId, filters)
      - usePost(postId)
   
   c) use-accounts.ts
      - useConnectAccount()
      - useDisconnectAccount()
      - useAccounts(teamId)
      - useRefreshToken(accountId)
   
   d) use-analytics.ts
      - useAnalytics(teamId, dateRange)

10. Authentication Context (src/components/providers/)

    a) auth-provider.tsx
       - User state management
       - Auto token refresh
       - Login/logout handlers
       - Protect routes

11. Environment Variables (.env.local)
    ```
    NEXT_PUBLIC_API_URL=http://localhost:8000/api/v2
    NEXT_PUBLIC_APP_URL=http://localhost:3000
    ```

REQUIREMENTS:
- Full TypeScript with strict mode
- All forms use react-hook-form + zod validation
- Responsive design (mobile-first)
- Dark mode support (next-themes)
- Loading states everywhere
- Error boundaries
- Toast notifications (sonner)
- Optimistic updates where applicable
- Accessible (ARIA labels, keyboard navigation)
- SEO-friendly (metadata)
- Production-ready

DELIVERABLES:
- All pages (auth, dashboard, compose, queue, accounts, analytics)
- All shared components
- API client + hooks
- Auth provider + guards
- Environment variables
- Complete, working, production-ready code
- No placeholders or TODOs

Generate all files with full implementations.
```

---

## 🎯 USAGE INSTRUCTIONS

### How to Use These Prompts:

1. **Do them in order** (Phase 1 → Phase 5)
2. **Copy entire prompt** (don't modify)
3. **Paste into Claude** or another AI
4. **Get complete module**
5. **Test immediately**:
   ```bash
   cd backend
   go build ./...
   make test
   make run
   ```
6. **Move to next phase** when tests pass

### After Each Phase:

```bash
# Verify compilation
cd backend && go build ./...

# Run tests
make test

# Start server
make run

# Test endpoints
curl http://localhost:8000/api/v2/teams
```

---

**Last Updated**: October 12, 2025  
**Status**: Ready to use sequentially