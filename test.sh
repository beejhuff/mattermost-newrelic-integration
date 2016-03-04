#!/bin/sh

set -u
set -e

token=$1

curl -i -X POST -d 'alert={ "created_at":"2012-10-25T18:49:53%2B00:00", "application_name":"Application name", "account_name":"Account name", "severity":"critical", "message":"Apdex score fell below critical level of 0.90", "short_description":"[application name] alert opened", "long_description":"Alert opened on [application name]: Apdex score fell below critical level of 0.90","alert_url":"https://rpm.newrelc.com/accounts/[account_id]/applications/[application_id]/incidents/[incident_id]" }' 127.0.0.1:8230/webhook/"$token"

curl -i -X POST -d 'deployment={ "created_at":"2012-10-25T18:49:53%2B00:00", "application_name":"Application name", "account_name":"Account name", "changelog":"Changelog for deployment", "description":"Information about deployment", "revision":"Revision number", "deployment_url":"https://rpm.newrelic.com/accounts/[account_id]/applications/[application_id]/deployments/[deployment_id]", "deployed_by":"Name of person deploying" }' 127.0.0.1:8230/webhook/"$token"
