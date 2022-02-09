import * as pulumi from "@pulumi/pulumi";
import * as containerregistry from "@pulumi/azure-native/containerregistry";
import * as resources from "@pulumi/azure-native/resources";
import * as web from "@pulumi/azure-native/web";

import * as docker from "@pulumi/docker";
import { UnauthenticatedClientAction } from "@pulumi/azure-native/web/v20150801";
import { UnauthenticatedClientActionV2 } from "@pulumi/azure-native/web/v20200601";

const productName = "request-a-dasher";
const stackName = pulumi.getStack();

const resourceGroup = new resources.ResourceGroup(productName);

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
                value: "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",
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
                clientId: "879682729138-b6pvks3oh0qid7it8v3llkf29f9ek86r.apps.googleusercontent.com",
                clientSecretSettingName: "GOOGLE_PROVIDER_AUTHENTICATION_SECRET",
            }
        }
    }
});

export const endpoint = pulumi.interpolate`https://${app.defaultHostName}`;
