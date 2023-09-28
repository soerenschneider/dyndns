# Configuration

## Client Configuration

| Field           | Type            | JSON Field                   | Environment Variable                |
|-----------------|-----------------|------------------------------|-------------------------------------|
| Host            | string          | host                         | DYNDNS_HOST                         |
| AddrFamilies    | []string        | address_families             | DYNDNS_ADDRESS_FAMILIES             |
| KeyPairPath     | string          | keypair_path                 | DYNDNS_KEYPAIR_PATH                 |
| MetricsListener | string          | metrics_listen               | DYNDNS_METRICS_LISTEN               |
| PreferredUrls   | []string        | http_resolver_preferred_urls | DYNDNS_HTTP_RESOLVER_PREFERRED_URLS |
| FallbackUrls    | []string        | http_resolver_fallback_urls  | DYNDNS_HTTP_RESOLVER_FALLBACK_URLS  |
| Once            | bool            | -                            | -                                   |
| MqttConfig      | MqttConfig      | -                            | -                                   |
| EmailConfig     | EmailConfig     | notifications                | -                                   |
| InterfaceConfig | InterfaceConfig | -                            | -                                   |

## MqttConfig

| Field          | Type     | JSON Field      | Environment Variable |
|----------------|----------|-----------------|----------------------|
| Brokers        | []string | brokers         | DYNDNS_BROKERS       |
| ClientId       | string   | client_id       | DYNDNS_CLIENT_ID     |
| CaCertFile     | string   | tls_ca_cert     | DYNDNS_TLS_CA        |
| ClientCertFile | string   | tls_client_cert | DYNDNS_TLS_CERT      |
| ClientKeyFile  | string   | tls_client_key  | DYNDNS_TLS_KEY       |
| TlsInsecure    | bool     | tls_insecure    | DYNDNS_TLS_INSECURE  |


## EmailConfig

| Field        | Type     | JSON Field | Environment Variable  |
|--------------|----------|------------|-----------------------|
| From         | string   | from       | DYNDNS_EMAIL_FROM     |
| To           | []string | to         | DYNDNS_EMAIL_TO       |
| SmtpHost     | string   | host       | DYNDNS_EMAIL_HOST     |
| SmtpPort     | int      | port       | DYNDNS_EMAIL_PORT     |
| SmtpUsername | string   | user       | DYNDNS_EMAIL_USER     |
| SmtpPassword | string   | password   | DYNDNS_EMAIL_PASSWORD |


## Server Config

| Field           | Type                | JSON Field     | Environment Variable |
|-----------------|---------------------|----------------|----------------------|
| KnownHosts      | map[string][]string | known_hosts    | -                    |
| HostedZoneId    | string              | hosted_zone_id | -                    |
| MetricsListener | string              | metrics_listen | -                    |
| MqttConfig      | MqttConfig          | -              | -                    |
| VaultConfig     | VaultConfig         | -              | -                    |
| EmailConfig     | EmailConfig         | notifications  | -                    |


## Vault Config
Here's a markdown table that displays the name, type, JSON field name, and environment variable name (if applicable) for each field in the `VaultConfig` struct:

| Field           | Type              | JSON Field            | Environment Variable |
|-----------------|-------------------|-----------------------|----------------------|
| VaultAddr       | string            | vault_addr            | -                    |
| AuthStrategy    | VaultAuthStrategy | vault_auth_strategy   | -                    |
| AwsRoleName     | string            | vault_aws_role_name   | -                    |
| AwsMountPath    | string            | vault_aws_mount_path  | -                    |
| AppRoleId       | string            | vault_app_role_id     | -                    |
| AppRoleSecretId | string            | vault_app_role_secret | -                    |
| VaultToken      | string            | vault_token           | -                    |

Please note that some fields do not have corresponding environment variable names as they are not specified in the `env` tag.

## Reference
| Keyword        | Description                                    | Example                      | Mandatory |
|----------------|------------------------------------------------|------------------------------|-----------|
| host           | FQDN of the host you want to update the IP for | https://vault:8200           | Y         |
| keypair_path   | The path to the ed25519 keypair                | /etc/dyndns/keypair          | Y         |
| metrics_listen | HTTP metrics handler listen address            | :9095                        | N         |
| brokers        | The MQTT brokers to connect to                 | ["tcp://host.tld":1883]      | Y         |
| client_id      | Client id for the MQTT connection              | crazy-horse                  | Y         |
| interface      | The network interface to check for IP updates  | eth0                         | N         |
