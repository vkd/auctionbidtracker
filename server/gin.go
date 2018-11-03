package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vkd/auctionbidtracker"
)

// NewServer - return server handler
func NewServer(a auctionbidtracker.Auctioner) http.Handler {
	e := gin.Default()

	items := e.Group("/items")
	{
		items.GET("", getItemsHandler(a))
		items.POST("", newItemHandler(a))
		item := items.Group("/:item")
		{
			item.GET("/winning-bid", getWinningBidHandler(a))
			bids := item.Group("/bids")
			{
				bids.GET("", getBidsByItemHandler(a))
				bids.POST("", newBidHandler(a))
			}
		}
	}

	return e
}

func getItemsHandler(a auctionbidtracker.Auctioner) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query 'user_id' is required not empty"})
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query 'user_id' must me number", "details": err.Error()})
			return
		}

		items, err := a.GetItemsByUser(auctionbidtracker.UserID(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on get items by user", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, items)
	}
}

func newItemHandler(a auctionbidtracker.Auctioner) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item auctionbidtracker.Item
		err := c.ShouldBindJSON(&item)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong json request", "details": err.Error()})
			return
		}

		new, err := a.AddItem(&item)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on add new item", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, new)
	}
}

func getWinningBidHandler(a auctionbidtracker.Auctioner) gin.HandlerFunc {
	return func(c *gin.Context) {
		item := c.Param("item")
		itemID, err := strconv.Atoi(item)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong item format: must be int", "details": err.Error()})
			return
		}

		bid, err := a.GetWinningBid(auctionbidtracker.ItemID(itemID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on get winning bid", "details": err.Error()})
			return
		}
		if bid == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}

		c.JSON(http.StatusOK, bid)
	}
}

func getBidsByItemHandler(a auctionbidtracker.Auctioner) gin.HandlerFunc {
	return func(c *gin.Context) {
		item := c.Param("item")
		itemID, err := strconv.Atoi(item)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong item format: must be int", "details": err.Error()})
			return
		}

		bids, err := a.GetAllBids(auctionbidtracker.ItemID(itemID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on get all bids", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, bids)
	}
}

func newBidHandler(a auctionbidtracker.Auctioner) gin.HandlerFunc {
	return func(c *gin.Context) {
		item := c.Param("item")
		itemID, err := strconv.Atoi(item)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong item format: must be int", "details": err.Error()})
			return
		}

		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query 'user_id' is required not empty"})
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query 'user_id' must me number", "details": err.Error()})
			return
		}

		var bid auctionbidtracker.Bid
		err = c.ShouldBindJSON(&bid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "wrong json request", "details": err.Error()})
			return
		}

		err = a.MakeBid(auctionbidtracker.ItemID(itemID), auctionbidtracker.UserID(userID), &bid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error on make bid", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
