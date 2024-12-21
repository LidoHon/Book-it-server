package libs

import (
	"net/http"
	"os"

	"github.com/shurcooL/graphql"
)

type customTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers to the request
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}
	return c.transport.RoundTrip(req)
}

func SetupGraphqlClient() *graphql.Client {

	// return client
	httpClient := &http.Client{
		Transport: &customTransport{
			transport: http.DefaultTransport,
			headers: map[string]string{
				"x-hasura-admin-secret": os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET"),
			},
		},
	}

	// Create a GraphQL client with the HTTP client
	client := graphql.NewClient(os.Getenv("HASURA_GRAPHQL_API_ENDPOINT"), httpClient)

	return client
}
