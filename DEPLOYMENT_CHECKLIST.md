# Deployment Checklist - 30-Minute Launch

Copy-paste commands. Launch in 30 minutes.

---

## 1. Install Wrangler (2 min)

```bash
npm install -g wrangler
wrangler login  # Log into Cloudflare account
```

---

## 2. Create Project (2 min)

```bash
wrangler init corsproxy-worker
cd corsproxy-worker

# Install dependencies
npm install itty-router

# Create D1 database
wrangler d1 create corsproxy
```

**⚠️ Copy the database ID from output:**
```
✨ Successfully created database. Uploaded as corsproxy
database_id: YOUR_DATABASE_ID_HERE
```

Save `YOUR_DATABASE_ID_HERE` - you'll need it in step 4.

---

## 3. Create Database Schema (3 min)

Create `schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS request_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id TEXT NOT NULL,
    target_url TEXT NOT NULL,
    status_code INTEGER,
    response_size INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_key ON request_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_logs_date ON request_logs(created_at);
```

Deploy schema:

```bash
wrangler d1 execute corsproxy --file=schema.sql
```

---

## 4. Add Worker Code (3 min)

Create `src/index.ts` - **Copy entire code from CLOUDFLARE_WORKER_CODE.md**

Update `wrangler.toml`:

```toml
name = "corsproxy"
main = "src/index.ts"
compatibility_date = "2024-01-01"
minify = true

[[d1_databases]]
binding = "DB"
database_name = "corsproxy"
database_id = "PASTE_YOUR_DB_ID_HERE"

[[kv_namespaces]]
binding = "KV"
id = "your-kv-id"
preview_id = "your-kv-preview-id"
```

**Replace `PASTE_YOUR_DB_ID_HERE` with your actual ID from step 2.**

---

## 5. Test Locally (5 min)

```bash
wrangler dev
```

Visit: http://localhost:8787

Should see the dashboard. Try signing up.

```bash
# In another terminal, test the API:
curl "http://localhost:8787/api/signup" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Should return: {"user_id":1,"email":"test@example.com"}
```

Press Ctrl+C to stop.

---

## 6. Deploy Worker (3 min)

```bash
wrangler deploy
```

Output:
```
✨ Successfully published your Worker
→ https://corsproxy-xyz.workers.dev
```

**Save this URL: `https://corsproxy-xyz.workers.dev`**

---

## 7. Test Deployed Worker (3 min)

```bash
# Replace with your actual Worker URL
WORKER_URL="https://corsproxy-xyz.workers.dev"

# Create account
curl "$WORKER_URL/api/signup" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Response: {"user_id":1,"email":"test@example.com"}
```

---

## 8. Create Pages Dashboard (5 min)

1. Push to GitHub:

```bash
git init
git add .
git commit -m "Initial CORS proxy worker"
git remote add origin https://github.com/YOUR_USERNAME/corsproxy-worker
git push -u origin main
```

2. Create `public/index.html` - **Copy entire code from CLOUDFLARE_WORKER_CODE.md**

3. Update the API base in `index.html`:

```javascript
// Line with:
const API_BASE = 'https://corsproxy-xyz.workers.dev'

// Change to YOUR Worker URL from step 6
```

4. Push dashboard code:

```bash
git add public/index.html
git commit -m "Add dashboard"
git push origin main
```

5. Deploy to Pages:
   - Go to https://dash.cloudflare.com → Pages
   - Click "Create application"
   - Select your GitHub repo
   - Build settings:
     - Framework: None
     - Build command: (leave empty)
     - Build output directory: `public`
   - Click "Save and Deploy"

**Your Pages URL:** `https://corsproxy-YOUR_GITHUB_USERNAME.pages.dev`

---

## 9. Connect Custom Domain (2 min)

Option A: Use Cloudflare nameservers
```
1. Buy domain (Namecheap, Google Domains, etc.)
2. Change nameservers to:
   - nathaniel.ns.cloudflare.com
   - oscar.ns.cloudflare.com
3. In Cloudflare dashboard → DNS → Add A record pointing to your site
```

Option B: Use Workers route (easier)
```
1. In Cloudflare dashboard → Workers → Triggers → Routes
2. Add route: cors.yourdomain.com/*
3. Select your corsproxy worker
```

---

## 10. Test the Full System (5 min)

1. Go to your Pages URL (dashboard)
2. Sign up with your email
3. Create an API key
4. Copy the key

Test the proxy:

```bash
# Replace with your actual Worker URL and key
curl "https://corsproxy-xyz.workers.dev/?url=https://api.github.com/users/octocat&key=cors_YOUR_KEY"

# Should return GitHub user data with CORS headers set
```

Test from JavaScript:

```javascript
// From any website's console:
fetch('https://corsproxy-xyz.workers.dev/?url=https://api.example.com', {
  headers: { 'Authorization': 'Bearer cors_YOUR_KEY' }
})
  .then(r => r.json())
  .then(data => console.log(data))
```

---

## ✅ You're Live!

Your CORS proxy is now live and accessible:

```
Dashboard:  https://corsproxy-YOUR_USERNAME.pages.dev
Proxy API:  https://corsproxy-xyz.workers.dev/?url=...
```

---

## Next Steps

### Monitor
Add this to your dashboard to track usage:

```javascript
// In index.html, after loadKeys():
async function loadStats() {
  const userId = localStorage.getItem('user_id')
  
  // You could add a /api/stats endpoint to return total users, requests, etc.
  // For now, just track locally
}
```

### Share
```
1. Tweet your launch
2. Post on Hacker News
3. Share on Indie Hackers
4. Post on Dev.to
```

Example:
```
Just launched a free CORS proxy! 🚀
- 1000 requests/day free
- No backend, all on Cloudflare
- Costs me $0 to operate

Try it: https://corsproxy-xyz.workers.dev

What API did you always want to access?
```

### Add Paid Tiers (Later)

Once you have users, add this to accept payments:

```bash
# Install Stripe
npm install stripe

# Add endpoint to wrangler
router.post('/api/checkout', async (req) => {
  const { user_id, plan } = await req.json()
  // Create Stripe checkout session
})
```

But for now, just focus on getting free users and measuring demand.

---

## Troubleshooting

### "Database not found"
```
wrangler d1 execute corsproxy --file=schema.sql
# Make sure you're in the correct directory
```

### "Worker deployment failed"
```
# Check wrangler.toml has correct database ID
# Run: wrangler deploy --verbose
```

### "Pages not building"
```
# Make sure public/index.html exists
# Check build settings match above (output: public)
```

### API key not working
```
# Make sure to copy full key (with cors_ prefix)
# Check Authorization header vs key query param
```

---

## Success Metrics (First Week)

Track these to know it's working:

```
✅ Dashboard loads
✅ Can sign up and create key
✅ Can make proxy request
✅ CORS headers returned (Access-Control-Allow-Origin)
✅ Usage logged to database
✅ First 10 external users sign up
✅ See real requests in logs
```

---

## Commands Quick Reference

```bash
# Development
wrangler dev

# Deploy
wrangler deploy

# View logs
wrangler tail

# Check database
wrangler d1 shell corsproxy
> SELECT * FROM users;

# Create new database (if needed)
wrangler d1 create corsproxy-prod

# Execute SQL file
wrangler d1 execute corsproxy --file=schema.sql
```

---

## You're Done! 🎉

Total time: 30-60 minutes
Total cost: $0
Monthly operating cost: $0-50 (for free tier)

Now go ship it and get users!

Once you have real usage data, we can add paid tiers and watch the revenue roll in.

Good luck! 🚀
