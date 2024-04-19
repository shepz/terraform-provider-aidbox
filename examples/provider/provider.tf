terraform {
  required_providers {
    aidboxlicense = {
      source = "hashicorp.com/edu/aidboxlicense"
      version = "~> 1.0"
    }
  }
}

provider "aidboxlicense" {
  endpoint = "http://localhost:8080"
}

resource "aidboxlicense_example" "example" {
  configurable_attribute = "some-value"
}
