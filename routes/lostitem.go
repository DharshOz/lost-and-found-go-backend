package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"lostfound-backend/db"
	"lostfound-backend/models"
	"lostfound-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddLostItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	// Upload to Cloudinary
	imageURL, err := utils.UploadImage(file)
	if err != nil || imageURL == "" {
		log.Println("UploadImage error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Handling locations as an array
	locationStrings := c.PostForm("locations")
	locations := []string{}
	if locationStrings != "" {
		locations = append(locations, locationStrings)
	}

	district := c.PostForm("district")
	state := c.PostForm("state")
	if district == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "District and state are required"})
		return
	}

	item := models.LostItem{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Category:    c.PostForm("category"),
		ImageURL:    imageURL,
		District:    district,
		State:       state,
		Locations:   locations,
		DateLost:    time.Now(),
		CreatedBy:   objID,
		CreatedAt:   time.Now(),
	}

	collection := db.GetCollection("lostitems")
	result, err := collection.InsertOne(context.Background(), item)
	if err != nil {
		log.Println("InsertOne error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Item created successfully",
		"itemId":   result.InsertedID,
		"imageURL": imageURL,
	})
}

func GetAllLostItems(c *gin.Context) {
	collection := db.GetCollection("lostitems")
	filter := bson.M{}

	// Add filters from query parameters
	if name := c.Query("name"); name != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: name, Options: "i"}}
	}
	if category := c.Query("category"); category != "" {
		filter["category"] = bson.M{"$regex": primitive.Regex{Pattern: category, Options: "i"}}
	}
	if district := c.Query("district"); district != "" {
		filter["district"] = bson.M{"$regex": primitive.Regex{Pattern: district, Options: "i"}}
	}
	if showOnlyUserItems := c.Query("userOnly"); showOnlyUserItems == "true" {
		if userID, exists := c.Get("userID"); exists {
			objID, err := primitive.ObjectIDFromHex(userID.(string))
			if err == nil {
				filter["createdBy"] = objID
			}
		}
	}

	// Add pagination options
	findOptions := options.Find()
	if limit := c.Query("limit"); limit != "" {
		limitInt, err := utils.StringToInt(limit)
		if err == nil {
			findOptions.SetLimit(int64(limitInt))
		}
	}
	if skip := c.Query("skip"); skip != "" {
		skipInt, err := utils.StringToInt(skip)
		if err == nil {
			findOptions.SetSkip(int64(skipInt))
		}
	}

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		log.Println("Find error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}
	defer cursor.Close(context.Background())

	var items []models.LostItem
	if err := cursor.All(context.Background(), &items); err != nil {
		log.Println("Cursor decode error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetLostItemByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := db.GetCollection("lostitems")
	var item models.LostItem
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&item)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Populate user information
	var user models.User
	userCollection := db.GetCollection("users")
	err = userCollection.FindOne(context.Background(), bson.M{"_id": item.CreatedBy}).Decode(&user)
	if err == nil {
		item.CreatedByUser = &models.User{
			Username: user.Username,
			Email:    user.Email,
		}
	}

	c.JSON(http.StatusOK, item)
}

func GetLostItemsByUser(c *gin.Context) {
	userID := c.Param("userId")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := db.GetCollection("lostitems")
	cursor, err := collection.Find(context.Background(), bson.M{"createdBy": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}
	defer cursor.Close(context.Background())

	var items []models.LostItem
	if err := cursor.All(context.Background(), &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetFilterOptions(c *gin.Context) {
	collection := db.GetCollection("lostitems")

	categories, err := collection.Distinct(context.Background(), "category", bson.M{})
	if err != nil {
		log.Println("Distinct category error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	districts, err := collection.Distinct(context.Background(), "district", bson.M{})
	if err != nil {
		log.Println("Distinct district error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch districts"})
		return
	}

	states, err := collection.Distinct(context.Background(), "state", bson.M{})
	if err != nil {
		log.Println("Distinct state error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch states"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"districts":  districts,
		"states":     states,
	})
}

func DeleteLostItem(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item ID is required"})
		return
	}

	// Get the authenticated user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID format"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	collection := db.GetCollection("lostitems")

	// First verify the item exists and belongs to the user
	var item models.LostItem
	err = collection.FindOne(context.Background(), bson.M{
		"_id":       objID,
		"createdBy": userObjID,
	}).Decode(&item)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or not owned by user"})
		return
	}

	// Delete the item
	result, err := collection.DeleteOne(context.Background(), bson.M{
		"_id":       objID,
		"createdBy": userObjID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item deleted successfully",
		"itemId":  itemID,
	})
}
