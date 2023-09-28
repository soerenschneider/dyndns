## Goals

### Resilience
#### Multiple Brokers
It's possible to specify multiple MQTT brokers at once. Instead of only connecting to one broker at a time, dyndns
tries to be connected to all configured brokers all the time.

#### Multiple Servers
Dyndns allows multiple servers to listen simultaneously for requests and try to upsert them. Instead of using complicated consensus mechanism, upserts for host records are expected to be idempotent.

#### Detection of IP updates
Currently, two methods are support to detect IP updates
1. Checking the IP address of the network interface connected to the internet. While this is very reliable, the client needs to run on your router.
2. Using external websites, such as ipinfo.net, that return your ip address. Trust is shifted to the website providing your IP address but it can be run on any computer within your network.

### Security
#### Pre-defined hostnames
Hosts have to be pre-defined in the dyndns server's configuration before requests are accepted for that host. This is done by specifying (multiple) [ed25519](https://ed25519.cr.yp.to/) [keypairs](https://en.wikipedia.org/wiki/Public-key_cryptography).

#### Spoofing protection
Update request payloads are signed using the host's configured private key. The signature is verified by the server component.

#### Replay attacks
A (signed) timestamp is included in the update request payload so even when you're using a public MQTT broker, replayed messages won't lead to updated host records.

#### Least privilege
The clients run without access to credentials of the DNS provider. The worst case scenario is the ability to change a single host record. Obviously this only makes sense if you have more than one host record that should be synchronized.

The dyndns server should only run a trusted and well-secured server though.

#### Support for dynamic IAM credentials
Dyndns supports acquiring dynamic IAM credentials through Vault's [AWS secret engine](https://developer.hashicorp.com/vault/docs/secrets/aws)

### Observability

#### Metrics
Prometheus metrics are supported and documented below. It's easy to define dashboards and alerts to make sure all is up and well.

#### Notifications
Notifications (such as emails) can be used to display events such as detection of updated IPs and applied updates.


(To reach some goals outlined above) the high level architecture is as follows.