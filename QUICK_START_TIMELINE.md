# 2-Week Launch Timeline

## Strategy
**Free tier only. Generous limits. Get users. Add billing later based on real demand.**

---

## Week 1: Build MVP

### Monday-Tuesday: Backend Setup

**Do:**
```bash
# 1. PostgreSQL setup (20 min)
# Go to Render.com → Create PostgreSQL instance
# Copy connection string → .env

# 2. Backend scaffolding (1 hour)
mkdir backend && cd backend
go mod init corsproxy
go get github.com/lib/pq github.com/joho/godotenv

# 3. Copy database.go, main.go from LAUNCH_CHECKLIST.md
# 4. Test locally
go run main.go
```

**Endpoints created:**
- POST `/api/signup` - Create user
- POST `/api/keys/create` - New API key
- GET `/api/keys` - List user's keys
- DELETE `/api/keys/delete` - Remove key
- POST `/api/keys/validate` - Check if key valid + limit check
- POST `/api/logs` - Log request

**Time: 2-3 hours**

### Wednesday: Frontend Setup

**Do:**
```bash
# 1. Create Next.js project (10 min)
npx create-next-app@latest frontend --typescript --tailwind
cd frontend

# 2. Copy page.tsx and dashboard/page.tsx from LAUNCH_CHECKLIST.md
# 3. Test locally
npm run dev

# Visit http://localhost:3000
```

**Pages created:**
- `/` - Landing page
- `/dashboard` - Main dashboard

**Time: 1-2 hours**

### Thursday-Friday: Integrate + Test

**Do:**
```bash
# 1. Test sign-up flow end-to-end
# Frontend → Backend API → Database

# 2. Test API key creation/deletion

# 3. Test requests via:
curl "http://localhost:8000/?url=https://api.github.com/users/melihbirim" \
  -H "Authorization: Bearer cors_xxx"

# 4. Fix bugs
```

**Time: 2-3 hours**

---

## Week 2: Deploy + Launch

### Monday: Deploy to Production

**Backend (Render):**
```bash
cd backend

# 1. Create GitHub repo
git init
git add .
git commit -m "Initial backend"
git remote add origin https://github.com/YOUR_USERNAME/corsproxy-backend
git push -u origin main

# 2. Go to Render.com
# → New Web Service
# → Connect GitHub repo
# → Set environment variables:
#    DATABASE_URL=postgresql://...
# → Deploy (auto on push)
```

**Frontend (Vercel):**
```bash
cd frontend

# 1. Create GitHub repo
git init
git add .
git commit -m "Initial frontend"
git remote add origin https://github.com/YOUR_USERNAME/corsproxy-frontend
git push -u origin main

# 2. Go to Vercel.com
# → Import GitHub repo
# → Auto-deploys in 30 seconds
```

**Update API URLs:**
```tsx
// frontend/app/api/signup.ts
// Change localhost:8000 to your Render backend URL
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'https://api.yourdomain.com'
```

**Time: 1 hour**

### Tuesday: Domain + Marketing Assets

**Domain:**
```
1. Buy short domain (cors.io, cors.me, free-cors.io)
   - Namecheap: $10/year
   - Point DNS to Vercel (for frontend)

2. Update API_BASE to point to Render backend

3. Update social links
```

**Marketing assets (30 min):**
- [x] Landing page (already in page.tsx)
- [x] Demo GIF (use Gifcap or ScreenFlow)
- [x] Twitter copy ("I built a free CORS proxy...")
- [x] HN Show HN title

**Time: 1-2 hours**

### Wednesday: Soft Launch (Tuesday Evening)

Test with 10 friends:
```
1. Send them link: https://yourname.com
2. Ask them to:
   - Sign up
   - Create API key
   - Try making a request
3. Record feedback
4. Fix critical bugs
```

**Time: 2 hours**

### Thursday-Friday: Public Launch

**Platform priority:**

**Morning (9-10am EST):**
- Post on Hacker News: "Show HN: Free CORS Proxy - No signup required, 1000 req/day"
- Link: https://yourname.com

**Noon:**
- Tweet announcement (tag @buildspace, @indiehackers if relevant)
- Example:
```
Just launched a free CORS proxy for indie developers 🚀
- 1000 requests/day free
- No backend needed
- Works from any frontend

Try it: https://cors.example.com/?url=https://api.example.com

What API did you always want to access from your frontend?
```

