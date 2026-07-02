package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func getFileMimeType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset read pointer if you plan to process the file further
	_, _ = file.Seek(0, 0)

	// Detect the content type
	mimeType := http.DetectContentType(buffer)
	fmt.Println("MIME Type:", mimeType)
	if mimeType == "" {
		return "", errors.New("couldn't get the mimetype")
	}
	return mimeType, nil
}

func getFileExtension(mimeType string) string {
	split := strings.Split(mimeType, "/")
	if len(split) < 2 {
		return mimeType
	}

	return split[1]
}

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse form", err)
		return
	}

	file, _, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mimeType, err := getFileMimeType(file)

	//fileDataBytes, err := io.ReadAll(file)
	//if err != nil {
	//	respondWithError(w, http.StatusInternalServerError, "Unable to read file", err)
	//	return
	//}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to read file", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Cannot access this video file", err)
		return
	}

	fileExtension := getFileExtension(mimeType)

	randomKey := generateRandomKey()

	fileName := randomKey + "." + fileExtension

	newThumbnailUrl := filepath.Join("assets", fileName)

	create, err := os.Create(newThumbnailUrl)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot upload video file", err)
		return
	}

	defer create.Close()

	prepended := "http://"

	if r.TLS != nil {
		prepended = "https://"
	}

	newThumbnailUrl = prepended + path.Join(r.Host, newThumbnailUrl)
	video.ThumbnailURL = &newThumbnailUrl

	_, err = io.Copy(create, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot upload video file", err)
		return
	}

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to read file", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
