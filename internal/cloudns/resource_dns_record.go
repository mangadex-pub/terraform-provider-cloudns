package cloudns

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/wolframite/cloudns-go"
	"strings"
)

func resourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A simple DNS record.",

		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,
		CustomizeDiff: resourceDnsRecordValidate,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDnsRecordImport,
		},

		// Naming **does not** follow the scheme used by ClouDNS, due to how comically misleading and unclear it is
		// see: https://www.cloudns.net/wiki/article/58/ for the relevant "vanilla" schema on ClouDNS side
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the record (eg: `[something].cloudns.net 600 in A 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"zone": {
				Description: "The zone on which to add the record (eg: `something.[cloudns.net] 600 in A 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"ttl": {
				Description: "The TTL to assign to the record (eg: `something.cloudns.net [600] in A 1.2.3.4`)",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    false,
			},
			"type": {
				Description: "The type of record (eg: `something.cloudns.net 600 in [A] 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"value": {
				Description: "Value of the record (eg: `something.cloudns.net 600 in A [1.2.3.4]`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"priority": {
				Description:      "Priority for MX record (eg: `something.cloudns.net 600 in MX [10] mail.example.com`)",
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         false,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 65535)),
			},
			"weight": {
				Description:      "Weight for SRV record",
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         false,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 65535)),
			},
			"port": {
				Description:      "Port for SRV record",
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         false,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(0, 65535)),
			},
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientConfig := meta.(ClientConfig)
	recordToCreate := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("CREATE %s.%s %d in %s %s", recordToCreate.Host, recordToCreate.Domain, recordToCreate.TTL, recordToCreate.Rtype, recordToCreate.Record))

	clientConfig.rateLimiter.Take()
	recordCreated, err := recordToCreate.Create(&clientConfig.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordCreated.ID)

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		d.SetId("")
		return nil
	}

	config := meta.(ClientConfig)
	lookup := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("READ Record#%s (%s.%s %d in %s %s)", lookup.ID, lookup.Host, lookup.Domain, lookup.TTL, lookup.Rtype, lookup.Record))

	config.rateLimiter.Take()
	zoneRead, err := cloudns.Zone{Domain: lookup.Domain}.List(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, zoneRecord := range zoneRead {
		wantedId := d.Id()
		actualId := zoneRecord.ID
		if wantedId == actualId {
			err = updateState(d, &zoneRecord)
			if err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("UPDATE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))

	config.rateLimiter.Take()
	updated, err := record.Update(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(updated.ID)

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("DELETE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))

	config.rateLimiter.Take()
	_, err := record.Destroy(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordValidate(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	rtype := d.Get("type").(string)
	_, isPriorityProvided := d.GetOkExists("priority")
	_, isWeightProvided := d.GetOkExists("weight")
	_, isPortProvided := d.GetOkExists("port")
	if rtype == "MX" && !isPriorityProvided {
		return fmt.Errorf("Priority is required for MX record")
	}

	if rtype == "SRV" && !isPriorityProvided {
		return fmt.Errorf("Priority is required for MX record")
	}

	if rtype == "SRV" && !isWeightProvided {
		return fmt.Errorf("Weight is required for MX record")
	}

	if rtype == "SRV" && !isPortProvided {
		return fmt.Errorf("Port is required for MX record")
	}

	return nil
}

func resourceDnsRecordImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ClientConfig)

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Bad ID format: %#v. Expected: \"zone/id\"", d.Id())
	}
	zone := parts[0]
	wantedId := parts[1]

	config.rateLimiter.Take()
	zoneRead, err := cloudns.Zone{Domain: zone}.List(&config.apiAccess)
	if err != nil {
		return nil, err
	}

	idx := -1
	for i, zoneRecord := range zoneRead {
		if wantedId == zoneRecord.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("Record not found: %#v", wantedId)
	}

	zoneRecord := zoneRead[idx]
	err = updateState(d, &zoneRecord)
	if err != nil {
		return nil, err
	}
	d.SetId(wantedId)

	tflog.Debug(ctx, fmt.Sprintf("IMPORT %s.%s %d in %s %s", zoneRecord.Host, zoneRecord.Domain, zoneRecord.TTL, zoneRecord.Rtype, zoneRecord.Record))

	return []*schema.ResourceData{d}, nil
}

func updateState(d *schema.ResourceData, zoneRecord *cloudns.Record) error {
	err := d.Set("name", zoneRecord.Host)
	if err != nil {
		return err
	}

	err = d.Set("zone", zoneRecord.Domain)
	if err != nil {
		return err
	}

	err = d.Set("type", zoneRecord.Rtype)
	if err != nil {
		return err
	}

	err = d.Set("value", zoneRecord.Record)
	if err != nil {
		return err
	}

	err = d.Set("ttl", zoneRecord.TTL)
	if err != nil {
		return err
	}

	if zoneRecord.Rtype == "MX" || zoneRecord.Rtype == "SRV" {
		err = d.Set("priority", zoneRecord.Priority)
		if err != nil {
			return err
		}
	}

	if zoneRecord.Rtype == "SRV" {
		err = d.Set("weight", zoneRecord.Weight)
		if err != nil {
			return err
		}
	}

	if zoneRecord.Rtype == "SRV" {
		err = d.Set("port", zoneRecord.Port)
		if err != nil {
			return err
		}
	}

	return nil
}

func toApiRecord(d *schema.ResourceData) cloudns.Record {
	id := d.Id()
	name := d.Get("name").(string)
	zone := d.Get("zone").(string)
	rtype := d.Get("type").(string)
	value := d.Get("value").(string)
	ttl := d.Get("ttl").(int)
	priority := 0
	weight := 0
	port := 0
	if rtype == "MX" {
		priority = d.Get("priority").(int)
	}

	if rtype == "SRV" {
		priority = d.Get("priority").(int)
		weight = d.Get("weight").(int)
		port = d.Get("port").(int)
	}

	return cloudns.Record{
		ID:       id,
		Host:     name,
		Domain:   zone,
		Rtype:    rtype,
		Record:   value,
		TTL:      ttl,
		Priority: priority,
		Weight:   weight,
		Port:     port,
	}
}
