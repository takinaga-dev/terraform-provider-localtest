package localtest

import (
	"github.com/facette/logger"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type localProvider struct {
	log *logger.Logger
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"localtest_file": resourceLocalFile(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"localtest_file": dataSourceLocalFile(),
		},
	}
}
