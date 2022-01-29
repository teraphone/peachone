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
    LAST DEPLOYED: Fri Jan 28 17:47:29 2022
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
export PORT="3000"
export SIGNING_KEY="secret"
```

Or they can be defined inline:

```
DB_HOST="127.0.0.1" DB_USER="postgres" DB_PASSWORD="pw" DB_NAME="peachone-dev" DB_PORT="5432" DB_AUTOMIGRATE="false" PORT="3000" SIGNING_KEY="secret" ./peachone
```
