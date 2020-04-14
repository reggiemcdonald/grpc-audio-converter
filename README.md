### A gRPC Microservice for Converting Audio

#### Environment Variables:
- The microservice saves audio to a specified S3 bucket, and so the following 
parameters must be specified:
    - `REGION`: Region of the bucket
    - `BUCKET_NAME`: The name of the bucket that will contain the converted audio
    - `ACCESS_KEY`: The AWS access key of the user account
    - `SECRET_ACCESS_KEY`: The secret access key of the user account
- The following are optional parameters:
    - `S3_ENDPOINT`: An optional parameter - specify this for local testing
    of the microservice using localstack
  