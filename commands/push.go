package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/gofileutils/fileutils"
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
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "/v3/apps", "-X", "POST", "-d",
		fmt.Sprintf(`{"name":"%s", "relationships": { "space": {"guid":"%s"}}, %s}`, fc.Args()[1], mySpace.Guid, lifecycle))
	FreakOut(err)
	app := V3AppModel{}
	err = json.Unmarshal([]byte(output[0]), &app)
	FreakOut(err)
	if app.Error_Code != "" {
		FreakOut(errors.New("Error creating v3 app: " + app.Error_Code))
	}

	//create package
	pack := V3PackageModel{}
	if dockerImage != "" {
		request := fmt.Sprintf(`{"type": "docker", "data": {"image": %s}}`, dockerImage)
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "-X", "POST", "-d", request)
		FreakOut(err)

		err = json.Unmarshal([]byte(output[0]), &pack)
		if err != nil {
			FreakOut(errors.New("Error creating v3 app package: " + app.Error_Code))
		}
	} else {
		//create the empty package to upload the app bits to
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "-X", "POST", "-d", "{\"type\": \"bits\"}")
		FreakOut(err)

		err = json.Unmarshal([]byte(output[0]), &pack)
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
		zipper := app_files.ApplicationZipper{}
		fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
			zipper.Zip(appDir, zipFile)
			_, upload := exec.Command("curl", fmt.Sprintf("%s/v3/packages/%s/upload", apiString, pack.Guid), "-F", fmt.Sprintf("bits=@%s", zipFile.Name()), "-H", fmt.Sprintf("Authorization: %s", token)).Output()
			FreakOut(upload)
		})

		//waiting for cc to pour bits into blobstore
		Poll(cliConnection, fmt.Sprintf("/v3/packages/%s", pack.Guid), "READY", 1*time.Minute, "Package failed to upload")
	}

	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/packages/%s/droplets", pack.Guid), "-X", "POST", "-d", "{}")
	FreakOut(err)
	droplet := V3DropletModel{}
	err = json.Unmarshal([]byte(output[0]), &droplet)
	if err != nil {
		FreakOut(errors.New("error marshaling the v3 droplet: " + err.Error()))
	}
	//wait for the droplet to be ready
	Poll(cliConnection, fmt.Sprintf("/v3/droplets/%s", droplet.Guid), "STAGED", 1*time.Minute, "Droplet failed to stage")

	//assign droplet to the app
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/current_droplet", app.Guid), "-X", "PUT", "-d", fmt.Sprintf("{\"droplet_guid\":\"%s\"}", droplet.Guid))
	FreakOut(err)

	//pick the first available shared domain, get the guid
	space, _ := cliConnection.GetCurrentSpace()
	nextUrl := "/v2/shared_domains"
	allDomains := DomainsModel{}
	for nextUrl != "" {
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", nextUrl)
		FreakOut(err)
		tmp := DomainsModel{}
		err = json.Unmarshal([]byte(output[0]), &tmp)
		FreakOut(err)
		allDomains.Resources = append(allDomains.Resources, tmp.Resources...)

		if tmp.NextUrl != "" {
			nextUrl = tmp.NextUrl
		} else {
			nextUrl = ""
		}
	}
	domainGuid := allDomains.Resources[0].Metadata.Guid
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", "v2/routes", "-X", "POST", "-d", fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, fc.Args()[1], domainGuid, space.Guid))

	var routeGuid string
	if strings.Contains(output[0], "CF-RouteHostTaken") {
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v2/routes?q=host:%s;domain_guid:%s", fc.Args()[1], domainGuid))
		routes := RoutesModel{}
		err = json.Unmarshal([]byte(output[0]), &routes)
		routeGuid = routes.Routes[0].Metadata.Guid
	} else {
		route := RouteModel{}
		err = json.Unmarshal([]byte(output[0]), &route)
		if err != nil {
			FreakOut(errors.New("error unmarshaling the route: " + err.Error()))
		}
		routeGuid = route.Metadata.Guid
	}

	FreakOut(err)
	route := RouteModel{}
	err = json.Unmarshal([]byte(output[0]), &route)
	if err != nil {
		FreakOut(errors.New("error unmarshaling the route: " + err.Error()))
	}

	//map the route to the app
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/route_mappings", app.Guid), "-X", "POST", "-d", fmt.Sprintf(`{"relationships": { "route": { "guid": "%s" }	} }`, routeGuid))
	FreakOut(err)

	//start the app
	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/start", app.Guid), "-X", "PUT")
	FreakOut(err)

	fmt.Println("Done pushing! Checkout your processes using 'cf apps'")
}
