import * as pulumi from '@pulumi/pulumi';
import * as gcp from '@pulumi/gcp';

const gcpConfig = new pulumi.Config('gcp');
const project = gcpConfig.require('project');
const region = gcpConfig.require('region');
const projectNumber = new pulumi.Config().requireNumber('projectNumber');

const saCloudRunGhpsync = new gcp.serviceaccount.Account('sa-cloud-run-ghpsync', {
  accountId: 'sa-cloud-run-ghpsync',
  displayName: 'Service Account for Cloud Run ghpsync',
});
const cloudRunGhpsyncRoles = ['roles/secretmanager.secretAccessor'];
applyIAMMember('sa-cloud-run-ghpsync', saCloudRunGhpsync, cloudRunGhpsyncRoles);

const saCloudSchedulerForKickGhpsync = new gcp.serviceaccount.Account('sa-scheduler-ghpsync', {
  accountId: 'sa-scheduler-ghpsync',
  displayName: 'Service Account for Cloud Scheduler to kick ghpsync',
});
const cloudSchedulerRoles = ['roles/run.invoker'];
applyIAMMember('sa-scheduler-ghpsync', saCloudSchedulerForKickGhpsync, cloudSchedulerRoles);

const runName = 'github-project-sync';
new gcp.cloudscheduler.Job('kick-ghpsync', {
  // every hours at minute 5
  schedule: '5 * * * *',
  timeZone: 'Asia/Tokyo',
  httpTarget: {
    httpMethod: 'POST',
    uri: `https://${region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${projectNumber}/jobs/${runName}:run`,
    headers: { 'Content-Type': 'application/json' },
    oauthToken: {
      serviceAccountEmail: saCloudSchedulerForKickGhpsync.email,
    },
  },
});

function applyIAMMember(key: string, sa: gcp.serviceaccount.Account, roles: string[]) {
  roles.forEach((role) => {
    new gcp.projects.IAMMember(`${key}-${role}`, {
      project,
      role,
      member: pulumi.interpolate`serviceAccount:${sa.email}`,
    });
  });
}
