package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePostResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Post    *Post  `json:"post,omitempty"`
}

type App struct {
	db *sql.DB
}

func main() {
	// Initialize database
	db, err := sql.Open("sqlite3", "./posts.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	app := &App{db: db}

	// Initialize database with schema
	err = app.initDB()
	if err != nil {
		log.Fatal(err)
	}

	// Setup router
	r := mux.NewRouter()

	// Serve static files
	r.HandleFunc("/", app.homeHandler).Methods("GET")
	r.HandleFunc("/slow", app.slowHomeHandler).Methods("GET")
	r.HandleFunc("/api/posts", app.getPostsHandler).Methods("GET")
	r.HandleFunc("/api/posts", app.createPostHandler).Methods("POST")
	r.HandleFunc("/api/posts/{id}", app.getPostHandler).Methods("GET")

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func (app *App) initDB() error {
	schema, err := os.ReadFile("data/schema.sql")
	if err != nil {
		return err
	}

	_, err = app.db.Exec(string(schema))
	return err
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
	rows, err := app.db.Query("SELECT id, title, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		w.Write([]byte("<p>Error loading posts: " + err.Error() + "</p>"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			continue
		}

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
	rows, err := app.db.Query("SELECT id, title, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		w.Write([]byte("<p>Error loading posts: " + err.Error() + "</p>"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			continue
		}

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
	rows, err := app.db.Query("SELECT id, title, content, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
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

	result, err := app.db.Exec("INSERT INTO posts (title, content) VALUES (?, ?)", req.Title, req.Content)
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

	id, err := result.LastInsertId()
	if err != nil {
		response := CreatePostResponse{
			Success: false,
			Error:   "Failed to get inserted ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the created post
	var post Post
	err = app.db.QueryRow("SELECT id, title, content, created_at FROM posts WHERE id = ?", id).
		Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)
	if err != nil {
		response := CreatePostResponse{
			Success: false,
			Error:   "Failed to retrieve created post",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreatePostResponse{
		Success: true,
		Post:    &post,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (app *App) getPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var post Post
	err := app.db.QueryRow("SELECT id, title, content, created_at FROM posts WHERE id = ?", id).
		Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
