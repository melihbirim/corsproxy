# Quick Start: Building the SaaS Version

This guide covers the **first 4-6 weeks** of implementation to launch the MVP.

---

## Week 1-2: Backend Foundation

### Step 1: Add Database Support

Install dependencies:
```bash
go get github.com/lib/pq              # PostgreSQL driver
go get github.com/joho/godotenv       # Environment variables
```

Create new files:

**`database.go`** - Database schema and migrations
```go
package main

import (
    "database/sql"
    "log"
    _ "github.com/lib/pq"
)

var db *sql.DB

func initDatabase(dsn string) error {
    var err error
    db, err = sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    
    // Create tables
    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR UNIQUE NOT NULL,
        oauth_id VARCHAR,
        tier VARCHAR DEFAULT 'free',
        created_at TIMESTAMP DEFAULT NOW(),
        stripe_customer_id VARCHAR
    );
    
    CREATE TABLE IF NOT EXISTS api_keys (
        id VARCHAR PRIMARY KEY,
        user_id INT REFERENCES users(id),
        name VARCHAR,
        created_at TIMESTAMP DEFAULT NOW(),
        last_used TIMESTAMP,
        rate_limit INT DEFAULT 100
    );
    
    CREATE TABLE IF NOT EXISTS request_logs (
        id BIGSERIAL PRIMARY KEY,
        user_id INT REFERENCES users(id),
        api_key_id VARCHAR REFERENCES api_keys(id),
        target_url VARCHAR,
        status_code INT,
        response_size INT,
        latency_ms INT,
        created_at TIMESTAMP DEFAULT NOW()
    );
    
    CREATE INDEX IF NOT EXISTS idx_logs_user_date ON request_logs(user_id, created_at);
    CREATE INDEX IF NOT EXISTS idx_api_key_user ON api_keys(user_id);
    `
    
    _, err = db.Exec(schema)
    return err
}
```

**`auth.go`** - API key validation
```go
package main

import (
    "crypto/rand"
    "encoding/hex"
    "strings"
    "time"
)

type APIKey struct {
    ID        string
    UserID    int
    Name      string
    CreatedAt time.Time
    LastUsed  *time.Time
    RateLimit int
}

// Generate random API key
func generateAPIKey() string {
    b := make([]byte, 32)
    rand.Read(b)
    return "cors_" + hex.EncodeToString(b)
}

// Validate API key and return user info
func validateAPIKey(key string) (*APIKey, error) {
    if key == "" || !strings.HasPrefix(key, "cors_") {
        return nil, ErrInvalidAPIKey
    }
    
    var apiKey APIKey
    err := db.QueryRow(
        "SELECT id, user_id, name, created_at, last_used, rate_limit FROM api_keys WHERE id = $1",
        key,
    ).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.CreatedAt, &apiKey.LastUsed, &apiKey.RateLimit)
    
    if err != nil {
        return nil, ErrInvalidAPIKey
    }
    
    // Update last used
    go db.Exec("UPDATE api_keys SET last_used = NOW() WHERE id = $1", key)
    
    return &apiKey, nil
}

var ErrInvalidAPIKey = errors.New("invalid API key")
```

### Step 2: Update Rate Limiting (Per-User)

**`ratelimit.go`** - User-based rate limiting
```go
package main

import (
    "sync"
    "time"
)

type UserRateLimit struct {
    count     int
    resetTime time.Time
    limit     int
}

var (
    userLimiters = make(map[int]*UserRateLimit)
    limitMutex   sync.RWMutex
)

func checkUserRateLimit(userID int, limit int) bool {
    limitMutex.Lock()
    defer limitMutex.Unlock()
    
    now := time.Now()
    rl, exists := userLimiters[userID]
    
    if !exists || now.After(rl.resetTime) {
        userLimiters[userID] = &UserRateLimit{
            count:     1,
            resetTime: now.Add(time.Hour),
            limit:     limit,
        }
        return true
    }
    
    if rl.count >= rl.limit {
        return false
    }
    
    rl.count++
    return true
}
```

### Step 3: Add Request Logging

**`logging.go`** - Log all requests for analytics
```go
package main

import (
    "context"
    "time"
)

