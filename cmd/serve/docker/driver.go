package docker

//Limitation: To use subpath we are limited to defining a new volume definition via alias

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/docker/go-plugins-helpers/volume"

	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/rc"
)

//Driver implement docker driver api
type Driver struct {
	root string
}

//NewDriver create a new docker driver
func NewDriver(root string) *Driver {
	return &Driver{
		root: root,
	}
}

//Create create and init the requested volume (add to rclone config file)
func (d *Driver) Create(r *volume.CreateRequest) error {
	if _, ok := r.Options["type"]; !ok {
		return errors.New("missing `type` option")
	}

	//Check local mountpoint
	mPath := filepath.Join(d.root, r.Name)
	_, err := os.Lstat(mPath) //Create folder if not exist. This will also failed if already exist
	if os.IsNotExist(err) {
		if err = os.MkdirAll(mPath, 0700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	isEmpty, err := folderIsEmpty(mPath)
	if err != nil {
		return err
	}
	if !isEmpty {
		return fmt.Errorf("%v already exist and is not empty", mPath)
	}

	params := rc.Params{}
	for id, val := range r.Options {
		if id == "type" { //Skip type
			continue
		}
		params[id] = val
	}

	return config.CreateRemote(context.Background(), r.Name, r.Options["type"], params, false, false)
}

//Remove remove the requested volume (remove from rclone config file)
func (d *Driver) Remove(r *volume.RemoveRequest) error {
	config.DeleteRemote(r.Name)
	return nil
}

//List volumes handled by the driver (configured volume in rclone file)
func (d *Driver) List() (*volume.ListResponse, error) {
	remotes := config.FileSections()
	sort.Strings(remotes)
	var volumes []*volume.Volume
	for _, vName := range remotes {
		volumes = append(volumes, &volume.Volume{Name: vName, Mountpoint: filepath.Join(d.root, vName)}) //TODO CreatedAt: v.CreatedAt
	}
	return &volume.ListResponse{Volumes: volumes}, nil
}

//Get get info on the requested volume (configured volume in rclone file)
func (d *Driver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	dump := config.DumpRcRemote(r.Name)
	if len(dump) == 0 {
		return nil, fmt.Errorf("volume %s not found", r.Name)
	}
	return &volume.GetResponse{Volume: &volume.Volume{Name: r.Name, Mountpoint: filepath.Join(d.root, r.Name), Status: dump}}, nil //TODO CreatedAt: v.CreatedAt
}

//Path get path of the requested volume
func (d *Driver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	return &volume.PathResponse{Mountpoint: filepath.Join(d.root, r.Name)}, nil
}

//Mount mount the requested volume
func (d *Driver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	//TODO
	return nil, nil
}

//Unmount unmount the requested volume
func (d *Driver) Unmount(r *volume.UnmountRequest) error {
	//TODO
	return nil
}

//Capabilities Send capabilities of the local driver
func (d *Driver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local", //We can only support `local` scope as `global` need a cluster controller logic.
		},
	}
}

//folderIsEmpty based on: http://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func folderIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
