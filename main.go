package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"endless/store"
	"endless/train"

	"github.com/gorilla/mux"
	"github.com/pkg/profile"
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

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	// Initialize database store
	postStore, err := store.NewSQLiteStore("./endless.db")
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

	// Serve static files
	r.HandleFunc("/", app.homeHandler).Methods("GET")
	// need to restrict these to only allow requests from localhost
	r.HandleFunc("/api/train", app.trainMarkovModelHandler).Methods("POST").Host("localhost")
	r.HandleFunc("/api/train/{id}", app.updateMarkovModelHandler).Methods("PUT").Host("localhost")
	r.HandleFunc("/post/{id}", app.generatePageStreamHandler).Methods("GET").Host("localhost")

	// Start pprof server on a separate port (optional but recommended)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	streamPage(w, r, 0, app)
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
		return nil, err
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no models found in database")
	}

	// Cache the first (most recent) model
	app.cachedModel = &models[0]
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
		http.Error(w, "Invalid ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	streamPage(w, r, seed, app)
}

func streamPage(w http.ResponseWriter, r *http.Request, seedInput int64, app *App) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/html")
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

	// Calculate timing to achieve ~10 second total load time
	totalTargetTime := 8 * time.Second   // Reduced from 10 to 8 seconds
	titleTime := 2000 * time.Millisecond // 1.5 seconds
	linksTime := 2000 * time.Millisecond // Increased to 2 seconds for word-by-word streaming
	contentTime := totalTargetTime - titleTime - linksTime

	// Calculate delays based on content length
	titleChars := len(story.Link.Title)
	titleDelay := titleTime / time.Duration(titleChars)

	words := strings.Fields(story.Content)
	wordDelay := contentTime / time.Duration(len(words))

	// Calculate link timing - we'll stream each word of each link title
	totalLinkWords := 0
	for _, link := range story.Links {
		totalLinkWords += len(strings.Fields(link.Title))
	}
	linkWordDelay := linksTime / time.Duration(totalLinkWords)

	// Helper function to add jitter to delays
	addJitter := func(baseDelay time.Duration) time.Duration {
		// Add ±30% jitter
		jitterRange := float64(baseDelay) * 0.3
		jitter := (prng.Float64()*2 - 1) * jitterRange // Random value between -0.3 and +0.3
		return baseDelay + time.Duration(jitter)
	}

	// Send the HTML header and styles first
	headerHTML := `<!DOCTYPE html>
<html>
<head>
    <title>` + story.Link.Title + `</title>
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
            margin-bottom: 20px;
            border-bottom: 2px solid #007cba;
            padding-bottom: 10px;
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
    </style>
</head>
<body>
    <div class="story">
        <h1 class="title">`

	w.Write([]byte(headerHTML))
	w.(http.Flusher).Flush()

	// Stream the title character by character
	for _, char := range story.Link.Title {
		w.Write([]byte(string(char)))
		w.(http.Flusher).Flush()
		time.Sleep(addJitter(titleDelay))
	}

	// Send the title closing and content div opening
	contentStart := `</h1>
        <div class="content">`

	w.Write([]byte(contentStart))
	w.(http.Flusher).Flush()

	// Split content into words and stream them
	for i, word := range words {
		// Add space before word (except for first word)
		if i > 0 {
			w.Write([]byte(" "))
		}
		w.Write([]byte(word))
		w.(http.Flusher).Flush()
		time.Sleep(addJitter(wordDelay))
	}

	// Send the content closing and links section opening
	linksStart := `</div>
        <div class="links-section">
            <h2 class="links-title">Related Stories</h2>
            <ul class="links-list">`

	w.Write([]byte(linksStart))
	w.(http.Flusher).Flush()

	// Stream links one by one with word-by-word streaming
	for _, link := range story.Links {
		// Start the list item and link opening
		w.Write([]byte(`
                <li><a href="` + link.Url + `">`))
		w.(http.Flusher).Flush()

		// Stream the link title word by word
		linkWords := strings.Fields(link.Title)
		for i, word := range linkWords {
			// Add space before word (except for first word)
			if i > 0 {
				w.Write([]byte(" "))
			}
			w.Write([]byte(word))
			w.(http.Flusher).Flush()
			time.Sleep(addJitter(linkWordDelay))
		}

		// Close the link and list item
		w.Write([]byte(`</a></li>`))
		w.(http.Flusher).Flush()
	}

	// Send the closing HTML
	footerHTML := `
            </ul>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(footerHTML))
	w.(http.Flusher).Flush()
}
