# JWT Revoke Go SDK

A Go SDK for the JWT Revoke API that provides easy integration with token revocation services. Built with resilience and reliability in mind using context support and configurable retry mechanisms.

## Installation

Install the package using go get:

go get github.com/jwtrevoke/go-sdk

## Features

- Automatic retry with exponential backoff
- Context support for request cancellation
- Configurable timeouts
- Built-in rate limit handling
- Structured error handling
- Functional options pattern
- Type-safe API

## Usage

### Basic Setup

import (
	"time"
	"github.com/jwtrevoke/go-sdk"
)

// Initialize with default options
client := jwtrevokeapi.NewClient("your_api_key_here")

// Or initialize with custom options
clientWithOptions := jwtrevokeapi.NewClient(
	"your_api_key_here",
	jwtrevokeapi.WithMaxRetries(3),
	jwtrevokeapi.WithTimeout(10*time.Second),
	jwtrevokeapi.WithRateLimitDelay(time.Second),
)

### List Revoked Tokens

tokens, err := client.ListRevokedTokens()
if err != nil {
	if clientErr, ok := err.(*jwtrevokeapi.ClientError); ok {
		fmt.Printf("API Error: %s (Status: %d)\n", clientErr.Message, clientErr.StatusCode)
		fmt.Printf("Error Data: %v\n", clientErr.Data)
		return
	}
	panic(err)
}

for _, token := range tokens {
	fmt.Printf("Token ID: %s, Reason: %s\n", token.ID, token.Reason)
}

### Revoke a Token

expiryDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
revokedToken, err := client.RevokeToken(
	"token_123",
	"Security breach",
	expiryDate,
)
if err != nil {
	if clientErr, ok := err.(*jwtrevokeapi.ClientError); ok {
		fmt.Printf("API Error: %s (Status: %d)\n", clientErr.Message, clientErr.StatusCode)
		fmt.Printf("Error Data: %v\n", clientErr.Data)
		return
	}
	panic(err)
}

### Delete a Revoked Token

err := client.DeleteRevokedToken("token_123")
if err != nil {
	if clientErr, ok := err.(*jwtrevokeapi.ClientError); ok {
		fmt.Printf("API Error: %s (Status: %d)\n", clientErr.Message, clientErr.StatusCode)
		fmt.Printf("Error Data: %v\n", clientErr.Data)
		return
	}
	panic(err)
}

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| MaxRetries | Maximum number of retry attempts | 3 |
| Timeout | Request timeout duration | 10 seconds |
| RateLimitDelay | Delay between rate limit retries | 1 second |
| BaseURL | API base URL | https://api.jwtrevoke.com |

## Error Handling

The SDK uses the ClientError type for error handling, which includes:

- Message: Human-readable error message
- StatusCode: HTTP status code
- Data: Raw response data from the API

## Types

### RevokedToken

type RevokedToken struct {
	ID            string    `json:"id"`
	JwtID         string    `json:"jwt_id"`
	Reason        string    `json:"reason"`
	ExpiryDate    time.Time `json:"expiry_date"`
	RevokedByEmail string   `json:"revoked_by_email,omitempty"`
}

### ClientError

type ClientError struct {
	StatusCode int
	Message    string
	Data       interface{}
}

## Best Practices

1. Context Usage: Always consider using context for request cancellation
2. Error Handling: Use type assertions to handle ClientError specifically
3. Timeout Configuration: Adjust timeouts based on your application's needs
4. Rate Limiting: The SDK handles rate limits automatically with exponential backoff

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue in the GitHub repository or contact our support team.