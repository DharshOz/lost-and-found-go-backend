package routes

import (
	"context"
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

func AddFoundItem(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	var foundItem models.FoundItem
	if err := c.ShouldBind(&foundItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Set additional fields
	foundItem.Image = imageURL
	foundItem.DateFound = time.Now()
	foundItem.FoundPerson = objID

	collection := db.GetCollection("founditems")
	result, err := collection.InsertOne(context.Background(), foundItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create found item"})
		return
	}

	// Add notification to the lost person
	if foundItem.LostPerson != primitive.NilObjectID {
		notification := models.Notification{
			Message:   "Your lost item has been reported as found",
			Read:      false,
			CreatedAt: time.Now(),
		}

		userCollection := db.GetCollection("users")
		_, err = userCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": foundItem.LostPerson},
			bson.M{"$push": bson.M{"notifications": notification}},
		)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Found item reported successfully",
		"itemId":  result.InsertedID,
	})
}

func GetAllFoundItems(c *gin.Context) {
	collection := db.GetCollection("founditems")
	filter := bson.M{}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var items []models.FoundItem
	if err := cursor.All(context.Background(), &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
		return
	}

	// Populate user information
	userCollection := db.GetCollection("users")
	for i, item := range items {
		var foundUser models.User
		err := userCollection.FindOne(context.Background(), bson.M{"_id": item.FoundBy}).Decode(&foundUser)
		if err == nil {
			items[i].FoundByUser = &models.User{
				Username: foundUser.Username,
				Email:    foundUser.Email,
			}
		}

		if item.LostPerson != primitive.NilObjectID {
			var lostUser models.User
			err := userCollection.FindOne(context.Background(), bson.M{"_id": item.LostPerson}).Decode(&lostUser)
			if err == nil {
				items[i].LostPersonUser = &models.User{
					Username: lostUser.Username,
					Email:    lostUser.Email,
				}
			}
		}
	}

	c.JSON(http.StatusOK, items)
}

func GetFoundItemByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	collection := db.GetCollection("founditems")
	var item models.FoundItem
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&item)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Populate user information
	userCollection := db.GetCollection("users")
	if item.FoundBy() != primitive.NilObjectID {
		var foundUser models.User
		err := userCollection.FindOne(context.Background(), bson.M{"_id": item.FoundBy}).Decode(&foundUser)
		if err == nil {
			item.FoundByUser = &models.User{
				Username: foundUser.Username,
				Email:    foundUser.Email,
			}
		}
	}

	if item.LostPerson != primitive.NilObjectID {
		var lostUser models.User
		err := userCollection.FindOne(context.Background(), bson.M{"_id": item.LostPerson}).Decode(&lostUser)
		if err == nil {
			item.LostPersonUser = &models.User{
				Username: lostUser.Username,
				Email:    lostUser.Email,
			}
		}
	}

	c.JSON(http.StatusOK, item)
}

func GetFoundItemsByLostItem(c *gin.Context) {
	lostItemID := c.Param("lostItemId")
	objID, err := primitive.ObjectIDFromHex(lostItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lost item ID"})
		return
	}

	collection := db.GetCollection("founditems")
	cursor, err := collection.Find(context.Background(), bson.M{"lostPerson": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var items []models.FoundItem
	if err := cursor.All(context.Background(), &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
		return
	}

	// Populate found by user information
	userCollection := db.GetCollection("users")
	for i, item := range items {
		var user models.User
		err := userCollection.FindOne(context.Background(), bson.M{"_id": item.FoundBy}).Decode(&user)
		if err == nil {
			items[i].FoundByUser = &models.User{
				Username: user.Username,
				Email:    user.Email,
			}
		}
	}

	c.JSON(http.StatusOK, items)
}

func GetFoundItemsByUser(c *gin.Context) {
	userID := c.Param("foundPersonId")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := db.GetCollection("founditems")
	cursor, err := collection.Find(context.Background(), bson.M{"foundBy": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.Background())

	var items []models.FoundItem
	if err := cursor.All(context.Background(), &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
		return
	}

	// Populate lost person information
	userCollection := db.GetCollection("users")
	for i, item := range items {
		if item.LostPerson != primitive.NilObjectID {
			var user models.User
			err := userCollection.FindOne(context.Background(), bson.M{"_id": item.LostPerson}).Decode(&user)
			if err == nil {
				items[i].LostPersonUser = &models.User{
					Username: user.Username,
					Email:    user.Email,
				}
			}
		}
	}

	c.JSON(http.StatusOK, items)
}

func UpdateFoundItem(c *gin.Context) {
	itemID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Get the authenticated user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var updateData models.FoundItem
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := db.GetCollection("founditems")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID, "foundBy": userObjID},
		bson.M{"$set": updateData},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or not owned by user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item updated successfully",
		"itemId":  itemID,
	})
}

func UpdateFoundStatus(c *gin.Context) {
	itemID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var request struct {
		Found bool `json:"found"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := db.GetCollection("founditems")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"found": request.Found}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
		"found":   request.Found,
	})
}

func DeleteFoundItem(c *gin.Context) {
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

	collection := db.GetCollection("founditems")

	// First verify the item exists and belongs to the user
	var item models.FoundItem
	err = collection.FindOne(context.Background(), bson.M{
		"_id":     objID,
		"foundBy": userObjID,
	}).Decode(&item)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found or not owned by user"})
		return
	}

	// Delete the item
	result, err := collection.DeleteOne(context.Background(), bson.M{
		"_id":     objID,
		"foundBy": userObjID,
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
