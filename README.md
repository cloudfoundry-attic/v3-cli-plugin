# V3_beta

This is a Cloud Foundry CLI plugin for v3 of the CF Cloud Controller API. Both it and the V3 api are currently under active development, so stability isn't guaranteed. Please use caution when using this plugin and the V3 api in general.

#Commands

####v3-push
#####Syntax: cf v3-push APPNAME /path/to/app.zip
Pushes, maps a route, and starts the zipped app as a V3 app and associated V3 processes. Currently tries to push to the default domain of th
