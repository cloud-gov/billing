#!/usr/bin/env bash
set -exo pipefail

# The CloudFoundry terraform provider has one issue affecting this repo: CloudFoundry applications that are updated do not always fully restage. This script restages the applications manually.
# Issue for app restaging: https://github.com/cloudfoundry/terraform-provider-cloudfoundry/issues/127

# Load values from state file
ORG=$(jq  -r '.outputs.org_name.value' terraform-state/terraform.tfstate)
SPACE=$(jq  -r '.outputs.space_name.value' terraform-state/terraform.tfstate)
APP=$(jq  -r '.outputs.app_name.value' terraform-state/terraform.tfstate)

cf api $CF_API_URL
(set +x; cf auth $CF_CLIENT_ID $CF_CLIENT_SECRET --client-credentials)

cf target -o $ORG -s $SPACE
cf restage $APP --strategy rolling
