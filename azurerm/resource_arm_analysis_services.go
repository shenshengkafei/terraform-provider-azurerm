package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/analysisservices"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jen20/riviera/azure"
)

func resourceArmAnalysisServices() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmArmAnalysisServicesCreate,
		Read:   resourceArmArmAnalysisServicesRead,
		Update: resourceArmArmAnalysisServicesUpdate,
		Delete: resourceArmArmAnalysisServicesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sku_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sku_tier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": locationSchema(),

			"tags": tagsSchema(),
		},
	}
}

func resourceArmArmAnalysisServicesUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).analysisClient

	log.Printf("[INFO] preparing arguments for AzureRM Analysis Sever creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})
	skuName := d.Get("sku_name").(string)
	skuTier := d.Get("sku_tier").(string)
	server := analysisservices.Server{
		Name:     &name,
		Location: &location,
		Sku: &analysisservices.ResourceSku{
			Name: analysisservices.SkuName(skuName),
			Tier: analysisservices.SkuTier(skuTier),
		},
		Tags: expandTags(tags),
	}

	_, error := client.Create(resGroup, name, server, make(chan struct{}))
	err := <-error
	if err != nil {
		return err
	}

	read, err := client.GetDetails(resGroup, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read  Analysis Server %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmArmAnalysisServicesRead(d, meta)
}

func resourceArmArmAnalysisServicesCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).analysisClient

	log.Printf("[INFO] preparing arguments for AzureRM Analysis Sever creation.")

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})
	skuName := d.Get("sku_name").(string)
	skuTier := d.Get("sku_tier").(string)
	server := analysisservices.Server{
		Name:     &name,
		Location: &location,
		Sku: &analysisservices.ResourceSku{
			Name: analysisservices.SkuName(skuName),
			Tier: analysisservices.SkuTier(skuTier),
		},
		Tags: expandTags(tags),
	}

	_, error := client.Create(resGroup, name, server, make(chan struct{}))
	err := <-error
	if err != nil {
		return err
	}

	read, err := client.GetDetails(resGroup, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read  Analysis Server %s (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmArmAnalysisServicesRead(d, meta)
}

func resourceArmArmAnalysisServicesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).analysisClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["servers"]

	return fmt.Errorf("In Read method and the servers name is %s", name)

	resp, err := client.GetDetails(resGroup, name)
	if err != nil {
		if responseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure Analysis Server Set %s: %s", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("location", azureRMNormalizeLocation(*resp.Location))
	d.Set("resource_group_name", resGroup)

	if resp.Sku != nil {
		d.Set("sku_name", resp.Sku.Name)
		d.Set("sku_tier", resp.Sku.Tier)
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmArmAnalysisServicesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	rivieraClient := client.rivieraClient

	deleteRequest := rivieraClient.NewRequestForURI(d.Id())
	deleteRequest.Command = &azure.DeleteResourceGroup{}

	deleteResponse, err := deleteRequest.Execute()
	if err != nil {
		return fmt.Errorf("Error deleting resource group: %s", err)
	}
	if !deleteResponse.IsSuccessful() {
		return fmt.Errorf("Error deleting resource group: %s", deleteResponse.Error)
	}

	return nil

}
