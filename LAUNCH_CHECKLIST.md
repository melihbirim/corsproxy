# Free Tier MVP - Launch Plan (1-2 weeks)

## Strategic Approach
**Get real users → Measure demand → Add billing based on data**

This is smarter than guessing what users want. Focus on:
- ✅ Simple sign-up (Google OAuth or email)
- ✅ API key generation
- ✅ Free tier with generous limits
- ✅ Usage tracking (to understand demand)
- ✅ Dead simple dashboard
- ✅ ZERO payment infrastructure (add later)

---

## Week 1: Backend + Database

### Day 1-2: Setup & Database

```bash
# Create new project structure
mkdir corsproxy-saas
cd corsproxy-saas

# Backend
mkdir backend
cd backend
go mod init corsproxy
go get github.com/lib/pq
go get github.com/joho/godotenv
go get github.com/golang-jwt/jwt/v5

# Frontend
cd ..
npx create-next-app@latest frontend --typescript --tailwind
```

### Database Schema (`backend/schema.sql`)

```sql
-- Super simple schema, nothing fancy
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL,
    name VARCHAR,
    created_at TIMESTAMP DEFAULT NOW(),
    requests_today INT DEFAULT 0,
    last_request_reset TIMESTAMP DEFAULT NOW()
);

CREATE TABLE api_keys (
    id VARCHAR PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE request_logs (
    id BIGSERIAL PRIMARY KEY,
    api_key_id VARCHAR REFERENCES api_keys(id) ON DELETE CASCADE,
    target_url VARCHAR,
    status_code INT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Index for fast lookups
CREATE INDEX idx_api_key ON api_keys(id);
CREATE INDEX idx_logs_key ON request_logs(api_key_id);
CREATE INDEX idx_logs_date ON request_logs(created_at);
```

Deploy PostgreSQL:
- Render: https://render.com (free tier, 1 shared database)
- Or Railway: https://railway.app (also free tier)
- Copy connection string to `.env`

### Backend: Main API (`backend/main.go`)

```go
package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "crypto/rand"
    "encoding/hex"
    
    _ "github.com/lib/pq"
)

var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
}

// Generate random API key (cors_xxx)
func generateAPIKey() string {
    b := make([]byte, 16)
    rand.Read(b)
    return "cors_" + hex.EncodeToString(b)
}

// User signup (email + Google OAuth handled on frontend)
func handleSignup(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `json:"email"`
        Name  string `json:"name"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    var userID int
    err := db.QueryRow(
        "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = $2 RETURNING id",
        req.Email, req.Name,
    ).Scan(&userID)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Return simple JWT-like token for frontend (or just use email for MVP)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "user_id": userID,
        "email":   req.Email,
    })
}

// Create API key
func handleCreateKey(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID int    `json:"user_id"`
        Name   string `json:"name"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    keyID := generateAPIKey()
    _, err := db.Exec(
        "INSERT INTO api_keys (id, user_id, name) VALUES ($1, $2, $3)",
        keyID, req.UserID, req.Name,
    )
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "id":  keyID,
        "key": keyID,
    })
}

// Get user's API keys
func handleGetKeys(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    
    rows, err := db.Query(
        "SELECT id, name, created_at FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC",
        userID,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer rows.Close()
    
    var keys []map[string]interface{}
    for rows.Next() {
        var id, name string
        var createdAt string
        rows.Scan(&id, &name, &createdAt)
        keys = append(keys, map[string]interface{}{
            "id":         id,
            "name":       name,
            "created_at": createdAt,
        })
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(keys)
}

// Delete API key
func handleDeleteKey(w http.ResponseWriter, r *http.Request) {
    keyID := r.URL.Query().Get("key_id")
    
    _, err := db.Exec("DELETE FROM api_keys WHERE id = $1", keyID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}

// Validate API key and count requests
func handleValidateKey(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Key string `json:"key"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    var userID int
    err := db.QueryRow(
        "SELECT user_id FROM api_keys WHERE id = $1",
        req.Key,
    ).Scan(&userID)
    
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    
    // Check daily limit (1000 requests/day for free)
    var count int
    db.QueryRow(
        "SELECT COUNT(*) FROM request_logs WHERE api_key_id = $1 AND DATE(created_at) = CURRENT_DATE",
        req.Key,
    ).Scan(&count)
    
    if count > 1000 {
        http.Error(w, "Daily limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "valid":       true,
        "user_id":     userID,
        "limit":       1000,
        "used_today":  count,
    })
}

