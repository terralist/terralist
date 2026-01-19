# RBAC Configuration

The RBAC feature enables restrictions of access to Terralist resources. Terralist does not have its own user management system, delegating this job to one (or more) OAuth 2.0 providers. If the provider authenticates the user, Terralist asks the provider for some metadata and takes the user as being authenticated under those claims. Depending on the provider implementation, those claims can differ.

There are two main components where RBAC configuration can be defined:

- The server-side (global) RBAC configuration;
- The API Key RBAC configuration;

## Basic Built-in Roles

Terralist has three pre-defined roles. Not all of them support expansion, but you are free to define new roles as you please (see below).

- `role:anonymous`: has access to no resources (unless specified otherwise in the server-side configuration);
- `role:readonly`<sup>*</sup>: read-only access to all resources;
- `role:admin`<sup>*</sup>: unrestricted access to all resources;

<sup>*</sup> This role cannot be extended.

The `role:anonymous` is a special role that is assigned to unauthenticated users. This role can be customized from the server-side configuration and through those modifications users are able to expose (publicly) resources from the registry. By default, this role has no grant attached.

## Default Policy for Authenticated Users

When a user is authenticated in Terralist, it will be granted the role specified by the `rbac-default-role` configuration option, if there is no other role specified for the given user.

## RBAC Model Structure

The model syntax is based on [Casbin](https://casbin.org/docs/overview) and highly inspired from the [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) implementation. There are two different types of syntax: one of assigning policies, and another one for assigning users to internal roles.

**Group**: Used to assign users or groups to internal roles.

Syntax: `g, <username/useremail/group>, <role>`

- `<username/useremail/group>`: The entity to whom the role will be assigned. Depending on the OAuth provider implementation those values can represent different things; Usually, the `username` refers to the `sub` claim, while the `useremail` and `group` refers to a custom claims, which might not even be supported by the provider you are using. Check the OAuth provider documentation for more details.
- `<role>`: The internal role to which the entity will be assigned.

<!-- TODO: Add proper oauth provider docs -->

Below is a table that defines claims meaning for each OAuth provider.

| Provider\Claim | `username`  | `useremail`    | `group`                                                                                                      |
| -------------- | ----------- | -------------- | ------------------------------------------------------------------------------------------------------------ |
| BitBucket      | Username    | User E-mail    | Not supported.                                                                                               |
| GitHub         | Username    | User E-mail    | GitHub Organization Teams slugs that the user is part of (if `gh-organization` configuration option is set). |
| GitLab         | Username    | User E-mail    | GitLab User Group names.                                                                                     |
| OIDC           | `sub` claim | Not supported. | Not supported.                                                                                               |

**Policy**: Allows to assign permissions to an entity.

Syntax: `p, <role/username/useremail/group>, <resource>, <action>, <object>, <effect>`

- `<role/username/useremail/group>`: The entity to whom the policy will be assigned
- `<resource>`<sup>*</sup>: The type of resource on which the action is performed. Can be one of: `modules`, `providers`, `authorities`. Supports glob matching (e.g. )
- `<action>`<sup>*</sup>: The operation that is being performed on the resource. Can be one of: `get`, `create`, `update`, `delete`. Supports glob matching.
- `<object>`<sup>*</sup>: The object identifier representing the resource on which the action is performed. Supports glob matching. Depending on the resource, the object's format will vary. 
- `<effect>`: Whether this policy should grant or restrict the operation on the target object. One of `allow` or `deny`.

<sup>*</sup> This attribute supports glob matching. For example, for resources `*` will match all 3 resources, `mod*` will match only `modules`, while for objects `my-authority/my-module/aws` will match only one module, while `my-authority/*/*` will match all modules within the authority `my-authority`.

Below is a table that defines the correct object syntax for each resource group.

| Resource Group | Object Syntax                                    |
| -------------- | ------------------------------------------------ |
| `authorities`  | `<authority-name>`                               |
| `modules`      | `<authority-name>/<module-name>/<provider-name>` |
| `providers`    | `<authority-name>/<provider-name>`               |

 For example, an object c

## API Key Authority Isolation

When using API keys for authentication, Terralist enforces strict authority isolation:

- **API keys are bound to their issuing authority**: An API key can only access modules and providers belonging to the authority that issued the key.
- **Cross-authority access is denied**: API keys from one authority cannot access resources from other authorities.
- **Case-insensitive matching**: Authority names are compared case-insensitively.

This is a security boundary that prevents API key privilege escalation in multi-tenant environments.

**Example:**
- API key issued by authority `my-org` can access `my-org/my-module/aws`
- API key issued by authority `my-org` **cannot** access `other-org/their-module/aws`

Note: This isolation only applies to API key authentication. Users authenticated via OAuth session can access resources based on their RBAC policies.
