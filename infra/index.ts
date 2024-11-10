import { apply } from './module';

const { outputs } = apply();

export const saEmailForCloudRunGhpsync = outputs.saEmail;
export const bucketNameForGithubProjectSync = outputs.bucketName;
export const registryId = outputs.registry;
