# SocialQueue - Complete Module Development Prompts

## 🎯 How to Use These Prompts

Each prompt gives you a **COMPLETE MODULE** that you can build, test, and verify independently.

**DON'T** ask for "step-by-step guides" or partial implementations.
**DO** use these prompts to get complete, working modules.

---

## 📦 MODULE 1: Complete Team Management

### Prompt: Team Module - Complete Implementation

```
I need a COMPLETE team management module for my SocialQueue application using Clean Architecture.

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Following Clean Architecture (Domain → Application → Infrastructure → Presentation)
- Reference: Look at my existing User module for patterns

WHAT TO CREATE:

1. SQLC QUERIES (backend/sql/teams.sql):
   - CreateTeam
   - GetTeam (by ID)
   - GetTeamsByUserID
   - UpdateTeam
   - DeleteTeam (soft delete)
   - AddTeamMember
   - RemoveTeamMember
   - UpdateMemberRole
   - GetTeamMembers
   - GetTeamMemberByUserID
   - CountTeams
   - CountTeamMembers

2. TEAM REPOSITORY (backend/internal/infrastructure/persistence/team_repository.go):
   - Implement ALL methods from domain/team/repository.go interface
   - Use SQLC generated queries
   - Handle transactions for member operations
   - Map between domain entities and database models
   - Include proper error handling

3. TEAM USE CASES (backend/internal/application/team/):
   Create 8 complete use cases:
   
   a) create_team.go
      - Input: name, description, plan, ownerID
      - Validation: name required, owner exists
      - Business logic: create team, add owner as admin
      - Output: TeamDTO with members
   
   b) get_team.go
      - Input: teamID, userID (for authorization)
      - Validation: user is team member
      - Output: TeamDTO with members
   
   c) update_team.go
      - Input: teamID, name, description, settings
      - Authorization: only admins can update
      - Output: Updated TeamDTO
   
   d) delete_team.go
      - Input: teamID, userID
      - Authorization: only owner can delete
      - Business logic: soft delete, remove all members
   
   e) invite_member.go
      - Input: teamID, email, role, inviterID
      - Authorization: only admins can invite
      - Business logic: check seat limits, send invitation email
      - Output: MemberDTO
   
   f) remove_member.go
      - Input: teamID, userID, removerID
      - Authorization: admins can remove, owner cannot be removed
      - Business logic: check if last admin
   
   g) update_member_role.go
      - Input: teamID, userID, newRole, updaterID
      - Authorization: only owner can change roles
      - Business logic: validate role, prevent removing last admin
   
   h) list_teams.go
      - Input: userID, pagination
      - Output: List of teams user belongs to

4. TEAM HANDLER (backend/internal/handlers/team_handler.go):
   - POST /api/v2/teams - CreateTeam
   - GET /api/v2/teams/:id - GetTeam
   - PUT /api/v2/teams/:id - UpdateTeam
   - DELETE /api/v2/teams/:id - DeleteTeam
   - POST /api/v2/teams/:id/members - InviteMember
   - DELETE /api/v2/teams/:id/members/:userId - RemoveMember
   - PATCH /api/v2/teams/:id/members/:userId/role - UpdateMemberRole
   - GET /api/v2/teams - ListTeams

5. UPDATE CONTAINER (backend/cmd/api/container.go):
   - Add TeamRepository
   - Add all 8 use cases
   - Add TeamHandler
   - Wire dependencies

6. UPDATE ROUTER (backend/cmd/api/router.go):
   - Register all team routes
   - Apply auth middleware
   - Add admin checks where needed

7. INTEGRATION TESTS (backend/tests/integration/team_test.go):
   - Test complete team lifecycle
   - Test member management
   - Test authorization checks
   - Test error cases

REQUIREMENTS:
- Use existing domain/team entities (they're already defined)
- Follow same patterns as domain/user implementation
- All DTOs should have json tags
- Include comprehensive error handling
- Add logging for important operations
- Use transactions where needed
- Include inline documentation
- Make it production-ready

GENERATE:
- All 7 files listed above
- Runnable, testable code
- No TODOs or placeholders
- Complete implementation
```

---

## 📦 MODULE 2: Complete Post Scheduling

### Prompt: Post Module - Complete Implementation

