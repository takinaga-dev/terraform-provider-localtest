package localtest

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceLocalFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalFileCreate,
		Read:   resourceLocalFileRead,
		Delete: resourceLocalFileDelete,
		Update: resourceLocalFileUpdate,

		Schema: map[string]*schema.Schema{
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"sensitive_content", "content_base64", "source"},
			},
			"sensitive_content": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"content", "content_base64", "source"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content", "source"},
			},
			"filename": {
				Type:        schema.TypeString,
				Description: "Path to the output file",
				Required:    true,
				ForceNew:    true,
			},
			"file_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for the output file",
				Optional:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"directory_permission": {
				Type:         schema.TypeString,
				Description:  "Permissions to set for directories created",
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"source": {
				Type:          schema.TypeString,
				Description:   "Path to file to use as source for content of output file",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"content", "sensitive_content", "content_base64"},
			},
		},
	}
}

func resourceLocalFileUpdate(d *schema.ResourceData, m interface{}) error {
	outputPath := d.Get("filename").(string)
	filePerm := d.Get("file_permission").(string)
	content := d.Get("content").(string)
	perm32, _ := strconv.ParseUint(filePerm, 8, 32)

	err := ioutil.WriteFile(outputPath, []byte(content), os.FileMode(perm32))
	if err != nil {
		return nil
	}

	os.Chmod(outputPath, os.FileMode(perm32))

	return resourceLocalFileRead(d, m)
}

func resourceLocalFileRead(d *schema.ResourceData, _ interface{}) error {
	// If the output file doesn't exist, mark the resource for creation.
	outputPath := d.Get("filename").(string)
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	// Verify that the content of the destination file matches the content we
	// expect. Otherwise, the file might have been modified externally and we
	// must reconcile.
	outputContent, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return err
	}

	checksum := sha1.Sum([]byte(outputContent))
	d.SetId(hex.EncodeToString(checksum[:]))
	// do not check content
	/*if hex.EncodeToString(outputChecksum[:]) != d.Id() {
		d.SetId("")
		return nil
	}*/

	// insert file_permission and content to ResourceData
	d.Set("file_permission", fmt.Sprintf("%#o", info.Mode()))
	d.Set("content", string(outputContent))

	return nil
}

func resourceLocalFileContent(d *schema.ResourceData) ([]byte, error) {
	if content, sensitiveSpecified := d.GetOk("sensitive_content"); sensitiveSpecified {
		return []byte(content.(string)), nil
	}
	if b64Content, b64Specified := d.GetOk("content_base64"); b64Specified {
		return base64.StdEncoding.DecodeString(b64Content.(string))
	}

	if v, ok := d.GetOk("source"); ok {
		source := v.(string)
		return ioutil.ReadFile(source)
	}

	content := d.Get("content")
	return []byte(content.(string)), nil
}

func resourceLocalFileCreate(d *schema.ResourceData, _ interface{}) error {
	content, err := resourceLocalFileContent(d)
	if err != nil {
		return err
	}

	destination := d.Get("filename").(string)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		dirPerm := d.Get("directory_permission").(string)
		dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)
		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return err
		}
	}

	filePerm := d.Get("file_permission").(string)

	fileMode, _ := strconv.ParseInt(filePerm, 8, 64)

	if err := ioutil.WriteFile(destination, []byte(content), os.FileMode(fileMode)); err != nil {
		return err
	}

	checksum := sha1.Sum([]byte(content))
	d.SetId(hex.EncodeToString(checksum[:]))

	return nil
}

func resourceLocalFileDelete(d *schema.ResourceData, _ interface{}) error {
	os.Remove(d.Get("filename").(string))
	return nil
}