**Afternoon:**
- Post on Indie Hackers
- Post on Dev.to
- Share in relevant Discord servers

**Evening & Weekend:**
- Answer HN comments
- Fix bugs as reported
- Send thank you tweets to early users
- Share early wins: "10 signups in first hour!"

**Time: Ongoing**

---

## What You'll Have at End of Week 2

✅ Live SaaS product
✅ Real users trying it
✅ Database of usage patterns
✅ Feedback for improvement
✅ Zero payment complexity
✅ Data to decide: "Do people want more? Are they willing to pay?"

---

## Metrics Dashboard (Build This Week 2)

Simple backend endpoint to track:

```go
// GET /api/metrics (admin only for now)
{
  "total_signups": 145,
  "today_signups": 45,
  "total_keys": 212,
  "today_keys": 67,
  "total_requests": 1240000,
  "today_requests": 340000,
  "daily_active_users": 48,
  "avg_requests_per_user": 2604
}
```

Add to dashboard for you to monitor:
```tsx
// frontend/app/admin/page.tsx
fetch('/api/metrics')
  .then(r => r.json())
  .then(d => console.log(d))
```

---

## Critical Path - Don't Do These Yet

🚫 Stripe/billing integration
🚫 OAuth (email is fine)
🚫 Rate limiting per IP
🚫 Advanced analytics dashboard
🚫 Custom domains feature
🚫 Team collaboration
🚫 Self-hosted version
🚫 Fancy UI (Tailwind basic is enough)

These come AFTER you have real users asking for them.

---

## Post-Launch Priorities (Based on Usage Data)

**If 100+ users with 10k+ daily requests:**
→ Add billing ($5/month for 100k/month)
→ Email: "People loved it! Here's a paid plan"
→ Expect 2-5% conversion

**If low usage but high sign-ups:**
→ Something's wrong with onboarding
→ Interview users: "Why not using it?"
→ Fix based on feedback

**If high usage but no sign-ups:**
→ Landing page issue
→ Improve copy

**If flatline:**
→ Maybe CORS proxying isn't what people need
→ Pivot or double down on marketing
→ Use data to decide

---

## Success Definition

**Week 1 (by Friday):**
- ✅ Product works end-to-end
- ✅ Can sign up and create key
- ✅ Can make requests
- ✅ Backend + Frontend deployed

**Week 2 (by Friday):**
- ✅ Live on custom domain
- ✅ Launched on HN, Twitter, etc.
- ✅ 100+ signups
- ✅ 10+ daily active users
- ✅ Real usage patterns emerging

If you hit these, you've won. Everything else is iteration.

---

## Daily Standup Checklist

**Day 1 (Mon):** ⬜ Backend scaffolding
**Day 2 (Tue):** ⬜ DB setup + API endpoints
**Day 3 (Wed):** ⬜ Frontend scaffolding
**Day 4 (Thu):** ⬜ Integration testing
**Day 5 (Fri):** ⬜ Deploy to production
**Day 6 (Mon):** ⬜ Domain setup + marketing assets
**Day 7 (Tue):** ⬜ Soft launch with friends
**Day 8 (Wed):** ⬜ Hacker News launch
**Day 9 (Thu):** ⬜ Twitter/Social media
**Day 10 (Fri):** ⬜ Community posts

---

## One More Thing: Monitor Early

Add simple logging from day 1:

```go
// Log every signup, key creation, request
log.Printf("NEW_SIGNUP: %s", email)
log.Printf("NEW_KEY: user_id=%d, key_id=%s", userID, keyID)
log.Printf("REQUEST: key=%s, url=%s, status=%d", keyID, url, status)
```

Tail logs in real-time while launching. See what's happening. Celebrate wins. Fix issues immediately.

It's way more fun to watch real people use your product in real-time than to wonder "did anyone sign up?"

---

## Resources You'll Need

- Render.com (backend hosting) - Free
- PostgreSQL (database) - Free tier available
- Vercel.com (frontend) - Free
- Namecheap/Google Domains (domain) - $10/year
- GitHub (code hosting) - Free

**Total cost: $0 to launch, ~$10/year for domain**

---

## You've Got This

This is a doable 2-week sprint. The key is:
1. **Keep it simple** - No fancy stuff
2. **Ship fast** - Done is better than perfect
3. **Get real users** - Data beats opinions
4. **Add billing when demand is clear** - No guessing

Good luck! 🚀
