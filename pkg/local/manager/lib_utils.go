package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kubernetes-sigs/alibaba-cloud-csi-driver/pkg/local/lib"
	"github.com/kubernetes-sigs/alibaba-cloud-csi-driver/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func getRegionNamespaceInfo(region string) (*PmemRegions, error) {
	listCmd := fmt.Sprintf("%s ndctl list -RN -r %s", NsenterCmd, region)

	out, err := utils.Run(listCmd)
	if err != nil {
		log.Errorf("List NameSpace for region %s error: %v", region, err)
		return nil, err
	}
	regions := &PmemRegions{}
	err = json.Unmarshal(([]byte)(out), regions)
	if len(regions.Regions) == 0 {
		log.Errorf("list Namespace for region %s get 0 region", region)
		return nil, errors.New("list Namespace get 0 region by " + region)
	}

	if len(regions.Regions[0].Namespaces) != 1 {
		log.Errorf("list Namespace for region %s get 0 or multi namespaces", region)
		return nil, errors.New("list Namespace for region get 0 or multi namespaces" + region)
	}
	return regions, nil
}

func checkProjQuotaNamespaceValid(region string) (devicePath string, namespaceName string, err error) {
	regions, err := getRegionNamespaceInfo(region)
	if err != nil {
		return "", "", err
	}
	namespace := regions.Regions[0].Namespaces[0]
	if namespace.Mode != "fsdax" {
		log.Errorf("projectQuota namespace mode %s wrong", namespace.Mode)
		return "", "", errors.New("projectQuota namespace wrong mode" + namespace.Mode)
	}
	return "/dev/" + namespace.BlockDev, namespace.Dev, nil
}

func checkKMEMCreated(chardev string) (bool, error) {
	listCmd := fmt.Sprintf("%s daxctl list", NsenterCmd)
	out, err := utils.Run(listCmd)
	if err != nil {
		log.Errorf("List daxctl error: %v", err)
		return false, err
	}
	memList := []*DaxctrlMem{}
	err = json.Unmarshal(([]byte)(out), &memList)
	if err != nil {
		return false, err
	}
	for _, mem := range memList {
		if mem.Chardev == chardev && mem.Mode == "system-ram" {
			return true, nil
		}
	}
	return false, nil
}

func checkKMEMNamespaceValid(region string) (string, error) {
	regions, err := getRegionNamespaceInfo(region)
	if err != nil {
		return "", err
	}
	namespaceMode := regions.Regions[0].Namespaces[0].Mode
	if namespaceMode != "devdax" {
		log.Errorf("KMEM namespace mode %s wrong", namespaceMode)
		return "", errors.New("KMEM namespace wrong mode" + namespaceMode)
	}
	return regions.Regions[0].Namespaces[0].CharDev, nil
}

func makeNamespaceMemory(chardev string) error {
	makeCmd := fmt.Sprintf("%s daxctl reconfigure-device -m system-ram %s", NsenterCmd, chardev)
	_, err := utils.Run(makeCmd)
	return err
}

func createNameSpace(region, pmemType string) error {
	var createCmd string
	if pmemType == "lvm" {
		createCmd = fmt.Sprintf("%s ndctl create-namespace -r %s", NsenterCmd, region)
	} else {
		createCmd = fmt.Sprintf("%s ndctl create-namespace -r %s --mode=devdax", NsenterCmd, region)
	}
	_, err := utils.Run(createCmd)
	if err != nil {
		log.Errorf("Create NameSpace for region %s error: %v", region, err)
		return err
	}
	log.Infof("Create NameSpace for region %s successful", region)
	return nil
}

// device used in pv
// device used in block
func checkNameSpaceUsed(devicePath string) bool {
	pvCheckCmd := fmt.Sprintf("%s pvs %s 2>&1 | grep -v \"Failed to \" | grep /dev | awk '{print $2}' | wc -l", NsenterCmd, devicePath)
	out, err := utils.Run(pvCheckCmd)
	if err == nil && strings.TrimSpace(out) != "0" {
		log.Infof("NameSpace %s used for pv", devicePath)
		return true
	}

	out, err = checkFSType(devicePath)
	if err == nil && strings.TrimSpace(out) != "" {
		log.Infof("NameSpace %s format as %s", devicePath, out)
		return true
	}
	return false
}

