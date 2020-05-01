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
- [x] Containerized environment
- [ ] Converting buffers? (would only work for audio that doesn't need seekable files)
- [ ] Limited concurrency, channel listeners
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
 "sourceUrl": "<string>"
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
  "id": "<string>",
  "status": "<string>",
  "url": "<string>"
}
```
where:
- `id`: job ID string
- `status`: current job status, one of `CONVERTING` | `COMPLETED` | `FAILED`
    - `QUEUED` status will becoming soon
- `url`: URL string to dwonload the converted audio - this is a presigned URL
that is valid for 24h from the time of conversion

### Deployment
You will need the following installed:
- Docker
- Docker-compose
- Terraform CLI

A dev and prod deployment have been created. The dev deployment runs entirely locally, and does not 
make use of AWS. The prod deployment creates an S3 storage bucket.

Begin by cloning the repo. Once complete, continue with the following steps.

Setup:
1. Create the `.env` file in the project root directory:
    - Locate your `AWS_ACCESS_KEY` and add it to the `.env`. Alternatively, you may omit this if you only intend to deploy locally.
    - Locate `AWS_SECRET_KEY` and add it to the `.env`. Alternatively, you may omit this if you only intend to deploy locally.
    - Create a username for the database, and add it to the `.env` file with the key `POSTGRES_USER` 
    - Create a password for the database, and add it to the `.env` file under the key `POSTGRES_PASSWORD` 

#### To deploy locally:
1. `cd` to the project root directory
2. Run `./run`
3. Select option (4) to build and run dev

To close, `ctrl+c`. Then select option (6) to remove the images from your docker.

#### To deploy to AWS:
1. `cd` to the project
2. If you dont have the AWS CLI installed, then you should create a `terraform.tfvars` file and add `AWS_ACCESS_KEY` 
and `AWS_SECRET_KEY` to the file
3. The default region is `us-west-2`. You can change this in `terraform.tfvars`
4. Run `./run`
5. Select option (1) to deploy a new S3 bucket for the microservice
6. Select (3) to build and run grpc-audio-converter

To close, `ctrl+c`. Then select option (5) to remove the images. Select option (2) to fully delete the 
S3 bucket (including any data that is in the bucket).