```
I need a COMPLETE post scheduling module for my SocialQueue application using Clean Architecture.

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Following Clean Architecture
- Reference: My User and Team modules for patterns

WHAT TO CREATE:

1. SQLC QUERIES (backend/sql/posts.sql):
   - CreatePost
   - GetPost (by ID)
   - GetPostsByTeamID
   - GetPostsByUserID
   - GetScheduledPosts (due for publishing)
   - GetPostQueue (scheduled posts for a team)
   - UpdatePost
   - UpdatePostStatus
   - DeletePost (soft delete)
   - CountPosts (by team, by status)
   - GetPostsByDateRange

2. POST REPOSITORY (backend/internal/infrastructure/persistence/post_repository.go):
   - Implement ALL methods from domain/post/repository.go interface
   - Use SQLC generated queries
   - Handle post attachments (media URLs)
   - Map between domain entities and database models

3. POST USE CASES (backend/internal/application/post/):
   Create 7 complete use cases:
   
   a) create_draft.go
      - Input: content, platforms[], attachments[], teamID, authorID
      - Validation: content not empty, valid platforms, team member
      - Business logic: create draft, save attachments
      - Output: PostDTO
   
   b) schedule_post.go
      - Input: postID, scheduledAt, timezone
      - Authorization: author or team admin
      - Business logic: validate future time, check rate limits
      - Output: PostDTO with schedule
   
   c) update_post.go
      - Input: postID, content, platforms, attachments
      - Authorization: author or admin
      - Validation: not published yet
      - Output: Updated PostDTO
   
   d) delete_post.go
      - Input: postID, userID
      - Authorization: author or admin
      - Business logic: cancel if scheduled
   
   e) get_post.go
      - Input: postID, userID
      - Authorization: team member
      - Output: PostDTO with full details
   
   f) list_posts.go
      - Input: teamID, status filter, pagination
      - Authorization: team member
      - Output: List of PostDTOs
   
   g) publish_now.go
      - Input: postID, userID
      - Authorization: author or admin
      - Business logic: validate ready to publish, mark as queued
      - Output: PostDTO

4. POST HANDLER (backend/internal/handlers/post_handler.go):
   - POST /api/v2/posts - CreateDraft
   - GET /api/v2/posts/:id - GetPost
   - PUT /api/v2/posts/:id - UpdatePost
   - DELETE /api/v2/posts/:id - DeletePost
   - POST /api/v2/posts/:id/schedule - SchedulePost
   - POST /api/v2/posts/:id/publish - PublishNow
   - GET /api/v2/teams/:teamId/posts - ListPosts
   - GET /api/v2/teams/:teamId/queue - GetPostQueue

5. UPDATE CONTAINER & ROUTER:
   - Add PostRepository
   - Add all 7 use cases
   - Add PostHandler
   - Register routes

6. INTEGRATION TESTS:
   - Test post lifecycle (draft → schedule → publish)
   - Test permissions
   - Test date/time handling
   - Test media attachments

REQUIREMENTS:
- Use domain/post entities
- Handle timezones correctly
- Support multiple platforms per post
- Character count per platform
- Media URL validation
- Schedule validation (future dates only)
- Production-ready

GENERATE:
- All files complete and working
- No placeholders
- Full error handling
```

---

## 📦 MODULE 3: Complete Social OAuth & Publishing

### Prompt: Social Module - Complete Implementation

