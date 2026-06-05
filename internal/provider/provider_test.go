package provider

import (
	"context"
	"os"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"technitium": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func testAccClientFromEnv() (*client.Client, error) {
	serverURL := os.Getenv("TECHNITIUM_SERVER_URL")
	apiToken := os.Getenv("TECHNITIUM_API_TOKEN")
	if apiToken != "" {
		return client.NewWithToken(serverURL, apiToken, 0)
	}
	return client.NewWithCredentials(
		serverURL,
		os.Getenv("TECHNITIUM_USERNAME"),
		os.Getenv("TECHNITIUM_PASSWORD"),
		0,
	)
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	// Verify the provider can produce a protocol 6 server without error.
	_, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("unexpected error creating provider server: %s", err)
	}
}

func TestProviderSchema_HasRequiredAttributes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := New("test")()

	schemaResp := &fwprovider.SchemaResponse{}
	p.Schema(ctx, fwprovider.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors getting provider schema: %s", schemaResp.Diagnostics)
	}

	s := schemaResp.Schema

	// Verify server_url exists and is optional (env var fallback).
	serverURL, ok := s.Attributes["server_url"]
	if !ok {
		t.Fatal("expected server_url attribute in schema")
	}
	if !serverURL.IsOptional() {
		t.Error("expected server_url to be optional")
	}

	// Verify username exists and is optional.
	username, ok := s.Attributes["username"]
	if !ok {
		t.Fatal("expected username attribute in schema")
	}
	if !username.IsOptional() {
		t.Error("expected username to be optional")
	}
	if !username.IsSensitive() {
		t.Error("expected username to be sensitive")
	}

	// Verify password exists and is optional.
	password, ok := s.Attributes["password"]
	if !ok {
		t.Fatal("expected password attribute in schema")
	}
	if !password.IsOptional() {
		t.Error("expected password to be optional")
	}
	if !password.IsSensitive() {
		t.Error("expected password to be sensitive")
	}

	// Verify api_token exists and is optional.
	apiToken, ok := s.Attributes["api_token"]
	if !ok {
		t.Fatal("expected api_token attribute in schema")
	}
	if !apiToken.IsOptional() {
		t.Error("expected api_token to be optional")
	}
	if !apiToken.IsSensitive() {
		t.Error("expected api_token to be sensitive")
	}
}