func logRequest(ctx context.Context, userID int, apiKeyID string, url string, statusCode int, respSize int, latencyMs int64) {
    go func() {
        _, err := db.Exec(`
            INSERT INTO request_logs (user_id, api_key_id, target_url, status_code, response_size, latency_ms)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, userID, apiKeyID, url, statusCode, respSize, latencyMs)
        
        if err != nil {
            log.Printf("Error logging request: %v", err)
        }
    }()
}
```

### Step 4: Update Main Handler

Modify `corsProxyHandler` to use API keys:

```go
func corsProxyHandlerWithAuth(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    
    // 1. Extract API key from Authorization header (or query for web apps)
    apiKeyStr := r.Header.Get("Authorization")
    if apiKeyStr == "" {
        // Allow query param for easier browser testing
        apiKeyStr = r.URL.Query().Get("key")
    } else {
        // Strip "Bearer " prefix if present
        if strings.HasPrefix(apiKeyStr, "Bearer ") {
            apiKeyStr = strings.TrimPrefix(apiKeyStr, "Bearer ")
        }
    }
    
    var userID int
    var apiKeyID string
    var userTier string
    var rateLimit int
    
    if apiKeyStr != "" {
        apiKey, err := validateAPIKey(apiKeyStr)
        if err != nil {
            http.Error(w, `{"error":"Invalid API key"}`, http.StatusUnauthorized)
            return
        }
        userID = apiKey.UserID
        apiKeyID = apiKey.ID
        rateLimit = apiKey.RateLimit
        
        // Get user tier for SLA
        db.QueryRow("SELECT tier FROM users WHERE id = $1", userID).Scan(&userTier)
    } else {
        // Anonymous user - very limited rate limit
        userID = 0
        rateLimit = 10 // 10 requests per hour for anon
    }
    
    // 2. Check rate limit
    if !checkUserRateLimit(userID, rateLimit) {
        http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
        return
    }
    
    // 3. Call original handler
    corsProxyHandler(w, r)
    
    // 4. Log request (in background)
    statusCode := 200 // This is tricky - need to wrap ResponseWriter
    respSize := 0
    latency := time.Since(startTime).Milliseconds()
    logRequest(context.Background(), userID, apiKeyID, r.URL.Query().Get("url"), statusCode, respSize, latency)
}
```

---

## Week 2-3: Frontend Dashboard

Use **Next.js** for speed (better than building in React from scratch):

```bash
npx create-next-app@latest dashboard --typescript --tailwind
cd dashboard
```

**`app/layout.tsx`** - Main layout with auth
```tsx
import { getServerSession } from 'next-auth'
import { authOptions } from './api/auth/[...nextauth]/route'

export default async function RootLayout({ children }) {
  const session = await getServerSession(authOptions)
  
  return (
    <html>
      <body>
        <header>
          {session ? (
            <>
              <span>{session.user.email}</span>
              <a href="/api/auth/signout">Sign out</a>
            </>
          ) : (
            <a href="/api/auth/signin">Sign in</a>
          )}
        </header>
        {children}
      </body>
    </html>
  )
}
```

**`app/dashboard/page.tsx`** - Dashboard home
```tsx
import { getServerSession } from 'next-auth'

export default async function Dashboard() {
  const session = await getServerSession()
  
  if (!session) redirect('/api/auth/signin')
  
  const stats = await fetch(`/api/stats?userId=${session.user.id}`)
  const data = await stats.json()
  
  return (
    <main>
      <h1>Welcome, {session.user.name}</h1>
      <div className="grid grid-cols-4 gap-4">
        <Card title="Requests Today" value={data.requestsToday} />
        <Card title="Rate Limit" value={`${data.usedToday}/${data.rateLimit}`} />
        <Card title="Plan" value={data.tier.toUpperCase()} />
        <Card title="Status" value="🟢 Healthy" />
      </div>
    </main>
  )
}
```

**`app/dashboard/keys/page.tsx`** - API Key management
```tsx
'use client'

import { useState } from 'react'

export default function APIKeysPage() {
  const [keys, setKeys] = useState([])
  const [newKeyName, setNewKeyName] = useState('')
  
  const createKey = async () => {
    const res = await fetch('/api/keys', {
      method: 'POST',
      body: JSON.stringify({ name: newKeyName })
    })
    const data = await res.json()
    setKeys([...keys, data])
  }
  
  return (
    <div>
      <h1>API Keys</h1>
      <input 
        placeholder="Key name (e.g., 'Frontend App')"
        value={newKeyName}
        onChange={(e) => setNewKeyName(e.target.value)}
      />
      <button onClick={createKey}>Create Key</button>
      
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Key</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {keys.map(key => (
            <tr key={key.id}>
              <td>{key.name}</td>
              <td>
                <code>{key.id}</code>
                <button onClick={() => navigator.clipboard.writeText(key.id)}>
                  Copy
                </button>
              </td>
              <td>{new Date(key.created_at).toLocaleDateString()}</td>
              <td>
                <button onClick={() => deleteKey(key.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
```

---

## Week 3: Stripe Integration

Install Stripe:
```bash
go get github.com/stripe/stripe-go/v72
```

**`billing.go`** - Stripe subscription management
```go
import "github.com/stripe/stripe-go/v72"

func createCheckoutSession(userID int, tier string) (string, error) {
    stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
    
    priceMap := map[string]string{
        "pro":      "price_xxx_pro",
        "business": "price_xxx_business",
    }
    
    params := &stripe.CheckoutSessionParams{
        SuccessURL: stripe.String("https://mysite.com/billing/success"),
        CancelURL:  stripe.String("https://mysite.com/billing"),
        PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {
                Price:    stripe.String(priceMap[tier]),
                Quantity: stripe.Int64(1),
            },
        },
    }
    
    session, _ := session.New(params)
    return session.URL, nil
}
```

---

## Week 4: Deploy to Production

### Option A: Render (Recommended for fast launch)
1. Push code to GitHub
2. Connect Render to GitHub repo
3. Set environment variables
4. Deploy (auto-redeploy on push)

### Option B: Railway
Similar flow, also simple

### Environment Variables
```env
# Database
DATABASE_URL=postgresql://user:pass@host:5432/corsproxy

# Auth
NEXTAUTH_SECRET=<random string>
NEXTAUTH_URL=https://yourdomain.com
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx

# Stripe
STRIPE_SECRET_KEY=sk_xxx
STRIPE_PUBLIC_KEY=pk_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

---

## Week 5-6: Launch & Iterate

### Before Launch
- [ ] Test API key creation/deletion
- [ ] Test Stripe integration
- [ ] Test rate limiting
- [ ] Load test (100 concurrent users)
- [ ] Create documentation
- [ ] Set up support email
- [ ] Create landing page

### Launch Channels
1. **Hacker News** - Post Monday 11am EST (async, but hits early users)
2. **Product Hunt** - Schedule launch, ask HN users to upvote
3. **Twitter** - Announce with demo gif
4. **Indie Hackers** - Ask for feedback
5. **Email friends** - Get early momentum

### Metrics to Track
- Signups per day
- Free → Pro conversion rate
- API key creation rate (engagement)
- Average requests per user

---

## Tech Stack Summary

**Backend:**
- Go 1.21+
- PostgreSQL
- Redis (optional, add later)
- Stripe API

**Frontend:**
- Next.js 14+
- TypeScript
- Tailwind CSS
- NextAuth (authentication)

**Deployment:**
- Render or Railway (host both backend + frontend)
- PostgreSQL managed database
- GitHub for CI/CD

**Total Setup Time:** ~2 weeks for basic MVP

---

## Cost Estimate (Monthly)

| Service | Cost | Notes |
|---------|------|-------|
| Render/Railway | $10-30 | Backend + Frontend |
| PostgreSQL | $15+ | Managed database |
| Stripe | 2.9% + $0.30 | Only on paid tiers |
| Domain | $10 | dns.com or Namecheap |
| **Total** | **$35-50** | Grows with scale |

**Revenue breakeven:** ~10 Pro users ($50/month)

---

## Key Advice

1. **Start simple** - Don't build everything. MVP = Auth + API keys + basic dashboard
2. **Talk to users** - Interview 10 people before launch, adjust based on feedback
3. **Launch fast** - Ship MVP in 4 weeks, iterate based on real user feedback
4. **Automate operations** - Set up monitoring/alerts immediately to sleep better
5. **Use SaaS templates** - Leverage existing libraries (NextAuth, Stripe) instead of building from scratch

---

## Next Immediate Step

1. Create PostgreSQL database on Render/Railway
2. Add the Go files (database.go, auth.go, logging.go, billing.go)
3. Create Next.js dashboard project
4. Deploy to staging and test manually
5. Then go live!

Questions? This is a lot, but each step is small and doable.
