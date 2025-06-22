# Endless

A Markov chain-based story generator that creates endless narratives with a beautiful grid-based home page showcasing daily generated stories.

## Features

- **Markov Chain Story Generation**: Uses trained models to generate coherent stories
- **Daily Story Grid**: Home page displays 12 unique stories generated daily using time-based seeds
- **Streaming HTML Output**: Real-time word-by-word story generation with visual effects
- **Database Storage**: SQLite backend for storing trained models
- **RESTful API**: Endpoints for training and updating models
- **SEO Optimized**: Comprehensive search engine optimization features
- **Responsive Design**: Beautiful grid layout that works on all devices

## How It Works

### Daily Story Generation

The home page generates 12 unique stories every day using a time-based seed system:

- **Daily Seed**: Uses `time.Now().Unix() / 86400` to create a consistent daily seed
- **Unique Posts**: Each post gets a unique seed derived from the daily seed
- **Consistent Experience**: Same stories appear throughout the day, refreshing at midnight
- **Deterministic**: Same seed always produces the same story

### Story Grid Layout

- **Responsive Grid**: 3x4 layout on desktop, single column on mobile
- **Card Design**: Each story displayed in an attractive card with hover effects
- **Story Excerpts**: First 150 characters of each story as preview
- **Author Attribution**: Each story attributed to a random author
- **Publication Dates**: Realistic dates within the last 2 years

## SEO Features

The application includes comprehensive SEO optimizations:

### Meta Tags

- **Description**: Auto-generated from story content (truncated to 160 characters)
- **Keywords**: Relevant tags including author name and content type
- **Author**: Story author attribution
- **Robots**: Proper indexing instructions

### Open Graph Tags

- **og:type**: Article type for social media sharing
- **og:title**: Story title for social previews
- **og:description**: Story excerpt for social descriptions
- **og:url**: Canonical URL for social sharing
- **og:site_name**: Site branding
- **article:author**: Author attribution
- **article:published_time**: Publication timestamp

### Twitter Card Tags

- **twitter:card**: Large image card format
- **twitter:title**: Story title
- **twitter:description**: Story excerpt
- **twitter:site**: Site Twitter handle
- **twitter:creator**: Author Twitter handle

### Structured Data (JSON-LD)

- **Article schema**: Complete article markup for search engines
- **Author schema**: Author information
- **Publisher schema**: Site organization details
- **Word count**: Content length information
- **Keywords**: Content categorization

### Technical SEO

- **Canonical URLs**: Prevents duplicate content issues
- **Breadcrumb Navigation**: Site structure for search engines
- **Semantic HTML**: Proper use of article, nav, and section elements
- **Schema.org Microdata**: Inline structured data attributes
- **Sitemap.xml**: Auto-generated sitemap for search engines
- **Robots.txt**: Crawling instructions for search bots
- **Mobile Optimization**: Responsive viewport meta tags
- **Favicon Support**: Site branding icons

### Social Media Optimization

- **Facebook Sharing**: Open Graph tags for Facebook
- **Twitter Sharing**: Twitter Card tags for Twitter
- **LinkedIn Sharing**: Professional network compatibility
- **Image Placeholders**: Social media image references

## API Endpoints

- `GET /` - Homepage with daily story grid
- `GET /post/{id}` - Generate story with specific seed
- `POST /api/train` - Train new Markov model (localhost only)
- `PUT /api/train/{id}` - Update existing model (localhost only)
- `GET /health` - Health check (localhost only)
- `GET /sitemap.xml` - SEO sitemap with homepage and example posts
- `GET /robots.txt` - SEO robots file

## Usage

1. **Start the server**:

   ```bash
   go run main.go
   ```

2. **View the home page**:

   ```bash
   curl http://localhost:8080/
   ```

   This will show a beautiful grid of 12 daily-generated stories.

3. **Train a model** (from localhost):

   ```bash
   curl -X POST http://localhost:8080/api/train \
     -H "Content-Type: text/plain" \
     -d "Your training text here..."
   ```

4. **Generate a specific story**:

   ```bash
   curl http://localhost:8080/post/123
   ```

5. **View SEO files**:
   ```bash
   curl http://localhost:8080/sitemap.xml
   curl http://localhost:8080/robots.txt
   ```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `SQLITE_DB_DIR` - Database directory (default: current directory)
- `PUBLIC_HOST` - Public hostname for canonical URLs (e.g., https://example.com)

## Development

The application uses:

- **Gorilla Mux** for routing
- **SQLite** for data storage
- **gomarkov** for Markov chain generation
- **HTML streaming** for real-time content delivery
- **CSS Grid** for responsive layout
- **Time-based seeding** for consistent daily generation

## Daily Story Generation Algorithm

```go
// Daily seed changes every day at midnight
baseSeed := time.Now().Unix() / 86400

// Each post gets a unique seed
for i := 0; i < 12; i++ {
    postSeed := baseSeed + int64(i*1000)
    // Generate story with postSeed
}
```

This ensures:

- **Consistency**: Same stories throughout the day
- **Uniqueness**: Each story has a different seed
- **Daily Refresh**: New stories every day at midnight
- **Deterministic**: Reproducible results for the same day

## SEO Best Practices Implemented

1. **Content Optimization**

   - Descriptive titles and meta descriptions
   - Proper heading hierarchy (H1, H2)
   - Semantic HTML structure
   - Word count and reading time indicators

2. **Technical SEO**

   - Fast loading with streaming content
   - Mobile-responsive design
   - Clean URL structure
   - Proper HTTP status codes

3. **Social Media SEO**

   - Rich social media previews
   - Optimized sharing cards
   - Author attribution
   - Publication timestamps

4. **Search Engine Optimization**
   - Structured data markup
   - Sitemap generation
   - Robots.txt configuration
   - Canonical URL handling

## Customization Ideas

1. **Grid Layout**: Adjust the number of stories or grid columns
2. **Story Categories**: Add tags or categories to stories
3. **User Interactions**: Add likes, shares, or comments
4. **Search Functionality**: Search through generated stories
5. **Story Archives**: View stories from previous days
6. **Author Pages**: Dedicated pages for each author
7. **Story Recommendations**: Suggest similar stories
8. **Export Options**: Download stories as PDF or EPUB
