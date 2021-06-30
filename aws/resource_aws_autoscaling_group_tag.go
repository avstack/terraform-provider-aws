package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAutoscalingGroupTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAutoscalingGroupTagCreate,
		Read:   resourceAwsAutoscalingGroupTagRead,
		Update: resourceAwsAutoscalingGroupTagUpdate,
		Delete: resourceAwsAutoscalingGroupTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"asg_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func extractAutoscalingGroupNameAndKeyFromTagID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource ID; cannot look up resource: %s", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsAutoscalingGroupTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	asgName := d.Get("asg_name").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, nil, map[string]string{key: value}); err != nil {
		return fmt.Errorf("error updating Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", asgName, key))

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromTagID(d.Id())

	if err != nil {
		return err
	}

	exists, value, err := keyvaluetags.AutoscalingGetTag(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, key)

	if err != nil {
		return fmt.Errorf("error reading Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	if !exists {
		log.Printf("[WARN] Autoscaling Tag (%s) for resource (%s) not found, removing from state", key, asgName)
		d.SetId("")
		return nil
	}

	d.Set("key", key)
	d.Set("asg_name", asgName)
	d.Set("value", value)

	return nil
}

func resourceAwsAutoscalingGroupTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromTagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, nil, map[string]string{key: d.Get("value").(string)}); err != nil {
		return fmt.Errorf("error updating Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	return resourceAwsAutoscalingGroupTagRead(d, meta)
}

func resourceAwsAutoscalingGroupTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn
	asgName, key, err := extractAutoscalingGroupNameAndKeyFromTagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.AutoscalingUpdateTags(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, map[string]string{key: d.Get("value").(string)}, nil); err != nil {
		return fmt.Errorf("error deleting Autoscaling Tag (%s) for resource (%s): %w", key, asgName, err)
	}

	return nil
}
