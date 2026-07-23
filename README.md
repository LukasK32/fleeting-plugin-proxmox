# Fleeting plugin for Proxmox Virtual Environment

This is a [fleeting](https://gitlab.com/gitlab-org/fleeting/fleeting) plugin for [Proxmox Virtual Environment](https://www.proxmox.com/en/proxmox-virtual-environment/overview).

## Installation

See [Releases](https://github.com/LukasK32/fleeting-plugin-proxmox/releases) for available versions and installation instructions.

## Configuration

### Plugin settings

| Parameter                    | Type                      | Default value                      | Description                                                                                  |
| ---------------------------- | ------------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------- |
| `url`                        | string                    | N/A (required)                     | Proxmox VE URL.                                                                              |
| `insecure_skip_tls_verify`   | bool                      | `false`                            | If `true` then TLS certificate verification is disabled.                                     |
| `credentials_file_path`      | string                    | N/A (required)                     | Path to Proxmox VE credentials file.                                                         |
| `pool`                       | string                    | N/A (required)                     | Name of the Proxmox VE pool to use.                                                          |
| `storage`                    | string                    | N/A (required if template is a VM) | Name of the Proxmox VE storage to use.                                                       |
| `template_id`                | int                       | N/A (required)                     | ID of the Proxmox VE VM to create instances from.                                            |
| `max_instances`              | int                       | N/A (required)                     | Maximum instances than can be deployed.                                                      |
| `instance_network_interface` | string                    | `ens18`                            | Network interface to read instance's IPv4 address from.                                      |
| `instance_network_protocol`  | `any` or `ipv4` or `ipv6` | `ipv4`                             | Network protocol to look for when discovering instance's IP address. `any` prioritizes IPv6. |
| `instance_name_creating`     | string                    | `fleeting-creating`                | Name to set for instances during creation.                                                   |
| `instance_name_running`      | string                    | `fleeting-running`                 | Name to set for running instances.                                                           |
| `instance_name_removing`     | string                    | `fleeting-removing`                | Name to set for instances during removal.                                                    |

### Credentials file

The credentials file selects one of two mutually exclusive authentication
methods.

**API token (recommended).** Tokens can be scoped to the plugin's pool,
don't expire, and need no session refresh, so no privileged password is
held in memory or on disk. Create one under the dedicated user (see
[Proxmox configuration](#proxmox-configuration)) with privilege separation
enabled, and grant it the same roles as the user.

| Parameter      | Type   | Description                                           |
| -------------- | ------ | ----------------------------------------------------- |
| `token_id`     | string | Full token identifier, in `user@realm!tokenname` form |
| `token_secret` | string | The token's secret value                              |

**Username/password (legacy).** Drives an hourly session-ticket refresh.

<!-- TODO: Document `path` and `privs`  -->
| Parameter  | Type   | Description               |
| ---------- | ------ | ------------------------- |
| `realm`    | string | Authentication Realm      |
| `username` | string | User name                 |
| `password` | string | User password             |
| `otp`      | string | One-time password for 2FA |

`credentials.example.json` shows the API-token form. For username/password authentication, replace its contents with the legacy fields listed above.

### Template VM configuration

The template must be a bootable VM with enabled DHCP and QEMU guest agent installed. See [Proxmox documentation](https://pve.proxmox.com/wiki/Qemu-guest-agent) for more details.

### Proxmox configuration

You **MUST** create a **DEDICATED** user, pool and storage for usage with this plugin. Any other configuration is untested and unsupported.

After creating a **DEDICATED** user, pool and storage follow procedure below to add required permissions:

1. Add template VM as a member to the pool.
2. Add storage as a member to the pool.
3. Add following roles for the user to the pool:
   * `PVEVMAdmin`,
   * `PVEPoolUser`,
   * `PVEDatastoreUser`.
4. Add following role for the user to the network that you will use for deployed VMs:
    * `PVESDNAdmin`.
5. Add following role for the user to the node with the storage, network, template etc.:
    * `PVEAuditor` without propagation.

If you are using **API token authentication** (recommended), also perform the following:

1. Create an API token for the dedicated user (*Datacenter → Permissions → API Tokens → Add*). Leave **Privilege Separation** enabled so the token only ever has the permissions you grant it explicitly.
2. Grant the token the same roles on the same paths as in steps 3–5 above. A privilege-separated token starts with no permissions of its own, so it needs its own role assignments in addition to the user's. (Alternatively, disable Privilege Separation when creating the token so it inherits the user's permissions — less isolated, but no separate role assignments are needed.)

## Development

### Integration tests

1. Create and fill out `credentials.json` file in project root (see [Credentials file](#credentials-file) for details)
    ```bash
    cp credentials.example.json credentials.json
    ```
2. Create and fill out `config.json` file in project root (see [Plugin settings](#plugin-settings) for details)
    ```bash
    cp config.example.json config.json
    ```
3. Run integration tests:
    ```bash
    make integration-test
    ```
