### A gRPC Microservice for Audio Encoding Conversion

#### What is it? 
This microservice provides audio file conversion, where a public URL is provided
and converted to the desired format before being uploaded to an S3 bucket. The converted audio
is persisted for 24h, during which the user can request a presigned GET url 
to the object.

#### Supported Formats
The list is going to be a lot bigger shortly once I update the ProtoBuff. Currently supported encodings are:
- WAV
- M4A
- MP3
- FLAC
Again, since audio conversion is done through the FFMPEG tool, this list will get longer shortly (or become a link to the FFMPEG docs).

#### A note on gRPC and Protocol Buffers
The audio conversion microservice uses gRPC and Protocol Buffers for communication.
gRPC provides a convenient and performant interface for communication between microservices.
Protocol Buffers allow for a language-agnostic interface to facilitate gRPC.

A REST interface was provided, primarily for local testing.

#### Deployment
This process will update as I continue to dockerize this microservice. Currently, its relatively manual.
Terraform is used to deploy the necessary S3 bucket

1. Clone the repo 
2. Install `terraform cli`, `aws cli`, and `ffmpeg`
3. Create `.env` and add `POSTGRES_USER`, `POSTGRES_PASSWORD` with values of your choosing. Add `REGION` and specify the AWS region that you plan on deploying to
4. In the project root, `docker-compose up -d` to start the database
5. If you do not have an AWS configuration for the CLI, you may create one with `aws configure`
    - Optionally, add `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` to the `terraform.tfvars` and `.env` files instead
6. Decide whether to deploy the S3 bucket locally or online: 
    - You can deploy locally if you have `localstack` running
    - If your region specified in step 3 is different than `us-west-2`, then specify this region in `terraform.tfvars`
7. Run `./run` and select the appropriate deploy. You will need to say `yes`
8. The bucket name will be shown once complete, add this as a line in a `.env`:
    - `BUCKET_NAME=<paste name here>`
9. Select the run option
10. The converter service will be running on :3000, the rest interface on :4000 and the database
on :5432
