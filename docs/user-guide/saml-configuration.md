# SAML 2.0 Configuration

Terralist supports SAML 2.0 authentication through the `saml` OAuth provider. This guide covers configuring SAML authentication with your Identity Provider (IdP), including basic setup and specific examples for popular providers.

## Overview

SAML 2.0 authentication in Terralist works through the following flow:

1. **SP-Initiated SSO**: User clicks login, Terralist generates SAML AuthnRequest and redirects to IdP
2. **Authentication**: User authenticates with the IdP
3. **Response**: IdP sends SAML Assertion back to Terralist's Assertion Consumer Service (ACS) endpoint
4. **Validation**: Terralist validates the SAML response and extracts user information
5. **Authorization**: User is granted access based on SAML attributes and RBAC configuration

## Basic Configuration

The minimum required configuration for SAML authentication includes:

- `oauth-provider`: Set to `saml`
- Either `saml-idp-metadata-url`, `saml-idp-metadata-file`, or `saml-idp-entity-id` and `saml-idp-sso-url` and `saml-idp-sso-certificate`

!!! note "Service Provider Entity ID"
    The Service Provider (SP) entity ID is automatically generated as the SP metadata URL: `https://your-terralist-instance.com/v1/api/auth/saml/metadata`. This is the standard SAML convention and ensures proper integration with Identity Providers.

### Required Configuration

| Configuration | Description | Example |
|--------------|-------------|---------|
| `oauth-provider` | Set to `saml` | `saml` |
| `saml-idp-metadata-url` | URL to fetch IdP metadata | `https://idp.example.com/metadata` |

### Optional Configuration

| Configuration | Default | Description |
|--------------|---------|-------------|
| `saml-name-attribute` | `displayName` | SAML attribute containing user's display name |
| `saml-email-attribute` | `email` | SAML attribute containing user's email |
| `saml-groups-attribute` | - | SAML attribute containing user's groups |
| `saml-http-client-timeout` | `30s` | Timeout for metadata HTTP requests |
| `saml-assertion-clock-skew` | `5m` | Allowed clock difference between SP and IdP |
| `saml-max-assertion-age` | `1h` | Maximum age of SAML assertions |
| `saml-allow-idp-initiated` | `false` | Whether to allow IdP-initiated SSO |

### IdP Metadata Sources

Terralist supports three ways to configure IdP information:

#### Option 1: Metadata URL (Recommended)

```bash
terralist server \
  --oauth-provider saml \
  --saml-idp-metadata-url "https://idp.example.com/saml/metadata"
```

#### Option 2: Metadata File

```bash
terralist server \
  --oauth-provider saml \
  --saml-idp-metadata-file "/path/to/idp-metadata.xml"
```

#### Option 3: Direct Configuration

```bash
terralist server \
  --oauth-provider saml \
  --saml-idp-entity-id "https://idp.example.com" \
  --saml-idp-sso-url "https://idp.example.com/saml/sso" \
  --saml-idp-sso-certificate "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
```

#### YAML Configuration

```yaml
oauth-provider: saml
saml-idp-metadata-url: "https://idp.example.com/saml/metadata"
saml-name-attribute: "displayName"
saml-email-attribute: "email"
token-signing-secret: "your-signing-secret"
```

For direct IdP configuration:

```yaml
oauth-provider: saml
saml-idp-entity-id: "https://idp.example.com"
saml-idp-sso-url: "https://idp.example.com/saml/sso"
saml-idp-sso-certificate: |
  -----BEGIN CERTIFICATE-----
  MIICiTCCAg+gAwIBAgIJAJ8l4HnPq7F1MAOGA1UEBhMCVVMxCzAJBgNVBAgTAkNB
  ...
  -----END CERTIFICATE-----
token-signing-secret: "your-signing-secret"
```

## Service Provider Metadata

Terralist automatically generates Service Provider (SP) metadata and serves it at:

