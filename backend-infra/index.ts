import * as pulumi from "@pulumi/pulumi";
import * as containerregistry from "@pulumi/azure-native/containerregistry";
import * as resources from "@pulumi/azure-native/resources";
import * as web from "@pulumi/azure-native/web";

// Create an Azure Resource Group
const resourceGroup = new resources.ResourceGroup("deliverate-backend");

const plan = new web.AppServicePlan("deliverate-plan", {
    resourceGroupName: resourceGroup.name,
    kind: "Linux",
    reserved: true,
    sku: {
        name: "B1",
        tier: "Basic",
    },
});