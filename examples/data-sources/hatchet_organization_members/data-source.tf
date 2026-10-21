terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 1.1.3"
    }
  }
}

provider "hatchet" {
  # Token is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable.
}

data "hatchet_organization_members" "current" {}

output "accepted_member_emails" {
  value = [for member in data.hatchet_organization_members.current.members : member.email]
}
