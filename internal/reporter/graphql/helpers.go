package graphql

import (
	"context"
	"fmt"
	
	"github.com/guidewire-oss/fern-platform/internal/reporter/graphql/dataloader"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// getLoaders gets the dataloader from context
func getLoaders(ctx context.Context) *dataloader.Loaders {
	if ctx == nil {
		return nil
	}
	
	loadersVal := ctx.Value("loaders")
	if loadersVal == nil {
		return nil
	}
	
	loaders, ok := loadersVal.(*dataloader.Loaders)
	if !ok {
		return nil
	}
	return loaders
}

// getCurrentUser gets the current user from context
func getCurrentUser(ctx context.Context) (*database.User, error) {
	user, ok := ctx.Value("user").(*database.User)
	if !ok {
		return nil, fmt.Errorf("user not authenticated")
	}
	return user, nil
}

// getRequestID gets the request ID from context
func getRequestID(ctx context.Context) string {
	reqID, _ := ctx.Value("request_id").(string)
	return reqID
}

// convertPtrString converts a *string to string
func convertPtrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// convertStringPtr converts a string to *string
func convertStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// paginateSlice applies pagination to a slice
func paginateSlice[T any](items []T, first int, after string) ([]T, bool) {
	start := 0
	if after != "" {
		// Simple cursor implementation - in production, use proper cursor encoding
		fmt.Sscanf(after, "%d", &start)
	}
	
	if start >= len(items) {
		return []T{}, false
	}
	
	end := start + first
	hasMore := false
	
	if end > len(items) {
		end = len(items)
	} else {
		hasMore = true
	}
	
	return items[start:end], hasMore
}