# Configuration

## Server Configuration

### CLI Reference
```
Usage of ./dyndns-server:
  -config string
        Path to the config file (default "/etc/dyndns/config.json")
  -version
        Print version and exit
```

### Reference
| Keyword               | Description                                                                                               | Example                                                                                       | Mandatory |
|-----------------------|-----------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------|-----------|
| known_hosts           | A map of known hostnames and their public key                                                             | {"example.domain.tld":"AAAAC3NzaC1lZDI1NTE5AAAAIBfJ2Qjt5GPi7DKRPGxJCkvk8xNsG9dA607tnWagOk2D"} | Y         |
| hosted_zone_id        | [AWS Route53 hosted zone id](https://docs.aws.amazon.com/Route53/latest/APIReference/API_HostedZone.html) | Z119WBBTVP5WFX                                                                                | Y         |
| metrics_listen        | HTTP metrics handler listen address                                                                       | :9095                                                                                         | N         |
| broker                | The MQTT broker you connect to                                                                            | tcp://host.tld                                                                                | Y         |
| client_id             | Client id for the MQTT connection                                                                         | crazy-horse                                                                                   | Y         |
| vault_addr            | Vault URL                                                                                                 | https://vault.tld:8200                                                                        | Y         |
| vault_app_role_id     | [AppRole role id](https://www.vaultproject.io/docs/auth/approle) to authenticate against vault            | 988a9dfd-ea69-4a53-6cb6-9d6b86474bba                                                          | Y         |
| vault_app_role_secret | [AppRole secret id](https://www.vaultproject.io/docs/auth/approle) to authenticate against vault          | 37b74931-c4cd-d49a-9246-ccc62d682a25                                                          | Y         |

## Client Configuration

### CLI Reference
```
Usage of ./dyndns-client:
  -config string
        Path to the config file (default "/etc/dyndns/client.json")
  -once
        Path to the config file
  -version
        Print version and exit

```

### Reference
| Keyword        | Description                                    | Example             | Mandatory |
|----------------|------------------------------------------------|---------------------|-----------|
| host           | FQDN of the host you want to update the IP for | https://vault:8200  | Y         |
| keypair_path   | The path to the ed25519 keypair                | /etc/dyndns/keypair | Y         |
| metrics_listen | HTTP metrics handler listen address            | :9095               | N         |
| broker         | The MQTT broker you connect to                 | tcp://host.tld      | Y         |
| client_id      | Client id for the MQTT connection              | crazy-horse         | Y         |
| interface      | The network interface to check for IP updates  | eth0                | N         |

# Available Metrics

| Subsystem | Metric                                    | Type    | Description         | Labels         |
|-----------|-------------------------------------------|---------|---------------------|----------------|
|           | version                                   | gauge   | Version information | version, hash  |
| server    | dns_propagation_requests_total            | counter |                     |                |
| server    | dns_propagation_request_timestamp_seconds | gauge   |                     |                |
| server    | dns_propagations_total                    | counter |                     | host           |
| server    | dns_propagations_errors_total             | counter |                     | host           |
| server    | messages_received_total                   | counter |                     |                |
| server    | signature_verifications_errors_total      | counter |                     | host           |
| server    | public_keys_missing_total                 | counter |                     | host           |
| server    | message_validations_failed_total          | counter |                     | host, reason   |
| server    | vault_token_expiry_time_seconds           | gauge   |                     |                |
|           | mqtt_connections_lost_total               | counter |                     |                |
| server    | config_public_key_errors_total            | counter |                     |                |
| server    | message_parsing_failed_total              | counter |                     |                |
| client    | ip_resolves_errors_total                  | counter |                     | host, resolver |
| client    | ip_resolves_invalid_total                 | counter |                     | host, resolver |
| client    | ip_resolves_success_total                 | counter |                     | host, resolver |
| client    | ip_resolves_last_check_timestamp_seconds  | gauge   |                     | host, resolver |
| client    | updates_dispatch_errors_total             | counter |                     | host           |
| client    | updates_dispatched_total                  | counter |                     |                |
| client    | state_changed_timestamp                   | gauge   |                     | host, from, to |