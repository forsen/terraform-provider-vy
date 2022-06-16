package provider

import (
	"context"
	"fmt"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type appClientResourceType struct{}

func (t appClientResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "An app client, used to access resource servers.",

		Attributes: map[string]tfsdk.Attribute{
			// id is required by the SDKv2 testing framework.
			// See https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			// TODO: Later, this is probably going to be the generated ID from cognito.
			"id": {
				Type: types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Computed: true,
			},
			"name": {
				MarkdownDescription: "The name of this app client",
				Required:            true,
				Type:                types.StringType,
			},
			"scopes": {
				MarkdownDescription: "Scopes that this client has access to",
				Optional:            true,
				Type:                types.ListType{ElemType: types.StringType},
			},
		},
	}, nil
}

func (t appClientResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return appClientResource{
		provider: provider,
	}, diags
}

type appClientResourceData struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Scopes []string     `tfsdk:"scopes"`
}

type appClientResource struct {
	provider provider
}

func (ac appClientResourceData) toDomain(domain *central_cognito.AppClient) {
	domain.Name = ac.Name.Value
	domain.Scopes = ac.Scopes
}

func appClientResourceDataFromDomain(domain central_cognito.AppClient, state *appClientResourceData) {
	state.Id.Value = domain.Name
	state.Id.Null = false
	state.Name.Value = domain.Name
	state.Scopes = domain.Scopes
}

func (r appClientResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data appClientResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id.Value = data.Name.Value
	data.Id.Null = false

	var appClient central_cognito.AppClient
	data.toDomain(&appClient)

	err := r.provider.Client.CreateAppClient(appClient)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Could not create app client",
			fmt.Sprintf("App client with name %s could not be created: %s", appClient.Name, err.Error()),
		)
		resp.Diagnostics.Append(diags...)

		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r appClientResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data appClientResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var server central_cognito.AppClient
	err := r.provider.Client.ReadAppClient(data.Name.Value, &server)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to read resource server",
			fmt.Sprintf("Can't read resource server %s from remote: %s ", data.Name.Value, err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	appClientResourceDataFromDomain(server, &data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r appClientResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data appClientResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id.Value = data.Name.Value
	data.Id.Null = false

	var appClient central_cognito.AppClient
	data.toDomain(&appClient)

	err := r.provider.Client.UpdateAppClient(central_cognito.AppClientUpdateRequest{
		Name:   appClient.Name,
		Scopes: appClient.Scopes,
	})
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to update app client",
			fmt.Sprintf("Can't update app client %s in remote: %s ", data.Name.Value, err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r appClientResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data appClientResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.Client.DeleteAppClient(data.Name.Value)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to delete app client",
			fmt.Sprintf("Can't delete app client %s in remote: %s ", data.Name.Value, err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r appClientResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	panic("implement me")
}