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
