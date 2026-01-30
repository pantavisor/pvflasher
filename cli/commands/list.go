package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"pvflasher/internal/device"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := device.NewManager()
		devs, err := mgr.List()
		if err != nil {
			return err
		}

		if len(devs) == 0 {
			fmt.Println("No devices found.")
			return nil
		}

		fmt.Println("Available devices:")
		for _, d := range devs {
			removable := ""
			if d.Removable {
				removable = "(Removable)"
			}
			mounted := ""
			if len(d.MountPoints) > 0 {
				mounted = fmt.Sprintf("[Mounted: %s]", strings.Join(d.MountPoints, ", "))
			}
			fmt.Printf("- %s: %s %s %s %s [%d bytes]\n", d.Name, d.Vendor, d.Model, removable, mounted, d.Size)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
