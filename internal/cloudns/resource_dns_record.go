package cloudns

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sta-travel/cloudns-go"
	"time"
)

func resourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A simple DNS record.",

		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,

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
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientConfig := meta.(ClientConfig)
	recordToCreate := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("CREATE %s.%s %d in %s %s", recordToCreate.Host, recordToCreate.Domain, recordToCreate.TTL, recordToCreate.Rtype, recordToCreate.Record))
	recordCreated, err := recordToCreate.Create(&clientConfig.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordCreated.ID)

	timeoutErr := resource.RetryContext(ctx, 30*time.Second, func() *resource.RetryError {
		recordRead, lookupError := resourceSimpleRead(ctx, d, meta)

		if lookupError != nil {
			return resource.NonRetryableError(*lookupError)
		}

		if recordRead == nil {
			return resource.RetryableError(errors.New("record wasn't visible yet"))
		}

		return nil
	})

	if timeoutErr != nil {
		return diag.FromErr(timeoutErr)
	}

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	recordRead, lookupError := resourceSimpleRead(ctx, d, meta)

	if lookupError != nil {
		return diag.FromErr(*lookupError)
	}

	if recordRead == nil {
		d.SetId("")
		return nil
	}

	err := d.Set("name", recordRead.Host)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("zone", recordRead.Domain)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("type", recordRead.Rtype)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("value", recordRead.Record)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("ttl", recordRead.TTL)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSimpleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (*cloudns.Record, *error) {
	config := meta.(ClientConfig)
	lookup := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("READ %s.%s %d in %s %s", lookup.Host, lookup.Domain, lookup.TTL, lookup.Rtype, lookup.Record))

	zoneRead, err := cloudns.Zone{Domain: lookup.Domain}.List(&config.apiAccess)
	if err != nil {
		return nil, &err
	}

	for _, record := range zoneRead {
		if record.ID == d.Id() {
			return &record, nil
		}
	}

	return nil, nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("UPDATE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))
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
	_, err := record.Destroy(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDnsRecordRead(ctx, d, meta)
}

func toApiRecord(d *schema.ResourceData) cloudns.Record {
	id := d.Id()
	name := d.Get("name").(string)
	zone := d.Get("zone").(string)
	rtype := d.Get("type").(string)
	value := d.Get("value").(string)
	ttl := d.Get("ttl").(int)

	return cloudns.Record{
		ID:     id,
		Host:   name,
		Domain: zone,
		Rtype:  rtype,
		Record: value,
		TTL:    ttl,
	}
}