```
I need a COMPLETE social media OAuth and publishing module for SocialQueue using Clean Architecture.

PROJECT CONTEXT:
- Backend: Go with Chi router, PostgreSQL, SQLC
- Module path: github.com/techappsUT/social-queue
- Need OAuth for: Twitter (X), LinkedIn, Facebook
- Following Clean Architecture

WHAT TO CREATE:

1. SOCIAL ADAPTER INTERFACE (backend/internal/adapters/social/adapter.go):
   ```go
   type Adapter interface {
       // OAuth
       GetAuthURL(state string, scopes []string) string
       ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
       RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error)
       
       // Publishing
       PublishPost(ctx context.Context, token *OAuthToken, post *Post) (*PublishResult, error)
       DeletePost(ctx context.Context, token *OAuthToken, postID string) error
       
       // Analytics
       GetPostAnalytics(ctx context.Context, token *OAuthToken, postID string) (*Analytics, error)
       
       // Validation
       ValidateToken(ctx context.Context, token *OAuthToken) (bool, error)
   }
   ```

2. PLATFORM ADAPTERS:
   
   a) Twitter Adapter (backend/internal/adapters/social/twitter/):
      - client.go: OAuth 2.0 PKCE flow implementation
      - publisher.go: Tweet creation, media upload
      - Use Twitter API v2
   
   b) LinkedIn Adapter (backend/internal/adapters/social/linkedin/):
      - client.go: LinkedIn OAuth 2.0
      - publisher.go: Post creation (text + images)
      - Use LinkedIn Marketing API
   
   c) Facebook Adapter (backend/internal/adapters/social/facebook/):
      - client.go: Facebook OAuth
      - publisher.go: Page post creation
      - Use Graph API v18

3. SQLC QUERIES (backend/sql/social.sql):
   - CreateSocialAccount
   - GetSocialAccount
   - GetSocialAccountsByTeamID
   - GetSocialAccountsByPlatform
   - UpdateSocialAccount
   - DeleteSocialAccount
   - StoreTokens (encrypted)
   - GetTokens
   - CountAccountsByPlatform

4. SOCIAL REPOSITORY (backend/internal/infrastructure/persistence/social_repository.go):
   - Implement domain/social/repository.go interface
   - Encrypt/decrypt OAuth tokens
   - Handle token refresh

5. SOCIAL USE CASES (backend/internal/application/social/):
   
   a) connect_account.go
      - Input: teamID, platform, authCode, state
      - Business logic: exchange code, store encrypted tokens
      - Output: SocialAccountDTO
   
   b) disconnect_account.go
      - Input: accountID, userID
      - Authorization: team admin
      - Business logic: revoke tokens, delete account
   
   c) refresh_tokens.go
      - Input: accountID
      - Business logic: refresh OAuth tokens, update storage
   
   d) list_accounts.go
      - Input: teamID
      - Output: List of connected accounts
   
   e) publish_post.go
      - Input: accountID, postID
      - Business logic: get post, get account, call adapter.PublishPost
      - Output: PublishResult
   
   f) get_analytics.go
      - Input: accountID, postID, platformPostID
      - Business logic: fetch from platform, cache results

6. SOCIAL HANDLER (backend/internal/handlers/social_handler.go):
   - GET /api/v2/social/auth/:platform - Initiate OAuth
   - GET /api/v2/social/auth/:platform/callback - OAuth callback
   - POST /api/v2/social/accounts - Connect account
   - GET /api/v2/teams/:teamId/social/accounts - List accounts
   - DELETE /api/v2/social/accounts/:id - Disconnect
   - POST /api/v2/social/accounts/:id/refresh - Refresh tokens
   - POST /api/v2/posts/:postId/publish - Publish post
   - GET /api/v2/posts/:postId/analytics - Get analytics

7. TOKEN ENCRYPTION SERVICE (backend/internal/infrastructure/services/encryption.go):
   - AES-256 encryption for OAuth tokens
   - Secure key management
   - Encrypt/Decrypt functions

8. UPDATE CONTAINER & ROUTER

9. INTEGRATION TESTS:
   - Test OAuth flow (mock external APIs)
   - Test token refresh
   - Test publishing
   - Test multiple platforms

REQUIREMENTS:
- Production OAuth credentials handling
- Secure token storage (encrypted)
- Token refresh before expiry
- Rate limiting per platform
- Error handling for API failures
- Retry logic with exponential backoff
- Platform-specific character limits
- Media upload support

GENERATE:
- Complete implementation for all 3 platforms
- No API keys hardcoded (use env vars)
- Full error handling
- Production-ready
```

---

## 📦 MODULE 4: Complete Worker System

### Prompt: Worker System - Complete Implementation

```
I need a COMPLETE background worker system for processing scheduled posts in SocialQueue.

PROJECT CONTEXT:
- Backend: Go
- Module path: github.com/techappsUT/social-queue
- Need to process scheduled posts automatically
- Use Redis for queue/locks

WHAT TO CREATE:

1. REDIS CACHE SERVICE (backend/internal/infrastructure/services/redis_cache.go):
   - Implement common.CacheService interface
   - Connection pooling
   - Methods: Get, Set, Delete, Lock, Unlock
   - TTL support

2. WORKER QUEUE SERVICE (backend/internal/infrastructure/services/worker_queue.go):
   - Enqueue job
   - Dequeue job
   - Mark job complete/failed
   - Retry logic
   - Dead letter queue

3. WORKER MAIN (backend/cmd/worker/main.go):
   ```go
   func main() {
       // Initialize dependencies
       // Start job processors
       // Graceful shutdown
   }
   ```

4. JOB PROCESSORS (backend/cmd/worker/jobs/):
   
   a) publish_post.go:
      - Query due posts (scheduled_at <= now)
      - Acquire lock per post
      - Get social accounts for post
      - Call social adapter to publish
      - Update post status
      - Record analytics
      - Handle errors (retry up to 3 times)
   
   b) fetch_analytics.go:
      - Query published posts (24hrs+ old)
      - Fetch analytics from each platform
      - Store in database
      - Update post metrics
   
   c) cleanup.go:
      - Delete old draft posts (30+ days)
      - Clean up failed jobs
      - Archive old analytics

5. WORKER CONFIGURATION:
   - Poll interval (1 minute)
   - Batch size (10 posts per run)
   - Retry attempts (3)
   - Retry delay (exponential backoff)
   - Concurrency (5 workers)

6. UPDATE DOCKER COMPOSE:
   ```yaml
   worker:
     build:
       context: ./backend
       dockerfile: Dockerfile.worker
     depends_on:
       - postgres
       - redis
     environment:
       - DATABASE_URL=...
       - REDIS_URL=...
   ```

7. OBSERVABILITY:
   - Structured logging
   - Metrics (posts processed, failures, latency)
   - Health check endpoint

8. INTEGRATION TESTS:
   - Test post scheduling → publishing flow
   - Test retry logic
   - Test concurrent processing
   - Test dead letter queue

REQUIREMENTS:
- Idempotent job processing
- At-most-once delivery
- Graceful shutdown
- No race conditions
- Production-ready

GENERATE:
- All worker files
- Dockerfile.worker
- Updated docker-compose.yml
- Tests
```

