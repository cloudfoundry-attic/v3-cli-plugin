package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	"github.com/cloudfoundry/v3-cli-plugin/util"
)

func UnassignIsolationSegment(cliConnection plugin.CliConnection, args []string) {
	orgName := args[1]
	isoSegName := args[2]

	fmt.Printf("Unassigning isolation segment %s from org %s...\n", isoSegName, orgName)

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v2/organizations")
	orgs := OrgsModel{}
	org := OrgModel{}
	output := strings.Join(rawOutput, "")
	err := json.Unmarshal([]byte(output), &orgs)
	util.FreakOut(err)

	for _, v := range orgs.Orgs {
		if v.Entity.Name == orgName {
			org = v
			break
		}
	}

	if org.Entity.Name != orgName {
		fmt.Println("Unable to find org ", orgName)
		return
	}

	urlValues := url.Values{}
	urlValues.Add("names", isoSegName)

	rawOutput, _ = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/isolation_segments?%s", urlValues.Encode()))
	isoSegs := V3IsolationSegmentsModel{}
	output = strings.Join(rawOutput, "")
	err = json.Unmarshal([]byte(output), &isoSegs)
	util.FreakOut(err)

	if len(isoSegs.IsoSegs) == 0 {
		fmt.Printf("Isolation segment %s not found\n", isoSegName)
		return
	}
	isoSeg := isoSegs.IsoSegs[0]

	rawOutput, _ = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations/%s", isoSeg.Guid, org.Metadata.Guid),
		"-X", "DELETE",
	)
	output = strings.Join(rawOutput, "")
	if len(output) > 0 {
		fmt.Println("Error unassigning iso seg: ", output)
		return
	}
	fmt.Println("OK")
}
