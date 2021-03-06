package cli

import (
	"encoding/json"
	"fmt"

	humanize "github.com/dustin/go-humanize"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/spf13/cobra"
)

func newImageDescribeCommand(cli *CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "describe [FLAGS] IMAGE",
		Short:                 "Describe an image",
		Args:                  cobra.ExactArgs(1),
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
		PreRunE:               cli.ensureToken,
		RunE:                  cli.wrap(runImageDescribe),
	}
	addOutputFlag(cmd, outputOptionJSON(), outputOptionFormat())
	return cmd
}

func runImageDescribe(cli *CLI, cmd *cobra.Command, args []string) error {
	outputFlags := outputFlagsForCommand(cmd)

	idOrName := args[0]
	image, resp, err := cli.Client().Image.Get(cli.Context, idOrName)
	if err != nil {
		return err
	}
	if image == nil {
		return fmt.Errorf("image not found: %s", idOrName)
	}

	switch {
	case outputFlags.IsSet("json"):
		return imageDescribeJSON(resp)
	case outputFlags.IsSet("format"):
		return describeFormat(image, outputFlags["format"][0])
	default:
		return imageDescribeText(cli, image)
	}
}

func imageDescribeText(cli *CLI, image *hcloud.Image) error {
	fmt.Printf("ID:\t\t%d\n", image.ID)
	fmt.Printf("Type:\t\t%s\n", image.Type)
	fmt.Printf("Status:\t\t%s\n", image.Status)
	fmt.Printf("Name:\t\t%s\n", na(image.Name))
	fmt.Printf("Description:\t%s\n", image.Description)
	if image.ImageSize != 0 {
		fmt.Printf("Image size:\t%.1f GB\n", image.ImageSize)
	} else {
		fmt.Printf("Image size:\t%s\n", na(""))
	}
	fmt.Printf("Disk size:\t%.0f GB\n", image.DiskSize)
	fmt.Printf("Created:\t%s (%s)\n", datetime(image.Created), humanize.Time(image.Created))
	fmt.Printf("OS flavor:\t%s\n", image.OSFlavor)
	fmt.Printf("OS version:\t%s\n", na(image.OSVersion))
	fmt.Printf("Rapid deploy:\t%s\n", yesno(image.RapidDeploy))
	fmt.Printf("Protection:\n")
	fmt.Printf("  Delete:\t%s\n", yesno(image.Protection.Delete))

	fmt.Print("Labels:\n")
	if len(image.Labels) == 0 {
		fmt.Print("  No labels\n")
	} else {
		for key, value := range image.Labels {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

func imageDescribeJSON(resp *hcloud.Response) error {
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	if image, ok := data["image"]; ok {
		return describeJSON(image)
	}
	if images, ok := data["images"].([]interface{}); ok {
		return describeJSON(images[0])
	}
	return describeJSON(data)
}