func createPmemVG(deviceList []string, vgName string) error {
	localDeviceStr := strings.Join(deviceList, " ")
	vgAddCmd := fmt.Sprintf("%s vgcreate --force %s %s", NsenterCmd, vgName, localDeviceStr)
	_, err := utils.Run(vgAddCmd)
	if err != nil {
		log.Errorf("Create VG (%v) with PV (%v) error: %s", vgName, localDeviceStr, err.Error())
		return err
	}

	log.Infof("Successful add Local Disks to VG (%s): %s", vgName, localDeviceStr)
	return nil
}

// GetRegions get regions info
func GetRegions() (*PmemRegions, error) {
	regions := &PmemRegions{}
	getRegionCmd := fmt.Sprintf("%s ndctl list -RN", NsenterCmd)
	regionOut, err := utils.Run(getRegionCmd)
	if err != nil {
		return regions, err
	}
	err = json.Unmarshal(([]byte)(regionOut), regions)
	if err != nil {
		if strings.HasPrefix(regionOut, "[") {
			regionList := []PmemRegion{}
			err = json.Unmarshal(([]byte)(regionOut), &regionList)
			if err != nil {
				return regions, err
			}
			regions.Regions = regionList
		} else {
			return regions, err
		}
	}

	return regions, nil
}

// GetNameSpaceCapacity get namespace size
func GetNameSpaceCapacity(ns *PmemNameSpace) int64 {
	expect := (ns.Size + ns.Align) * 4096 / 4032
	return expect
}

// GetNameSpace get namespace info
func GetNameSpace(namespaceName string) (*PmemNameSpace, error) {
	namespace := &PmemNameSpace{}
	namespaceList := []*PmemNameSpace{}
	getRegionCmd := fmt.Sprintf("%s ndctl list -n %s", NsenterCmd, namespaceName)
	regionOut, err := utils.Run(getRegionCmd)
	if err != nil {
		return namespace, err
	}
	err = json.Unmarshal(([]byte)(regionOut), &namespaceList)
	if err != nil {
		return namespace, err
	}
	if len(namespaceList) == 1 {
		return namespaceList[0], nil
	}
	return namespace, fmt.Errorf("namespace found error")
}

// ToProto build NameSpace object
func (pns *PmemNameSpace) ToProto() *lib.NameSpace {
	new := &lib.NameSpace{}
	new.CharDev = pns.CharDev
	new.Name = pns.Name
	new.Dev = pns.Dev
	new.Mode = pns.Mode
	new.Size = pns.Size
	new.Uuid = pns.UUID
	new.Align = pns.Align
	new.MapType = pns.MapType
	new.SectorSize = pns.SectorSize
	return new
}

func checkFSType(devicePath string) (string, error) {
	// We use `file -bsL` to determine whether any filesystem type is detected.
	// If a filesystem is detected (ie., the output is not "data", we use
	// `blkid` to determine what the filesystem is. We use `blkid` as `file`
	// has inconvenient output.
	// We do *not* use `lsblk` as that requires udev to be up-to-date which
	// is often not the case when a device is erased using `dd`.
	output, err := exec.Command("file", "-bsL", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "data" {
		return "", nil
	}
	output, err = exec.Command("blkid", "-c", "/dev/null", "-o", "export", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "=")
		if len(fields) != 2 {
			return "", ErrParse
		}
		if fields[0] == "TYPE" {
			return fields[1], nil
		}
	}
	return "", ErrParse
}

// EnsureFolder ...
func EnsureFolder(target string) error {
	mdkirCmd := "mkdir"
	_, err := exec.LookPath(mdkirCmd)
	if err != nil {
		if err == exec.ErrNotFound {
			return fmt.Errorf("%q executable not found in $PATH", mdkirCmd)
		}
		return err
	}

	mkdirFullPath := fmt.Sprintf("%s mkdir -p %s", NsenterCmd, target)
	_, err = utils.Run(mkdirFullPath)
	if err != nil {
		log.Errorf("Create path error: %v", err)
		return err
	}
	return nil
}
