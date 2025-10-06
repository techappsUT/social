socialqueue/
├── .gitignore
├── .editorconfig
├── README.md
├── docker-compose.yml              # Dev environment orchestration
├── Makefile                        # Root-level commands
├── .vscode/
│   ├── settings.json
│   ├── extensions.json
│   └── launch.json
│
├── frontend/
│   ├── .eslintrc.json
│   ├── .prettierrc
│   ├── .prettierignore
│   ├── next.config.js
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── package.json
│   ├── .env.example
│   ├── .env.local
│   ├── .husky/
│   │   ├── pre-commit
│   │   └── pre-push
│   ├── public/
│   │   ├── favicon.ico
│   │   └── images/
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── globals.css
│   │   │   ├── (auth)/
│   │   │   │   ├── login/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── register/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── layout.tsx
│   │   │   ├── (dashboard)/
│   │   │   │   ├── dashboard/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── compose/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── queue/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── analytics/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── accounts/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── layout.tsx
│   │   │   └── api/
│   │   │       └── health/
│   │   │           └── route.ts
│   │   ├── components/
│   │   │   ├── ui/                 # shadcn/ui components
│   │   │   │   ├── button.tsx
│   │   │   │   ├── card.tsx
│   │   │   │   ├── dialog.tsx
│   │   │   │   └── ...
│   │   │   ├── layout/
│   │   │   │   ├── header.tsx
│   │   │   │   ├── sidebar.tsx
│   │   │   │   └── footer.tsx
│   │   │   ├── posts/
│   │   │   │   ├── post-composer.tsx
│   │   │   │   ├── post-card.tsx
│   │   │   │   └── post-calendar.tsx
│   │   │   └── providers/
│   │   │       ├── query-provider.tsx
│   │   │       └── auth-provider.tsx
│   │   ├── lib/
│   │   │   ├── api.ts              # API client
│   │   │   ├── auth.ts             # Auth utilities
│   │   │   ├── utils.ts            # Helper functions
│   │   │   └── constants.ts
│   │   ├── hooks/
│   │   │   ├── use-auth.ts
│   │   │   ├── use-posts.ts
│   │   │   └── use-accounts.ts
│   │   ├── types/
│   │   │   ├── index.ts
│   │   │   ├── post.ts
│   │   │   ├── account.ts
│   │   │   └── user.ts
│   │   └── styles/
│   │       └── themes/
│   └── next-env.d.ts
│
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── .golangci.yml
│   ├── .pre-commit-config.yaml
│   ├── .env.example
│   ├── .env
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   ├── Makefile
│   ├── README.md
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go             # Main API server
│   │   └── worker/
│   │       └── main.go             # Background job worker
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # App configuration
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   └── ratelimit.go
│   │   ├── handlers/
│   │   │   ├── auth.go
│   │   │   ├── posts.go
│   │   │   ├── accounts.go
│   │   │   ├── analytics.go
│   │   │   └── health.go
│   │   ├── services/
│   │   │   ├── auth/
│   │   │   │   └── auth.go
│   │   │   ├── posts/
│   │   │   │   └── posts.go
│   │   │   ├── social/
│   │   │   │   ├── twitter.go
│   │   │   │   ├── facebook.go
│   │   │   │   ├── linkedin.go
│   │   │   │   └── instagram.go
│   │   │   └── queue/
│   │   │       └── queue.go
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── post.go
│   │   │   ├── account.go
│   │   │   └── analytics.go
│   │   ├── repository/
│   │   │   ├── user.go
│   │   │   ├── post.go
│   │   │   └── account.go
│   │   ├── database/
│   │   │   ├── db.go
│   │   │   └── migrations.go
│   │   └── utils/
│   │       ├── jwt.go
│   │       ├── hash.go
│   │       └── validator.go
│   ├── migrations/
│   │   ├── 000001_create_users_table.up.sql
│   │   ├── 000001_create_users_table.down.sql
│   │   ├── 000002_create_posts_table.up.sql
│   │   ├── 000002_create_posts_table.down.sql
│   │   ├── 000003_create_accounts_table.up.sql
│   │   └── 000003_create_accounts_table.down.sql
│   ├── pkg/
│   │   └── logger/
│   │       └── logger.go
│   └── scripts/
│       ├── seed.go
│       └── setup.sh
│
├── infra/
│   ├── docker-compose.dev.yml
│   ├── docker-compose.prod.yml
│   ├── .env.example
│   ├── nginx/
│   │   └── nginx.conf
│   └── k8s/                        # Future Kubernetes configs
│       ├── api-deployment.yaml
│       └── worker-deployment.yaml
│
└── docs/
    ├── API.md
    ├── ARCHITECTURE.md
    ├── DEPLOYMENT.md
    └── CONTRIBUTING.md