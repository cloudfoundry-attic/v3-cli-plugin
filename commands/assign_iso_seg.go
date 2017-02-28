package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
)

func AssignIsolationSegment(cliConnection plugin.CliConnection, args []string) {
	orgName := args[1]
	isoSegName := args[2]

	fmt.Printf("Assigning isolation segment %s to org %s...\n", isoSegName, orgName)

	rawOutput, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v2/organizations")
	orgs := OrgsModel{}
	org := OrgModel{}
	output := strings.Join(rawOutput, "")
	json.Unmarshal([]byte(output), &orgs)

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
	json.Unmarshal([]byte(output), &isoSegs)

	if len(isoSegs.IsoSegs) == 0 {
		fmt.Printf("Isolation segment %s not found\n", isoSegName)
		return
	}
	isoSeg := isoSegs.IsoSegs[0]

	rawOutput, _ = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/isolation_segments/%s/relationships/organizations", isoSeg.Guid),
		"-X", "POST",
		"-d", fmt.Sprintf(`'{"data": [{"guid": "%s"}]}'`, org.Metadata.Guid),
	)
	//check response?
	fmt.Println("OK")
}
