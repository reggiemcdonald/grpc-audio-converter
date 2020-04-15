### A gRPC Microservice for Audio Encoding Conversion

### What is it? 
This microservice provides audio file conversion, where a public URL is provided
and converted to the desired format before being uploaded to an S3 bucket. The converted audio
is persisted for 24h, during which the user can request a presigned GET url 
to the object.

Its still a work in progress - see next section.
#### TODOs
- [x] Ability to convert full files from public URL
- [x] Ability to retrieve status updates and presigned URL when complete
- [ ] Converting buffers? (would only work for audio that doesn't need seekable files)
- [ ] Limiting concurrency, channel listeners
- [ ] Complete dockerization
- [ ] Proper instructions on use (will post once I get the docker containers setup)

### Supported Encodings
The list is going to be a lot bigger shortly once I update the ProtoBuff. Currently supported encodings are:
- WAV
- M4A
- MP3
- FLAC

Again, since audio conversion is done through the FFMPEG tool, this list will get longer shortly (or become a link to the FFMPEG docs).

### A note on gRPC and Protocol Buffers
The audio conversion microservice uses gRPC and Protocol Buffers for communication.
gRPC provides a convenient and performant interface for communication between microservices.
Protocol Buffers allow for a language-agnostic interface to facilitate gRPC. Right now I've compiled them for 
Go, but they can be compiled for different languages.

A REST interface was provided, primarily for local testing. The endpoints are listed below.

### REST Endpoints
#### `POST /convert-file`: creates a new conversion job.
Query Params: 
- `src`: The encoding of the source file
- `dest`: The desired converted encoding

Body:
```json
{
 "sourceUrl": <string>
}
``` 
where:
- `sourceUrl` is a URL string to download the audio file
Returns: 
`202` and the ID of your request on successful call
---
#### `GET /convert-file`: Gets the status of a job
Query params:
- `id`: The ID of the job
Returns:
- `200`
 ```json
{
  "id": <string>,
  "status": <string>,
  "url": <string>
}
```
where:
- `id`: job ID string
- `status`: current job status, one of `CONVERTING` | `COMPLETED` | `FAILED`
    - `QUEUED` status will becoming soon
- `url`: URL string to dwonload the converted audio - this is a presigned URL
that is valid for 24h from the time of conversion

### Deployment
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
