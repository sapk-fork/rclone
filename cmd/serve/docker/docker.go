// Package docker serves a remote suitable for use with docker volume api
//+build linux

package docker

import (
	"path/filepath"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/spf13/cobra"

	"github.com/rclone/rclone/cmd"
	"github.com/rclone/rclone/fs/config/flags"
)

var (
	//PluginAlias is the name of the docker plugin (can be override for test)
	PluginAlias   = "rclone"
	baseDirectory string
	gid           int
)

func init() {
	flagSet := Command.Flags()
	flags.IntVarP(flagSet, &gid, "gid", "g", 0, "GID to use for mountpoint")
	flags.StringVarP(flagSet, &baseDirectory, "basedir", "b", filepath.Join(volume.DefaultDockerRootDirectory, PluginAlias), "base directory for volume")
}

// Command definition for cobra
var Command = &cobra.Command{
	Use:   "docker",
	Short: `Serve any remote on docker's volume plugin API.`,
	Long: `rclone serve docker implements docker's volume plugin API.
This allows docker to use rclone as a data storage mechanism for various cloud providers.`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(0, 0, command, args)
		cmd.Run(false, false, command, func() error { //TODO cmd.ShowStats() ?
			d := NewDriver(baseDirectory)
			h := volume.NewHandler(d)
			return h.ServeUnix(PluginAlias, gid)
		})
	},
}
