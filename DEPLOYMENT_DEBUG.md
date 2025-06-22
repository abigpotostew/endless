# Deployment Debugging Guide

## Issue: 404 Errors on `/post/*` Pages

The application works locally but returns 404s in deployment. This guide will help you identify and fix the issue.

## Root Cause Analysis

The most likely cause is **no trained model in the database**. The application requires a Markov chain model to generate pages. Without a model, the `getLatestModel()` function fails, causing 404 errors.

## Debugging Steps

### 1. Test Your Deployment

Run the test script against your deployment:

```bash
# Replace with your actual deployment URL
./test_streaming.sh https://your-deployment-url.com
```

### 2. Check Application Logs

Look for these error messages in your deployment logs:

- `"Failed to retrieve model: no models found in database"`
- `"Failed to load model"`
- `"Failed to generate page"`

### 3. Verify Database State

The application needs at least one trained model in the database. Check if your database has models:

```sql
-- If you can access the SQLite database directly
SELECT COUNT(*) FROM markov_chain_model;
```

## Solutions

### Solution 1: Train a Model via API (Recommended)

If your deployment allows localhost API access, train a model:

```bash
# Replace with your deployment URL
curl -X POST \
  -H "Content-Type: text/plain" \
  -d "This is a sample story. It contains multiple sentences. Each sentence will be processed by the Markov chain. The model will learn patterns from this text and generate new stories. This is how the system works." \
  http://your-deployment-url.com/api/train
```

### Solution 2: Initialize Database with Sample Data

Create a script to initialize your database with a trained model:

```bash
#!/bin/bash
# init-db.sh

# Sample training text (you can replace with your own)
TRAINING_TEXT="The quick brown fox jumps over the lazy dog. This is a sample story for training the Markov chain model. The model will learn from this text and generate new stories. Each sentence should end with proper punctuation. This ensures the model works correctly."

# Train the model
curl -X POST \
  -H "Content-Type: text/plain" \
  -d "$TRAINING_TEXT" \
  http://localhost:8080/api/train

echo "Database initialized with sample model"
```

### Solution 3: Docker Compose with Initialization

Modify your `docker-compose.yml` to include initialization:

```yaml
services:
  endless:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8890:8080"
    environment:
      - PORT=8080
      - SQLITE_DB_DIR=/data
    volumes:
      - endless_data:/data
    restart: unless-stopped
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--no-verbose",
          "--tries=1",
          "--spider",
          "http://localhost:8080/health",
        ]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    # Add initialization command
    command: >
      sh -c "
        ./endless &
        sleep 5 &&
        curl -X POST -H 'Content-Type: text/plain' -d 'This is a sample story for training. It has multiple sentences. The model will learn from this text.' http://localhost:8080/api/train &&
        wait
      "

volumes:
  endless_data:
    driver: local
```

## Common Deployment Issues

### 1. Environment Variables

Ensure these environment variables are set in your deployment:

- `SQLITE_DB_DIR`: Directory for SQLite database (default: current directory)
- `PORT`: Port to listen on (default: 8080)
- `PUBLIC_HOST`: Public hostname for canonical URLs

### 2. Database Permissions

Ensure the application has write permissions to the database directory:

```bash
# For Docker deployments
chmod 755 /data
chown appuser:appgroup /data
```

### 3. Reverse Proxy Configuration

If using nginx or another reverse proxy, ensure it forwards all paths correctly:

```nginx
location / {
    proxy_pass http://localhost:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 4. Health Check Issues

The health check endpoint is restricted to localhost. If your deployment uses external health checks, you may need to modify the restriction:

```go
// In main.go, line 64
r.HandleFunc("/health", app.healthHandler).Methods("GET") // Remove .Host("localhost")
```

## Testing After Fix

1. **Test homepage**: Should show generated posts
2. **Test post pages**: Should generate content based on seed
3. **Test API endpoints**: Should work for training models

## Monitoring

Add logging to track model availability:

```go
// In getLatestModel() function
log.Printf("Retrieved model ID: %d", model.ID)
```

## Emergency Fix

If you need a quick fix for production, you can temporarily modify the application to return a default page when no model is available:

```go
func (app *App) generatePageStreamHandler(w http.ResponseWriter, r *http.Request) {
    // Get the {id} from the url
    vars := mux.Vars(r)
    idStr := strings.SplitN(vars["id"], "-", 2)[0]
    seed, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Try to get model, if fails, return default page
    model, err := app.getLatestModel()
    if err != nil {
        // Return a default page instead of 404
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write([]byte(`<html><body><h1>Coming Soon</h1><p>Content is being generated...</p></body></html>`))
        return
    }

    streamPage(w, r, seed, app)
}
```

## Contact

If you continue to have issues, check:

1. Application logs for specific error messages
2. Database connectivity and permissions
3. Network configuration and firewall rules
4. Reverse proxy configuration
