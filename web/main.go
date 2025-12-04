package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	miniflux "miniflux.app/v2/client"
)

type App struct {
	client    *miniflux.Client
	templates *template.Template
}

type PageData struct {
	Categories     []*miniflux.Category
	Feeds          []*miniflux.Feed
	Entries        []*miniflux.Entry
	SelectedEntry  *miniflux.Entry
	SelectedFeed   int64
	SelectedCat    int64
	EntriesCount   int
	Density        string
	ActiveTitle    string
	OpenCategories map[int64]bool
}

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get Miniflux credentials from environment variables
	apiURL := os.Getenv("MINIFLUX_API_URL")
	apiKey := os.Getenv("MINIFLUX_API_KEY")

	if apiURL == "" || apiKey == "" {
		log.Fatal("MINIFLUX_API_URL and MINIFLUX_API_KEY environment variables must be set")
	}

	// Initialize Miniflux client
	client := miniflux.New(apiURL, apiKey)

	// Test connection
	_, err := client.Me()
	if err != nil {
		log.Printf("Failed to connect to Miniflux API: %v", err)
		log.Printf("API URL: %s", apiURL)
		log.Fatalf("Please verify your MINIFLUX_API_URL and MINIFLUX_API_KEY in .env file")
	}

	log.Println("Successfully connected to Miniflux API")

	// Create template functions
	funcMap := template.FuncMap{
		"formatDate": formatDate,
		"safeHTML":   func(s string) template.HTML { return template.HTML(s) },
	}

	// Parse templates
	templates, err := template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	app := &App{
		client:    client,
		templates: templates,
	}

	// Set up routes
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/entry/", app.handleEntry)
	http.HandleFunc("/mark-read/", app.handleMarkRead)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// getDensity reads the density cookie value, defaults to "1"
func getDensity(r *http.Request) string {
	cookie, err := r.Cookie("density")
	if err != nil || cookie.Value == "" {
		return "1"
	}
	return cookie.Value
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	categoryIDStr := r.URL.Query().Get("category")
	feedIDStr := r.URL.Query().Get("feed")

	var categoryID, feedID int64
	var err error

	if categoryIDStr != "" {
		categoryID, err = strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
	}

	if feedIDStr != "" {
		feedID, err = strconv.ParseInt(feedIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid feed ID", http.StatusBadRequest)
			return
		}
	}

	// Fetch categories and feeds
	categories, err := app.client.Categories()
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		log.Printf("Error fetching categories: %v", err)
		return
	}

	feeds, err := app.client.Feeds()
	if err != nil {
		http.Error(w, "Failed to fetch feeds", http.StatusInternalServerError)
		log.Printf("Error fetching feeds: %v", err)
		return
	}

	// Fetch entries based on filter
	var entriesResult *miniflux.EntryResultSet
	filter := &miniflux.Filter{
		Status:    miniflux.EntryStatusUnread,
		Limit:     100,
		Order:     "published_at",
		Direction: "desc",
	}

	if feedID > 0 {
		entriesResult, err = app.client.FeedEntries(feedID, filter)
	} else if categoryID > 0 {
		entriesResult, err = app.client.CategoryEntries(categoryID, filter)
	} else {
		entriesResult, err = app.client.Entries(filter)
	}

	if err != nil {
		http.Error(w, "Failed to fetch entries", http.StatusInternalServerError)
		log.Printf("Error fetching entries: %v", err)
		return
	}

	// Determine active title and open categories
	activeTitle := "All Items"
	openCategories := make(map[int64]bool)

	if feedID > 0 {
		// Find feed title and its category
		for _, feed := range feeds {
			if feed.ID == feedID {
				activeTitle = feed.Title
				openCategories[feed.Category.ID] = true
				break
			}
		}
	} else if categoryID > 0 {
		// Find category title
		for _, cat := range categories {
			if cat.ID == categoryID {
				activeTitle = cat.Title
				openCategories[categoryID] = true
				break
			}
		}
	} else {
		// Open all categories for "All Items" view
		for _, cat := range categories {
			openCategories[cat.ID] = true
		}
	}

	// Prepare page data
	data := PageData{
		Categories:     categories,
		Feeds:          feeds,
		Entries:        entriesResult.Entries,
		SelectedCat:    categoryID,
		SelectedFeed:   feedID,
		EntriesCount:   entriesResult.Total,
		Density:        getDensity(r),
		ActiveTitle:    activeTitle,
		OpenCategories: openCategories,
	}

	// Render template
	err = app.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}

