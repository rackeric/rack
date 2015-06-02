package servercommands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/fatih/structs"
	"github.com/jrperritt/rackcli/auth"
	"github.com/jrperritt/rackcli/output"
	"github.com/jrperritt/rackcli/util"
	"github.com/olekukonko/tablewriter"
	osServers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

var create = cli.Command{
	Name:        "create",
	Usage:       fmt.Sprintf("%s %s create [flags]", util.Name, commandPrefix),
	Description: "Creates a new server",
	Action:      commandCreate,
	Flags:       flagsCreate(),
}

func flagsCreate() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "imageRef",
			Usage: "[optional; required if imageName and bootFromVolume flags are not provided] The image ID from which to create the server.",
		},
		cli.StringFlag{
			Name:  "imageName",
			Usage: "[optional; required if imageRef and bootFromVolume flags are not provided] The name of the image from which to create the server.",
		},
		cli.StringFlag{
			Name:  "flavorRef",
			Usage: "[optional; required if flavorName is not provided] The flavor ID that the server should have.",
		},
		cli.StringFlag{
			Name:  "flavorName",
			Usage: "[optional; required if flavorRef is not provided] The name of the flavor that the server should have.",
		},
		cli.StringFlag{
			Name:  "securityGroups",
			Usage: "[optional] A comma-separated string of names of the security groups to which this server should belong.",
		},
		cli.StringFlag{
			Name:  "userData",
			Usage: "[optional] Configuration information or scripts to use after the server boots.",
		},
		cli.StringFlag{
			Name:  "networks",
			Usage: "[optional] A comma-separated string of IDs of the networks to attach to this server. If not provided, a public and private network will be attached.",
		},
		cli.StringFlag{
			Name:  "metadata",
			Usage: "[optional] A comma-separated string a key=value pairs.",
		},
		cli.StringFlag{
			Name:  "adminPass",
			Usage: "[optional] The root password for the server. If not provided, one will be randomly generated and returned in the output.",
		},
		cli.StringFlag{
			Name:  "keypair",
			Usage: "[optional] the name of the SSH KeyPair to be injected into this server.",
		},
	}
}

func commandCreate(c *cli.Context) {
	util.CheckArgNum(c, 1)
	serverName := c.Args()[0]
	opts := &servers.CreateOpts{
		Name:           serverName,
		ImageRef:       c.String("imageRef"),
		ImageName:      c.String("imageName"),
		FlavorRef:      c.String("flavorRef"),
		FlavorName:     c.String("flavorName"),
		SecurityGroups: strings.Split(c.String("securityGroups"), ","),
		AdminPass:      c.String("adminPass"),
		KeyPair:        c.String("keypair"),
	}

	if c.IsSet("userData") {
		s := c.String("userData")
		userData, err := ioutil.ReadFile(s)
		if err != nil {
			opts.UserData = userData
		} else {
			opts.UserData = []byte(s)
		}
	}

	if c.IsSet("networks") {
		netIDs := strings.Split(c.String("networks"), ",")
		networks := make([]osServers.Network, len(netIDs))
		for i, netID := range netIDs {
			networks[i] = osServers.Network{
				UUID: netID,
			}
		}
		opts.Networks = networks
	}

	if c.IsSet("metadata") {
		metadata := make(map[string]string)
		metaStrings := strings.Split(c.String("metadata"), ",")
		for _, metaString := range metaStrings {
			temp := strings.Split(metaString, "=")
			if len(temp) != 2 {
				fmt.Printf("Error parsing metadata: Expected key=value format but got %s\n", metaString)
				os.Exit(1)
			}
			metadata[temp[0]] = temp[1]
		}
		opts.Metadata = metadata
	}

	client := auth.NewClient("compute")
	o, err := servers.Create(client, opts).Extract()
	if err != nil {
		fmt.Printf("Error creating server: %s\n", err)
		os.Exit(1)
	}
	output.Print(c, o, tableCreate)
}

func tableCreate(c *cli.Context, i interface{}) {
	m := structs.Map(i)
	t := tablewriter.NewWriter(c.App.Writer)
	t.SetHeader([]string{"property", "value"})
	for k, v := range m {
		t.Append([]string{k, fmt.Sprint(v)})
	}
	t.Render()
}