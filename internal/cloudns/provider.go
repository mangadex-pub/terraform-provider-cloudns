package cloudns

import (
	"context"
	"github.com/sta-travel/cloudns-go"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"auth_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CLOUDNS_AUTH_ID", nil),
				},
				"sub_auth_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("CLOUDNS_SUB_AUTH_ID", nil),
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("CLOUDNS_PASSWORD", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{},
			ResourcesMap: map[string]*schema.Resource{
				"cloudns_dns_record": resourceDnsRecord(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type ClientConfig struct {
	apiAccess cloudns.Apiaccess
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		authId := d.Get("auth_id").(string)
		subAuthId := d.Get("sub_auth_id").(string)
		password := d.Get("password").(string)

		if len(password) == 0 {
			return nil, diag.Errorf("Expected password to be defined but it wasn't")
		}

		if (len(authId) > 0) == (len(subAuthId) > 0) {
			var golangSucks = "not defined"
			if len(authId) > 0 {
				golangSucks = "defined"
			}
			return nil, diag.Errorf("Exactly one of auth_id or sub_auth_id must be set, but both were %s", golangSucks)
		}

		if len(authId) > 0 {
			authIdInt, err := strconv.Atoi(authId)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			return &ClientConfig{
				apiAccess: cloudns.Apiaccess{
					Authid:       authIdInt,
					Authpassword: password,
				},
			}, nil
		} else {
			subAuthIdInt, err := strconv.Atoi(subAuthId)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			return &ClientConfig{
				apiAccess: cloudns.Apiaccess{
					Subauthid:    subAuthIdInt,
					Authpassword: password,
				},
			}, nil
		}

	}
}
