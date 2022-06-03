# Kubernetes PostgreSQL
ArtifactHub link [here](https://artifacthub.io/packages/helm/bitnami/postgresql).

1. Create secret with:

    ```
    kubectl create secret generic postgresql-creds \
     --from-literal=postgresql-password=pw
    ```
    
2. Install with 

    ```
    helm install postgresql -f postgresql-values.yaml bitnami/postgresql
    ```

    Output:

    ```
    NAME: postgresql
    LAST DEPLOYED: Mon Feb  7 19:24:16 2022
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    NOTES:
    CHART NAME: postgresql
    CHART VERSION: 10.16.2
    APP VERSION: 11.14.0

    ** Please be patient while the chart is being deployed **

    PostgreSQL can be accessed via port 5432 on the following DNS names from within your cluster:

        postgresql.default.svc.cluster.local - Read/Write connection

    To get the password for "postgres" run:

        export POSTGRES_PASSWORD=$(kubectl get secret --namespace default postgresql-creds -o jsonpath="{.data.postgresql-password}" | base64 --decode)

    To connect to your database run the following command:

        kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace default --image docker.io/bitnami/postgresql:11.14.0-debian-10-r28 --env="PGPASSWORD=$POSTGRES_PASSWORD" --command -- psql --host postgresql -U postgres -d peachone-dev -p 5432

    To connect to your database from outside the cluster execute the following commands:

        kubectl port-forward --namespace default svc/postgresql 5432:5432 &
        PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U postgres -d peachone-dev -p 5432
    ```

# GORM connection

For gorm to open a connection to the database it needs to know the following: host, username, password, dbname, port... 

```
const DNS = "host=127.0.0.1 user=postgres password=pw dbname=peachone-dev port=5432 sslmode=disable TimeZone=US/Pacific"
```

These need to passed in as environment variables...

# Environment variables

```
export DB_HOST="127.0.0.1"
export DB_USER="postgres"
export DB_PASSWORD="pw"
export DB_NAME="peachone-dev"
export DB_PORT="5432"
export DB_AUTOMIGRATE="false"
export PORT="8080"
export SIGNING_KEY="secret"
export LIVEKIT_KEY=<secret-key>
export LIVEKIT_SECRET=<secret-value>
export LIVEKIT_HOST="sfu-demo.teraphone.app"
export SERVICE_ACCOUNT_JSON=<path-to-service-account-key-file>
```

Or they can be defined inline:

```
DB_HOST="127.0.0.1" DB_USER="postgres" DB_PASSWORD="pw" DB_NAME="peachone-dev" DB_PORT="5432" DB_AUTOMIGRATE="false" PORT="8080" SIGNING_KEY="secret" LIVEKIT_KEY=<secret-key> LIVEKIT_SECRET=<secret-value> LIVEKIT_HOST="sfu-demo.teraphone.app" SERVICE_ACCOUNT_JSON="livekit-demo-firebase-adminsdk-7dp4y-85e675b567.json" ./peachone
```

Note: the SERVICE_ACCOUNT_JSON environment variable is necessary for local development only. If the service is running in gcloud then the variable should be empty. SERVICE_ACCOUNT_JSON should be a path to the service account key for the Firebase Admin SDK available [here](https://console.firebase.google.com/project/livekit-demo/settings/serviceaccounts/adminsdk). Warning: this key should be kept secret.

# REST API Endpoints

## /v1/public
/signup
- POST: create an account and get private API auth token
  
/login
- POST: login and get private API auth token

## /v1/private (requires auth token)
/auth
- GET: returns a new auth token with refreshed expiration

/groups
- GET: returns list of group objects that the user has access to. each group object contains details about the group
- POST: create new group

/groups/id (must have access to group)
- GET: returns details about the group
- PATCH: (requires group admin) modify group properties
- DELETE: (requires group admin) “deletes” the group… tbd

/groups/id/users
- GET: returns list of user objects that are members of the group (from GroupUsers table)
- POST: (requires group admin) add user to group with role: "base"

/groups/id/users/id
- GET: returns user object (from GroupUsers table)
- DELETE: (requires group admin) removes user from group 
- PATCH: (requires group admin) modifies user properties (e.g. role) 

/groups/id/invites
- GET: (requires group admin) returns list of invite objects for the group (from GroupInviteCodes table)
- POST: (requires group admin) creates an invite code to the group

/groups/id/invites/id
- GET: (requires group admin) returns the invite object with the corresponding id
- DELETE: (requires group admin) deletes the invite

/groups/id/rooms
- GET: returns a list of room objects that the user has access to
- POST: create a new room

/groups/id/rooms/id
- GET: (requires room access) returns details of the room
- DELETE: (requires group admin) “deletes” the room… tbd
- PATCH: (requires room admin) modifies room properties

/groups/id/rooms/id/users (RoomUsers entries are created for all group members when room is created)
- GET: returns list of user objects for the room (from RoomUsers table)
- POST: (requires room admin) adds user to room. this may not be necessary since room users are created when room is created or when a user is added to the group

/groups/id/rooms/id/users/id
- GET: returns user object (from RoomUsers table)
- PATCH: (requires room admin) update RoomUser properties

## /v1/roomservice (interacting with voice chat server)
/rooms
- GET: return a list of rooms that are active
- POST: start a room if it's inactive (this appears to be unnecessary)

/rooms/:group_id/:room_id 
- GET: return list of room participants
- PATCH: (requires room admin) room moderation (kick, or room-wide mute/unmute) (is this necessary? do through desktop client instead?)
- DELETE: (requires room admin) drop all participants and terminate the room (probably don't need this)

/rooms/:group_id/:room_id/join
- GET: returns the join token for the room

# Docker Image
## Build & Push Docker Image

Building locally and pushing using Docker. (NOTE: if you're aren't on an Intel computer you need to use buildx. See "Deploy to Cloud Run" below.)

1. Navigate to the peachone directory (dir of this readme).
2. Run the command:

    ```
    docker build . --tag us-west1-docker.pkg.dev/livekit-demo/peachone/peachone:latest
    ```

3. If you have not yet configured Docker to use the gcloud command-line tool to authenticate requests to Container Registry, do so now using the command:

    ```
    gcloud auth configure-docker
    ```

    You need to do this before you can push or pull images using Docker. You only need to do it once.

4. Push the container image to Container Registry:

    ```
    docker push us-west1-docker.pkg.dev/livekit-demo/peachone/peachone:latest
    ```

### Deploy to Cloud Run

Documentation can be found [here](https://cloud.google.com/run/docs/deploying).

Ran into issues due to image being built on my Apple Silicon Mac (arm64), causing Cloud Run to fail. Solution is to build with multi-arch support. Details [here](https://docs.docker.com/desktop/multi-arch/).

    ```
    docker buildx build --platform linux/amd64,linux/arm64 -t us-west1-docker.pkg.dev/livekit-demo/peachone/peachone:latest . --push
    ```

If you don't have an Artifact Registry repo yet you can create one [here](https://console.cloud.google.com/artifacts).

## Run Docker Image Locally

Instructions for testing locally can be found [here](https://cloud.google.com/run/docs/testing/local#docker).

1. Use the Docker command:

    ```
    PORT="8080" && docker run -p 8080:${PORT} -e PORT=${PORT} -e DB_HOST="127.0.0.1" -e DB_USER="postgres" -e DB_PASSWORD="pw" -e DB_NAME="peachone-dev" -e DB_PORT="5432" -e DB_AUTOMIGRATE="false" -e SIGNING_KEY="secret" -e LIVEKIT_KEY=<secret-key> -e LIVEKIT_SECRET=<secret-value> -e LIVEKIT_HOST="sfu-demo.teraphone.app" -e SERVICE_ACCOUNT_JSON=<path-to-service-account-key-file> [image-url]
    ```

    Replace [image-url] with the URL of the image you just pushed to Container Registry: `us-west1-docker.pkg.dev/livekit-demo/peachone/peachone:latest`. Don't forget to set the other necessary environment variables!

2. Open [http://localhost:8080/v1/public](http://localhost:8080/v1/public) in your browser. __NOTE__: you will __not__ be able to use any private endpoints since the container wont be able to find the Postgres database. (On Mac you can get around this by setting the DB_HOST environment var to DB_HOST="docker.for.mac.localhost" when you start the container.)

# Connecting to the remote DB with vscode client

Follow instructions [here](https://cloud.google.com/sql/docs/postgres/connect-admin-ip).

For secure access you will need to follow the instructions [here](https://cloud.google.com/sql/docs/postgres/configure-ssl-instance#new-client).