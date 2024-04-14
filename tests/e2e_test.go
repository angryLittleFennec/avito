package tests

import (
	"avito/internal/generated"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestUrl() string {
	if value, ok := os.LookupEnv("API_URL"); ok {
		return value
	}
	return "http://localhost:8080"
}

func TestBannerLifecycle(t *testing.T) {
	client, err := generated.NewClientWithResponses(getTestUrl())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	adminToken := "admin1"

	postResp, err := client.PostBannerWithResponse(ctx, &generated.PostBannerParams{Token: &adminToken}, generated.PostBannerJSONRequestBody{
		Content:   &map[string]interface{}{"message": "New Year Sale"},
		FeatureId: ptrToInt(100),
		IsActive:  ptrToBool(true),
		TagIds:    &[]int{100, 200},
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, postResp.HTTPResponse.StatusCode)
	fmt.Println("Banner created with ID:", *postResp.JSON201.BannerId)

	getResp, err := client.GetBannerWithResponse(ctx, &generated.GetBannerParams{
		FeatureId: ptrToInt(100),
		TagId:     ptrToInt(100),
		Token:     &adminToken,
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, getResp.HTTPResponse.StatusCode)
	for _, banner := range *getResp.JSON200 {
		fmt.Printf("Retrieved banner: %+v\n", banner)
	}

}

func TestDeleteBannerLifecycle(t *testing.T) {
	client, err := generated.NewClientWithResponses(getTestUrl())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	adminToken := "admin1"

	postBody := generated.PostBannerJSONRequestBody{
		Content:   &map[string]interface{}{"title": "Test Banner", "description": "This is a test banner."},
		IsActive:  ptrToBool(true),
		FeatureId: ptrToInt(122),
		TagIds:    &[]int{112, 102},
	}
	postResp, err := client.PostBannerWithResponse(ctx, &generated.PostBannerParams{Token: &adminToken}, postBody)
	assert.NoError(t, err)
	assert.NotNil(t, postResp.JSON201)
	assert.Equal(t, http.StatusCreated, postResp.HTTPResponse.StatusCode)

	bannerID := *postResp.JSON201.BannerId

	deleteResp, err := client.DeleteBannerIdWithResponse(ctx, bannerID, &generated.DeleteBannerIdParams{Token: &adminToken})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, deleteResp.HTTPResponse.StatusCode)

	getResp, err := client.GetBannerWithResponse(ctx, &generated.GetBannerParams{
		Token: &adminToken,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.HTTPResponse.StatusCode)

	found := false
	for _, banner := range *getResp.JSON200 {
		if *banner.BannerId == bannerID {
			found = true
			break
		}
	}
	assert.False(t, found, "Banner should not be found after deletion")
}

func TestPatchAndUpdateBanner(t *testing.T) {
	client, err := generated.NewClientWithResponses(getTestUrl())
	require.NoError(t, err, "Failed to create client")

	adminToken := "admin1"
	postResp, err := client.PostBannerWithResponse(context.Background(), &generated.PostBannerParams{Token: &adminToken}, generated.PostBannerJSONRequestBody{
		Content:   &map[string]interface{}{"title": "Initial Title", "text": "Initial Text"},
		FeatureId: ptrToInt(10),
		IsActive:  ptrToBool(true),
		TagIds:    &[]int{110},
	})
	require.NoError(t, err, "Failed to create banner")
	require.Equal(t, http.StatusCreated, postResp.HTTPResponse.StatusCode, "Banner creation failed")
	bannerID := *postResp.JSON201.BannerId

	updateBody := generated.PatchBannerIdJSONRequestBody{
		Content:   &map[string]interface{}{"title": "Updated Title", "text": "Updated Text", "url": "http://new-url.com"},
		FeatureId: ptrToInt(10),
		IsActive:  ptrToBool(false),
		TagIds:    &[]int{110, 111},
	}

	patchResp, err := client.PatchBannerIdWithResponse(context.Background(), bannerID, &generated.PatchBannerIdParams{Token: &adminToken}, updateBody)
	require.NoError(t, err, "Error while patching banner")
	require.Equal(t, http.StatusOK, patchResp.HTTPResponse.StatusCode, "Banner patching failed")

	getResp, err := client.GetBannerWithResponse(context.Background(), &generated.GetBannerParams{Token: &adminToken, FeatureId: ptrToInt(10), TagId: ptrToInt(110)})
	require.NoError(t, err, "Failed to retrieve banner post-update")
	assert.Equal(t, http.StatusOK, getResp.HTTPResponse.StatusCode, "Failed to retrieve updated banner")

	assert.False(t, *(*getResp.JSON200)[0].IsActive, "Banner active flag did not update correctly")
	assert.Equal(t, *(*getResp.JSON200)[0].TagIds, []int{110, 111})
	assert.Equal(t, *(*getResp.JSON200)[0].FeatureId, 10)
}

func TestGetUserBanner(t *testing.T) {
	adminToken := "admin1"
	client, _ := generated.NewClientWithResponses(getTestUrl())
	ctx := context.Background()
	_, err := client.PostBannerWithResponse(ctx, &generated.PostBannerParams{Token: &adminToken}, generated.PostBannerJSONRequestBody{
		Content:   &map[string]interface{}{"message": "New Year Sale"},
		FeatureId: ptrToInt(1),
		IsActive:  ptrToBool(true),
		TagIds:    &[]int{1, 2},
	})

	if err != nil {
		t.Errorf("Error during the banner post: %v", err)
	}

	params := generated.GetUserBannerParams{
		TagId:     1,
		FeatureId: 1,
	}

	resp, err := client.GetUserBannerWithResponse(context.Background(), &params, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Token", "admin1")
		return nil
	})
	if err != nil || resp.JSON200 == nil {
		t.Errorf("Expected successful retrieval, got error: %v", err)
	}

	assert.Equal(t, resp.JSON200, &map[string]interface{}{"message": "New Year Sale"})

	resp, err = client.GetUserBannerWithResponse(context.Background(), &params, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Token", "user1")
		return nil
	})
	if err != nil || resp.JSON200 == nil {
		t.Errorf("Expected successful retrieval, got error: %v", err)
	}
}

func TestPostDuplicateBanner(t *testing.T) {
	client, err := generated.NewClientWithResponses(getTestUrl())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	adminToken := "admin1"

	bannerDetails := generated.PostBannerJSONRequestBody{
		Content:   &map[string]interface{}{"title": "Summer Sale", "description": "Get your summer gear at great prices!"},
		FeatureId: ptrToInt(4),
		IsActive:  ptrToBool(true),
		TagIds:    &[]int{104, 105},
	}

	firstResp, err := client.PostBannerWithResponse(ctx, &generated.PostBannerParams{Token: &adminToken}, bannerDetails)
	if err != nil {
		t.Errorf("Error during the first banner post: %v", err)
	} else {
		assert.Equal(t, http.StatusCreated, firstResp.StatusCode())
	}

	secondResp, _ := client.PostBannerWithResponse(ctx, &generated.PostBannerParams{Token: &adminToken}, bannerDetails)
	assert.Equal(t, http.StatusConflict, secondResp.StatusCode())

}

func ptrToInt(i int) *int {
	return &i
}

func ptrToBool(b bool) *bool {
	return &b
}
