# V3_beta

This is a Cloud Foundry CLI plugin for v3 of the CF Cloud Controller API. Both it and the V3 api are currently under active development, so stability isn't guaranteed. Please use caution when using this plugin and the V3 api in general.

#Commands

####v3-push
#####Syntax: cf v3-push APPNAME /path/to/app.zip
Pushes, maps a route, and starts the zipped app as a V3 app and associated V3 processes. Currently tries to push to the default domain of your currently targeted Cloud Foundry.

#####Syntax: cf v3-push APPNAME --docker-image "path/to/docker/image"
Pushes, maps a route, and starts the provided docker image as a V3 app and associated V3 processes. Currently tries to push to the default domain of your currently targeted Cloud Foundry.

####v3-delete
#####Syntax: cf v3-delete APPNAME
Deletes the specified V3 app

####v3-apps
#####Syntax: cf v3-apps
Shows v3 apps running in the currently targeted space.

####v3-processes
#####Syntax: cf v3-processes
Shows all v3 processes and v2 apps running in the currently targeted space.

####v3-logs
#####Syntax: cf v3-logs APPNAME
Tails logs for all processes of the specified V3 app.

####v3-tasks
#####Syntax: cf v3-tasks APPNAME
Shows tasks for the specified V3 app.

####v3-run-task
#####Syntax: cf v3-run-task APPNAME TASKNAME COMMAND
Run a task for an app.

####v3-bind-service
#####Syntax: cf v3-bind-service APPNAME SERVICEINSTANCE [-c PARAMETERS_AS_JSON]
Binds a service instance to a v3 app.

#Contributing
Members of the CF CAPI team should have access. If you do end up making changes, please follow these steps:

1. Make your changes
2. Bump the version number in v3.go
3. Compile a new binary via `go build v3.go`, move the binary to bin/OSX_binary
4. Commit your changes

##Running tests

1. Install [Ginkgo](https://github.com/onsi/ginkgo)
1. Run `ginkgo -r` in the root directory
