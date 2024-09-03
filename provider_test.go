package main

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"hlb": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HLB_API_KEY"); v == "" {
		t.Fatal("HLB_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("HLB_API_URL"); v == "" {
		t.Fatal("HLB_API_URL must be set for acceptance tests")
	}
}