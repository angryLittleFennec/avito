package server

import (
	"avito/internal/db"
	"avito/internal/generated"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	UniqueViolationErr = pq.ErrorCode("23505")
)

type CustomBannerResponse struct {
	ID        uint            `json:"banner_id"`
	Content   json.RawMessage `json:"content"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	IsActive  bool            `json:"is_active"`
	FeatureID int             `json:"feature_id"`
	TagIds    []int           `json:"tag_ids,"`
}

type BannerPostResponseCreated struct {
	BannerId *uint `json:"banner_id,omitempty"`
}

// GetBanner handles GET /banner endpoint
func (s *Server) GetBanner(ctx echo.Context, params generated.GetBannerParams) error {
	slog.Info("Starting GetBanner request", "params", params)

	var banners []db.Banner

	query := s.DB.Model(&db.Banner{})

	if params.FeatureId != nil && params.TagId != nil {
		slog.Debug("Filtering banners by both feature and tag", "feature", *params.FeatureId, "tag", *params.TagId)
		query = query.Joins("join banner_feature_tags on banner_feature_tags.banner_id = banners.id").
			Where("banner_feature_tags.feature_id = ? AND banner_feature_tags.tag_id = ?", *params.FeatureId, *params.TagId)
	} else if params.FeatureId != nil {
		slog.Debug("Filtering banners by feature", "feature", *params.FeatureId)
		query = query.Joins("join banner_feature_tags on banner_feature_tags.banner_id = banners.id").
			Where("banner_feature_tags.feature_id = ?", *params.FeatureId)
	} else if params.TagId != nil {
		slog.Debug("Filtering banners by tag", "tag", *params.TagId)
		query = query.Joins("join banner_feature_tags on banner_feature_tags.banner_id = banners.id").
			Where("banner_feature_tags.tag_id = ?", *params.TagId)
	}

	if params.Limit != nil {
		query = query.Limit(*params.Limit)
		slog.Debug("Applying limit to query", "limit", *params.Limit)
	}
	if params.Offset != nil {
		query = query.Offset(*params.Offset)
		slog.Debug("Applying offset to query", "offset", *params.Offset)
	}

	if err := query.Find(&banners).Error; err != nil {
		slog.Error("Failed to fetch banners from database", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch banners from database")
	}

	response := make([]CustomBannerResponse, len(banners))
	for i, banner := range banners {
		response[i] = CustomBannerResponse{
			ID:        banner.ID,
			Content:   banner.Content,
			CreatedAt: banner.CreatedAt,
			UpdatedAt: banner.UpdatedAt,
			IsActive:  banner.IsActive,
		}
		var bannerFeatureTags []db.BannerFeatureTag
		if err := s.DB.Where("banner_id = ?", banner.ID).Find(&bannerFeatureTags).Error; err != nil {
			slog.Error("Failed to fetch banner relations", "bannerID", banner.ID, "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch banner relations")
		}
		for _, bft := range bannerFeatureTags {
			if bft.FeatureID != 0 {
				response[i].FeatureID = bft.FeatureID
			}
			response[i].TagIds = append(response[i].TagIds, bft.TagID)
		}
	}

	slog.Info("Successfully retrieved banners", "count", len(banners))
	return ctx.JSON(http.StatusOK, response)
}

func (s *Server) PostBanner(ctx echo.Context, params generated.PostBannerParams) error {
	var jsonBody generated.PostBannerJSONBody
	if err := ctx.Bind(&jsonBody); err != nil {
		slog.Error("Failed to bind JSON body for new banner", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if jsonBody.IsActive == nil || jsonBody.Content == nil || jsonBody.FeatureId == nil || jsonBody.TagIds == nil {
		slog.Warn("Missing one or more required fields for new banner", "IsActive", jsonBody.IsActive, "Content", jsonBody.Content, "FeatureId", jsonBody.FeatureId, "TagIds", jsonBody.TagIds)
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields: IsActive, Content, FeatureId, and TagIds must be provided")
	}

	slog.Info("Starting transaction to create new banner")
	tx := s.DB.Begin()
	if tx.Error != nil {
		slog.Error("Failed to start transaction for new banner", "error", tx.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start database transaction")
	}

	banner := db.Banner{
		IsActive: *jsonBody.IsActive,
		Content:  getJsonFromPointer(jsonBody.Content),
	}

	if err := tx.Create(&banner).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to save new banner", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save banner: "+err.Error())
	}

	slog.Info("Banner created successfully", "bannerID", banner.ID)
	for _, tagId := range *jsonBody.TagIds {
		bftEntry := db.BannerFeatureTag{
			BannerID:  banner.ID,
			FeatureID: *jsonBody.FeatureId,
			TagID:     tagId,
		}
		if err := tx.Create(&bftEntry).Error; err != nil {
			tx.Rollback()
			if isDuplicateEntryError(err) {
				slog.Warn("Attempted to create a duplicate feature tag combination", "feature", *jsonBody.FeatureId, "tag", tagId)
				return echo.NewHTTPError(http.StatusConflict, "Duplicate feature and tag combination")
			}
			slog.Error("Failed to create banner feature tag", "feature", *jsonBody.FeatureId, "tag", tagId, "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create banner feature tag: "+err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction for new banner", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}

	slog.Info("Banner creation and association completed successfully", "bannerID", banner.ID)
	return ctx.JSON(http.StatusCreated, BannerPostResponseCreated{BannerId: &banner.ID})
}

func isDuplicateEntryError(err error) bool {
	return strings.Contains(err.Error(), "23505")
}

// DeleteBannerId handles DELETE /banner/{id} endpoint
func (s *Server) DeleteBannerId(ctx echo.Context, id int, params generated.DeleteBannerIdParams) error {
	if err := s.DB.Delete(&db.Banner{}, id).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete banner")
	}
	return ctx.NoContent(http.StatusNoContent)
}

// PatchBannerId handles PATCH /banner/{id} endpoint
func (s *Server) PatchBannerId(ctx echo.Context, id int, params generated.PatchBannerIdParams) error {
	var jsonBody generated.PatchBannerIdJSONBody
	if err := ctx.Bind(&jsonBody); err != nil {
		slog.Error("Failed to bind JSON body", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	slog.Info("Starting transaction for patching banner", "bannerID", id)
	tx := s.DB.Begin()
	if tx.Error != nil {
		slog.Error("Failed to start database transaction", "error", tx.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start database transaction")
	}

	var banner db.Banner
	if err := tx.First(&banner, id).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("Banner not found during patch operation", "bannerID", id)
			return echo.NewHTTPError(http.StatusNotFound, "Banner not found")
		}
		slog.Error("Database error on retrieving banner", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: "+err.Error())
	}

	if jsonBody.IsActive != nil {
		banner.IsActive = *jsonBody.IsActive
	}
	if jsonBody.Content != nil {
		banner.Content = getJsonFromPointer(jsonBody.Content)
	}
	if err := tx.Save(&banner).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to update banner", "bannerID", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update banner: "+err.Error())
	}

	slog.Info("Banner updated successfully", "bannerID", id)

	var existingTags []db.BannerFeatureTag
	if err := tx.Where("banner_id = ?", id).Find(&existingTags).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to retrieve existing feature and tag associations", "bannerID", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve existing feature and tag associations: "+err.Error())
	}

	featureId := jsonBody.FeatureId
	var tagIds []int
	if featureId == nil && len(existingTags) > 0 {
		featureId = &existingTags[0].FeatureID
	}
	if jsonBody.TagIds == nil {
		for _, tag := range existingTags {
			tagIds = append(tagIds, tag.TagID)
		}
	} else {
		tagIds = *jsonBody.TagIds
	}

	if err := tx.Where("banner_id = ?", id).Delete(&db.BannerFeatureTag{}).Error; err != nil {
		tx.Rollback()
		slog.Error("Failed to delete existing banner feature tags", "bannerID", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete existing banner feature tags: "+err.Error())
	}

	if featureId != nil {
		for _, tagId := range tagIds {
			bftEntry := db.BannerFeatureTag{
				BannerID:  uint(id),
				FeatureID: *featureId,
				TagID:     tagId,
			}
			if err := tx.Create(&bftEntry).Error; err != nil {
				tx.Rollback()
				if isDuplicateEntryError(err) {
					slog.Warn("Duplicate feature and tag combination detected", "feature", *featureId, "tag", tagId)
					return echo.NewHTTPError(http.StatusConflict, "Duplicate feature and tag combination")
				}
				slog.Error("Failed to create new banner feature tag", "error", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create new banner feature tag: "+err.Error())
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("Failed to commit transaction", "bannerID", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
	}

	slog.Info("Banner patch operation completed successfully", "bannerID", id)
	return ctx.String(http.StatusOK, "OK")
}

func (s *Server) GetUserBanner(ctx echo.Context, params generated.GetUserBannerParams) error {
	slog.Info("Attempting to retrieve banner", "featureID", params.FeatureId, "tagID", params.TagId)

	redisKey := fmt.Sprintf("banner:%d:%d", params.FeatureId, params.TagId)

	if params.UseLastRevision == nil || !*params.UseLastRevision {
		slog.Info("Checking cache for banner", "redisKey", redisKey)
		result, err := s.Redis.Get(context.Background(), redisKey).Result()
		if err == nil {
			slog.Info("Cache hit for banner", "redisKey", redisKey)
			return ctx.JSONBlob(http.StatusOK, []byte(result))
		} else if err != redis.Nil {
			slog.Error("Redis error occurred", "error", err)
		} else {
			slog.Info("Cache miss for banner", "redisKey", redisKey)
		}
	}

	var banner db.Banner
	if err := s.DB.Model(&db.Banner{}).Joins("join banner_feature_tags on banner_feature_tags.banner_id = banners.id").
		Where("banner_feature_tags.feature_id = ? AND banner_feature_tags.tag_id = ?", params.FeatureId, params.TagId).First(&banner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Warn("Banner not found in database", "featureID", params.FeatureId, "tagID", params.TagId)
			return echo.NewHTTPError(http.StatusNotFound, "Banner not found or is not active")
		}
		slog.Error("Database error when fetching banner", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error: "+err.Error())
	}

	slog.Info("Banner retrieved from database", "bannerID", banner.ID)

	var tags []db.BannerFeatureTag
	if err := s.DB.Where("banner_id = ?", banner.ID).Find(&tags).Error; err != nil {
		slog.Error("Failed to fetch banner feature tags from database", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch banner feature tags: "+err.Error())
	}

	respBytes, err := json.Marshal(banner.Content)
	if err != nil {
		slog.Error("Failed to serialize banner response for caching", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to serialize response: "+err.Error())
	}

	if err := s.Redis.Set(context.Background(), redisKey, respBytes, 5*time.Minute).Err(); err != nil {
		slog.Error("Failed to cache banner data in Redis", "error", err)
	} else {
		slog.Info("Banner data cached in Redis successfully", "redisKey", redisKey)
	}

	return ctx.JSON(http.StatusOK, banner.Content)
}

func getJsonFromPointer(p *map[string]interface{}) json.RawMessage {
	if p != nil {
		jsonData, err := json.Marshal(*p)
		if err != nil {
			panic("Failed to marshal JSON")
		}
		return jsonData
	}
	return nil
}
