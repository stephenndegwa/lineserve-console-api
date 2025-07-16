package main

//We should not use gophercloud v1 because it is not maintained and has many issues.
//We should use gophercloud v2 instead.
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	baseURL = "http://localhost:8080"
)

var (
	token string
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Define commands
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	loginUsername := loginCmd.String("username", "", "OpenStack username")
	loginPassword := loginCmd.String("password", "", "OpenStack password")

	listInstancesCmd := flag.NewFlagSet("list-instances", flag.ExitOnError)
	createInstanceCmd := flag.NewFlagSet("create-instance", flag.ExitOnError)
	instanceName := createInstanceCmd.String("name", "", "Instance name")
	flavorID := createInstanceCmd.String("flavor", "", "Flavor ID")
	imageID := createInstanceCmd.String("image", "", "Image ID")
	networkID := createInstanceCmd.String("network", "", "Network ID")
	keyName := createInstanceCmd.String("key", "", "Key name (optional)")

	listImagesCmd := flag.NewFlagSet("list-images", flag.ExitOnError)
	listFlavorsCmd := flag.NewFlagSet("list-flavors", flag.ExitOnError)
	listNetworksCmd := flag.NewFlagSet("list-networks", flag.ExitOnError)

	// Check if a command is provided
	if len(os.Args) < 2 {
		fmt.Println("Expected a command")
		fmt.Println("Available commands: login, list-instances, create-instance, list-images, list-flavors, list-networks")
		os.Exit(1)
	}

	// Load token from environment variable
	token = os.Getenv("LINESERVE_TOKEN")

	// Parse command
	switch os.Args[1] {
	case "login":
		loginCmd.Parse(os.Args[2:])
		if *loginUsername == "" {
			*loginUsername = os.Getenv("OS_USERNAME")
		}
		if *loginPassword == "" {
			*loginPassword = os.Getenv("OS_PASSWORD")
		}
		if *loginUsername == "" || *loginPassword == "" {
			fmt.Println("Username and password are required")
			os.Exit(1)
		}
		login(*loginUsername, *loginPassword)

	case "list-instances":
		listInstancesCmd.Parse(os.Args[2:])
		if token == "" {
			fmt.Println("Not authenticated. Run 'login' command first or set LINESERVE_TOKEN environment variable")
			os.Exit(1)
		}
		listInstances()

	case "create-instance":
		createInstanceCmd.Parse(os.Args[2:])
		if token == "" {
			fmt.Println("Not authenticated. Run 'login' command first or set LINESERVE_TOKEN environment variable")
			os.Exit(1)
		}
		if *instanceName == "" || *flavorID == "" || *imageID == "" || *networkID == "" {
			fmt.Println("Name, flavor, image, and network are required")
			os.Exit(1)
		}
		createInstance(*instanceName, *flavorID, *imageID, *networkID, *keyName)

	case "list-images":
		listImagesCmd.Parse(os.Args[2:])
		if token == "" {
			fmt.Println("Not authenticated. Run 'login' command first or set LINESERVE_TOKEN environment variable")
			os.Exit(1)
		}
		listImages()

	case "list-flavors":
		listFlavorsCmd.Parse(os.Args[2:])
		if token == "" {
			fmt.Println("Not authenticated. Run 'login' command first or set LINESERVE_TOKEN environment variable")
			os.Exit(1)
		}
		listFlavors()

	case "list-networks":
		listNetworksCmd.Parse(os.Args[2:])
		if token == "" {
			fmt.Println("Not authenticated. Run 'login' command first or set LINESERVE_TOKEN environment variable")
			os.Exit(1)
		}
		listNetworks()

	default:
		fmt.Println("Unknown command")
		fmt.Println("Available commands: login, list-instances, create-instance, list-images, list-flavors, list-networks")
		os.Exit(1)
	}
}

func login(username, password string) {
	// Create request body
	reqBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

	// Create request
	req, err := http.NewRequest("POST", baseURL+"/login", strings.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Parse response
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	// Save token
	fmt.Printf("Logged in successfully. Token: %s\n", loginResp.Token)
	fmt.Println("Set the LINESERVE_TOKEN environment variable to use this token in future commands:")
	fmt.Printf("export LINESERVE_TOKEN=%s\n", loginResp.Token)
}

func listInstances() {
	// Create request
	req, err := http.NewRequest("GET", baseURL+"/api/instances", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(prettyJSON.String())
}

func createInstance(name, flavorID, imageID, networkID, keyName string) {
	// Create request body
	reqBody := fmt.Sprintf(`{"name":"%s","flavor_id":"%s","image_id":"%s","network_id":"%s"`, name, flavorID, imageID, networkID)
	if keyName != "" {
		reqBody += fmt.Sprintf(`,"key_name":"%s"`, keyName)
	}
	reqBody += "}"

	// Create request
	req, err := http.NewRequest("POST", baseURL+"/api/instances", strings.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(prettyJSON.String())
}

func listImages() {
	// Create request
	req, err := http.NewRequest("GET", baseURL+"/api/images", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(prettyJSON.String())
}

func listFlavors() {
	// Create request
	req, err := http.NewRequest("GET", baseURL+"/api/flavors", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(prettyJSON.String())
}

func listNetworks() {
	// Create request
	req, err := http.NewRequest("GET", baseURL+"/api/networks", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(prettyJSON.String())
}