// Log request
func handleLogRequest(w http.ResponseWriter, r *http.Request) {
    var req struct {
        APIKeyID   string `json:"api_key_id"`
        TargetURL  string `json:"target_url"`
        StatusCode int    `json:"status_code"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    _, err := db.Exec(
        "INSERT INTO request_logs (api_key_id, target_url, status_code) VALUES ($1, $2, $3)",
        req.APIKeyID, req.TargetURL, req.StatusCode,
    )
    
    if err != nil {
        log.Printf("Error logging: %v", err)
    }
    
    w.WriteHeader(http.StatusOK)
}

func main() {
    // Routes
    http.HandleFunc("/api/signup", handleSignup)
    http.HandleFunc("/api/keys/create", handleCreateKey)
    http.HandleFunc("/api/keys", handleGetKeys)
    http.HandleFunc("/api/keys/delete", handleDeleteKey)
    http.HandleFunc("/api/keys/validate", handleValidateKey)
    http.HandleFunc("/api/logs", handleLogRequest)
    
    log.Println("Backend listening on :8000")
    http.ListenAndServe(":8000", nil)
}
```

### Update Main CORS Proxy

Modify the existing `main.go` to add key validation:

```go
// Add to corsProxyHandler
func corsProxyHandler(w http.ResponseWriter, r *http.Request) {
    // Extract API key
    apiKey := r.Header.Get("Authorization")
    if apiKey == "" {
        apiKey = r.URL.Query().Get("key")
    } else if strings.HasPrefix(apiKey, "Bearer ") {
        apiKey = strings.TrimPrefix(apiKey, "Bearer ")
    }
    
    // Validate key with backend API
    var valid bool
    if apiKey != "" {
        resp, err := http.Post(
            "http://localhost:8000/api/keys/validate",
            "application/json",
            strings.NewReader(fmt.Sprintf(`{"key":"%s"}`, apiKey)),
        )
        valid = resp.StatusCode == 200
        if err != nil {
            valid = false // Allow anonymous
        }
    } else {
        valid = true // Allow anonymous (but rate limited)
    }
    
    // Continue with original logic...
    // (existing CORS proxy code)
}
```

---

## Week 1: Frontend

### Simple Landing Page + Dashboard (`frontend/`)

```bash
cd frontend
npm install -D shadcn-ui @radix-ui/react-dialog
```

**`app/page.tsx`** - Landing page

```tsx
'use client'
import { useState } from 'react'
import Link from 'next/link'

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 to-slate-800">
      {/* Header */}
      <header className="border-b border-slate-700">
        <div className="max-w-6xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="text-2xl font-bold text-white">CORS Proxy</div>
          <Link href="/dashboard">
            <button className="bg-blue-600 text-white px-4 py-2 rounded">
              Get Started Free
            </button>
          </Link>
        </div>
      </header>

      {/* Hero */}
      <main className="max-w-6xl mx-auto px-6 py-20 text-center">
        <h1 className="text-5xl font-bold text-white mb-6">
          Unlimited CORS Proxying
        </h1>
        <p className="text-xl text-slate-400 mb-12">
          Access any API from your frontend. No backend needed.
        </p>
        <p className="text-lg text-slate-300 mb-8">
          <code className="bg-slate-700 px-3 py-1 rounded">
            ?url=https://api.example.com
          </code>
        </p>

        {/* Features */}
        <div className="grid grid-cols-3 gap-8 my-20">
          <div className="bg-slate-700 p-6 rounded">
            <h3 className="text-lg font-bold mb-2">⚡ Fast</h3>
            <p className="text-slate-400">Global edge distribution</p>
          </div>
          <div className="bg-slate-700 p-6 rounded">
            <h3 className="text-lg font-bold mb-2">🔓 Open</h3>
            <p className="text-slate-400">1000 requests/day free</p>
          </div>
          <div className="bg-slate-700 p-6 rounded">
            <h3 className="text-lg font-bold mb-2">📊 Tracked</h3>
            <p className="text-slate-400">See your usage anytime</p>
          </div>
        </div>

        <Link href="/dashboard">
          <button className="bg-blue-600 text-white px-8 py-4 rounded text-lg font-bold">
            Sign Up Free
          </button>
        </Link>

        <div className="mt-12 text-slate-400 text-sm">
          No credit card required
        </div>
      </main>
    </div>
  )
}
```

**`app/dashboard/page.tsx`** - Simple dashboard

```tsx
'use client'
import { useState, useEffect } from 'react'

export default function Dashboard() {
  const [userID, setUserID] = useState<number | null>(null)
  const [email, setEmail] = useState('')
  const [keys, setKeys] = useState<any[]>([])
  const [newKeyName, setNewKeyName] = useState('')
  const [stats, setStats] = useState({ used: 0, limit: 1000 })

  // Simple email-based login for MVP (no OAuth yet)
  const handleSignup = async () => {
    const res = await fetch('/api/signup', {
      method: 'POST',
      body: JSON.stringify({ email, name: email.split('@')[0] })
    })
    const data = await res.json()
    setUserID(data.user_id)
    localStorage.setItem('user_id', data.user_id)
  }

  useEffect(() => {
    const storedID = localStorage.getItem('user_id')
    if (storedID) {
      setUserID(parseInt(storedID))
      fetchKeys(parseInt(storedID))
    }
  }, [])

  const fetchKeys = async (id: number) => {
    const res = await fetch(`/api/keys?user_id=${id}`)
    const data = await res.json()
    setKeys(data || [])
  }

  const createKey = async () => {
    const res = await fetch('/api/keys/create', {
      method: 'POST',
      body: JSON.stringify({ user_id: userID, name: newKeyName })
    })
    const data = await res.json()
    setKeys([...keys, data])
    setNewKeyName('')
  }

  if (!userID) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="bg-slate-800 p-8 rounded max-w-md w-full">
          <h1 className="text-2xl font-bold text-white mb-6">Sign Up</h1>
          <input
            type="email"
            placeholder="your@email.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-2 rounded bg-slate-700 text-white mb-4"
          />
          <button
            onClick={handleSignup}
            className="w-full bg-blue-600 text-white px-4 py-2 rounded font-bold"
          >
            Sign Up Free
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      {/* Header */}
      <header className="border-b border-slate-700 px-6 py-4">
        <div className="max-w-6xl mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold">CORS Proxy</h1>
          <div>
            <span className="text-slate-400">{email}</span>
            <button
              onClick={() => {
                localStorage.clear()
                setUserID(null)
              }}
              className="ml-4 text-blue-400"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Dashboard */}
      <main className="max-w-6xl mx-auto px-6 py-12">
        {/* Stats */}
        <div className="grid grid-cols-2 gap-6 mb-12">
          <div className="bg-slate-800 p-6 rounded">
            <div className="text-slate-400 mb-2">Usage Today</div>
            <div className="text-4xl font-bold">
              {stats.used} <span className="text-lg text-slate-400">/ 1000</span>
            </div>
            <div className="w-full bg-slate-700 mt-4 h-2 rounded">
              <div
                className="bg-blue-600 h-full rounded"
                style={{ width: `${(stats.used / 1000) * 100}%` }}
              ></div>
            </div>
          </div>
          <div className="bg-slate-800 p-6 rounded">
            <div className="text-slate-400 mb-2">API Keys</div>
            <div className="text-4xl font-bold">{keys.length}</div>
            <div className="text-slate-400 mt-4">Active keys</div>
          </div>
        </div>

        {/* API Keys */}
        <div className="bg-slate-800 p-6 rounded mb-12">
          <h2 className="text-2xl font-bold mb-6">API Keys</h2>

          <div className="flex gap-4 mb-6">
            <input
              type="text"
              placeholder="Key name (e.g., 'Frontend App')"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              className="flex-1 px-4 py-2 rounded bg-slate-700 text-white"
            />
            <button
              onClick={createKey}
              className="bg-blue-600 text-white px-6 py-2 rounded font-bold"
            >
              Create
            </button>
          </div>

          {keys.length === 0 ? (
            <p className="text-slate-400">No API keys yet. Create one to get started!</p>
          ) : (
            <div className="space-y-4">
              {keys.map((key) => (
                <div key={key.id} className="bg-slate-700 p-4 rounded flex justify-between items-center">
                  <div>
                    <div className="font-bold">{key.name}</div>
                    <code className="text-slate-400 text-sm">{key.id}</code>
                  </div>
                  <button
                    onClick={() => navigator.clipboard.writeText(key.id)}
                    className="bg-blue-600 px-4 py-2 rounded text-sm"
                  >
                    Copy
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Usage Example */}
        <div className="bg-slate-800 p-6 rounded">
          <h2 className="text-2xl font-bold mb-4">How to Use</h2>
          <p className="text-slate-400 mb-4">
            Make requests with your API key in the Authorization header:
          </p>
          <code className="bg-slate-900 p-4 rounded block text-sm text-green-400">
            {`curl "https://cors.yourdomain.com/?url=https://api.example.com" \\
  -H "Authorization: Bearer YOUR_KEY"`}
          </code>
        </div>
      </main>
    </div>
  )
}
```

---

## Week 2: Deploy

### Backend Deployment (Render)

```bash
cd backend
# Create render.yaml
```

**`render.yaml`**
```yaml
services:
  - type: web
    name: cors-proxy-api
    env: go
    plan: free
    buildCommand: go build -o cors-proxy-api main.go
    startCommand: ./cors-proxy-api
    envVars:
      - key: DATABASE_URL
        value: postgresql://user:pass@db:5432/corsproxy
```

Push to GitHub, connect to Render → auto-deploy

### Frontend Deployment (Vercel)

```bash
cd frontend
# Push to GitHub
git push origin main
```

Go to vercel.com → Import GitHub repo → Deploy (5 seconds)

---

## Week 2: Launch Channels

### Pre-Launch (Day 1-2)
- [ ] Set up custom domain (cors.io or similar)
- [ ] Test sign-up flow end-to-end
- [ ] Write README for product page
- [ ] Create 3-minute demo GIF

### Launch Day (Day 3)
- [ ] Post on **Hacker News** (9-10am EST)
- [ ] Share on **Twitter**
- [ ] Post on **Indie Hackers**
- [ ] Post on **Dev.to**

### Week 2 (Days 4-7)
- [ ] Respond to comments
- [ ] Fix bugs reported by users
- [ ] Share early user stories
- [ ] Request Product Hunt review votes

---

## Metrics to Track NOW (Before Adding Billing)

Create simple dashboard showing:
```
├─ Total signups (per day)
├─ API keys created (per day)
├─ Requests made (per day)
├─ Daily active users
├─ Churn rate (who comes back)
└─ Average requests per user
```

Add to backend:
```go
// Simple daily metrics
func getDailyMetrics() {
    var signups, keyCreations, totalRequests, activeUsers int
    
    db.QueryRow(`
        SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE
    `).Scan(&signups)
    
    db.QueryRow(`
        SELECT COUNT(*) FROM api_keys WHERE DATE(created_at) = CURRENT_DATE
    `).Scan(&keyCreations)
    
    db.QueryRow(`
        SELECT COUNT(*) FROM request_logs WHERE DATE(created_at) = CURRENT_DATE
    `).Scan(&totalRequests)
    
    db.QueryRow(`
        SELECT COUNT(DISTINCT user_id) FROM request_logs WHERE DATE(created_at) = CURRENT_DATE
    `).Scan(&activeUsers)
    
    // Log or return these metrics
}
```

---

## Free Tier Limits (Start Generous)

```
├─ Requests per day: 1,000 (very generous)
├─ API keys per user: 3
├─ Max request size: 10MB
├─ Request timeout: 30s
├─ No authentication required
└─ No rate limiting per IP (track by key only)
```

This is intentionally generous to attract users and see who actually needs more.

---

## What NOT to Include Yet

❌ Stripe/Billing
❌ OAuth (email signup is fine for MVP)
❌ Advanced analytics
❌ Custom domains
❌ Team collaboration
❌ API documentation (README is fine)
❌ Support chat

---

## Success Metrics (First 2 Weeks)

If you get:
- ✅ **100 signups** = People care
- ✅ **50 keys created** = They're actually using it
- ✅ **10k requests** = Real usage pattern emerging
- ✅ **10 daily active users** = You have product-market fit signal

If you get these numbers, THEN add billing.

---

## After Launch: Billing Phase

Once you hit 100+ users with real usage:
1. Add Stripe (5 hours)
2. Set simple pricing: $5/month for 100k requests
3. Email users: "We built billing if you want more"
4. Convert 2-5% to paid (expected)

But first: **Ship. Get users. Measure demand.**

---

## Quick Start Commands

```bash
# Backend
cd backend
go run main.go

# Frontend (new terminal)
cd frontend
npm run dev

# Deploy when ready
git push origin main  # Auto-deploys to Render + Vercel
```

That's it. You're live.
