# goFunction-Cli

A CLI tool for creating GCP Cloud Functions in Gen2 using the "functions-framework-go"

Often I'm required to build out a quick serverless function to give a website access to a server side function in Genesys Cloud or another service and to front end a Client Credentials server side request with allowing CORS etc. To do this with Google Cloud Platform you can leverage the [Functions framework](https://cloud.google.com/functions/docs/functions-framework) to run and test the serverless function on your localhost machine before deploying it to the cloud. To do this though you need to follow a specific folder structure as well as leverage the SDK.

This CLI tool is designed to make it SUPER easy to build and deploy these with the Genesys Cloud GO SDK with your own Basic Auth to make demo development easy.

## Installing the CLI

You first of all need to ensure you have [GO installed](https://go.dev/doc/install) based on your operating system that may differ in process but its well documented in that link.

Now you can simply run:

```
go install github.com/mcphee11/goFunction-Cli@latest
```

This will then install the CLI tool in your GOPATH. if this is setup in your system paths then you can now simply use the CLI directly. To test this type:

```
goFunction-Cli --version
```

And you should get an output like: `goFunction-Cli version version 0.4` NOTE: as I update the versions this number may be higher.

## Using the CLI

To create a new project go to a location in your terminal eg: `~/Documents` and then type in the below command ensuring to change the `--flags` with your own information

```
goFunction-Cli buildGenesys --region=YOUR_REGION --clientId=YOUR_CLIENT_ID --secret=YOUR_SECRET --username=YOUR_USERNAME --password=YOUR_PASSWORD --name=YOUR_FUNCTION_NAME
```

I have put some more details on what each of these `flags` mean. This information can also be found by going

```
goFunction-Cli buildGenesys --help
```

Here is the output about these flags

    --clientId string    OAuth 2.0 ClientId
    --name string        A name for the cloud function
    --password string   A Basic password for external auth
    --region string      The Genesys Cloud Region eg: mypurecloud.com.au
    --secret string      OAuth 2.0 secret
    --username string    A Basic username for external auth

## Running the project locally

Now there will be a few files and folders created to run this project locally simply cd into the folder created that will be the --name you gave the project and run the `run.sh` file

```
./run.sh
```

By default this runs the server on port 8081

This default config is reading a POST request with a body of {id: xyz} where xyz = conversationId. As the default is also using a Basic Auth you will need to pass in the -u if using curl and example is below:

```
curl -u username:password http://localhost:8081 -H "Content-Type: application/json" -d '{"Id":"YOUR_CONVERSATION_ID"}'
```

## Deploying the project to GCP

To deploy this to GCP as a serverless function, in the same root folder of the project where you ran the `run.sh` file you will find a `deploy.sh` file so like above run this file.

```
./deploy.sh
```

### NOTE:

While I am passing in the `--flags` when building this file you will require to have [gcloud](https://cloud.google.com/sdk/docs/install) installed which is the Google Cloud CLI tool. As well as a default `Project` set in gcloud, I have set the GCP region to deploy as `--region=australia-southeast1` you may want to change that as well to your own local region.

Right now I am also only storing the clientId and secret as env-vars. This should really be using GCP secret storage which I plan to move this to in the future. Also im using the `--allow-unauthenticated` and setting my own 'Basic Auth' in the function. In a prod env you may want to use the GCP function Auth that they provide.

This method is designed for me to quickly spin up a test then destroy it later.

## Changing the request

In the `genesys.go` file is where the request to the server side endpoint happens so when your changing it to call endpoint X this is where you can edit the code to suit your use case.

## Final thoughts

I may add more into this in the future but for now I hope this enabled you to spin up demo test functions much faster as it does for me.
