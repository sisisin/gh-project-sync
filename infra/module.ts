import * as pulumi from '@pulumi/pulumi';
import * as gcp from '@pulumi/gcp';
const gcpConfig = new pulumi.Config('gcp');
const project = gcpConfig.require('project');
const region = gcpConfig.require('region');
const projectNumber = new pulumi.Config().requireNumber('projectNumber');

export function apply() {
  const saCloudRunGhpsync = new gcp.serviceaccount.Account('sa-cloud-run-ghpsync', {
    accountId: 'sa-cloud-run-ghpsync',
    displayName: 'Service Account for Cloud Run ghpsync',
  });
  const cloudRunGhpsyncRoles = ['roles/secretmanager.secretAccessor'];
  applyIAMMember(saCloudRunGhpsync, cloudRunGhpsyncRoles);

  const saCloudSchedulerForKickGhpsync = new gcp.serviceaccount.Account('sa-scheduler-ghpsync', {
    accountId: 'sa-scheduler-ghpsync',
    displayName: 'Service Account for Cloud Scheduler to kick ghpsync',
  });
  const cloudSchedulerRoles = ['roles/run.invoker'];
  applyIAMMember(saCloudSchedulerForKickGhpsync, cloudSchedulerRoles);

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

  const githubProjectSyncBucket = new gcp.storage.Bucket('github-project-sync', {
    location: 'us-west1',
    name: 'github-project-sync',
  });
  applyBucketIAMMember(githubProjectSyncBucket, saCloudRunGhpsync, ['roles/storage.objectUser']);

  const dataset = new gcp.bigquery.Dataset('github_project_sync', {
    datasetId: 'github_project_sync',
    friendlyName: 'Dataset for GitHub Project Sync',
    location: 'US',
  });

  const projectItemsTable = new gcp.bigquery.Table('project_items', {
    datasetId: dataset.datasetId,
    tableId: 'project_items',
    deletionProtection: false,
    timePartitioning: {
      type: 'HOUR',
      // 7 days
      expirationMs: 60 * 60 * 24 * 7 * 1000,
    },

    description:
      'Table for GitHub Project Items. GitHub API v4を非正規化したレコードを入れます。itemsの最新の情報は最新のパーティションを参照して取得できます',

    schema: JSON.stringify([
      { name: 'organization_id', type: 'STRING', mode: 'REQUIRED', description: 'GitHub API - Organization.id value' },
      { name: 'project_number', type: 'INT64', mode: 'REQUIRED', description: 'GitHub API - ProjectV2.number value' },
      { name: 'project_title', type: 'STRING', description: 'GitHub API - ProjectV2.name value' },
      { name: 'item', type: 'JSON', description: 'GitHub API ProjectV2Item.items.nodes value' },
      { name: 'created_at', type: 'TIMESTAMP', description: 'GitHub API - ProjectV2Item.createdAt value' },
      { name: 'updated_at', type: 'TIMESTAMP', description: 'GitHub API - ProjectV2Item.updatedAt value' },
      { name: 'creator', type: 'STRING', description: 'GitHub API - ProjectV2Item.creator.login value' },
    ]),
  });
  const ghSyncBigQueryTableUserRole = new gcp.projects.IAMCustomRole('gh-sync-bigquery-user', {
    roleId: 'gh_sync_bigquery_user',
    title: 'gh project sync bigquery user',
    permissions: [
      'bigquery.tables.create',
      'bigquery.tables.updateData',
      'bigquery.tables.update',
      'bigquery.tables.list',
      'bigquery.tables.getData',
    ],
  });
  const ghSyncBigQueryJobUserRole = new gcp.projects.IAMCustomRole('gh-sync-bigquery-job-user', {
    roleId: 'gh_sync_bigquery_job_user',
    title: 'gh project sync bigquery job user',
    permissions: ['bigquery.jobs.create'],
  });

  pulumi.all([ghSyncBigQueryTableUserRole.roleId]).apply((roleIds) => {
    applyBigQueryDatasetIAMMember(
      dataset,
      saCloudRunGhpsync,
      roleIds.map((roleId) => `projects/${project}/roles/${roleId}`),
    );
  });

  pulumi.all([ghSyncBigQueryJobUserRole.roleId]).apply((roleIds) => {
    applyIAMMember(
      saCloudRunGhpsync,
      roleIds.map((roleId) => `projects/${project}/roles/${roleId}`),
    );
  });

  return {
    outputs: {
      saEmail: saCloudRunGhpsync.email,
      bucketName: githubProjectSyncBucket.name,
    },
  };
}

function applyIAMMember(sa: gcp.serviceaccount.Account, roles: string[]) {
  roles.forEach((role) => {
    sa.accountId.apply((accountId) => {
      new gcp.projects.IAMMember(`${accountId}-${role}`, {
        project,
        role,
        member: pulumi.interpolate`serviceAccount:${sa.email}`,
      });
    });
  });
}

function applyBigQueryDatasetIAMMember(dataset: gcp.bigquery.Dataset, sa: gcp.serviceaccount.Account, roles: string[]) {
  roles.forEach((role) => {
    dataset.datasetId.apply((datasetId) => {
      new gcp.bigquery.DatasetIamMember(`${datasetId}-${role}`, {
        datasetId: datasetId,
        role,
        member: pulumi.interpolate`serviceAccount:${sa.email}`,
      });
    });
  });
}

function applyBucketIAMMember(bucket: gcp.storage.Bucket, sa: gcp.serviceaccount.Account, roles: string[]) {
  roles.forEach((role) => {
    bucket.name.apply((bucketName) => {
      new gcp.storage.BucketIAMMember(`${bucketName}-${role}`, {
        bucket: bucketName,
        role,
        member: pulumi.interpolate`serviceAccount:${sa.email}`,
      });
    });
  });
}
