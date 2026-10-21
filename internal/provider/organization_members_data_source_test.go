// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationMembersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "hatchet_organization_members" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hatchet_organization_members.test", "id"),
					resource.TestCheckResourceAttrSet("data.hatchet_organization_members.test", "members.#"),
				),
			},
		},
	})
}
