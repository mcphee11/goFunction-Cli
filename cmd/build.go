/*
Copyright © 2024 https://github.com/mcphee11
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Google Cloud Go function gen2 template",
	Long: `This uses the Go cloud functions framework to allow 
	for localhost testing before easy cloud deploying. For example:

Middleware between a website or mobile app and Genesys Cloud Platform
API calls using client credentials on the server and Basic auth on the client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("build called")
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			fmt.Printf("Missing the --name flag")
			return
		}
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			fmt.Printf("Missing the --region flag")
			return
		}
		clientId, _ := cmd.Flags().GetString("clientId")
		if clientId == "" {
			fmt.Printf("Missing the --clientId flag")
			return
		}
		secret, _ := cmd.Flags().GetString("secret")
		if secret == "" {
			fmt.Printf("Missing the --secret flag")
			return
		}
		username, _ := cmd.Flags().GetString("username")
		if username == "" {
			fmt.Printf("Missing the --username flag")
			return
		}
		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			fmt.Printf("Missing the --password flag")
			return
		}
		buildDirAndFiles(name, region, clientId, secret, username, password)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.PersistentFlags().String("name", "", "A name for the cloud function")
	buildCmd.PersistentFlags().String("region", "", "The Genesys Cloud Region eg: mypurecloud.com.au")
	buildCmd.PersistentFlags().String("clientId", "", "OAuth 2.0 ClientId")
	buildCmd.PersistentFlags().String("secret", "", "OAuth 2.0 secret")
	buildCmd.PersistentFlags().String("username", "", "A Basic username for external auth")
	buildCmd.PersistentFlags().String("password", "", "A Basic password for external auth")
}

func buildDirAndFiles(flagName string, flagRegion string, flagClientId string, flagSecret string, flagUsername string, flagPassword string) {

	// Template code for file building
	var run = fmt.Sprintf(`
#!/bin/bash

REGION=%s CLIENT_ID=%s SECRET=%s USERNAME=%s PASSWORD=%s FUNCTION_TARGET=start LOCAL_ONLY=true go run cmd/main.go
`, flagRegion, flagClientId, flagSecret, flagUsername, flagPassword)

	var deploy = fmt.Sprintf(`
#!/bin/bash

gcloud functions deploy %s --gen2 --runtime=go122 --region=australia-southeast1 --source=. --entry-point=start --trigger-http --allow-unauthenticated --set-env-vars REGION=%s,CLIENT_ID=%s,SECRET=%s,USERNAME=%s,PASSWORD=%s
`, flagName, flagRegion, flagClientId, flagSecret, flagUsername, flagPassword)

	var main = `
package main

import (
	"log"
	"os"

	// Blank-import the function package so the init() runs
	_ "example.com"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	// Use PORT environment variable, or default to 8081.
	port := "8081"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	// By default, listen on all interfaces. If testing locally, run with
	// LOCAL_ONLY=true to avoid triggering firewall warnings and
	// exposing the server outside of your own machine.
	hostname := ""
	if localOnly := os.Getenv("LOCAL_ONLY"); localOnly == "true" {
		hostname = "127.0.0.1"
	}
	if err := funcframework.StartHostPort(hostname, port); err != nil {
		log.Fatalf("funcframework.StartHostPort: %v\n", err)
	}
}
`

	var function = `
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
`

	var genesys = `
package start

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

type RequestBody struct {
// build json struct
Id string ` + "`" + `json:"Id"` + "`" + `
}

func genesysStart(w http.ResponseWriter, config *platformclientv2.Configuration, requestBody RequestBody) {
	// Set the API your wanting to use eg Conversation
	apiInstanceConversation := platformclientv2.NewConversationsApiWithConfig(config)

	reply, postError := getConversationId(*apiInstanceConversation, requestBody.Id)	// set request body object
	if postError != nil {
		response(w, postError.Error(), 400)
	}
	response(w, reply, 200)
}

func getConversationId(apiInstanceConversation platformclientv2.ConversationsApi, conversationId string) (reply string, error error) {
	data, response, err := apiInstanceConversation.GetConversation(conversationId)
	if err != nil {
		fmt.Printf("Error calling GetConversation: %v\n", err)
		return "", err
	} else {
		fmt.Printf("Response:  Success: %v  Status code: %v  Correlation ID: %v\n", response.IsSuccess, response.StatusCode, response.CorrelationID)
		json, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("response not json")
		}
		return string(json), nil
	}
}
`

	fmt.Printf("Creating Dir: %s\n", flagName)
	err := os.Mkdir(flagName, 0777)
	if err != nil {
		fmt.Printf("Error creating directory %s", flagName)
	}
	// created cmd dir
	err = os.Mkdir(fmt.Sprintf("%s/cmd", flagName), 0777)
	if err != nil {
		fmt.Println("Error creating directory cmd")
	}
	fmt.Println("Created directory cmd")

	// create functions.go file
	err = os.WriteFile(fmt.Sprintf("%s/function.go", flagName), []byte(function), 0777)
	if err != nil {
		fmt.Printf("Error creating function.go : %s", err.Error())
	}
	fmt.Println("function.go created")

	// create genesys.go file
	err = os.WriteFile(fmt.Sprintf("%s/genesys.go", flagName), []byte(genesys), 0777)
	if err != nil {
		fmt.Printf("Error creating genesys.go : %s", err.Error())
	}
	fmt.Println("genesys.go created")

	// create main.go file
	err = os.WriteFile(fmt.Sprintf("%s/cmd/main.go", flagName), []byte(main), 0777)
	if err != nil {
		fmt.Printf("Error creating main.go : %s", err.Error())
	}
	fmt.Println("main.go created")

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("error getting working dir: %s", err)
	}
	// go mod init creation
	cmd := exec.Command("go", "mod", "init", "example.com")
	cmd.Dir = fmt.Sprintf("%s/%s", currentDir, flagName)

	// capture the exec.Command errors in more detail
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("go mode init error: %s : %s", err.Error(), stderr.String())
	}
	fmt.Println("go init completed... starting go mod tidy...")
	// go mod tidy process
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = fmt.Sprintf("%s/%s", currentDir, flagName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("go mod tidy error: %s : %s", err.Error(), stderr.String())
	}
	fmt.Println("go mod tidy completed")

	// create run.sh file
	err = os.WriteFile(fmt.Sprintf("%s/run.sh", flagName), []byte(run), 0777)
	if err != nil {
		fmt.Printf("Error creating run.sh : %s", err.Error())
	}
	fmt.Println("run.sh created")

	// create deploy.sh file
	err = os.WriteFile(fmt.Sprintf("%s/deploy.sh", flagName), []byte(deploy), 0777)
	if err != nil {
		fmt.Printf("Error creating deploy.sh : %s", err.Error())
	}
	fmt.Println("deploy.sh created")
}
