package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Ideally, the following would be separated into a client.go file.

// This is a small CLI program for simplifying interaction with the IGDB: https://www.igdb.com.
// Refer to these docs to get started: https://api-docs.igdb.com/#getting-started.
// And these docs for examples of the endpoints and queries supported: https://api-docs.igdb.com/?shell#examples.
const (
	// Constants used for authentication with the Twitch developer API.
	TWITCH_AUTH_URL                = "https://id.twitch.tv/oauth2/token"
	TWITCH_CLIENT_ID_ENV_VAR       = "CLIENT_ID"
	TWICTH_CLIENT_SECRET_ENV_VAR   = "CLIENT_SECRET"
	DEFAULT_TWITCH_AUTH_GRANT_TYPE = "client_credentials"

	// Constants for interacting with the IGDB developer API.
	IGDB_BASE_URL          = "https://api.igdb.com/v4"
	IGDB_CLIENT_ID_HEADER  = "Client-ID"
	IGDB_AUTH_TOKEN_HEADER = "Authorization"

	// Defined exit codes for context when the program errors.
	BAD_USAGE_EXIT_CODE      = 1
	INTERNAL_ERROR_EXIT_CODE = 2
)

// DatabaseClient is a client for interacting with the IGDB.
type DatabaseClient struct {
	clientID  string
	authToken string
}

// NewDatabaseClient instantiates a new instance of the database client.
func NewDatabaseClient(clientID string, authToken string) *DatabaseClient {
	return &DatabaseClient{
		clientID:  clientID,
		authToken: authToken,
	}
}

// newRequest instantiates a new request with the necessary headers.
func (d *DatabaseClient) newRequest(endpoint string, query string) (*http.Request, error) {
	reqBody := bytes.NewReader([]byte(query))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", IGDB_BASE_URL, endpoint), reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add(IGDB_CLIENT_ID_HEADER, d.clientID)
	req.Header.Add(IGDB_AUTH_TOKEN_HEADER, fmt.Sprintf("Bearer %s", d.authToken))
	return req, nil
}

// parseResponse parses the response body into a JSON string.
func (d *DatabaseClient) parseResponse(resp *http.Response) (string, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

// Query queries the client database and returns the parsed JSON response.
func (d *DatabaseClient) Query(endpoint string, query string) (string, error) {
	req, err := d.newRequest(endpoint, query)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s", err.Error())
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %s", err.Error())
	}

	parsedResp, err := d.parseResponse(resp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %s", err.Error())
	}

	return parsedResp, nil
}

// Ideally, the following would be separated into the main.go file.

// Start point of program execution.
func main() {
	// Validate the user input an endpoint and query.
	if len(os.Args) != 3 {
		printUsage(BAD_USAGE_EXIT_CODE)
	}

	// Initiliaze client data and get auth token.
	clientID, clientSecret, err := getClientIDAndSecret()
	if err != nil {
		handleErr("failed to retrieve client ID and secret", err, INTERNAL_ERROR_EXIT_CODE)
	}
	authToken, err := getAuthToken(clientID, clientSecret)
	if err != nil {
		handleErr("failed to get auth token", err, INTERNAL_ERROR_EXIT_CODE)
	}

	// Get input from the user for the query.
	endpoint := os.Args[1]
	query := os.Args[2]

	// Submit the query and display the results.
	databaseClient := NewDatabaseClient(clientID, authToken)
	queryResult, err := databaseClient.Query(endpoint, query)
	if err != nil {
		handleErr("failed to query the internet games database", err, INTERNAL_ERROR_EXIT_CODE)
	}

	fmt.Printf("Query result: \n%s\n", queryResult)
}

// twitchAuthBody represents the JSON request body for Twitch developer authentication.
type twitchAuthBody struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

// twitchAuthResponse represents the JSON response body for Twitch developer authentication.
type twitchAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int32  `json:"expires_in"`
	TokenType   string `json:"expires_in"`
}

// getClientIDAndSecret retrieves the client data from the local environment.
func getClientIDAndSecret() (string, string, error) {
	clientID := os.Getenv(TWITCH_CLIENT_ID_ENV_VAR)
	if clientID == "" {
		return "", "", fmt.Errorf("%s must be initialized", TWITCH_CLIENT_ID_ENV_VAR)
	}

	clientSecret := os.Getenv(TWICTH_CLIENT_SECRET_ENV_VAR)
	if clientSecret == "" {
		return "", "", fmt.Errorf("%s must be initialized", TWICTH_CLIENT_SECRET_ENV_VAR)
	}

	return clientID, clientSecret, nil
}

// getAuthToken retrieves a valid auth token from the Twitch developer API.
func getAuthToken(clientID string, clientSecret string) (string, error) {
	// Setup the request body.
	reqBody := &twitchAuthBody{
		ClientID:     os.Getenv(TWITCH_CLIENT_ID_ENV_VAR),
		ClientSecret: os.Getenv(TWICTH_CLIENT_SECRET_ENV_VAR),
		GrantType:    DEFAULT_TWITCH_AUTH_GRANT_TYPE,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	// Perform the request.
	resp, err := http.Post(TWITCH_AUTH_URL, "application/json", bodyReader)
	if err != nil {
		return "", err
	}

	// Parse the response body.
	respBody := &twitchAuthResponse{}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(respBytes, respBody)
	if err != nil {
		return "", err
	}

	return respBody.AccessToken, nil
}

// printUsage prints the program's usage to the console and exits.
func printUsage(exitCode int) {
	fmt.Printf("Usage: gamers-console \"<endpoint>\" \"<query>\"\n")
	os.Exit(exitCode)
}

// handleErr is a helper function for handling errors and exiting.
func handleErr(message string, err error, exitCode int) {
	fmt.Printf("%s with error: %s", message, err.Error())
	os.Exit(exitCode)
}
