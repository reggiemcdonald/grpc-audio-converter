CREATE TABLE convert_jobs (
    id varchar(50) PRIMARY KEY,
    status varchar(30),
    curr_url text,
    last_updated timestamp
);