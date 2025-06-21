# Go Web Server with SQLite Database

This project demonstrates how to create a web server in Go that serves HTTP content generated from a SQLite database, including a Markov chain text generator with streaming page rendering.

## Features

- **SQLite Database Integration**: Store and retrieve posts with automatic timestamps
- **Markov Chain Text Generation**: Generate stories using trained models
- **Streaming Page Rendering**: Watch content appear gradually over ~10 seconds
- **RESTful API**: Create posts and train models via HTTP endpoints
- **Real-time Content Generation**: Dynamic story generation with related links

## Prerequisites

Before starting, ensure you have the following installed:

- **Go** (version 1.19 or later) - [Download here](https://golang.org/dl/)
- **Git** - [Download here](https://git-scm.com/downloads)
- **SQLite3** - [Download here](https://www.sqlite.org/download.html)

## API Endpoints

### Posts

- `GET /` - Home page with post creation form
- `GET /api/posts` - Get all posts
- `POST /api/posts` - Create a new post
- `GET /api/posts/{id}` - Get a specific post

### Markov Chain Models

- `POST /api/train` - Train a new model with text data
- `PUT /api/train/{id}` - Update an existing model with additional text

### Generated Pages

- `GET /post/{id}` - View a generated page (instant load)
- `GET /post/{id}/stream` - View a generated page with streaming content (~8 seconds with natural jitter)

## Streaming Page Feature

The streaming page feature (`/post/{id}/stream`) provides a unique viewing experience where content appears gradually with natural timing variations:

1. **Title Animation**: Characters appear one by one over ~1.5 seconds with random jitter
2. **Content Streaming**: Words appear sequentially over ~5 seconds with natural variations
3. **Links Generation**: Related story links appear one by one over ~1.5 seconds with jitter

Total viewing time is approximately 8 seconds with ±30% jitter on each delay, creating a more natural and engaging "typewriter" effect that feels less mechanical and more human-like.

## Step-by-Step Instructions

### Step 1: Initialize the Go Module

Create a new directory for your project and initialize a Go module:

```bash
mkdir go-sqlite-webserver
cd go-sqlite-webserver
go mod init go-sqlite-webserver
```

### Step 2: Install Required Dependencies

Add the necessary Go packages for SQLite and web server functionality:

```bash
go get github.com/mattn/go-sqlite3
go get github.com/gorilla/mux
```

### Step 3: Create the Database Schema

Create a file called `schema.sql` with your database schema:

```sql
-- schema.sql
CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert some sample data
INSERT INTO posts (title, content) VALUES
    ('First Post', 'This is the content of the first post.'),
    ('Second Post', 'This is the content of the second post.'),
    ('Third Post', 'This is the content of the third post.');
```

### Step 4: Create the Main Application File

Create `main.go` with the following content:

```go
package main

import (
    "database/sql"
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "os"

    _ "github.com/mattn/go-sqlite3"
    "github.com/gorilla/mux"
)

type Post struct {
    ID        int    `json:"id"`
    Title     string `json:"title"`
    Content   string `json:"content"`
    CreatedAt string `json:"created_at"`
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
    r.HandleFunc("/api/posts", app.getPostsHandler).Methods("GET")
    r.HandleFunc("/api/posts/{id}", app.getPostHandler).Methods("GET")

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func (app *App) initDB() error {
    schema, err := os.ReadFile("schema.sql")
    if err != nil {
        return err
    }

    _, err = app.db.Exec(string(schema))
    return err
}

func (app *App) homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Go SQLite Web Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .post { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .title { color: #333; font-size: 1.2em; }
        .content { color: #666; margin-top: 10px; }
        .date { color: #999; font-size: 0.8em; }
    </style>
</head>
<body>
    <h1>Posts from SQLite Database</h1>
    <div id="posts"></div>

    <script>
        fetch('/api/posts')
            .then(response => response.json())
            .then(posts => {
                const postsDiv = document.getElementById('posts');
                posts.forEach(post => {
                    postsDiv.innerHTML += `
                        <div class="post">
                            <div class="title">${post.title}</div>
                            <div class="content">${post.content}</div>
                            <div class="date">${post.created_at}</div>
                        </div>
                    `;
                });
            });
    </script>
</body>
</html>`

    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(tmpl))
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
```

### Step 5: Create a Go Module File

Your `go.mod` file should look like this:

```go
module go-sqlite-webserver

go 1.19

require (
    github.com/gorilla/mux v1.8.0
    github.com/mattn/go-sqlite3 v1.14.16
)
```

### Step 6: Build and Run the Application

1. **Build the application:**

   ```bash
   go build -o webserver main.go
   ```

2. **Run the application:**

   ```bash
   ./webserver
   ```

   Or run directly with Go:

   ```bash
   go run main.go
   ```

### Step 7: Test the Application

1. **Open your web browser** and navigate to `http://localhost:8080`
2. **View the posts** displayed on the homepage
3. **Test the API endpoints:**
   - `http://localhost:8080/api/posts` - Get all posts (JSON)
   - `http://localhost:8080/api/posts/1` - Get specific post (JSON)

## Project Structure

After following all steps, your project should have this structure:

```
go-sqlite-webserver/
├── main.go
├── schema.sql
├── go.mod
├── go.sum
└── posts.db (created automatically)
```

## Features

- **Web Interface**: Clean HTML page displaying posts from the database
- **REST API**: JSON endpoints for retrieving posts
- **SQLite Database**: Lightweight, file-based database
- **Gorilla Mux**: Powerful HTTP router for Go
- **Template Rendering**: Dynamic HTML generation

## API Endpoints

- `GET /` - Homepage with posts display
- `GET /api/posts` - Returns all posts as JSON
- `GET /api/posts/{id}` - Returns a specific post as JSON

## Customization Ideas

1. **Add CRUD operations**: Implement POST, PUT, DELETE endpoints
2. **Add user authentication**: Implement login/logout functionality
3. **Add search functionality**: Search posts by title or content
4. **Add pagination**: Limit results and add page navigation
5. **Add categories/tags**: Organize posts with metadata
6. **Add file uploads**: Allow image uploads for posts

## Troubleshooting

### Common Issues

1. **SQLite driver not found**: Ensure you have the correct import and CGO enabled
2. **Port already in use**: Change the port number in `main.go` (line with `:8080`)
3. **Database file permissions**: Ensure the application has write permissions in the directory

### Building for Production

For production deployment, consider:

1. **Environment variables** for configuration
2. **HTTPS/TLS** for secure connections
3. **Database migrations** for schema changes
4. **Logging** with structured logging
5. **Monitoring** and health checks

## Next Steps

Once you have the basic application running, you can:

1. Add more database tables and relationships
2. Implement user authentication
3. Add form handling for creating/editing posts
4. Implement caching for better performance
5. Add unit tests for your handlers
6. Containerize the application with Docker

This provides a solid foundation for building web applications with Go and SQLite!
