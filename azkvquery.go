package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/pborman/getopt/v2"
)

func main() {

	vaultURI, mySecretName := getConfig()

	// Create a credential using the NewDefaultAzureCredential type.
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	// Establish a connection to the Key Vault client
	client, err := azsecrets.NewClient(vaultURI, cred, nil)
	if err != nil {
		log.Fatalf("key vault connection failed: %v", err)
	}

	// Get a secret. An empty string version gets the latest version of the secret.
	version := ""
	resp, err := client.GetSecret(context.TODO(), mySecretName, version, nil)

	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.StatusCode {
		case http.StatusNotFound:
			log.Println("Secret not found")
			listSecrets(client)
		case http.StatusForbidden:
			log.Fatalln("No permission to access this secret")
		default:
			log.Fatalf("Unrecognized http error code %d", respErr.StatusCode)
		}
	} else {
		log.Fatalf("Error getting secret from keyvault: %v", err)
	}

	fmt.Printf("%s: %s\n", mySecretName, *resp.Value)

}

func listSecrets(c *azsecrets.Client) {
	pager := c.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
		for _, secret := range page.Value {
			fmt.Printf("Secret ID: %s\n", *secret.ID)
		}
	}
}

func printUsage() {
	getopt.Usage()
	os.Exit(0)
}

func getConfig() (string, string) {
	optKvault := getopt.StringLong("keyvault", 'v', "", "Keyvault URI, optionally env AZURE_KEY_VAULT_URI")
	optSname := getopt.StringLong("secret", 's', "", "Keyvault secret name, also env AZURE_SECRET_NAME")
	optHelp := getopt.BoolLong("help", 0, "Help")
	getopt.Parse()

	if *optHelp {
		printUsage()
	}

	if *optKvault == "" {
		vaultURI, ok := os.LookupEnv("AZURE_KEY_VAULT_URI")
		if !ok {
			printUsage()
		}
		*optKvault = vaultURI
	}

	if *optSname == "" {
		mySecretName, ok := os.LookupEnv("AZURE_SECRET_NAME")
		if !ok {
			printUsage()
		}
		*optSname = mySecretName
	}

	return *optKvault, *optSname
}
