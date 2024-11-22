/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

//go:embed _templates/*
var templates embed.FS

// buildGenesysCmd represents the buildGenesys command
var buildGenesysCmd = &cobra.Command{
	Use:   "buildGenesys",
	Short: "Build a Google Cloud Go function gen2 template",
	Long: `This uses the Go cloud functions framework to allow 
	for localhost testing before easy cloud deploying. For example:

Middleware between a website or mobile app and Genesys Cloud Platform
API calls using client credentials on the server and Basic auth on the client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("buildGenesys called")
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
	rootCmd.AddCommand(buildGenesysCmd)
	buildGenesysCmd.PersistentFlags().String("name", "", "A name for the cloud function")
	buildGenesysCmd.PersistentFlags().String("region", "", "The Genesys Cloud Region eg: mypurecloud.com.au")
	buildGenesysCmd.PersistentFlags().String("clientId", "", "OAuth 2.0 ClientId")
	buildGenesysCmd.PersistentFlags().String("secret", "", "OAuth 2.0 secret")
	buildGenesysCmd.PersistentFlags().String("username", "", "A Basic username for external auth")
	buildGenesysCmd.PersistentFlags().String("password", "", "A Basic password for external auth")
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
	function, err := templates.ReadFile("_templates/function.go")
	if err != nil {
		fmt.Println(err.Error())
		_ = os.RemoveAll(flagName)
		return
	}
	err = os.WriteFile(fmt.Sprintf("%s/function.go", flagName), []byte(function), 0777)
	if err != nil {
		fmt.Println(err.Error())
		_ = os.RemoveAll(flagName)
		return
	}
	fmt.Println("function.go created")

	// create genesys.go file
	genesys, err := templates.ReadFile("_templates/genesys.go")
	if err != nil {
		fmt.Println(err.Error())
		_ = os.RemoveAll(flagName)
		return
	}
	err = os.WriteFile(fmt.Sprintf("%s/genesys.go", flagName), []byte(genesys), 0777)
	if err != nil {
		fmt.Printf("Error creating genesys.go : %s", err.Error())
	}
	fmt.Println("genesys.go created")

	// create main.go file
	main, err := templates.ReadFile("_templates/cmd/main.go")
	if err != nil {
		fmt.Printf("Error reading embedded file main.go : %s", err.Error())
	}
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
