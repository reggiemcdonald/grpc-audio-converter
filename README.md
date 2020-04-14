### A gRPC Microservice for Audio Encoding Conversion

#### What is it? 
This microservice provides audio file conversion, where a public URL is provided
and converted to the desired format before being uploaded to an S3 bucket. The converted audio
is persisted for 24h, during which the user can request a presigned GET url 
to the object.

#### A not on gRPC and Protocol Buffers
The audio conversion microservice uses gRPC and Protocol Buffers for communication.
gRPC provides a convenient and performant interface for communication between microservices.
Protocol Buffers allow for a language-agnostic interface to facilitate gRPC.

A REST interface was provided, primarily for local testing.

#### Deployment
This process will update as I continue to dockerize this microservice. Currently, its relatively manual.
Terraform is used to deploy the necessary S3 bucket

1. Clone the repo 
2. Install `terraform cli`, `aws cli`, and `ffmpeg`
3. In the project root, `docker-compose up -d` to start the database
4. If you do not have an AWS configuration for the CLI, you may create one with `aws configure`
    - Optionally, add `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` to the `terraform.tfvars` and `.env` files instead
5. Decide whether to deploy the S3 bucket locally or online: 
    - You can deploy locally if you have `localstack` running
6. Select the appropriate deploy, you must say `yes`
7. The bucket name will be shown once complete, add this as a line in a `.env`:
    - `BUCKET_NAME=<paste name here>`
8. Run `./run` and select the deployment option you want. Then run the server
9. The converter service will be running on :3000, the rest interface on :4000 and the database
on :5432