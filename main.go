package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"endless/store"
	"endless/train"

	"github.com/gorilla/mux"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePostResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Post    *store.Post `json:"post,omitempty"`
}

type CreateMarkovModelRequest struct {
	Success bool                    `json:"success"`
	Error   string                  `json:"error,omitempty"`
	Model   *store.MarkovChainModel `json:"model,omitempty"`
}

type App struct {
	store store.PostStore
}

func main() {
	// Initialize database store
	postStore, err := store.NewSQLiteStore("./posts.db")
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
	// r.HandleFunc("/slow", app.slowHomeHandler).Methods("GET")
	r.HandleFunc("/api/posts", app.getPostsHandler).Methods("GET")
	r.HandleFunc("/api/posts", app.createPostHandler).Methods("POST")
	r.HandleFunc("/api/posts/{id}", app.getPostHandler).Methods("GET")
	r.HandleFunc("/api/train", app.trainMarkovModelHandler).Methods("POST")
	r.HandleFunc("/api/train/{id}", app.updateMarkovModelHandler).Methods("PUT")
	r.HandleFunc("/post/{id}", app.generatePageHandler).Methods("GET")

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send the HTML header and styles first
	headerHTML := `<!DOCTYPE html>
<html>
  <head>
    <title>Go SQLite Web Server</title>
    <style>
      body { font-family: Arial, sans-serif; margin: 40px; }
      .post { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
      .title { color: #333; font-size: 1.2em; }
      .content { color: #666; margin-top: 10px; }
      .date { color: #999; font-size: 0.8em; }
      .form { margin: 20px 0; padding: 20px; border: 1px solid #ccc; border-radius: 5px; }
      .form input, .form textarea { width: 100%; margin: 5px 0; padding: 8px; }
      .form button { background: #007cba; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
    </style>
  </head>
  <body>
    <h1>Posts from SQLite Database</h1>
    <div class="form">
      <h3>Add New Post</h3>
      <input type="text" id="title" placeholder="Post Title" />
      <textarea id="content" placeholder="Post Content" rows="4"></textarea>
      <button onclick="addPost()">Add Post</button>
    </div>
    <div id="posts">`

	w.Write([]byte(headerHTML))
	w.(http.Flusher).Flush()

	// Stream posts from database
	posts, err := app.store.GetAllPosts()
	if err != nil {
		w.Write([]byte("<p>Error loading posts: " + err.Error() + "</p>"))
		return
	}

	for _, post := range posts {
		// Stream each post as it's read
		postHTML := fmt.Sprintf(`
			<div class="post">
				<div class="title">%s</div>
				<div class="content">%s</div>
				<div class="date">%s</div>
			</div>`, post.Title, post.Content, post.CreatedAt)

		w.Write([]byte(postHTML))
		w.(http.Flusher).Flush()
	}

	// Send the closing HTML and JavaScript
	footerHTML := `
    </div>
    <script>
      function addPost() {
        var title = document.getElementById("title").value;
        var content = document.getElementById("content").value;
        if (!title || !content) {
          alert("Please fill in both title and content");
          return;
        }
        fetch("/api/posts", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ title: title, content: content }),
        })
        .then((response) => response.json())
        .then((data) => {
          if (data.success) {
            document.getElementById("title").value = "";
            document.getElementById("content").value = "";
            location.reload();
          } else {
            alert("Error: " + data.error);
          }
        })
        .catch((error) => alert("Error: " + error));
      }
    </script>
  </body>
</html>`

	w.Write([]byte(footerHTML))
}

