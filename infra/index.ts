import * as pulumi from "@pulumi/pulumi";
import * as containerregistry from "@pulumi/azure-native/containerregistry";
import * as resources from "@pulumi/azure-native/resources";
import * as web from "@pulumi/azure-native/web";

import * as docker from "@pulumi/docker";
import { UnauthenticatedClientActionV2 } from "@pulumi/azure-native/web/v20200601";

const productName = "request-a-dasher";
const stackName = pulumi.getStack();

const resourceGroup = new resources.ResourceGroup(productName);
const cfg = new pulumi.Config();

// Build and publish container
const registry = new containerregistry.Registry("registry", {
    resourceGroupName: resourceGroup.name,
    sku: {
        name: "Basic",
    },
    adminUserEnabled: true,
});

const credentials = containerregistry.listRegistryCredentialsOutput({
    resourceGroupName: resourceGroup.name,
    registryName: registry.name,
});

const adminUsername = credentials.apply(credentials => credentials.username!);
const adminPassword = credentials.apply(credentials => credentials.passwords![0].value!);

const image = new docker.Image(productName + "-" + stackName, {
    imageName: pulumi.interpolate`${registry.loginServer}/${productName}:latest`,
    build: { context: `../app/` },
    registry: {
        server: registry.loginServer,
        username: adminUsername,
        password: adminPassword,
    },
});

// Publish webapp 
const plan = new web.AppServicePlan(productName + "-" + stackName, {
    resourceGroupName: resourceGroup.name,
    kind: "Linux",
    reserved: true,
    sku: {
        name: "B1",
        tier: "Basic",
    },
});

const app = new web.WebApp(productName + "-" + stackName, {
    resourceGroupName: resourceGroup.name,
    serverFarmId: plan.id,
    siteConfig: {
        appSettings: [
            {
                name: "DOCKER_REGISTRY_SERVER_URL",
                value: pulumi.interpolate`https://${registry.loginServer}`,
            },
            {
                name: "DOCKER_REGISTRY_SERVER_USERNAME",
                value: adminUsername,
            },
            {
                name: "DOCKER_REGISTRY_SERVER_PASSWORD",
                value: adminPassword,
            },
            {
                name: "WEBSITES_PORT",
                value: "8080",
            },
            {
                name: "GOOGLE_PROVIDER_AUTHENTICATION_SECRET",
                value: cfg.requireSecret("googleProviderAuthenticationSecret"),
            },
            {
                name: "DOORDASH_DEVELOPER_ID",
                value: cfg.require("developerId"),
            },
            {
                name: "DOORDASH_KEY_ID",
                value: cfg.require("keyId"),
            },
            {
                name: "DOORDASH_SIGNING_SECRET",
                value: cfg.require("signingSecret"),
            },
            {
                name: "GOOGLE_API_KEY",
                value: cfg.require("googleApiKey"),
            },
            {
                name: "STACK_NAME",
                value: stackName,
            }
        ],
        alwaysOn: true,
        linuxFxVersion: pulumi.interpolate`DOCKER|${image.imageName}`,
    },
    httpsOnly: true,
});

const authSettings = new web.WebAppAuthSettingsV2(productName + "-" + stackName, {
    name: app.name,
    resourceGroupName: resourceGroup.name,
    globalValidation: {
        requireAuthentication: true,
        unauthenticatedClientAction: UnauthenticatedClientActionV2.RedirectToLoginPage,
        redirectToProvider: "google",
    },
    identityProviders: {
        google: {
            enabled: true,
            registration: {
                clientId: cfg.requireSecret("googleProviderAuthenticationClientId"),
                clientSecretSettingName: "GOOGLE_PROVIDER_AUTHENTICATION_SECRET",
            }
        }
    }
});

export const endpoint = pulumi.interpolate`https://${app.defaultHostName}`;