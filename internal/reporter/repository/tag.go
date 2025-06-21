// Package repository provides data access layer for tags
package repository

import (
	"gorm.io/gorm"
	"github.com/guidewire-oss/fern-platform/pkg/database"
)

// TagRepository handles tag data operations
type TagRepository struct {
	*database.BaseRepository
	db *gorm.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{
		BaseRepository: database.NewBaseRepository(db),
		db:             db,
	}
}

// CreateTag creates a new tag
func (r *TagRepository) CreateTag(tag *database.Tag) error {
	return r.db.Create(tag).Error
}

// GetTagByID retrieves a tag by ID
func (r *TagRepository) GetTagByID(id uint) (*database.Tag, error) {
	var tag database.Tag
	err := r.db.First(&tag, id).Error
	return &tag, err
}

// GetTagByName retrieves a tag by name
func (r *TagRepository) GetTagByName(name string) (*database.Tag, error) {
	var tag database.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	return &tag, err
}

// UpdateTag updates an existing tag
func (r *TagRepository) UpdateTag(tag *database.Tag) error {
	return r.db.Save(tag).Error
}

// DeleteTag soft deletes a tag
func (r *TagRepository) DeleteTag(id uint) error {
	return r.db.Delete(&database.Tag{}, id).Error
}

// ListTags retrieves all tags with optional filtering
func (r *TagRepository) ListTags(search string, limit, offset int) ([]*database.Tag, int64, error) {
	query := r.db.Model(&database.Tag{})
	
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	query = query.Order("name ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	var tags []*database.Tag
	err := query.Find(&tags).Error
	return tags, total, err
}

// GetTagsByTestRun retrieves all tags associated with a test run
func (r *TagRepository) GetTagsByTestRun(testRunID uint) ([]*database.Tag, error) {
	var tags []*database.Tag
	err := r.db.Joins("JOIN test_run_tags ON tags.id = test_run_tags.tag_id").
		Where("test_run_tags.test_run_id = ?", testRunID).
		Find(&tags).Error
	return tags, err
}

// GetTagUsageStats returns usage statistics for tags
func (r *TagRepository) GetTagUsageStats() ([]*TagUsage, error) {
	var tagUsage []*TagUsage
	err := r.db.Model(&database.Tag{}).
		Select("tags.id, tags.name, tags.description, tags.color, COUNT(test_run_tags.test_run_id) as usage_count").
		Joins("LEFT JOIN test_run_tags ON tags.id = test_run_tags.tag_id").
		Group("tags.id, tags.name, tags.description, tags.color").
		Order("usage_count DESC").
		Scan(&tagUsage).Error
	return tagUsage, err
}

// TagUsage represents tag usage statistics
type TagUsage struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	UsageCount  int64  `json:"usage_count"`
}

// GetOrCreateTag gets a tag by name or creates it if it doesn't exist
func (r *TagRepository) GetOrCreateTag(name, description, color string) (*database.Tag, error) {
	var tag database.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	
	if err == gorm.ErrRecordNotFound {
		// Tag doesn't exist, create it
		tag = database.Tag{
			Name:        name,
			Description: description,
			Color:       color,
		}
		if err := r.db.Create(&tag).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	
	return &tag, nil
}

// GetPopularTags returns the most frequently used tags
func (r *TagRepository) GetPopularTags(limit int) ([]*TagUsage, error) {
	var tagUsage []*TagUsage
	query := r.db.Model(&database.Tag{}).
		Select("tags.id, tags.name, tags.description, tags.color, COUNT(test_run_tags.test_run_id) as usage_count").
		Joins("INNER JOIN test_run_tags ON tags.id = test_run_tags.tag_id").
		Group("tags.id, tags.name, tags.description, tags.color").
		Having("COUNT(test_run_tags.test_run_id) > 0").
		Order("usage_count DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Scan(&tagUsage).Error
	return tagUsage, err
}