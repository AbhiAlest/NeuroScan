package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Image struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ResultURL   string `json:"result_url"`
}

var images []Image

func main() {
	router := mux.NewRouter()

	// Serve uploaded files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Handle image upload
	router.HandleFunc("/api/images", handleImageUpload).Methods("POST")

	// Handle image retrieval
	router.HandleFunc("/api/images/{id}", getImageByID).Methods("GET")

	// Enable CORS
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	// Start server
	log.Println("Server started on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

func handleImageUpload(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error uploading image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file to disk
	filename := handler.Filename
	ext := filepath.Ext(filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		http.Error(w, "Only JPG, JPEG, and PNG files are allowed", http.StatusBadRequest)
		return
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error uploading image", http.StatusInternalServerError)
		return
	}
	newPath := filepath.Join("./uploads", filename)
	err = ioutil.WriteFile(newPath, fileBytes, 0644)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error uploading image", http.StatusInternalServerError)
		return
	}

	// Create new Image struct and add to images slice
	newImage := Image{
		ID:          fmt.Sprintf("%d", len(images)+1),
		OriginalURL: newPath,
		ResultURL:   "",
	}
	images = append(images, newImage)

	// Return JSON response with image ID
	response, err := json.Marshal(map[string]string{"id": newImage.ID})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func getImageByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Find image by ID
	var image Image
	for _, img := range images {
		if img.ID == id {
			image = img
			break
		}
	}
	if image.ID == "" {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Read image from file
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert image to bytes
	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save image to database
	imageData := models.Image{
		Name: image.Filename,
		Data: buffer.Bytes(),
	}
	err = db.SaveImage(&imageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]string{
		"status": "success",
		"message": "Image uploaded successfully",
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

	
