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