func (app *App) handleEntry(w http.ResponseWriter, r *http.Request) {
	// Get entry ID from URL
	entryIDStr := r.URL.Path[len("/entry/"):]
	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	// Get filter parameters to maintain context
	categoryIDStr := r.URL.Query().Get("category")
	feedIDStr := r.URL.Query().Get("feed")

	var categoryID, feedID int64

	if categoryIDStr != "" {
		categoryID, _ = strconv.ParseInt(categoryIDStr, 10, 64)
	}

	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	// Fetch categories and feeds
	categories, err := app.client.Categories()
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	feeds, err := app.client.Feeds()
	if err != nil {
		http.Error(w, "Failed to fetch feeds", http.StatusInternalServerError)
		return
	}

	// Fetch entries based on filter
	var entriesResult *miniflux.EntryResultSet
	filter := &miniflux.Filter{
		Status:    miniflux.EntryStatusUnread,
		Limit:     100,
		Order:     "published_at",
		Direction: "desc",
	}

	if feedID > 0 {
		entriesResult, err = app.client.FeedEntries(feedID, filter)
	} else if categoryID > 0 {
		entriesResult, err = app.client.CategoryEntries(categoryID, filter)
	} else {
		entriesResult, err = app.client.Entries(filter)
	}

	if err != nil {
		http.Error(w, "Failed to fetch entries", http.StatusInternalServerError)
		return
	}

	// Fetch selected entry
	entry, err := app.client.Entry(entryID)
	if err != nil {
		http.Error(w, "Failed to fetch entry", http.StatusInternalServerError)
		log.Printf("Error fetching entry: %v", err)
		return
	}

	// Determine active title and open categories
	activeTitle := "All Items"
	openCategories := make(map[int64]bool)

	if feedID > 0 {
		// Find feed title and its category
		for _, feed := range feeds {
			if feed.ID == feedID {
				activeTitle = feed.Title
				openCategories[feed.Category.ID] = true
				break
			}
		}
	} else if categoryID > 0 {
		// Find category title
		for _, cat := range categories {
			if cat.ID == categoryID {
				activeTitle = cat.Title
				openCategories[categoryID] = true
				break
			}
		}
	} else {
		// Open all categories for "All Items" view
		for _, cat := range categories {
			openCategories[cat.ID] = true
		}
	}

	// Prepare page data
	data := PageData{
		Categories:     categories,
		Feeds:          feeds,
		Entries:        entriesResult.Entries,
		SelectedEntry:  entry,
		SelectedCat:    categoryID,
		SelectedFeed:   feedID,
		EntriesCount:   entriesResult.Total,
		Density:        getDensity(r),
		ActiveTitle:    activeTitle,
		OpenCategories: openCategories,
	}

	// Render template
	err = app.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}

func (app *App) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get entry ID from URL
	entryIDStr := r.URL.Path[len("/mark-read/"):]
	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	// Mark entry as read
	err = app.client.UpdateEntries([]int64{entryID}, miniflux.EntryStatusRead)
	if err != nil {
		http.Error(w, "Failed to mark entry as read", http.StatusInternalServerError)
		log.Printf("Error marking entry as read: %v", err)
		return
	}

	// Get redirect parameters
	categoryIDStr := r.URL.Query().Get("category")
	feedIDStr := r.URL.Query().Get("feed")

	// Redirect back to index with filters
	redirectURL := "/"
	if categoryIDStr != "" {
		redirectURL += "?category=" + categoryIDStr
	} else if feedIDStr != "" {
		redirectURL += "?feed=" + feedIDStr
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// formatDate formats a time.Time for display
func formatDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return strconv.Itoa(mins) + " minutes ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return strconv.Itoa(hours) + " hours ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return strconv.Itoa(days) + " days ago"
	default:
		return t.Format("Jan 2, 2006")
	}
}
