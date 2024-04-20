terraform {
   required_providers {
      aidbox = {
         source = "hashicorp.com/edu/aidbox"
         version = "~> 1.0"
      }
   }
}

provider "aidbox" {
   endpoint = "http://localhost:8080"
}
