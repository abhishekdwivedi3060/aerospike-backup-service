package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/service"
)

const storageNameNotSpecifiedMsg = "Storage name is not specified"

// addStorage
// @Summary     Adds a storage to the config.
// @ID	        addStorage
// @Tags        Configuration
// @Router      /v1/config/storage/{name} [post]
// @Accept      json
// @Param       name path string true "Backup storage name"
// @Param       storage body model.Storage true "Backup storage details"
// @Success     201
// @Failure     400 {string} string
//
//nolint:dupl
func (ws *HTTPServer) addStorage(w http.ResponseWriter, r *http.Request) {
	var newStorage model.Storage
	err := json.NewDecoder(r.Body).Decode(&newStorage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, storageNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err = service.AddStorage(ws.config, name, &newStorage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ConfigurationManager.WriteConfiguration(ws.config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// readAllStorage reads all storage from the configuration.
// @Summary     Reads all storage from the configuration.
// @ID 	        readAllStorage
// @Tags        Configuration
// @Router      /v1/config/storage [get]
// @Produce     json
// @Success  	200 {object} map[string]model.Storage
// @Failure     400 {string} string
func (ws *HTTPServer) readAllStorage(w http.ResponseWriter, _ *http.Request) {
	storage := ws.config.Storage
	jsonResponse, err := json.Marshal(storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		slog.Error("failed to write response", "err", err)
	}
}

// readStorage  reads a specific storage from the configuration given its name.
// @Summary     Reads a specific storage from the configuration given its name.
// @ID	        readStorage
// @Tags        Configuration
// @Router      /v1/config/storage/{name} [get]
// @Param       name path string true "Backup storage name"
// @Produce     json
// @Success  	200 {object} model.Storage
// @Response    400 {string} string
// @Failure     404 {string} string "The specified storage could not be found"
func (ws *HTTPServer) readStorage(w http.ResponseWriter, r *http.Request) {
	storageName := r.PathValue("name")
	if storageName == "" {
		http.Error(w, storageNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	storage, ok := ws.config.Storage[storageName]
	if !ok {
		http.Error(w, fmt.Sprintf("Storage %s could not be found", storageName), http.StatusNotFound)
		return
	}
	jsonResponse, err := json.Marshal(storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		slog.Error("failed to write response", "err", err)
	}
}

// updateStorage updates an existing storage in the configuration.
// @Summary     Updates an existing storage in the configuration.
// @ID	        updateStorage
// @Tags        Configuration
// @Router      /v1/config/storage/{name} [put]
// @Accept      json
// @Param       name path string true "Backup storage name"
// @Param       storage body model.Storage true "Backup storage details"
// @Success     200
// @Failure     400 {string} string
func (ws *HTTPServer) updateStorage(w http.ResponseWriter, r *http.Request) {
	var updatedStorage model.Storage
	err := json.NewDecoder(r.Body).Decode(&updatedStorage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	storageName := r.PathValue("name")
	if storageName == "" {
		http.Error(w, storageNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err = service.UpdateStorage(ws.config, storageName, &updatedStorage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ConfigurationManager.WriteConfiguration(ws.config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// deleteStorage
// @Summary     Deletes a storage from the configuration by name.
// @ID	        deleteStorage
// @Tags        Configuration
// @Router      /v1/config/storage/{name} [delete]
// @Param       name path string true "Backup storage name"
// @Success     204
// @Failure     400 {string} string
func (ws *HTTPServer) deleteStorage(w http.ResponseWriter, r *http.Request) {
	storageName := r.PathValue("name")
	if storageName == "" {
		http.Error(w, storageNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err := service.DeleteStorage(ws.config, storageName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = ConfigurationManager.WriteConfiguration(ws.config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
