package handlers

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
)

func IsUserAllowed(ctx context.Context, userId int64) bool {
	// Check if the user is a bot
	if userId < 0 {
		log.Printf("user %v is a bot and isn't allowed to use the bot", userId)
		return false
	}
	// Create a slice of allowed user ids from env var
	allowedUserIds := os.Getenv("ALLOWED_USER_IDS")
	allowedUserIdsSlice := strings.Split(allowedUserIds, ",")
	allowedUserIdsInt64 := make([]int64, len(allowedUserIdsSlice))
	for i, id := range allowedUserIdsSlice {
		var err error
		allowedUserIdsInt64[i], err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Printf("failed to parse user id: %v", err)
			return false
		}
	}

	// Check if the user id is in the allowed user ids slice
	for _, id := range allowedUserIdsInt64 {
		if id == userId {
			log.Printf("user %v is allowed to use the bot", userId)
			return true
		}
	}
	log.Printf("user %v isn't allowed to use the bot", userId)

	return false
}