```bash
https://your-terralist-instance.com/v1/api/auth/saml/metadata
```

This URL should be provided to your IdP administrator for SP registration. The metadata includes:

- Entity ID (automatically set to the metadata URL)
- Assertion Consumer Service (ACS) URL
- Single Logout Service (SLS) URL (if configured)
- SP certificate (if configured)

## Certificate Configuration

For production deployments, it's recommended to configure SP certificates for request signing:

```bash
terralist server \
  --saml-cert-file "/path/to/sp-cert.pem" \
  --saml-key-file "/path/to/sp-key.pem"
```

If your private key is encrypted:

```bash
terralist server \
  --saml-cert-file "/path/to/sp-cert.pem" \
  --saml-key-file "/path/to/sp-key.pem" \
  --saml-private-key-secret "your-key-password"
```

## User Attribute Mapping

Terralist extracts user information from SAML assertions using configurable attribute names:

### Basic Attributes

```bash
terralist server \
  --saml-name-attribute "displayName" \
  --saml-email-attribute "email"
```

### Attribute Templating

You can combine multiple SAML attributes into a single field using Go template syntax:

```bash
terralist server \
  --saml-name-attribute "{{.givenName}} {{.sn}}" \
  --saml-email-attribute "email"
```

This example combines `givenName` and `sn` (surname) attributes into a full name like "John Doe".

**Template Syntax:**

- Use `{{.attributeName}}` to reference SAML attributes
- Templates support all standard Go template functions
- If a referenced attribute doesn't exist, the template falls back to standard attribute lookup
- Invalid templates will log an error and fall back to standard lookup

**Examples:**

- `"{{.givenName}} {{.sn}}"` - Combine first and last name
- `"{{.displayName}} ({{.department}})"` - Add department info
- `"Dr. {{.givenName}} {{.sn}}"` - Add title prefix

### Group-based Authorization

For RBAC with groups, configure the groups attribute:

```bash
terralist server \
  --saml-groups-attribute "memberOf"
```

The groups attribute should contain a list of group names or identifiers that will be used for RBAC policy matching.

## Google Workspace SAML Example

This section provides a complete example of configuring SAML authentication with Google Workspace as the Identity Provider.

### Prerequisites

1. **Google Workspace Admin Access**: You need admin access to configure SAML applications
2. **Custom Domain**: Your Google Workspace should be configured with a custom domain
3. **HTTPS Transport**: SAML requires secure HTTPS transport. Ensure Terralist is accessible over HTTPS either through application-level TLS certificates or a TLS-terminating reverse proxy

### Step 1: Configure Terralist

First, configure Terralist with basic SAML settings:

```bash
terralist server \
  --url "https://terralist.example.com" \
  --oauth-provider saml \
  --saml-name-attribute "displayName" \
  --saml-email-attribute "email" \
  --saml-groups-attribute "groups"
```

### Step 2: Get Google Workspace SAML Metadata

