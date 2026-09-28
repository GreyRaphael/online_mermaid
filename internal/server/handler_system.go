package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"online_mermaid/internal/config"
	"online_mermaid/internal/updater"
)

type applyUpdateRequest struct {
	DownloadURL string `json:"download_url,omitempty"`
	SHA256URL   string `json:"sha256_url,omitempty"`
}

func checkUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := os.Getenv("ONLINE_MERMAID_UPDATE_REPO")
		proxy := os.Getenv("ONLINE_MERMAID_UPDATE_PROXY")
		forceRefresh := r.URL.Query().Get("refresh") == "true" || r.URL.Query().Get("refresh") == "1"

		res, err := updater.CheckUpdate(r.Context(), config.Version, repo, proxy, forceRefresh)
		if err != nil {
			slog.Warn("check update failed", "error", err)
			writeError(w, http.StatusBadGateway, "update_check_failed", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, res)
	}
}

func applyUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := os.Getenv("ONLINE_MERMAID_UPDATE_REPO")
		proxy := os.Getenv("ONLINE_MERMAID_UPDATE_PROXY")

		var req applyUpdateRequest
		if r.Body != nil && r.ContentLength > 0 {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		downloadURL := req.DownloadURL
		sha256URL := req.SHA256URL

		// If download URL is not explicitly passed in body, fetch latest release automatically
		if downloadURL == "" {
			checkRes, err := updater.CheckUpdate(r.Context(), config.Version, repo, proxy, false)
			if err != nil {
				writeError(w, http.StatusBadGateway, "update_check_failed", "Failed to query release info: "+err.Error())
				return
			}
			if !checkRes.CanAutoUpdate {
				reason := checkRes.UpdateMessage
				if reason == "" {
					reason = "Auto update is not supported for current environment"
				}
				writeError(w, http.StatusBadRequest, "cannot_auto_update", reason)
				return
			}
			downloadURL = checkRes.AssetURL
			// sha256 URL convention
			if downloadURL != "" {
				sha256URL = downloadURL + ".sha256"
			}
		}

		if downloadURL == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "No download URL available for update")
			return
		}

		slog.Info("applying update", "download_url", downloadURL, "sha256_url", sha256URL)
		if err := updater.ApplyUpdate(r.Context(), downloadURL, sha256URL, proxy); err != nil {
			slog.Error("apply update failed", "error", err)
			writeError(w, http.StatusInternalServerError, "update_failed", "Failed to apply update: "+err.Error())
			return
		}

		execPath, err := os.Executable()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "restart_failed", "Update installed but cannot locate executable to restart: "+err.Error())
			return
		}
		execPath, _ = filepath.EvalSymlinks(execPath)

		// Respond with 200 OK before restarting
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "更新已应用，服务正在重启...",
		})

		// Trigger server restart in background
		updater.RestartServer(execPath)
	}
}
