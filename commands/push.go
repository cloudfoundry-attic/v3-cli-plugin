package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/gofileutils/fileutils"
	. "github.com/cloudfoundry/v3-cli-plugin/models"
	. "github.com/cloudfoundry/v3-cli-plugin/util"
	"github.com/simonleung8/flags"
)

func Push(cliConnection plugin.CliConnection, args []string) {
	appDir := "."
	buildpack := "null"
	dockerImage := ""

	fc := flags.New()
	fc.NewStringFlag("filepath", "p", "path to app dir or zip to upload")
	fc.NewStringFlag("buildpack", "b", "the buildpack to use")
	fc.NewStringFlag("docker-image", "di", "the docker image to use")
	fc.Parse(args...)
	if fc.IsSet("p") {
		appDir = fc.String("p")
	}
	if fc.IsSet("b") {
		buildpack = fmt.Sprintf(`"%s"`, fc.String("b"))
	}
	if fc.IsSet("di") {
		dockerImage = fmt.Sprintf(`"%s"`, fc.String("di"))
	}

	mySpace, _ := cliConnection.GetCurrentSpace()

	lifecycle := ""
	if dockerImage != "" {
		lifecycle = `"lifecycle": { "type": "docker", "data": {} }`
	} else {
		lifecycle = fmt.Sprintf(`"lifecycle": { "type": "buildpack", "data": { "buildpack": %s } }`, buildpack)
	}

	//create the app
	rawOutput, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/apps", "-X", "POST", "-d",
		fmt.Sprintf(`{"name":"%s", "relationships": { "space": {"guid":"%s"}}, %s}`, fc.Args()[1], mySpace.Guid, lifecycle))
	FreakOut(err)
	output := strings.Join(rawOutput, "")
	app := V3AppModel{}
	err = json.Unmarshal([]byte(output), &app)
	FreakOut(err)
	if app.Error_Code != "" {
		FreakOut(errors.New("Error creating v3 app: " + app.Error_Code))
	}
	time.Sleep(2 * time.Second) // wait for app to settle before kicking off the log streamer
	go Logs(cliConnection, args)
	time.Sleep(2 * time.Second) // b/c sharing the cliConnection makes things break

	//create package
	pack := V3PackageModel{}
	if dockerImage != "" {
		request := fmt.Sprintf(`{"type": "docker", "data": {"image": %s}}`, dockerImage)
		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "-X", "POST", "-d", request)
		FreakOut(err)
		output = strings.Join(rawOutput, "")

		err = json.Unmarshal([]byte(output), &pack)
		if err != nil {
			FreakOut(errors.New("Error creating v3 app package: " + app.Error_Code))
		}
	} else {
		//create the empty package to upload the app bits to
		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "-X", "POST", "-d", "{\"type\": \"bits\"}")
		FreakOut(err)
		output = strings.Join(rawOutput, "")

		err = json.Unmarshal([]byte(output), &pack)
		if err != nil {
			FreakOut(errors.New("Error creating v3 app package: " + app.Error_Code))
		}

		token, err := cliConnection.AccessToken()
		FreakOut(err)
		api, apiErr := cliConnection.ApiEndpoint()
		FreakOut(apiErr)
		apiString := fmt.Sprintf("%s", api)
		if strings.Index(apiString, "s") == 4 {
			apiString = apiString[:4] + apiString[5:]
		}

		//gather files
		zipper := appfiles.ApplicationZipper{}
		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			zipper.Zip(appDir, zipFile)
			_, upload := exec.Command("curl", fmt.Sprintf("%s/v3/packages/%s/upload", apiString, pack.Guid), "-F", fmt.Sprintf("bits=@%s", zipFile.Name()), "-H", fmt.Sprintf("Authorization: %s", token), "-H", "Expect:").Output()
			FreakOut(upload)
		})

		//waiting for cc to pour bits into blobstore
		Poll(cliConnection, fmt.Sprintf("/v3/packages/%s", pack.Guid), "READY", 5*time.Minute, "Package failed to upload")
	}

	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/packages/%s/droplets", pack.Guid), "-X", "POST", "-d", "{}")
	FreakOut(err)
	output = strings.Join(rawOutput, "")
	droplet := V3DropletModel{}
	err = json.Unmarshal([]byte(output), &droplet)
	if err != nil {
		FreakOut(errors.New("error marshaling the v3 droplet: " + err.Error()))
	}
	//wait for the droplet to be ready
	Poll(cliConnection, fmt.Sprintf("/v3/droplets/%s", droplet.Guid), "STAGED", 10*time.Minute, "Droplet failed to stage")

	//assign droplet to the app
	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/droplets/current", app.Guid), "-X", "PUT", "-d", fmt.Sprintf("{\"droplet_guid\":\"%s\"}", droplet.Guid))
	FreakOut(err)
	output = strings.Join(rawOutput, "")

	//pick the first available shared domain, get the guid
	space, _ := cliConnection.GetCurrentSpace()
	nextUrl := "/v2/shared_domains"
	allDomains := DomainsModel{}
	for nextUrl != "" {
		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", nextUrl)
		FreakOut(err)
		output = strings.Join(rawOutput, "")
		tmp := DomainsModel{}
		err = json.Unmarshal([]byte(output), &tmp)
		FreakOut(err)
		allDomains.Resources = append(allDomains.Resources, tmp.Resources...)

		if tmp.NextUrl != "" {
			nextUrl = tmp.NextUrl
		} else {
			nextUrl = ""
		}
	}
	domainGuid := allDomains.Resources[0].Metadata.Guid
	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "v2/routes", "-X", "POST", "-d", fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, fc.Args()[1], domainGuid, space.Guid))
	output = strings.Join(rawOutput, "")

	var routeGuid string
	if strings.Contains(output, "CF-RouteHostTaken") {
		rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v2/routes?q=host:%s;domain_guid:%s", fc.Args()[1], domainGuid))
		output = strings.Join(rawOutput, "")
		routes := RoutesModel{}
		err = json.Unmarshal([]byte(output), &routes)
		routeGuid = routes.Routes[0].Metadata.Guid
	} else {
		route := RouteModel{}
		err = json.Unmarshal([]byte(output), &route)
		if err != nil {
			FreakOut(errors.New("error unmarshaling the route: " + err.Error()))
		}
		routeGuid = route.Metadata.Guid
	}

	FreakOut(err)
	route := RouteModel{}
	err = json.Unmarshal([]byte(output), &route)
	if err != nil {
		FreakOut(errors.New("error unmarshaling the route: " + err.Error()))
	}

	//map the route to the app
	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/route_mappings", "-X", "POST", "-d", fmt.Sprintf(`{"relationships": { "route": { "guid": "%s" }, "app": { "guid": "%s" } }`, routeGuid, app.Guid))
	FreakOut(err)
	output = strings.Join(rawOutput, "")

	//start the app
	rawOutput, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/start", app.Guid), "-X", "PUT")
	FreakOut(err)
	output = strings.Join(rawOutput, "")

	fmt.Println("Done pushing! Checkout your processes using 'cf apps'")
}