1. Go to [Google Admin Console](https://admin.google.com)
2. Navigate to **Apps** → **Web and mobile apps**
3. Click **Add app** → **Add custom SAML app**
4. Enter app name: "Terralist"
5. Download the **IdP metadata** file

### Step 3: Configure Terralist with Google Metadata

```bash
terralist server \
  --url "https://terralist.example.com" \
  --oauth-provider saml \
  --saml-idp-metadata-file "/path/to/GoogleIDPMetadata.xml" \
  --saml-name-attribute "displayName" \
  --saml-email-attribute "email" \
  --saml-groups-attribute "groups"
```

### Step 4: Configure Google Workspace SAML App

In the Google Admin Console SAML app configuration:

#### Basic Settings

- **App name**: Terralist
- **Description**: Terraform Module Registry

#### Google Identity Provider details

These are automatically filled from the metadata file you downloaded.

#### Service Provider details

- **ACS URL**: `https://terralist.example.com/v1/api/auth/saml/acs`
- **Entity ID**: `https://terralist.example.com/v1/api/auth/saml/metadata`
- **Start URL**: `https://terralist.example.com` (optional)
- **Signed Response**: Enabled
- **Name ID format**: EMAIL

#### Attribute Mapping

Configure these attribute mappings:

| Google Directory attribute | App attribute |
|----------------------------|---------------|
| Primary email | email |
| Display name | displayName |
| Groups | groups |

!!! note "Google Workspace groups attribute requires additional configuration. You may need to contact Google support to enable group information in SAML assertions."

### Step 5: Enable the App

1. In Google Admin Console, go to the SAML app
2. Click **User access** and enable the app for your users or groups
3. Save the configuration

### Step 6: Test the Configuration

1. Access your Terralist instance
2. Click the login button
3. You should be redirected to Google for authentication
4. After successful authentication, you should be redirected back to Terralist

### Troubleshooting Google Workspace SAML

**Common Issues:**

1. **"App not configured for user"**: Ensure the SAML app is enabled for the user's account in Google Admin Console.
2. **"Invalid SAML response"**: Check the following:
    - ACS URL is correct: `https://terralist.example.com/v1/api/auth/saml/acs`
    - Entity ID matches: `https://terralist.example.com/v1/api/auth/saml/metadata`
    - HTTPS is properly configured
3. **"Groups not working"**: Google Workspace may not include groups in SAML assertions by default. Contact Google support to enable this feature.
4. **Certificate issues**: Ensure Google's IdP certificate is properly loaded. You can extract it from the metadata file or use the direct certificate configuration option.

## Kubernetes Deployment

When deploying Terralist with SAML authentication in Kubernetes environments, consider these additional configuration requirements:

### Request ID Validation

In multi-pod Kubernetes deployments, SAML request ID validation may fail because the in-memory request tracker doesn't persist across pods. To resolve this, disable request ID validation:

```bash
terralist server \
  --oauth-provider saml \
  --saml-idp-metadata-file "/path/to/metadata.xml" \
  --saml-disable-request-id-validation
```

!!! warning "Security Trade-off"
    Disabling request ID validation removes protection against SAML replay attacks. Ensure other security measures are in place, such as:
    - Short assertion validity periods (`--saml-max-assertion-age`)
    - Proper TLS/HTTPS configuration
    - Regular certificate rotation

### Load Balancer Configuration

Ensure your load balancer or ingress controller:

- **HTTPS Termination**: Properly terminates TLS if Terralist isn't handling certificates directly
- **Session Affinity**: Consider sticky sessions for SAML flows (optional, not required with request ID validation disabled)
- **Header Forwarding**: Forward original client IPs for security logging

### Pod Resources

SAML processing can be CPU-intensive due to XML parsing and cryptographic operations. Consider:

- **CPU Requests**: At least 100m CPU per pod
- **Memory Limits**: At least 256Mi per pod
- **Readiness Probes**: Include SAML metadata endpoint checks

### Secrets Management

Store SAML certificates and secrets securely:

- Use Kubernetes secrets for certificates and private keys
- Mount secrets as files rather than environment variables for certificates
- Rotate secrets regularly according to your security policy

## Advanced Configuration

### Security Settings

For enhanced security, consider these configuration options:

```bash
terralist server \
  --saml-assertion-clock-skew "3m" \        # Tighter clock skew tolerance
  --saml-max-assertion-age "30m" \         # Shorter assertion validity
  --saml-allow-idp-initiated false \       # Disable IdP-initiated SSO
  --saml-request-id-expiration "30m"       # Shorter request ID lifetime
```

### Metadata Refresh

For production environments with metadata URLs:

```bash
terralist server \
  --saml-idp-metadata-url "https://idp.example.com/metadata" \
  --saml-metadata-refresh-interval "12h" \
  --saml-metadata-refresh-check-interval "1h"
```

This ensures certificate updates are automatically picked up without service restart.

### Environment Variables

All SAML configuration can also be set via environment variables:

```bash
export TERRALIST_OAUTH_PROVIDER="saml"
export TERRALIST_SAML_IDP_METADATA_URL="https://idp.example.com/metadata"
export TERRALIST_SAML_NAME_ATTRIBUTE="displayName"
export TERRALIST_SAML_EMAIL_ATTRIBUTE="email"
export TERRALIST_SAML_HTTP_CLIENT_TIMEOUT="45s"
export TERRALIST_SAML_ASSERTION_CLOCK_SKEW="5m"
```

## RBAC Integration

SAML integrates with Terralist's RBAC system. Configure user roles based on SAML groups:

```bash
# In your RBAC policy file
g, engineering@company.com, role:admin
g, developers@company.com, role:contributor
g, viewers@company.com, role:readonly
```

The `saml-groups-attribute` configuration determines which SAML attribute contains the group information.

## Troubleshooting

### Common Issues

1. **"SAML authentication failed: invalid SAML response"**:
    - Check ACS URL configuration
    - Verify HTTPS is enabled
    - Ensure IdP metadata is correct and up-to-date
2. **"Certificate validation failed"**:
    - Verify IdP certificate is correctly configured
    - Check certificate format (should be PEM)
    - Ensure certificate hasn't expired
3. **"User attributes not found"**:
    - Check attribute names in SAML assertion
    - Verify attribute mapping configuration
    - Enable SAML response debugging if needed
4. **"Clock skew errors"**:
    - Increase `saml-assertion-clock-skew` value
    - Ensure system clocks are synchronized
5. **"SAML authentication failed: invalid or replayed request ID"** (Kubernetes):
    - This occurs in multi-pod Kubernetes deployments where request state isn't shared
    - Solution: Set `--saml-disable-request-id-validation` or `TERRALIST_SAML_DISABLE_REQUEST_ID_VALIDATION=true`
    - Note: This disables replay attack protection but allows SAML to work in distributed environments

### Debugging

Enable detailed SAML logging by setting log level to debug:

```bash
terralist server --log-level debug
```

When debug logging is enabled, Terralist will log:

- **Extracted Attributes**: All attributes found in the SAML assertion with their values
- **User Groups**: Final groups assigned to the user after attribute mapping

!!! warning "Security Note"
    Debug logs contain sensitive SAML response data. Only enable debug logging temporarily for troubleshooting and disable it in production environments.

Example debug output:

```json
{"level":"debug","saml_attributes":{"email":["user@example.com"],"groups":["developers","admins"]},"attribute_count":2,"message":"Extracted SAML attributes from assertion"}
{"level":"debug","user_groups":["developers","admins"],"groups_count":2,"message":"Final user groups after SAML authentication"}
```

### Testing SAML Configuration

You can test SAML configuration without affecting production:

1. Use a test/staging Terralist instance
2. Configure with test IdP credentials
3. Verify login flow works end-to-end
4. Check that user attributes are correctly extracted
5. Test RBAC policies with SAML groups

## Security Considerations

- **Always use HTTPS**: SAML requires secure TLS/HTTPS transport (can be provided by reverse proxy or application certificates)
- **Validate certificates**: Ensure IdP certificates are valid and up-to-date
- **Monitor assertions**: Regularly review SAML assertion logs for anomalies
- **Limit IdP-initiated SSO**: Keep `saml-allow-idp-initiated` disabled unless required
- **Regular metadata refresh**: Use metadata refresh for automatic certificate updates
- **Short assertion validity**: Use reasonable `saml-max-assertion-age` values

## Support

For additional help with SAML configuration:

- Check the [configuration reference](../configuration.md) for all SAML options
- Review the SAML implementation analysis for technical details
- Contact your IdP administrator for IdP-specific configuration help
- Check Terralist logs for detailed error information
