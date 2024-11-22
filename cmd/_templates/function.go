package start

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func init() {
	functions.HTTP("start", startHere)
}

func startHere(w http.ResponseWriter, r *http.Request) {
	// Set default headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method == http.MethodPost {
		//Get and check for variables
		region, clientId, secret, err := checkEnvironmentVariables()
		if err != nil {
			response(w, err.Error(), 400)
			return
		}

		// Check for Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok {
			response(w, "\"Auth missing\"", 400)
			return
		} else {
			if err := checkPasswordVariable(username, password); err != nil {
				response(w, err.Error(), 401)
				return
			}
			fmt.Println("Authorized")
		}

		//Check for valid POST body formatting
		respBytes, err := io.ReadAll(r.Body)
		if err != nil {
			response(w, "\"Error reading POST body\"", 400)
			return
		}
		var requestBody RequestBody
		if err := json.Unmarshal(respBytes, &requestBody); err != nil {
			response(w, "\"Bad Request\"", 400)
			return
		}

		//Do Genesys Cloud OAuth
		config := platformclientv2.GetDefaultConfiguration()
		config.BasePath = "https://api." + region
		if err := config.AuthorizeClientCredentials(clientId, secret); err != nil {
			response(w, err.Error(), 400)
		}
		fmt.Println("Logged In to Genesys Cloud")
		//Setup the actual request in genesys.go file
		genesysStart(w, config, requestBody)
		return
	} else {
		response(w, "Method not supported", 400)
	}
}

func checkEnvironmentVariables() (returnRegion string, returnClientId string, returnSecret string, err error) {
	region := os.Getenv("REGION")
	if region == "" {
		return "", "", "", fmt.Errorf("\"REGION not set\"")
	}
	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		return "", "", "", fmt.Errorf("\"CLIENT_ID not set\"")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		return "", "", "", fmt.Errorf("\"SECRET not set\"")
	}
	return region, clientId, secret, nil
}

func checkPasswordVariable(user string, pass string) (err error) {
	username := os.Getenv("USERNAME")
	if username != user {
		return fmt.Errorf("\"username wrong\"")
	}
	password := os.Getenv("PASSWORD")
	if password != pass {
		return fmt.Errorf("\"password wrong\"")
	}
	return nil
}

func response(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	fmt.Printf("Status: %v, Message: %s\n", status, msg)
	fmt.Fprintf(w, "{\"response\": %s}", msg)
}
