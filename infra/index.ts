import { apply } from './module';
import * as pulumi from '@pulumi/pulumi';
const gcpConfig = new pulumi.Config('gcp');

const { outputs } = apply();

export const saEmailForCloudRunGhpsync = outputs.saEmail;
export const bucketNameForGithubProjectSync = outputs.bucketName;
export const imageName = outputs.imageName;
export const registryDomain = outputs.registryDomain;
export const region = gcpConfig.require('region');
export const runName = outputs.runName;
