package executor

import (
	"github.com/akutz/gofig"
	"github.com/akutz/goof"

	"github.com/emccode/libstorage/api/registry"
	"github.com/emccode/libstorage/api/types"
	"github.com/emccode/libstorage/drivers/storage/openstack"

	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"fmt"
)

type driver struct {
	config gofig.Config
}

func init() {
	registry.RegisterStorageExecutor(openstack.Name, newDriver)
}

func newDriver() types.StorageExecutor {
	return &driver{}
}

func (d *driver) Init(ctx types.Context, config gofig.Config) error {
	d.config = config
	return nil
}

func (d *driver) Name() string {
	return openstack.Name
}

func (d *driver) InstanceID(
	ctx types.Context,
	opts types.Store) (*types.InstanceID, error) {

	uuid, err := getInstanceIDFromMetadataServer()
	if err != nil {
		metadataServerErr := err
		uuid, err = getInstanceIDWithDMIDecode()
		if err != nil {
			return nil, fmt.Errorf("%v ; %v", metadataServerErr, err)
		}
	}

	iid := &types.InstanceID{Driver: openstack.Name, ID: strings.ToLower(uuid)}

	return iid, nil
}

func getInstanceIDFromMetadataServer() (string, error) {
	conn, err := net.DialTimeout("tcp", "169.254.169.254:80", 50*time.Millisecond)
	if err != nil {
		return "", goof.WithError("error reaching OpenStack metadata server at 169.254.169.254", err)
	}
	defer conn.Close()

	url := "http://169.254.169.254/openstack/latest/meta_data.json"
	resp, err := http.Get(url)
	if err != nil {
		return "", goof.WithError("error getting metadata from " + url, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", goof.WithError("io error reading metadata", err)
	}
	var decodedJSON interface{}
	err = json.Unmarshal(data, &decodedJSON)
	if err != nil {
		return "", goof.WithError("error unmarshalling metadata", err)
	}
	decodedJSONMap, ok := decodedJSON.(map[string]interface{})
	if !ok {
		return "", goof.New("error casting metadata decoded JSON")
	}
	uuid, ok := decodedJSONMap["uuid"].(string)
	if !ok {
		return "", goof.New("error casting metadata uuid field")
	}

	return uuid, nil
}

func getInstanceIDWithDMIDecode() (string, error) {
	cmd := exec.Command("dmidecode", "-t",  "system")
	cmdOut, err := cmd.Output()

	if err != nil {
		return "", goof.WithError("error calling dmidecode", err)
	}

	rp := regexp.MustCompile("UUID:(.*)")
	uuid := strings.Replace(rp.FindString(string(cmdOut)), "UUID: ", "", -1)

	return strings.ToLower(uuid), nil
}

func (d *driver) NextDevice(
	ctx types.Context,
	opts types.Store) (string, error) {
	return "", types.ErrNotImplemented
}

func (d *driver) LocalDevices(
	ctx types.Context,
	opts *types.LocalDevicesOpts) (*types.LocalDevices, error) {
	devicesMap := make(map[string]string)

	file := "/proc/partitions"
	contentBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil,
			goof.WithFieldsE(
				goof.Fields{"file": file}, "error reading file", err)
	}

	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			devicePath := "/dev/" + fields[3]
			devicesMap[devicePath] = ""
		}
	}

	return &types.LocalDevices{
		Driver:    openstack.Name,
		DeviceMap: devicesMap,
	}, nil
}
