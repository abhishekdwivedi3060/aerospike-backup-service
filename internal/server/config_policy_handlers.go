package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/model"
	"github.com/abhishekdwivedi3060/aerospike-backup-service/pkg/service"
)

const policyNameNotSpecifiedMsg = "Policy name is not specified"

// addPolicy
// @Summary     Adds a policy to the config.
// @ID          addPolicy
// @Tags        Configuration
// @Router      /v1/config/policies/{name} [post]
// @Accept      json
// @Param       name path string true "Backup policy name"
// @Param       policy body model.BackupPolicy true "Backup policy details"
// @Success     201
// @Failure     400 {string} string
//
//nolint:dupl
func (ws *HTTPServer) addPolicy(w http.ResponseWriter, r *http.Request) {
	var newPolicy model.BackupPolicy
	err := json.NewDecoder(r.Body).Decode(&newPolicy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, policyNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err = service.AddPolicy(ws.config, name, &newPolicy)
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

// readPolicies reads all backup policies from the configuration.
// @Summary     Reads all policies from the configuration.
// @ID	        readPolicies
// @Tags        Configuration
// @Router      /v1/config/policies [get]
// @Produce     json
// @Success  	200 {object} map[string]model.BackupPolicy
// @Failure     400 {string} string
func (ws *HTTPServer) readPolicies(w http.ResponseWriter, _ *http.Request) {
	jsonResponse, err := json.Marshal(ws.config.BackupPolicies)
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

// readPolicy reads a specific backup policy from the configuration given its name.
// @Summary     Reads a backup policy from the configuration given its name.
// @ID	        readPolicy
// @Tags        Configuration
// @Router      /v1/config/policies/{name} [get]
// @Param       name path string true "Backup policy name"
// @Produce     json
// @Success  	200 {object} model.BackupPolicy
// @Response    400 {string} string
// @Failure     404 {string} string "The specified policy could not be found"
func (ws *HTTPServer) readPolicy(w http.ResponseWriter, r *http.Request) {
	policyName := r.PathValue("name")
	if policyName == "" {
		http.Error(w, policyNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	policy, ok := ws.config.BackupPolicies[policyName]
	if !ok {
		http.Error(w, fmt.Sprintf("Policy %s could not be found", policyName), http.StatusNotFound)
		return
	}
	jsonResponse, err := json.Marshal(policy)
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

// updatePolicy updates an existing policy in the configuration.
// @Summary     Updates an existing policy in the configuration.
// @ID 	        updatePolicy
// @Tags        Configuration
// @Router      /v1/config/policies/{name} [put]
// @Accept      json
// @Param       name path string true "Backup policy name"
// @Param       policy body model.BackupPolicy true "Backup policy details"
// @Success     200
// @Failure     400 {string} string
//
//nolint:dupl
func (ws *HTTPServer) updatePolicy(w http.ResponseWriter, r *http.Request) {
	var updatedPolicy model.BackupPolicy
	err := json.NewDecoder(r.Body).Decode(&updatedPolicy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, policyNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err = service.UpdatePolicy(ws.config, name, &updatedPolicy)
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

// deletePolicy
// @Summary     Deletes a policy from the configuration by name.
// @ID          deletePolicy
// @Tags        Configuration
// @Router      /v1/config/policies/{name} [delete]
// @Param       name path string true "Backup policy name"
// @Success     204
// @Failure     400 {string} string
func (ws *HTTPServer) deletePolicy(w http.ResponseWriter, r *http.Request) {
	policyName := r.PathValue("name")
	if policyName == "" {
		http.Error(w, policyNameNotSpecifiedMsg, http.StatusBadRequest)
		return
	}
	err := service.DeletePolicy(ws.config, policyName)
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
