package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvflasher/internal/bmap"
)

var outputBmap string
var blockSize int

var createCmd = &cobra.Command{
	Use:   "create [image]",
	Short: "Create a bmap file from an image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imagePath := args[0]
		
		if outputBmap == "" {
			outputBmap = imagePath + ".bmap"
		}

		fmt.Printf("Creating bmap for %s...\n", imagePath)
		
		bm, err := bmap.Create(imagePath, bmap.CreateOptions{
			BlockSize: blockSize,
		})
		if err != nil {
			return err
		}

		if err := bm.Save(outputBmap); err != nil {
			return err
		}

		fmt.Printf("Bmap file created: %s\n", outputBmap)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&outputBmap, "output", "o", "", "output bmap file (default image.bmap)")
	createCmd.Flags().IntVarP(&blockSize, "block-size", "b", 4096, "block size in bytes")
	rootCmd.AddCommand(createCmd)
}
