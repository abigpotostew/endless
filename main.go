package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/abigpotostew/endless/routes"
	"github.com/abigpotostew/endless/store"
	"github.com/abigpotostew/endless/train"

	"github.com/gorilla/mux"
)

type CreateMarkovModelRequest struct {
	Success bool                    `json:"success"`
	Error   string                  `json:"error,omitempty"`
	Model   *store.MarkovChainModel `json:"model,omitempty"`
}

type App struct {
	store       store.PostStore
	cachedModel *store.MarkovChainModel
}

const statsHtml = `<script data-goatcounter="https://stats.stewart.codes/count"
        async src="//stats.stewart.codes/count.js"></script>`

func main() {
	sqliteDbPath := os.Getenv("SQLITE_DB_DIR")
	if sqliteDbPath == "" {
		sqliteDbPath = "."
	}
	sqliteDbPath = filepath.Join(sqliteDbPath, "endless.db")
	// Initialize database store
	postStore, err := store.NewSQLiteStore(sqliteDbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer postStore.Close()

	// Test database connection
	err = postStore.Ping()
	if err != nil {
		log.Fatal(err)
	}

	app := &App{store: postStore}

	// Setup router
	r := mux.NewRouter()

	// Add logging middleware to all routes
	r.Use(routes.LoggingMiddleware)

	// Serve static files
	r.HandleFunc("/", app.homeHandler).Methods("GET")
	r.HandleFunc("/sitemap.xml", app.sitemapHandler).Methods("GET")
	r.HandleFunc("/robots.txt", app.robotsHandler).Methods("GET")
	r.HandleFunc("/post/{id}", app.generatePageStreamHandler).Methods("GET")
	// need to restrict these to only allow requests from localhost
	r.HandleFunc("/health", app.healthHandler).Methods("GET").Host("localhost")
	r.HandleFunc("/api/train", app.trainMarkovModelHandler).Methods("POST").Host("localhost")
	r.HandleFunc("/api/train/{id}", app.updateMarkovModelHandler).Methods("PUT").Host("localhost")

	// Start server
	//accept port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server starting on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for HTML response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get the latest model using cache
	model, err := app.getLatestModel()
	if err != nil {
		http.Error(w, "Failed to retrieve model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load the model from JSON data
	chain, err := train.LoadModel([]byte(model.ModelData))
	if err != nil {
		http.Error(w, "Failed to load model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate 12 posts for the grid (3x4 layout)
	posts, err := train.GenerateHomePagePosts(chain, 12)
	if err != nil {
		http.Error(w, "Failed to generate posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the HTML header with SEO meta tags
	headerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Endless Stories - Daily Collection</title>
    
    <!-- SEO Meta Tags -->
    <meta name="description" content="Discover endless stories generated daily. A collection of unique narratives created with AI-powered Markov chains.">
    <meta name="keywords" content="stories, fiction, narrative, creative writing, AI generated, markov chain, endless stories">
    <meta name="author" content="Endless Stories">
    <meta name="robots" content="index, follow">
    <meta name="language" content="English">
    <meta name="revisit-after" content="1 day">
    <meta name="distribution" content="global">
    <meta name="rating" content="general">
    
    <!-- Open Graph / Facebook -->
    <meta property="og:type" content="website">
    <meta property="og:url" content="` + html.EscapeString(getFullURL(r)) + `">
    <meta property="og:title" content="Endless Stories - Daily Collection">
    <meta property="og:description" content="Discover endless stories generated daily. A collection of unique narratives created with AI-powered Markov chains.">
    <meta property="og:site_name" content="Endless Stories">
    <meta property="og:locale" content="en_US">
    
    <!-- Twitter -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="Endless Stories - Daily Collection">
    <meta name="twitter:description" content="Discover endless stories generated daily. A collection of unique narratives created with AI-powered Markov chains.">
    <meta name="twitter:site" content="@endlessstories">
    
    <!-- Canonical URL -->
    <link rel="canonical" href="` + html.EscapeString(getFullURL(r)) + `">
    
    <!-- Favicon -->
    <link rel="icon" type="image/x-icon" href="/favicon.ico">
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
    
    <!-- Structured Data (JSON-LD) -->
    <script type="application/ld+json">
    {
        "@context": "https://schema.org",
        "@type": "WebSite",
        "name": "Endless Stories",
        "description": "Discover endless stories generated daily. A collection of unique narratives created with AI-powered Markov chains.",
        "url": "` + html.EscapeString(getFullURL(r)) + `",
        "publisher": {
            "@type": "Organization",
            "name": "Endless Stories",
            "logo": {
                "@type": "ImageObject",
                "url": "` + html.EscapeString(getFullURL(r)) + `/logo.png"
            }
        },
        "potentialAction": {
            "@type": "SearchAction",
            "target": "` + html.EscapeString(getFullURL(r)) + `/search?q={search_term_string}",
            "query-input": "required name=search_term_string"
        }
    }
    </script>
    
    <!-- Additional SEO Meta Tags -->
    <meta name="theme-color" content="#007cba">
    <meta name="msapplication-TileColor" content="#007cba">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="default">
    <meta name="apple-mobile-web-app-title" content="Endless Stories">
    
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
            background-color: #f5f5f5;
        }
        
        .header {
            text-align: center;
            margin-bottom: 40px;
            padding: 20px;
            background: linear-gradient(135deg, #007cba, #005a87);
            color: white;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        
        .header p {
            margin: 10px 0 0 0;
            font-size: 1.1em;
            opacity: 0.9;
        }
        
        .posts-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        
        .post-card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            transition: transform 0.2s ease, box-shadow 0.2s ease;
            text-decoration: none;
            color: inherit;
            display: block;
        }
        
        .post-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
        }
        
        .post-title {
            font-size: 1.3em;
            font-weight: bold;
            color: #333;
            margin-bottom: 10px;
            line-height: 1.3;
        }
        
        .post-excerpt {
            color: #666;
            font-size: 0.9em;
            line-height: 1.5;
            margin-bottom: 15px;
            display: -webkit-box;
            -webkit-line-clamp: 3;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }
        
        .post-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 0.8em;
            color: #888;
        }
        
        .post-author {
            font-weight: bold;
            color: #007cba;
        }
        
        .post-date {
            font-style: italic;
        }
        
        .footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            color: #666;
            font-size: 0.9em;
        }
        
        .refresh-info {
            background: #e8f4fd;
            border: 1px solid #007cba;
            border-radius: 5px;
            padding: 15px;
            margin-bottom: 20px;
            text-align: center;
            color: #005a87;
        }
        
        @media (max-width: 768px) {
            .posts-grid {
                grid-template-columns: 1fr;
            }
            
            .header h1 {
                font-size: 2em;
            }
        }
    </style>
	` + statsHtml + `
</head>
<body>
    <div class="header">
        <h1>Endless Stories</h1>
        <p>Discover unique narratives added daily by world class writers</p>
    </div>
    
    <div class="refresh-info">
        <strong>New stories added daily!</strong> The collection refreshes every day at midnight.
    </div>
    
    <div class="posts-grid">`

	w.Write([]byte(headerHTML))
	w.(http.Flusher).Flush()

	// Stream each post card
	for _, post := range posts {
		// Create excerpt from content (first 150 characters)
		excerpt := truncateString(post.Content, 150)

		postCard := `
        <a href="` + html.EscapeString(post.Link.Url) + `" class="post-card">
            <h2 class="post-title">` + html.EscapeString(post.Link.Title) + `</h2>
            <p class="post-excerpt">` + html.EscapeString(excerpt) + `</p>
            <div class="post-meta">
                <span class="post-author">` + html.EscapeString(post.Author) + `</span>
                <span class="post-date">` + post.LastUpdated.Format("Jan 2, 2006") + `</span>
            </div>
        </a>`

		w.Write([]byte(postCard))
		w.(http.Flusher).Flush()

		// Add a small delay for streaming effect
		time.Sleep(50 * time.Millisecond)
	}

	// Send the closing HTML
	footerHTML := `
    </div>
    
    <div class="footer">
        <p>Stories written daily • Explore unique narratives</p>
    </div>
</body>
</html>`

	w.Write([]byte(footerHTML))
	w.(http.Flusher).Flush()
}

// getLatestModel returns the latest model, using cache if available
func (app *App) getLatestModel() (*store.MarkovChainModel, error) {
	// Return cached model if available
	if app.cachedModel != nil {
		return app.cachedModel, nil
	}

	// Get the first available model from the database
	models, err := app.store.GetAllMarkovChainModels(1)
	if err != nil {
		log.Printf("Error retrieving models from database: %v", err)
		return nil, err
	}

	if len(models) == 0 {
		log.Printf("No models found in database - this is likely the cause of 404 errors")
		return nil, fmt.Errorf("no models found in database")
	}

	// Cache the first (most recent) model
	app.cachedModel = &models[0]
	log.Printf("Retrieved and cached model ID: %d", app.cachedModel.ID)
	return app.cachedModel, nil
}

// clearModelCache clears the cached model
func (app *App) clearModelCache() {
	app.cachedModel = nil
}

func (app *App) trainMarkovModelHandler(w http.ResponseWriter, r *http.Request) {
	// Read the plain text body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to read request body: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	// Check if body is empty
	if len(body) == 0 {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Request body cannot be empty",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert body to string for processing
	inputText := string(body)

	// Build the markov chain model
	chain, err := train.BuildModel(inputText)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to build model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Serialize the model to JSON
	modelData, err := train.SerializeModel(chain)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to serialize model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Save the model to the database
	model, err := app.store.SaveMarkovChainModel(modelData)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to save model to database: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Clear the cache since we have a new model
	app.clearModelCache()

	// Return success response
	response := CreateMarkovModelRequest{
		Success: true,
		Model:   model,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (app *App) updateMarkovModelHandler(w http.ResponseWriter, r *http.Request) {
	// Get the model ID from the URL
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Invalid model ID: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Read the plain text body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to read request body: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer r.Body.Close()

	// Check if body is empty
	if len(body) == 0 {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Request body cannot be empty",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert body to string for processing
	additionalText := string(body)

	// Get the existing model from the database
	existingModel, err := app.store.GetMarkovChainModel(id)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to retrieve model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Load the existing model from JSON data
	chain, err := train.LoadModel([]byte(existingModel.ModelData))
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to load existing model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Add the additional text to the existing model
	err = train.AddTextToModel(chain, additionalText)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to add text to model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Serialize the updated model to JSON
	modelData, err := train.SerializeModel(chain)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to serialize updated model: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update the model in the database
	updatedModel, err := app.store.UpdateMarkovChainModel(id, modelData)
	if err != nil {
		response := CreateMarkovModelRequest{
			Success: false,
			Error:   "Failed to update model in database: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Clear the cache since the model was updated
	app.clearModelCache()

	// Return success response
	response := CreateMarkovModelRequest{
		Success: true,
		Model:   updatedModel,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *App) generatePageStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Get the {id} from the url
	vars := mux.Vars(r)
	// example 123-this-is-a-post-title
	idStr := strings.SplitN(vars["id"], "-", 2)[0]
	// it should support parsing int64
	seed, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Invalid ID in URL %s: %v", r.URL.Path, err)
		http.Error(w, "Invalid ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	streamPage(w, r, seed, app)
}

// Helper function to truncate strings for meta descriptions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Try to truncate at a word boundary
	truncated := s[:maxLen]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLen*3/4 { // Only use word boundary if it's not too far back
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}

// Helper function to get the full URL for canonical and Open Graph tags
func getFullURL(r *http.Request) string {
	host := os.Getenv("PUBLIC_HOST")
	if host == "" {

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		host = scheme + "://" + r.Host
	}

	return host + r.URL.Path
}

func streamPage(w http.ResponseWriter, r *http.Request, seedInput int64, app *App) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Initialize random seed for jitter
	prng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get the latest model using cache
	model, err := app.getLatestModel()
	if err != nil {
		http.Error(w, "Failed to retrieve model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load the model from JSON data
	chain, err := train.LoadModel([]byte(model.ModelData))
	if err != nil {
		http.Error(w, "Failed to load model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate story with the seed
	story, err := train.GeneratePage(seedInput, chain)
	if err != nil {
		http.Error(w, "Failed to generate page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	words := strings.Fields(story.Content)
	wordDelay := 50 * time.Millisecond

	linkWordDelay := wordDelay

	// Helper function to add jitter to delays
	addJitter := func(baseDelay time.Duration) time.Duration {
		// Add ±30% jitter
		jitterRange := float64(baseDelay) * 0.3
		jitter := (prng.Float64()*2 - 1) * jitterRange // Random value between -0.3 and +0.3
		return baseDelay + time.Duration(jitter)
	}

	// Send the HTML header and styles first
	headerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + html.EscapeString(story.Link.Title) + `</title>
    
    <!-- SEO Meta Tags -->
    <meta name="description" content="` + html.EscapeString(truncateString(story.Content, 160)) + `">
    <meta name="keywords" content="story, fiction, narrative, creative writing, ` + html.EscapeString(story.Author) + `">
    <meta name="author" content="` + html.EscapeString(story.Author) + `">
    <meta name="robots" content="index, follow">
    <meta name="language" content="English">
    <meta name="revisit-after" content="7 days">
    <meta name="distribution" content="global">
    <meta name="rating" content="general">
    
    <!-- Open Graph / Facebook -->
    <meta property="og:type" content="article">
    <meta property="og:url" content="` + html.EscapeString(getFullURL(r)) + `">
    <meta property="og:title" content="` + html.EscapeString(story.Link.Title) + `">
    <meta property="og:description" content="` + html.EscapeString(truncateString(story.Content, 200)) + `">
    <meta property="og:site_name" content="Endless Stories">
    <meta property="og:locale" content="en_US">
    <meta property="article:author" content="` + html.EscapeString(story.Author) + `">
    <meta property="article:published_time" content="` + story.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + `">
    <meta property="article:modified_time" content="` + story.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + `">
    
    <!-- Twitter -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="` + html.EscapeString(story.Link.Title) + `">
    <meta name="twitter:description" content="` + html.EscapeString(truncateString(story.Content, 200)) + `">
    <meta name="twitter:site" content="@endlessstories">
    <meta name="twitter:creator" content="` + html.EscapeString(story.Author) + `">
    
    <!-- Canonical URL -->
    <link rel="canonical" href="` + html.EscapeString(getFullURL(r)) + `">
    
    <!-- Favicon -->
    <link rel="icon" type="image/x-icon" href="/favicon.ico">
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
    
    <!-- Structured Data (JSON-LD) -->
    <script type="application/ld+json">
    {
        "@context": "https://schema.org",
        "@type": "Article",
        "headline": "` + html.EscapeString(story.Link.Title) + `",
        "description": "` + html.EscapeString(truncateString(story.Content, 200)) + `",
        "image": "` + html.EscapeString(getFullURL(r)) + `/og-image.jpg",
        "author": {
            "@type": "Person",
            "name": "` + html.EscapeString(story.Author) + `"
        },
        "publisher": {
            "@type": "Organization",
            "name": "Endless Stories",
            "logo": {
                "@type": "ImageObject",
                "url": "` + html.EscapeString(getFullURL(r)) + `/logo.png"
            }
        },
        "datePublished": "` + story.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + `",
        "dateModified": "` + story.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + `",
        "mainEntityOfPage": {
            "@type": "WebPage",
            "@id": "` + html.EscapeString(getFullURL(r)) + `"
        },
        "wordCount": ` + strconv.Itoa(len(strings.Fields(story.Content))) + `,
        "articleSection": "Fiction",
        "keywords": "story, fiction, narrative, creative writing, ` + html.EscapeString(story.Author) + `"
    }
    </script>
	` + statsHtml + `
    
    <!-- Additional SEO Meta Tags -->
    <meta name="theme-color" content="#007cba">
    <meta name="msapplication-TileColor" content="#007cba">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="default">
    <meta name="apple-mobile-web-app-title" content="Endless Stories">
    
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
        }
        .story {
            background-color: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #007cba;
            margin: 20px 0;
        }
        .title {
            color: #333;
            font-size: 2em;
            text-align: center;
            margin-bottom: 10px;
            border-bottom: 2px solid #007cba;
            padding-bottom: 10px;
        }
        .last-updated {
            text-align: center;
            color: #666;
            font-size: 0.9em;
            font-style: italic;
            margin-bottom: 20px;
        }
        .author {
            text-align: center;
            color: #007cba;
            font-size: 1em;
            font-weight: bold;
            margin-bottom: 20px;
        }
        .content {
            font-size: 16px;
            color: #333;
            margin-bottom: 30px;
        }
        .links-section {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
        }
        .links-title {
            color: #333;
            font-size: 1.5em;
            margin-bottom: 15px;
        }
        .links-list {
            list-style: none;
            padding: 0;
        }
        .links-list li {
            margin: 10px 0;
        }
        .links-list a {
            color: #007cba;
            text-decoration: none;
            font-size: 16px;
            padding: 8px 12px;
            border: 1px solid #007cba;
            border-radius: 4px;
            display: inline-block;
            transition: background-color 0.3s, color 0.3s;
        }
        .links-list a:hover {
            background-color: #007cba;
            color: white;
        }
        
        /* SEO-friendly breadcrumb navigation */
        .breadcrumb {
            margin-bottom: 20px;
            font-size: 0.9em;
            color: #666;
        }
        .breadcrumb a {
            color: #007cba;
            text-decoration: none;
        }
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        
        /* Schema.org microdata support */
        .article-meta {
            border-top: 1px solid #eee;
            padding-top: 15px;
            margin-top: 20px;
            font-size: 0.8em;
            color: #666;
        }
    </style>
</head>
<body>
    <!-- Breadcrumb navigation for SEO -->
    <nav class="breadcrumb" aria-label="Breadcrumb">
        <a href="/">Home</a> &gt; 
        <span aria-current="page">` + html.EscapeString(story.Link.Title) + `</span>
    </nav>
    
    <article class="story" itemscope itemtype="https://schema.org/Article">
        <h1 class="title" itemprop="headline">`

	w.Write([]byte(headerHTML))
	w.(http.Flusher).Flush()

	// Stream the title character by character with jitter
	for _, char := range story.Link.Title {
		w.Write([]byte(html.EscapeString(string(char))))
		w.(http.Flusher).Flush()
		time.Sleep(addJitter(wordDelay / 3)) // Faster for individual characters
	}

	// Send the title closing and metadata
	metadataHTML := `</h1>
        <div class="last-updated" itemprop="dateModified" content="` + story.LastUpdated.Format("2006-01-02T15:04:05Z07:00") + `">Last updated: ` + story.LastUpdated.Format("January 2, 2006 at 3:04 PM") + `</div>
        <div class="author" itemprop="author" itemscope itemtype="https://schema.org/Person">
            <span itemprop="name">` + html.EscapeString(story.Author) + `</span>
        </div>
        <div class="content" itemprop="articleBody">`

	w.Write([]byte(metadataHTML))
	w.(http.Flusher).Flush()

	// Split content into words and stream them
	for i, word := range words {
		// Add space before word (except for first word)
		if i > 0 {
			w.Write([]byte(" "))
		}
		w.Write([]byte(html.EscapeString(word)))
		w.(http.Flusher).Flush()
		time.Sleep(addJitter(wordDelay))
	}

	// Send the content closing and links section opening
	linksStart := `</div>
        <div class="links-section">
            <h2 class="links-title">Related Stories</h2>
            <ul class="links-list" role="list">`

	w.Write([]byte(linksStart))
	w.(http.Flusher).Flush()

	// Stream links one by one with word-by-word streaming
	for _, link := range story.Links {
		// Start the list item and link opening
		w.Write([]byte(`
                <li role="listitem"><a href="` + html.EscapeString(link.Url) + `">`))
		w.(http.Flusher).Flush()

		// Stream the link title character by character
		for _, char := range link.Title {
			w.Write([]byte(html.EscapeString(string(char))))
			w.(http.Flusher).Flush()
			time.Sleep(addJitter(linkWordDelay / 3)) // Faster for individual characters
		}

		// Close the link and list item
		w.Write([]byte(`</a></li>`))
		w.(http.Flusher).Flush()
	}

	// Send the closing HTML
	footerHTML := `
            </ul>
        </div>
    </article>
</body>
</html>`

	w.Write([]byte(footerHTML))
	w.(http.Flusher).Flush()
}

func (app *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (app *App) sitemapHandler(w http.ResponseWriter, r *http.Request) {
	// Get the base URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host

	// Set content type for XML
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	// Get the latest model to generate some example posts for sitemap
	model, err := app.getLatestModel()
	if err != nil {
		// If no model available, just return homepage
		sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>` + baseURL + `/</loc>
        <lastmod>` + time.Now().Format("2006-01-02") + `</lastmod>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
</urlset>`
		w.Write([]byte(sitemapXML))
		return
	}

	// Load the model and generate some example posts
	chain, err := train.LoadModel([]byte(model.ModelData))
	if err != nil {
		// If model loading fails, just return homepage
		sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>` + baseURL + `/</loc>
        <lastmod>` + time.Now().Format("2006-01-02") + `</lastmod>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
</urlset>`
		w.Write([]byte(sitemapXML))
		return
	}

	// Generate 20 example posts for sitemap
	posts, err := train.GenerateHomePagePosts(chain, 20)
	if err != nil {
		// If post generation fails, just return homepage
		sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>` + baseURL + `/</loc>
        <lastmod>` + time.Now().Format("2006-01-02") + `</lastmod>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>
</urlset>`
		w.Write([]byte(sitemapXML))
		return
	}

	// Generate sitemap XML with homepage and posts
	sitemapXML := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    <url>
        <loc>` + baseURL + `/</loc>
        <lastmod>` + time.Now().Format("2006-01-02") + `</lastmod>
        <changefreq>daily</changefreq>
        <priority>1.0</priority>
    </url>`

	// Add post URLs
	for _, post := range posts {
		sitemapXML += `
    <url>
        <loc>` + baseURL + html.EscapeString(post.Link.Url) + `</loc>
        <lastmod>` + post.LastUpdated.Format("2006-01-02") + `</lastmod>
        <changefreq>monthly</changefreq>
        <priority>0.8</priority>
    </url>`
	}

	sitemapXML += `
</urlset>`

	w.Write([]byte(sitemapXML))
}

func (app *App) robotsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the base URL
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host

	// Set content type for text
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Generate robots.txt content
	robotsTxt := `User-agent: *
Allow: /
Disallow: /api/
Disallow: /health

User-agent: AI2Bot
User-agent: Ai2Bot-Dolma
User-agent: aiHitBot
User-agent: Amazonbot
User-agent: Andibot
User-agent: anthropic-ai
User-agent: Applebot
User-agent: Applebot-Extended
User-agent: bedrockbot
User-agent: Brightbot 1.0
User-agent: Bytespider
User-agent: CCBot
User-agent: ChatGPT-User
User-agent: Claude-SearchBot
User-agent: Claude-User
User-agent: Claude-Web
User-agent: ClaudeBot
User-agent: cohere-ai
User-agent: cohere-training-data-crawler
User-agent: Cotoyogi
User-agent: Crawlspace
User-agent: Diffbot
User-agent: DuckAssistBot
User-agent: EchoboxBot
User-agent: FacebookBot
User-agent: facebookexternalhit
User-agent: Factset_spyderbot
User-agent: FirecrawlAgent
User-agent: FriendlyCrawler
User-agent: Google-CloudVertexBot
User-agent: Google-Extended
User-agent: GoogleOther
User-agent: GoogleOther-Image
User-agent: GoogleOther-Video
User-agent: GPTBot
User-agent: iaskspider/2.0
User-agent: ICC-Crawler
User-agent: ImagesiftBot
User-agent: img2dataset
User-agent: ISSCyberRiskCrawler
User-agent: Kangaroo Bot
User-agent: meta-externalagent
User-agent: Meta-ExternalAgent
User-agent: meta-externalfetcher
User-agent: Meta-ExternalFetcher
User-agent: MistralAI-User/1.0
User-agent: MyCentralAIScraperBot
User-agent: NovaAct
User-agent: OAI-SearchBot
User-agent: omgili
User-agent: omgilibot
User-agent: Operator
User-agent: PanguBot
User-agent: Panscient
User-agent: panscient.com
User-agent: Perplexity-User
User-agent: PerplexityBot
User-agent: PetalBot
User-agent: PhindBot
User-agent: Poseidon Research Crawler
User-agent: QualifiedBot
User-agent: QuillBot
User-agent: quillbot.com
User-agent: SBIntuitionsBot
User-agent: Scrapy
User-agent: SemrushBot
User-agent: SemrushBot-BA
User-agent: SemrushBot-CT
User-agent: SemrushBot-OCOB
User-agent: SemrushBot-SI
User-agent: SemrushBot-SWA
User-agent: Sidetrade indexer bot
User-agent: TikTokSpider
User-agent: Timpibot
User-agent: VelenPublicWebCrawler
User-agent: Webzio-Extended
User-agent: wpbot
User-agent: YandexAdditional
User-agent: YandexAdditionalBot
User-agent: YouBot
Disallow: /

Sitemap: ` + baseURL + `/sitemap.xml`

	w.Write([]byte(robotsTxt))
}