func (app *App) slowHomeHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send the HTML header and styles first
	headerHTML := `<!DOCTYPE html>
<html>
  <head>
    <title>Go SQLite Web Server</title>
    <style>
      body { font-family: Arial, sans-serif; margin: 40px; }
      .post { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
      .title { color: #333; font-size: 1.2em; }
      .content { color: #666; margin-top: 10px; }
      .date { color: #999; font-size: 0.8em; }
      .form { margin: 20px 0; padding: 20px; border: 1px solid #ccc; border-radius: 5px; }
      .form input, .form textarea { width: 100%; margin: 5px 0; padding: 8px; }
      .form button { background: #007cba; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
    </style>
  </head>
  <body>
    <h1>Posts from SQLite Database</h1>
    <div class="form">
      <h3>Add New Post</h3>
      <input type="text" id="title" placeholder="Post Title" />
      <textarea id="content" placeholder="Post Content" rows="4"></textarea>
      <button onclick="addPost()">Add Post</button>
    </div>
    <div id="posts">`

	w.Write([]byte(headerHTML))
	w.(http.Flusher).Flush()

	// Stream posts from database
	posts, err := app.store.GetAllPosts()
	if err != nil {
		w.Write([]byte("<p>Error loading posts: " + err.Error() + "</p>"))
		return
	}

	for _, post := range posts {
		// Stream each post as it's read
		postHTML := fmt.Sprintf(`
			<div class="post">
				<div class="title">%s</div>
				<div class="content">%s</div>
				<div class="date">%s</div>
			</div>`, post.Title, post.Content, post.CreatedAt)

		w.Write([]byte(postHTML))
		w.(http.Flusher).Flush()
		time.Sleep(3 * time.Second)
	}

	// Send the closing HTML and JavaScript
	footerHTML := `
    </div>
    <script>
      function addPost() {
        var title = document.getElementById("title").value;
        var content = document.getElementById("content").value;
        if (!title || !content) {
          alert("Please fill in both title and content");
          return;
        }
        fetch("/api/posts", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ title: title, content: content }),
        })
        .then((response) => response.json())
        .then((data) => {
          if (data.success) {
            document.getElementById("title").value = "";
            document.getElementById("content").value = "";
            location.reload();
          } else {
            alert("Error: " + data.error);
          }
        })
        .catch((error) => alert("Error: " + error));
      }
    </script>
  </body>
</html>`

	w.Write([]byte(footerHTML))
}

func (app *App) getPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.store.GetAllPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (app *App) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CreatePostResponse{
			Success: false,
			Error:   "Invalid JSON",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Title == "" || req.Content == "" {
		response := CreatePostResponse{
			Success: false,
			Error:   "Title and content are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	post, err := app.store.SavePost(req.Title, req.Content)
	if err != nil {
		response := CreatePostResponse{
			Success: false,
			Error:   "Database error: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreatePostResponse{
		Success: true,
		Post:    post,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (app *App) getPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := app.store.GetPost(id)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
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

	// Return success response
	response := CreateMarkovModelRequest{
		Success: true,
		Model:   updatedModel,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (app *App) generatePageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the first available model from the database
	models, err := app.store.GetAllMarkovChainModels()
	if err != nil {
		http.Error(w, "Failed to retrieve models: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(models) == 0 {
		http.Error(w, "No models found in database", http.StatusNotFound)
		return
	}

	// Use the first (most recent) model
	model := models[0]

	// Load the model from JSON data
	chain, err := train.LoadModel([]byte(model.ModelData))
	if err != nil {
		http.Error(w, "Failed to load model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//get the {id} from the url
	vars := mux.Vars(r)
	// exaple 123-this-is-a-post-title
	idStr := strings.SplitN(vars["id"], "-", 2)[0]
	//it should support parsing int64
	seed, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generate story with "hello world" as seed
	// seedInt := int64(42)
	seedInt := seed
	story, err := train.GeneratePage(seedInt, chain)
	if err != nil {
		http.Error(w, "Failed to generate page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create basic HTML page
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
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
        <h1 class="title">%s</h1>
        <div class="content">%s</div>
        <div class="links-section">
            <h2 class="links-title">Related Stories</h2>
            <ul class="links-list">`, story.Link.Title, story.Link.Title, story.Content)

	// Add links to the HTML
	for _, link := range story.Links {
		htmlContent += fmt.Sprintf(`
                <li><a href="%s">%s</a></li>`, link.Url, link.Title)
	}

	// Close the HTML
	htmlContent += `
            </ul>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}
