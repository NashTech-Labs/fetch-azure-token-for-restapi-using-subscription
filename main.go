package main

import (
        "fmt"
        "os"
        "os/exec"
        "strings"
        "encoding/json"
        "net/http"
) 

var (
	globalEnvVars = make(map[string]string)
)


type ResourceGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}


func setTerraformVariables() (map[string]string, error) {
    // Getting enVars from environment variables
    ARM_CLIENT_ID := os.Getenv("AZURE_CLIENT_ID")
    ARM_CLIENT_SECRET := os.Getenv("AZURE_CLIENT_SECRET")
    ARM_TENANT_ID := os.Getenv("AZURE_TENANT_ID")
    ARM_SUBSCRIPTION_ID := os.Getenv("AZURE_SUBSCRIPTION_ID")

    // Print the exported values
    fmt.Println("Exported Values:")
    fmt.Printf("ARM_CLIENT_ID: %s\n", ARM_CLIENT_ID)
    fmt.Printf("ARM_CLIENT_SECRET: %s\n", ARM_CLIENT_SECRET)
    fmt.Printf("ARM_TENANT_ID: %s\n", ARM_TENANT_ID)
    fmt.Printf("ARM_SUBSCRIPTION_ID: %s\n", ARM_SUBSCRIPTION_ID)

    // Creating globalEnVars for terraform call through Terratest
    if ARM_CLIENT_ID != "" {
        globalEnvVars["ARM_CLIENT_ID"] = ARM_CLIENT_ID
        globalEnvVars["ARM_CLIENT_SECRET"] = ARM_CLIENT_SECRET
        globalEnvVars["ARM_SUBSCRIPTION_ID"] = ARM_SUBSCRIPTION_ID
        globalEnvVars["ARM_TENANT_ID"] = ARM_TENANT_ID
    }

    return globalEnvVars, nil
}


func getAccessToken(subscriptionID string) (string, error) {
    cmd := exec.Command("az", "account", "get-access-token", "--query", "accessToken", "--output", "tsv", "--subscription", subscriptionID)

    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(string(output)), nil
}



func main() {
	// Call setTerraformVariables to fetch the Azure environment variables and subscription ID
	envVars, err := setTerraformVariables()
	if err != nil {
		fmt.Printf("Error setting Terraform variables: %v\n", err)
		return
	}

	subscriptionID := envVars["ARM_SUBSCRIPTION_ID"]

	// Check if subscription ID is empty
	if subscriptionID == "" {
		fmt.Println("Azure subscription ID is not set")
		return
	}

	// Print the subscription ID
	fmt.Printf("Subscription ID: %s\n", subscriptionID)

	// Call the getAccessToken function to fetch the access token
	accessToken, err := getAccessToken(subscriptionID)
	if err != nil {
		fmt.Printf("Error getting access token: %v\n", err)
		return
	}

	// Print the access token
	fmt.Printf("Access Token: %s\n", accessToken)

	// Call the fetchResourceGroups function to fetch the resource groups
	resourceGroups, err := fetchResourceGroups(subscriptionID, accessToken)
	if err != nil {
		fmt.Printf("Error fetching resource groups: %v\n", err)
		return
	}

	// Print the resource groups
	fmt.Println("###########################################Resource Groups##################################################3")


    count := 1
    for _, rg := range resourceGroups {
        fmt.Printf("%d. Name: %s\n", count, rg.Name)
        fmt.Printf("   ID: %s\n", rg.ID)
        count++
    }
}


func fetchResourceGroups(subscriptionID, accessToken string) ([]ResourceGroup, error) {
	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourcegroups?api-version=2022-01-01", subscriptionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var response struct {
		Value []ResourceGroup `json:"value"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return response.Value, nil
}