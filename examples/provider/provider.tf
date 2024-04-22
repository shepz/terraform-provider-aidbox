terraform {
  required_providers {
    aidbox = {
      source = "hashicorp.com/edu/aidbox"
      version = "~> 1.0"
    }
  }
}

provider "aidbox" {
  endpoint = "https://aidbox.app/rpc"
}

resource "aidbox_license" "example" {
  name = "test-tf-guiguigui"
  type = "standard"
}
