# IAM Authentication & SSL

## IAM Database Authentication

Use AWS IAM roles/users instead of database passwords. Authentication tokens replace passwords and expire after 15 minutes.

### Enable

```sh
aws rds modify-db-instance \
  --db-instance-identifier mydb \
  --enable-iam-database-authentication \
  --apply-immediately
```

### Create Database User

```sql
-- Create user with rds_iam role (no password needed)
CREATE USER app_user;
GRANT rds_iam TO app_user;

-- Grant permissions as normal
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
```

### IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "rds-db:connect",
      "Resource": "arn:aws:rds-db:us-east-1:111111111111:dbuser:db-XXXXX/app_user"
    }
  ]
}
```

The resource ARN uses the DBI resource ID (not the instance identifier):

```sh
# Get DBI resource ID
aws rds describe-db-instances \
  --db-instance-identifier mydb \
  --query 'DBInstances[0].DbiResourceId'
```

### Generate Auth Token and Connect

```sh
# Generate token (valid for 15 minutes)
TOKEN=$(aws rds generate-db-auth-token \
  --hostname mydb.xxxxx.us-east-1.rds.amazonaws.com \
  --port 5432 \
  --username app_user)

# Connect with token as password
psql "host=mydb.xxxxx.us-east-1.rds.amazonaws.com \
  port=5432 dbname=mydb user=app_user \
  password=$TOKEN sslmode=verify-full \
  sslrootcert=us-east-1-bundle.pem"
```

### Go Example

```go
import (
    "context"
    "database/sql"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/rds/auth"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func connectWithIAM() (*sql.DB, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        return nil, err
    }

    endpoint := "mydb.xxxxx.us-east-1.rds.amazonaws.com:5432"
    token, err := auth.BuildAuthToken(context.TODO(), endpoint, "us-east-1", "app_user", cfg.Credentials)
    if err != nil {
        return nil, err
    }

    dsn := fmt.Sprintf("host=%s port=5432 dbname=mydb user=app_user password=%s sslmode=verify-full sslrootcert=us-east-1-bundle.pem",
        "mydb.xxxxx.us-east-1.rds.amazonaws.com", token)

    return sql.Open("pgx", dsn)
}
```

### Node.js Example

```typescript
import { RDS } from "@aws-sdk/client-rds";
import { Signer } from "@aws-sdk/rds-signer";
import pg from "pg";

const signer = new Signer({
  hostname: "mydb.xxxxx.us-east-1.rds.amazonaws.com",
  port: 5432,
  username: "app_user",
  region: "us-east-1",
});

const token = await signer.getAuthToken();

const client = new pg.Client({
  host: "mydb.xxxxx.us-east-1.rds.amazonaws.com",
  port: 5432,
  database: "mydb",
  user: "app_user",
  password: token,
  ssl: {
    rejectUnauthorized: true,
    ca: fs.readFileSync("us-east-1-bundle.pem"),
  },
});
```

### Token Caching

Tokens are valid for 15 minutes. Cache and refresh before expiry:

- Generate a new token every 10-12 minutes
- Connection pools: token is only used at connection creation, existing connections remain valid
- If using RDS Proxy with IAM auth, the proxy handles token refresh internally

### Limitations

- Max 256 new connections per second with IAM auth (throttling)
- Tokens are region-specific
- SSL is required (cannot use IAM auth without SSL)
- `max_connections` limit still applies

## SSL/TLS

### Force SSL on All Connections

```
# RDS parameter group
rds.force_ssl = 1
```

### SSL Modes

| Mode          | Certificate Validation  | MITM Protection | Use Case               |
| ------------- | ----------------------- | --------------- | ---------------------- |
| `disable`     | No                      | No              | Never (insecure)       |
| `allow`       | No                      | No              | Avoid                  |
| `prefer`      | No                      | No              | Default, not secure    |
| `require`     | No (encrypts only)      | No              | Minimum for production |
| `verify-ca`   | Validates CA            | Partial         | Good                   |
| `verify-full` | Validates CA + hostname | Yes             | Best — use this        |

### Download RDS CA Bundle

```sh
# Regional bundle
curl -o us-east-1-bundle.pem https://truststore.pki.rds.amazonaws.com/us-east-1/us-east-1-bundle.pem

# Global bundle (all regions)
curl -o global-bundle.pem https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
```

### Connect with SSL

```sh
psql "host=mydb.xxxxx.us-east-1.rds.amazonaws.com \
  dbname=mydb user=app_user \
  sslmode=verify-full \
  sslrootcert=global-bundle.pem"
```

### Verify SSL in Session

```sql
-- Check if current connection uses SSL
SELECT ssl, version, cipher FROM pg_stat_ssl WHERE pid = pg_backend_pid();

-- Check all connections for SSL
SELECT usename, ssl, client_addr
FROM pg_stat_ssl
JOIN pg_stat_activity USING (pid);
```

## RDS Proxy with IAM Auth

RDS Proxy handles IAM auth and connection pooling together.

```sh
# Create proxy with IAM auth
aws rds create-db-proxy \
  --db-proxy-name mydb-proxy \
  --engine-family POSTGRESQL \
  --auth '[{"AuthScheme":"SECRETS","SecretArn":"arn:aws:secretsmanager:...","IAMAuth":"REQUIRED"}]' \
  --role-arn arn:aws:iam::111111111111:role/rds-proxy-role \
  --vpc-subnet-ids subnet-xxx subnet-yyy

# Application connects to proxy endpoint with IAM token
# Proxy handles pooling and forwards to RDS
```

Advantages:

- Connection pooling without managing PgBouncer
- IAM auth at the proxy, password auth to the database
- Automatic failover handling (proxy pins to new primary)
- Multiplexes connections: many app connections → few database connections
