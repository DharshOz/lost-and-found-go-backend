package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"lostfound-backend/db"
	"lostfound-backend/routes"
	"lostfound-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get Mongo URI from env or fallback
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGODB_URI not set in environment variables")
	}

	// Connect to MongoDB
	db.ConnectMongoDB(mongoURI)
	defer db.DisconnectMongoDB()

	// Initialize Cloudinary
	utils.InitCloudinary()

	// Create Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // ← IMPORTANT
		MaxAge:           12 * time.Hour,
	}))

	// API Routes
	api := router.Group("/api")

	// Auth routes group
	auth := api.Group("/auth")
	{
		auth.POST("/signup", routes.Signup)
		auth.POST("/login", routes.Login)
		auth.GET("/check-session", routes.CheckSession)
	}

	// Protected routes (require auth)
	protected := api.Group("")
	protected.Use(routes.AuthMiddleware())
	{
		// User routes
		protected.GET("/auth/user/:id", routes.GetUserProfile) // Add this
		protected.PUT("/auth/user/:id", routes.UpdateUser)     // Add this

		// Lost items routes
		protected.POST("/lostitems", routes.AddLostItem)
		protected.GET("/lostitems", routes.GetAllLostItems)
		protected.GET("/lostitems/user/:userId", routes.GetLostItemsByUser)
		protected.GET("/lostitems/filters", routes.GetFilterOptions)
		protected.GET("/lostitems/:id", routes.GetLostItemByID)
		protected.DELETE("/lostitems/:id", routes.DeleteLostItem)

		// Found items routes
		protected.POST("/founditems", routes.AddFoundItem)
		protected.GET("/founditems", routes.GetAllFoundItems)
		protected.GET("/founditems/lostItem/:lostItemId", routes.GetFoundItemsByLostItem)
		protected.GET("/founditems/foundPerson/:foundPersonId", routes.GetFoundItemsByUser)
		protected.GET("/founditems/:id", routes.GetFoundItemByID)
		protected.PUT("/founditems/:id", routes.UpdateFoundItem)
		protected.PUT("/founditems/:id/found", routes.UpdateFoundStatus)
		protected.DELETE("/founditems/:id", routes.DeleteFoundItem)

		// Bookmark routes
		protected.POST("/bookmarks", routes.AddBookmark)
		protected.GET("/bookmarks", routes.GetBookmarks)
		protected.DELETE("/bookmarks/:id", routes.DeleteBookmark) // Add this
	}
	// Get port or fallback to 5000
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Printf("🚀 Server running on http://localhost:%s\n", port)
	log.Fatal(router.Run(":" + port))
}
