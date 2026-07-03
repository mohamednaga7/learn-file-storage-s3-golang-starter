package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generateRandomKey() string {
	tokenBytes := make([]byte, 32)

	_, _ = rand.Read(tokenBytes)

	return base64.RawURLEncoding.EncodeToString(tokenBytes)
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video, requestContext context.Context) (database.Video, error) {
	resultingVideo := video

	if resultingVideo.VideoURL != nil {
		videoUrlSplit := strings.Split(*resultingVideo.VideoURL, ",")
		if len(videoUrlSplit) < 2 {
			return resultingVideo, errors.New("Invalid video storage details")
		}

		videoBucket := videoUrlSplit[0]
		videoKey := videoUrlSplit[1]

		url, err := generatePresignedURL(cfg.s3Client, videoBucket, videoKey, 1*time.Hour, requestContext)
		if err != nil {
			return resultingVideo, err
		}

		resultingVideo.VideoURL = &url
	}

	return resultingVideo, nil
}