---

## 📦 MODULE 5: Complete Frontend

### Prompt: Frontend - Complete Implementation

```
I need a COMPLETE Next.js 14 frontend for SocialQueue (Buffer clone) using App Router, TypeScript, Tailwind, and shadcn/ui.

PROJECT CONTEXT:
- Framework: Next.js 14 with App Router
- Styling: Tailwind CSS + shadcn/ui
- Forms: react-hook-form + zod
- API: React Query
- Backend: http://localhost:8000/api/v2

WHAT TO CREATE:

1. AUTHENTICATION PAGES (src/app/(auth)/):
   
   a) login/page.tsx:
      - Email + password form
      - Form validation (zod)
      - Error handling
      - Redirect to /dashboard on success
      - Link to signup
   
   b) signup/page.tsx:
      - Email, username, password, firstName, lastName
      - Password strength indicator
      - Form validation
      - Call POST /api/v2/auth/signup
      - Redirect to /verify-email
   
   c) verify-email/page.tsx:
      - Show verification sent message
      - Resend verification link
      - Auto-redirect when verified

2. DASHBOARD LAYOUT (src/app/(dashboard)/):
   
   a) layout.tsx:
      - Sidebar navigation
      - Top header with user menu
      - Protected route wrapper
   
   b) dashboard/page.tsx:
      - Overview stats (posts scheduled, published, analytics)
      - Recent posts list
      - Quick actions

3. POST COMPOSER (src/app/(dashboard)/compose/page.tsx):
   - Rich text editor (Tiptap)
   - Character counter per platform
   - Platform selector (Twitter, LinkedIn, Facebook)
   - Media upload (drag & drop)
   - Image preview
   - Schedule picker (date + time + timezone)
   - Draft save
   - Publish / Schedule buttons

4. POST QUEUE (src/app/(dashboard)/queue/page.tsx):
   - Calendar view (react-big-calendar)
   - List view with filters
   - Drag-and-drop reordering
   - Quick actions: edit, delete, reschedule
   - Status indicators

5. SOCIAL ACCOUNTS (src/app/(dashboard)/accounts/page.tsx):
   - Connected accounts list
   - Connect new account buttons
   - OAuth flow handling
   - Disconnect account
   - Account health status

6. ANALYTICS (src/app/(dashboard)/analytics/page.tsx):
   - Charts (recharts):
     * Line chart: impressions over time
     * Bar chart: engagement by platform
     * Pie chart: post distribution
   - Top performing posts table
   - Export CSV

7. SHARED COMPONENTS (src/components/):
   
   - ui/: shadcn/ui components (button, input, card, form, dialog, etc.)
   - auth/login-form.tsx
   - auth/signup-form.tsx
   - posts/post-card.tsx
   - posts/post-composer.tsx
   - posts/schedule-picker.tsx
   - social/account-card.tsx
   - social/oauth-button.tsx
   - layout/sidebar.tsx
   - layout/header.tsx

8. API CLIENT (src/lib/api.ts):
   - Axios/fetch wrapper
   - Token management
   - Request interceptors
   - Error handling

9. REACT QUERY HOOKS (src/hooks/):
   - use-auth.ts: login, signup, logout
   - use-posts.ts: createPost, updatePost, deletePost, getPosts
   - use-accounts.ts: connectAccount, getAccounts, disconnect
   - use-analytics.ts: getAnalytics

10. AUTHENTICATION CONTEXT (src/components/providers/auth-provider.tsx):
    - User state management
    - Token refresh
    - Protected route HOC

REQUIREMENTS:
- Full TypeScript
- All forms use react-hook-form + zod
- Responsive design (mobile-first)
- Dark mode support
- Loading states everywhere
- Error boundaries
- Toast notifications (sonner)
- Optimistic updates
- Production-ready

GENERATE:
- All pages and components
- Full functionality
- No placeholders
- Works with backend API
```

---

## 🎯 USAGE TIPS

1. **Copy the entire prompt** for the module you want
2. **Paste into Claude** and get the complete module
3. **Test immediately** - it should compile and work
4. **Move to next module** when tests pass

---

## ✅ VALIDATION CHECKLIST

After receiving each module:

- [ ] All files compile without errors
- [ ] `make test` passes
- [ ] Integration tests pass
- [ ] API endpoints respond correctly
- [ ] No TODO comments
- [ ] Documentation is complete
- [ ] You can demonstrate the feature working

---

**Last Updated**: October 12, 2025