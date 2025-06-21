package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
)

// Client represents a GraphQL client
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	headers    map[string]string
}

// NewClient creates a new GraphQL client
func NewClient(baseURL string, options ...ClientOption) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &Client{
		baseURL: parsedURL.JoinPath("/graphql"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client, nil
}

// ClientOption represents a client configuration option
type ClientOption func(*Client)

// WithHeaders sets custom headers for the client
func WithHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		for key, value := range headers {
			c.headers[key] = value
		}
	}
}

// WithTimeout sets a custom timeout for the HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// Request represents a GraphQL request
type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response represents a GraphQL response
type Response struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message   string                 `json:"message"`
	Locations []GraphQLErrorLocation `json:"locations,omitempty"`
	Path      []interface{}          `json:"path,omitempty"`
}

// GraphQLErrorLocation represents an error location
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Query executes a GraphQL query
func (c *Client) Query(ctx context.Context, query string, variables ...map[string]interface{}) (*Response, error) {
	GinkgoHelper()

	var vars map[string]interface{}
	if len(variables) > 0 {
		vars = variables[0]
	}

	request := Request{
		Query:     query,
		Variables: vars,
	}

	return c.execute(ctx, request)
}

// Mutate executes a GraphQL mutation
func (c *Client) Mutate(ctx context.Context, mutation string, variables ...map[string]interface{}) (*Response, error) {
	GinkgoHelper()

	var vars map[string]interface{}
	if len(variables) > 0 {
		vars = variables[0]
	}

	request := Request{
		Query:     mutation,
		Variables: vars,
	}

	return c.execute(ctx, request)
}

// Introspect performs a GraphQL introspection query
func (c *Client) Introspect(ctx context.Context) (*Response, error) {
	GinkgoHelper()

	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	return c.Query(ctx, introspectionQuery)
}

// GetSchema retrieves the GraphQL schema
func (c *Client) GetSchema(ctx context.Context) (*SchemaResponse, error) {
	GinkgoHelper()

	response, err := c.Introspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("introspection failed: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("introspection errors: %v", response.Errors)
	}

	// Parse the schema from the response
	schemaData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid introspection response format")
	}

	schema, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing __schema in introspection response")
	}

	types, ok := schema["types"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing types in schema")
	}

	var typesList []TypeInfo
	for _, t := range types {
		typeMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := typeMap["name"].(string)
		kind, _ := typeMap["kind"].(string)
		
		typeInfo := TypeInfo{
			Name: name,
			Kind: kind,
		}

		if fields, ok := typeMap["fields"].([]interface{}); ok {
			for _, f := range fields {
				if fieldMap, ok := f.(map[string]interface{}); ok {
					fieldName, _ := fieldMap["name"].(string)
					typeInfo.Fields = append(typeInfo.Fields, FieldInfo{Name: fieldName})
				}
			}
		}

		if enumValues, ok := typeMap["enumValues"].([]interface{}); ok {
			for _, e := range enumValues {
				if enumMap, ok := e.(map[string]interface{}); ok {
					enumName, _ := enumMap["name"].(string)
					typeInfo.EnumValues = append(typeInfo.EnumValues, EnumValueInfo{Name: enumName})
				}
			}
		}

		typesList = append(typesList, typeInfo)
	}

	return &SchemaResponse{
		Types: typesList,
	}, nil
}

// GetType retrieves information about a specific type
func (c *Client) GetType(ctx context.Context, typeName string) (*TypeInfo, error) {
	GinkgoHelper()

	schema, err := c.GetSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	for _, t := range schema.Types {
		if t.Name == typeName {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("type not found: %s", typeName)
}

// HasSubscriptionSupport checks if the GraphQL server supports subscriptions
func (c *Client) HasSubscriptionSupport(ctx context.Context) (bool, error) {
	GinkgoHelper()

	response, err := c.Introspect(ctx)
	if err != nil {
		return false, fmt.Errorf("introspection failed: %w", err)
	}

	if len(response.Errors) > 0 {
		return false, fmt.Errorf("introspection errors: %v", response.Errors)
	}

	schemaData, ok := response.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid introspection response format")
	}

	schema, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("missing __schema in introspection response")
	}

	subscriptionType, exists := schema["subscriptionType"]
	return exists && subscriptionType != nil, nil
}

// RawRequest sends a raw GraphQL request
func (c *Client) RawRequest(ctx context.Context, requestBody string) (*Response, error) {
	GinkgoHelper()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL.String(), bytes.NewReader([]byte(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (c *Client) execute(ctx context.Context, request Request) (*Response, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// SchemaResponse represents the response from a schema introspection
type SchemaResponse struct {
	Types []TypeInfo `json:"types"`
}

// TypeInfo represents information about a GraphQL type
type TypeInfo struct {
	Name       string          `json:"name"`
	Kind       string          `json:"kind"`
	Fields     []FieldInfo     `json:"fields,omitempty"`
	EnumValues []EnumValueInfo `json:"enumValues,omitempty"`
}

// FieldInfo represents information about a GraphQL field
type FieldInfo struct {
	Name string `json:"name"`
}

// EnumValueInfo represents information about a GraphQL enum value
type EnumValueInfo struct {
	Name string `json:"name"`
}