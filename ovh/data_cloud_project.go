package ovh

import (
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudProject() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudProjectRead,
		Schema: map[string]*schema.Schema{
			// Either description or project_id can be used to identify a cloud project
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// Computed items
			"project_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCloudProjectRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	description := d.Get("description").(string)
	projectId := d.Get("project_id").(string)
	var projectIds []string
	project := CloudProject{}
	found := false

	if description == "" && projectId == "" {
		return fmt.Errorf("One of description or project_id must be provided")
	}

	// We can save ourselves an API call and an O(N) search if we know
	// the project ID, otherwise enumerate them all
	if projectId != "" {
		found = true
		projectIds = append(projectIds, projectId)
	} else {
		err := config.OVHClient.Get("/cloud/project", &projectIds)
		if err != nil {
			return fmt.Errorf("Error enumerating cloud projects:\n\t %q", err)
		}
	}

	// Linear search of projects, matching against "description"
	for _, projectId := range projectIds {
		endpoint := fmt.Sprintf("/cloud/project/%s", url.PathEscape(projectId))
		log.Printf("[DEBUG] GET cloud project: %s", endpoint)

		err := config.OVHClient.Get(endpoint, &project)
		if err != nil {
			return fmt.Errorf("Error calling %s:\n\t %q", endpoint, err)
		}

		if description != "" && *project.Description == description {
			found = true
		}

		if found {
			break
		}
	}

	// Check if we searched all projects
	if !found {
		return fmt.Errorf("Project not found")
	}

	for k, v := range project.ToMap() {
		log.Printf("[DEBUG]\t cloud project k: %s, v: %s", k, v)
		if k == "project_id" {
			d.SetId(fmt.Sprint(v))
		}

		d.Set(k, v)
	}

	return nil
}
