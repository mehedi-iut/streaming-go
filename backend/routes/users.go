package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"video_stream/models"
	"video_stream/utils"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = user.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	convertToJsonResponse(w, map[string]string{"message": "User created successfully"}, http.StatusCreated)

}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = user.ValidateCredentials()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Login successfully", "token": token}
	convertToJsonResponse(w, response, http.StatusOK)

}

func convertToJsonResponse(w http.ResponseWriter, response map[string]string, statusCode int) {
	//response := map[string]string{"message": msg}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		// Handle encoding error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
