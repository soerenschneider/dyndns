
# Available Metrics

## Common Metrics (shared in server and client mode)
Here's the markdown table for the updated variables:

| Metric Name                        | Description                                     | Labels                   |
| ---------------------------------- | ----------------------------------------------- | ------------------------ |
| version                            | Version metric                                  | version, hash            |
| start_time_seconds                 | Process start time in seconds                   | N/A                      |
| mqtt_brokers_configured_total      | Total count of configured MQTT brokers          | N/A                      |
| mqtt_brokers_connected_total       | Total count of connected MQTT brokers           | N/A                      |
| mqtt_connections_lost_total        | Total count of lost MQTT connections            | N/A                      |
| notification_errors                | Gauge indicating the number of notification errors | N/A                      |


## Server Metrics
| Metric Name                               | Description                                            | Labels                       |
| ----------------------------------------- | ------------------------------------------------------ | ---------------------------- |
| dyndns_heartbeat_timestamp_seconds        | Server heartbeat timestamp                             | N/A                          |
| dyndns_known_hosts_configuration_hash     | Known hosts configuration hash                        | N/A                          |
| dyndns_dns_propagation_requests_total     | Total count of DNS propagation requests                | N/A                          |
| dyndns_dns_propagation_request_timestamp_seconds | Timestamp of the latest DNS propagation request  | N/A                          |
| dyndns_dns_propagations_total             | Total count of successful DNS propagations             | host                         |
| dyndns_dns_propagations_errors_total      | Total count of DNS propagation errors                  | host                         |
| dyndns_messages_received_total            | Total count of received messages                       | N/A                          |
| dyndns_signature_verifications_errors_total | Total count of signature verification errors         | host                         |
| dyndns_messages_ignored_total              | Total count of ignored messages                         | host, reason                 |
| dyndns_message_validations_failed_total    | Total count of failed message validations              | host, reason                 |
| dyndns_vault_token_expiry_time_seconds    | Expiry time of the Vault token                          | N/A                          |
| dyndns_config_public_key_errors_total     | Total count of public key configuration errors          | N/A                          |
| dyndns_message_parsing_failed_total       | Total count of failed message parsing                   | N/A                          |


## Client Metrics

| Metric Name                                | Description                                            | Labels                            |
| ------------------------------------------ | ------------------------------------------------------ | --------------------------------- |
| dyndns_ip_resolves_errors_total             | Total count of IP resolve errors                       | host, resolver, name              |
| dyndns_ip_resolved_successful_total         | Total count of successful IP resolves                  | host, resolver, name              |
| dyndns_reconcilers_pending_changes_total    | Total count of pending changes for reconcilers          | host                              |
| dyndns_reconciler_timestamp_seconds         | Timestamp of reconciler activity                       | host                              |
| dyndns_ip_resolves_invalid_total            | Total count of invalid IP resolves                     | host, resolver, url               |
| dyndns_ip_resolves_success_total            | Total count of successful IP resolves                  | host, resolver                    |
| dyndns_ip_resolves_last_check_timestamp_seconds | Timestamp of the last IP resolve check          | host, resolver                    |
| dyndns_updates_dispatch_errors_total        | Total count of update dispatch errors                  | host                              |
| dyndns_updates_dispatched_total             | Total count of dispatched updates                      | N/A                               |
| dyndns_state_changed_timestamp              | Timestamp of state change                              | host, from, to                    |
| dyndns_current_state_bool                   | Current state as a boolean value                       | host, state                       |
| dyndns_resolver_response_time_seconds       | Histogram of resolver response time in seconds         | resolver                          |
