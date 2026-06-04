# Security

Terralist is designed to run as a private registry for trusted users. It is **not** hardened to be exposed directly to the public internet, and we strongly recommend against deploying it that way.

Run Terralist behind your own network boundary (VPN, private network, or an authenticating reverse proxy) and treat every authenticated tenant as a party that can reach the server, not just the registry API. Several capabilities, such as publishing a module from a remote URL, cause the server to act on tenant-supplied input.

## Threats and mitigations

### Server-side request forgery through artifact fetching

When a module or provider is published, the tenant supplies a download URL that Terralist fetches server-side and stores as a retrievable artifact. Without restrictions, a tenant with write access to a single authority can point that URL at an internal service or a cloud metadata endpoint (for example `http://169.254.169.254/`), and the response is exfiltrated back through the normal download endpoint.

Terralist refuses to fetch from private, loopback, link-local and unspecified addresses by default. The check runs against the resolved IP at connection time, so it also covers redirect targets and is not bypassable through DNS rebinding.

This behavior is controlled by [`fetch-allow-private-addresses`](../configuration.md#fetch-allow-private-addresses), which defaults to `false`.

#### Serving artifacts from a private network

If you host your module or provider artifacts on a host that only has a private address (for example an internal artifact server inside your cluster), the default behavior will block those fetches. In that case, set `fetch-allow-private-addresses` to `true`.

Enabling this removes the protection above for every fetch, so only do this when the network the registry can reach is itself trusted.
