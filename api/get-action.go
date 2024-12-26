// api/get-action.go
package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Získání parametrů z URL
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if owner == "" || repo == "" {
		http.Error(w, "Missing 'owner' or 'repo' query parameter", http.StatusBadRequest)
		return
	}

	// URL pro GitHub API (získání počtu běhů akcí pro daný repozitář)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Nastavení hlavičky pro autorizaci GitHub API
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GITHUB_TOKEN"))

	// Odeslání požadavku na GitHub API
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact GitHub API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Kontrola statusu odpovědi
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Error retrieving data from GitHub", http.StatusInternalServerError)
		return
	}

	// Načtení a parsování JSON odpovědi
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Failed to parse response body", http.StatusInternalServerError)
		return
	}

	// Vytvoření odpovědi ve formátu JSON
	response := map[string]interface{}{
		"owner":         owner,
		"repo":          repo,
		"actions_count": result["total_count"],
		"schemaVersion": 1,
		"label":         "uses",
		"message":       result["total_count"],
		"color":         "blue",
		"logo":          "githubactions",
	}

	// Nastavení správných hlaviček pro JSON odpověď
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Odeslání odpovědi
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	// Tento handler bude obsluhovat všechny požadavky na /api/get-action
	http.HandleFunc("/api/get-action", Handler)
	http.HandleFunc("/", Handler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
