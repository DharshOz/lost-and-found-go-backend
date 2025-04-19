package routes

import (
	"context"
	"net/http"
	"time"

	"lostfound-backend/db"
	"lostfound-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddBookmark(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var request struct {
		LostItemID string `json:"lostItemId"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate and convert IDs
	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	lostItemObjectID, err := primitive.ObjectIDFromHex(request.LostItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lost item ID"})
		return
	}

	// Create bookmark
	bookmark := models.Bookmark{
		User:      userObjectID,
		LostItem:  lostItemObjectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert into database
	collection := db.GetCollection("bookmarks")
	_, err = collection.InsertOne(context.Background(), bookmark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bookmark"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bookmark added successfully",
		"data":    bookmark,
	})
}
func GetBookmarks(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert userID string to ObjectID
	uid, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := db.GetCollection("bookmarks")
	lookupStage := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "lostitems"},
			{Key: "localField", Value: "lostItem"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "lostItem"},
		}},
	}
	unwindStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$lostItem"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "user", Value: uid},
	}}}

	opts := options.Aggregate().SetMaxTime(5 * time.Second)
	cursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{matchStage, lookupStage, unwindStage}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
		return
	}

	c.JSON(http.StatusOK, results)
}
func DeleteBookmark(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bookmark ID required"})
		return
	}

	// Convert IDs to ObjectID
	bookmarkObjectID, err := primitive.ObjectIDFromHex(bookmarkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bookmark ID"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete the bookmark (only if it belongs to the user)
	collection := db.GetCollection("bookmarks")
	result, err := collection.DeleteOne(context.Background(), bson.M{
		"_id":  bookmarkObjectID,
		"user": userObjectID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bookmark"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bookmark not found or not owned by user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bookmark deleted successfully"})
}
