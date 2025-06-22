# Solution: 404 Errors on `/post/*` Pages

## Problem Identified ✅

The 404 errors on `/post/*` pages in your deployment are caused by **no trained model in the database**. The application requires a Markov chain model to generate content, and without one, the `getLatestModel()` function fails.

## Root Cause

Looking at the code in `main.go`:

```go
func (app *App) generatePageStreamHandler(w http.ResponseWriter, r *http.Request) {
    // ... URL parsing ...
    streamPage(w, r, seed, app)  // This calls getLatestModel() internally
}

func (app *App) getLatestModel() (*store.MarkovChainModel, error) {
    // ... cache check ...
    models, err := app.store.GetAllMarkovChainModels(1)
    if err != nil {
        return nil, err
    }
    if len(models) == 0 {
        return nil, fmt.Errorf("no models found in database")  // ← This causes the 404
    }
    // ... return model ...
}
```

When `getLatestModel()` returns an error, the `streamPage()` function fails, causing a 404 response.

## Solution ✅

### Step 1: Train a Model

Use the provided initialization script to train a model:

```bash
# For localhost testing
./init-db.sh

# For your deployment (replace with your URL)
./init-db.sh https://your-deployment-url.com
```

### Step 2: Verify the Fix

After training the model, test your deployment:

```bash
# Test the homepage (should show generated posts)
curl https://your-deployment-url.com/

# Test a post page (should generate content)
curl https://your-deployment-url.com/post/123-test
```

## Alternative Solutions

### Option 1: Manual API Call

If the script doesn't work, manually train a model:

```bash
curl -X POST \
  -H "Content-Type: text/plain" \
  -d "This is a sample story. It has multiple sentences. The model will learn from this text and generate new stories." \
  https://your-deployment-url.com/api/train
```

### Option 2: Docker Compose with Auto-Initialization

Modify your `docker-compose.yml`:

```yaml
services:
  endless:
    # ... existing config ...
    command: >
      sh -c "
        ./endless &
        sleep 5 &&
        curl -X POST -H 'Content-Type: text/plain' -d 'Sample training text here.' http://localhost:8080/api/train &&
        wait
      "
```

### Option 3: Emergency Fallback Page

If you need an immediate fix, modify `main.go` to return a default page instead of 404:

```go
func (app *App) generatePageStreamHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...

    model, err := app.getLatestModel()
    if err != nil {
        // Return default page instead of 404
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write([]byte(`<html><body><h1>Coming Soon</h1><p>Content is being generated...</p></body></html>`))
        return
    }

    streamPage(w, r, seed, app)
}
```

## Prevention

To prevent this issue in future deployments:

1. **Always train a model** before deploying to production
2. **Add health checks** that verify model availability
3. **Use the initialization script** in your deployment pipeline
4. **Monitor application logs** for "no models found" errors

## Testing Tools

Use the provided test script to verify your deployment:

```bash
# Test all endpoints
./test_streaming.sh https://your-deployment-url.com
```

## Expected Behavior After Fix

✅ Homepage (`/`) shows generated posts with links  
✅ Post pages (`/post/*`) generate unique content  
✅ Sitemap (`/sitemap.xml`) includes post URLs  
✅ Robots.txt (`/robots.txt`) allows crawling

## Debugging Commands

If you still have issues:

```bash
# Check application logs
docker logs your-container-name

# Test database connectivity
docker exec your-container-name sqlite3 /data/endless.db "SELECT COUNT(*) FROM markov_chain_model;"

# Verify API endpoints
curl -v https://your-deployment-url.com/health
```

The solution is simple: **train a model first, then deploy**. The application works perfectly once a model is available in the database.
