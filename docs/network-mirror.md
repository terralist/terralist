# Network Mirror Protocol

Terralist supports the Terraform **Network Mirror Protocol** for providers, offering a simpler alternative to the standard Registry Protocol configuration.

## Overview

The Network Mirror Protocol provides:

- **Simpler configuration** - Less setup required in `.terraformrc`
- **Optional authentication** - Can work without authentication in isolated environments
- **Dual protocol support** - Works alongside the existing Registry Protocol
- **Provider mirroring** - Compatible with `terraform providers mirror` command
- **Provider-only support** - Note: Terraform's Network Mirror Protocol is currently designed for providers only, not modules

## Provider Network Mirror

### Benefits

- Reduced configuration complexity compared to host overrides
- Direct compatibility with Terraform's `providers mirror` command
- Ability to mirror providers from public registries

### Configuration

Configure Terraform to use Terralist as a network mirror for providers:

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url = "https://terralist.example.com/"
    include = ["registry.terraform.io/your-namespace/*"]
  }

  direct {
    exclude = ["registry.terraform.io/your-namespace/*"]
  }
}

credentials "terralist.example.com" {
  token = "x-api-key:YOUR_API_KEY"
}
```

### URL Format

The Network Mirror Protocol uses the following URL pattern:

```
{base_url}/{hostname}/{namespace}/{type}/...
```

**Example endpoints:**

- **List versions**: `https://terralist.example.com/registry.terraform.io/mycompany/aws/index.json`
- **Version details**: `https://terralist.example.com/registry.terraform.io/mycompany/aws/1.0.0.json`

### Response Format

#### List Provider Versions

**Endpoint**: `/{hostname}/{namespace}/{type}/index.json`

```json
{
  "versions": {
    "1.0.0": {},
    "1.0.1": {},
    "1.1.0": {}
  }
}
```

#### Get Version Details

**Endpoint**: `/{hostname}/{namespace}/{type}/{version}.json`

```json
{
  "archives": {
    "darwin_arm64": {
      "url": "https://example.com/provider.zip",
      "hashes": [
        "h1:abc123...",
        "zh:def456..."
      ]
    },
    "linux_amd64": {
      "url": "https://example.com/provider-linux.zip",
      "hashes": ["h1:xyz789..."]
    }
  }
}
```

### Using Providers with Network Mirror

Once configured, use providers normally in your Terraform configuration:

```hcl
terraform {
  required_providers {
    myapp = {
      source  = "registry.terraform.io/mycompany/myapp"
      version = "1.0.0"
    }
  }
}
```

Terraform will automatically fetch the provider from your Terralist network mirror.

### Mirroring Public Providers

You can mirror providers from the public registry to your Terralist instance:

```bash
# Mirror a provider using terraform providers mirror
terraform providers mirror .

# Upload the mirrored provider to Terralist
curl -X POST https://terralist.example.com/v1/api/providers/aws/5.0.0/upload \
     -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
     -d @provider-manifest.json
```

## Module Support

!!! warning "Not Supported by Terraform"
    Terraform's Network Mirror Protocol **only supports providers**, not modules. This is a limitation of the Terraform CLI itself.

    **Current Status (September 2025)**: Module Network Mirror Protocol is not implemented in Terraform. Track progress on [GitHub Issue #35892](https://github.com/hashicorp/terraform/issues/35892).

    Terralist continues to support modules through the standard **Module Registry Protocol** as documented in the [Getting Started](getting-started.md#upload-a-new-module) guide.

## Comparison: Network Mirror vs Registry Protocol

| Feature | Network Mirror | Registry Protocol |
|---------|---------------|------------------|
| Configuration complexity | Lower | Higher |
| Authentication | Optional | Required |
| Static JSON responses | Yes | No |
| `terraform providers mirror` support | Yes | No |
| Host override required | No | Yes |
| Terraform version required | 0.13.2+ | 0.13+ |

## Using Both Protocols

Terralist supports both protocols simultaneously. Choose based on your needs:

### Use Network Mirror when:

- You want simpler client configuration
- You're using `terraform providers mirror`
- You need offline/air-gapped environments
- Authentication is optional

### Use Registry Protocol when:

- You need dynamic provider/module discovery
- You require advanced authorization controls
- You're integrating with existing Terraform workflows
- You need fine-grained access control

## Configuration Examples

### Network Mirror Only

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url = "https://terralist.example.com/"
  }
}

credentials "terralist.example.com" {
  token = "x-api-key:YOUR_API_KEY"
}
```

### Mixed Configuration

Use network mirror for your company's providers, direct access for others:

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url = "https://terralist.example.com/"
    include = ["registry.terraform.io/mycompany/*"]
  }

  direct {
    exclude = ["registry.terraform.io/mycompany/*"]
  }
}

credentials "terralist.example.com" {
  token = "x-api-key:YOUR_API_KEY"
}
```

### Air-Gapped Environment

For completely offline environments, configure all providers through the network mirror:

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url = "https://terralist.internal/"
  }
}

# Authentication may be optional in isolated environments
credentials "terralist.internal" {
  token = "x-api-key:YOUR_API_KEY"
}
```

## Authentication

Network mirrors support the same authentication as the Registry Protocol:

- **API Keys**: `x-api-key:YOUR_API_KEY`
- **OAuth Tokens**: Standard bearer tokens
- **Anonymous Access**: Available when `providers-anonymous-read` is enabled

Configure anonymous access for isolated environments:

```yaml title="config.yaml"
providers-anonymous-read: true
```

## Troubleshooting

### Error: Provider not found

Ensure the provider is uploaded to Terralist and the namespace matches your configuration:

```bash
# Check if provider exists
curl https://terralist.example.com/registry.terraform.io/mycompany/myapp/index.json
```

### Error: Invalid hash

The provider's SHA256 hash doesn't match. Verify:

1. The provider binary is correctly uploaded
2. The hash in the manifest matches the actual file
3. The hash has the correct `h1:` prefix

### Error: Network mirror URL not accessible

Check:

1. Terralist is running and accessible
2. Network connectivity from your machine
3. HTTPS is properly configured (required for Terraform)
4. Credentials are correctly configured

## Security Considerations

When using Network Mirror Protocol:

- **HTTPS Required**: Terraform requires HTTPS for network mirrors
- **Authentication**: While optional, authentication is recommended for production
- **Hostname Validation**: Terralist validates hostname parameters to prevent injection
- **Rate Limiting**: Same rate limits apply as Registry Protocol
- **Hash Verification**: All providers must include valid SHA256 hashes

## Migration Guide

### From Registry Protocol to Network Mirror

1. **Keep existing setup** - Both protocols work simultaneously
2. **Update `.terraformrc`**:
   ```hcl
   # Add network mirror configuration
   provider_installation {
     network_mirror {
       url = "https://terralist.example.com/"
     }

     # Keep existing host override for transition
     direct {
       include = ["terralist.example.com/*/*"]
     }
   }
   ```
3. **Test with one provider** - Verify network mirror works
4. **Gradually migrate** - Move providers one namespace at a time
5. **Remove host override** - Once all providers are migrated

## References

- [Terraform Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol)
- [Terraform Provider Registry Protocol](https://developer.hashicorp.com/terraform/internals/provider-registry-protocol)
- [Provider Installation Methods](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation)
- [Terraform CLI Configuration](https://developer.hashicorp.com/terraform/cli/config/config-file)