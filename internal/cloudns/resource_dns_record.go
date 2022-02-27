package cloudns

import (
	"context"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sta-travel/cloudns-go"
)

func resourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A simple DNS record.",

		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,

		// Naming follows the scheme used by ClouDNS, despite how comically bad it is
		// see: https://www.cloudns.net/wiki/article/58/
		Schema: map[string]*schema.Schema{
			"host": {
				Description: "The name of the record (eg: `[something].cloudns.net 600 in A 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"domain_name": {
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
			"record_type": {
				Description: "The type of record (eg: `something.cloudns.net 600 in [A] 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"record": {
				Description: "Value of the record (eg: `something.cloudns.net 600 in A [1.2.3.4]`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientConfig := meta.(*ClientConfig)
	recordToCreate := toApiRecord(d)

	tflog.Trace(ctx, "Create %s.%s %d in %s %s", recordToCreate.Host, recordToCreate.Domain, recordToCreate.TTL, recordToCreate.Domain, recordToCreate.Record)
	recordCreated, err := recordToCreate.Create(&clientConfig.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordCreated.ID)

	return nil
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Trace(ctx, "Lookup %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record)
	recordRead, err := record.Read(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordRead.ID)
	d.Set("host", recordRead.Host)
	d.Set("ttl", recordRead.TTL)
	d.Set("domain_name", recordRead.Domain)
	d.Set("record_type", recordRead.Rtype)
	d.Set("record", recordRead.Record)

	return nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	_, err := record.Update(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	_, err := record.Destroy(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func toApiRecord(d *schema.ResourceData) cloudns.Record {
	id := d.Id()
	host := d.Get("host").(string)
	ttl := d.Get("ttl").(int)
	domainName := d.Get("domain_name").(string)
	recordType := d.Get("record_type").(string)
	recordValue := d.Get("record").(string)

	return cloudns.Record{
		Domain: domainName,
		ID:     id,
		Host:   host,
		Rtype:  recordType,
		Record: recordValue,
		TTL:    ttl,
	}
}
