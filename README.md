# Request a Dasher

A simple Go webapp for requesting a DoorDash Dasher via the [DoorDash Drive (v2) APIs](https://developer.doordash.com/en-US/api/drive).

## How to run

Make sure you've got the following environment variables set wherever you'll run the app.

- `DOORDASH_DEVELOPER_ID`: Get this from [Developer Portal](https://developer.doordash.com/portal/integration/drive/credentials)
- `DOORDASH_KEY_ID`: Get this from [Developer Portal](https://developer.doordash.com/portal/integration/drive/credentials)
- `DOORDASH_SIGNING_SECRET`: Get this from [Developer Portal](https://developer.doordash.com/portal/integration/drive/credentials)
- `GOOGLE_API_KEY`: Get a Google Maps API key from the [Google Cloud console](https://developers.google.com/maps/documentation/javascript/get-api-key)
```

Then, run the Go webserver:

```sh
cd app && go run main.go
```

Or, you can run it inside a Docker container:

```sh
cd app
docker build -t request-a-dasher:latest .
docker run -it -p 8080:8080 request-a-dasher:latest
```

## How to deploy to Azure

You can deploy this to an Azure App Service (optimized for minimal monthly cost) using the Pulumi program in `infra`. You'll need `pulumi` and the `az` CLI installed:

```sh
cd infra
az login
pulumi stack init dev
pulumi config set azure-native:location WestUS # or your region of choice
pulumi up
```
