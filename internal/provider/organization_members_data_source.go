// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ datasource.DataSource              = &OrganizationMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &OrganizationMembersDataSource{}
)

func NewOrganizationMembersDataSource() datasource.DataSource {
	return &OrganizationMembersDataSource{}
}

type OrganizationMembersDataSource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type OrganizationMembersDataSourceModel struct {
	ID      types.String                        `tfsdk:"id"`
	Members []OrganizationMemberDataSourceModel `tfsdk:"members"`
}

type OrganizationMemberDataSourceModel struct {
	Email    types.String `tfsdk:"email"`
	MemberID types.String `tfsdk:"member_id"`
	Role     types.String `tfsdk:"role"`
	Status   types.String `tfsdk:"status"`
}

func (d *OrganizationMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_members"
}

func (d *OrganizationMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches accepted members of the configured Hatchet organization. Pending invitations are not returned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Hatchet organization ID from the provider token.",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "Accepted members of the Hatchet organization.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							MarkdownDescription: "Email address of the accepted organization member.",
							Computed:            true,
						},
						"member_id": schema.StringAttribute{
							MarkdownDescription: "Hatchet organization member ID.",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Hatchet organization role.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Member status. Organization members returned by this data source are always ACCEPTED.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HatchetCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HatchetCloudClient, got: %T", req.ProviderData),
		)
		return
	}

	apiClient, err := createAPIClient(client.Endpoint, client.Token, client.ProviderVersion)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create API Client", err.Error())
		return
	}

	d.client = apiClient
	d.organizationID = client.OrganizationID
}

func (d *OrganizationMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	orgID, err := uuid.Parse(d.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	orgResp, err := d.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	if orgResp.StatusCode() < 200 || orgResp.StatusCode() >= 300 || orgResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read organization, got status: %d", orgResp.StatusCode()))
		return
	}

	members := make([]OrganizationMemberDataSourceModel, 0)
	if orgResp.JSON200.Members != nil {
		members = make([]OrganizationMemberDataSourceModel, 0, len(*orgResp.JSON200.Members))
		for _, member := range *orgResp.JSON200.Members {
			members = append(members, OrganizationMemberDataSourceModel{
				Email:    types.StringValue(string(member.Email)),
				MemberID: types.StringValue(string(member.Metadata.Id)),
				Role:     types.StringValue(string(member.Role)),
				Status:   types.StringValue("ACCEPTED"),
			})
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].Email.ValueString() < members[j].Email.ValueString()
	})

	state := OrganizationMembersDataSourceModel{
		ID:      types.StringValue(d.organizationID),
		Members: members,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
