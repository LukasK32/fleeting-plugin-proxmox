# Fleeting plugin for Proxmox Virtual Environment

This is a [fleeting](https://gitlab.com/gitlab-org/fleeting/fleeting) plugin for [Proxmox Virtual Environment](https://www.proxmox.com/en/proxmox-virtual-environment/overview).

## Configuration
### Plugin settings

| Parameter                    | Type   | Default value       | Description                                              |
| ---------------------------- | ------ | ------------------- | -------------------------------------------------------- |
| `url`                        | string | N/A (required)      | Proxmox VE URL.                                          |
| `insecure_skip_tls_verify`   | bool   | `false`             | If `true` then TLS certificate verification is disabled. |
| `credentials_file_path`      | string | N/A (required)      | Path to Proxmox VE credentials file.                     |
| `pool`                       | string | N/A (required)      | Name of the Proxmox VE pool to use.                      |
| `storage`                    | string | N/A (required)      | Name of the Proxmox VE storage to use.                   |
| `template_id`                | int    | N/A (required)      | ID of the Proxmox VE VM to create instances from.        |
| `max_instances`              | int    | N/A (required)      | Maximum instances than can be deployed.                  |
| `instance_network_interface` | string | `ens18`             | Network interface to read instance's IPv4 address from.  |
| `instance_name_creating`     | string | `fleeting-creating` | Name to set for instances during creation.               |
| `instance_name_running`      | string | `fleeting-running`  | Name to set for running instances.                       |
| `instance_name_removing`     | string | `fleeting-removing` | Name to set for instances during removal.                |

### Credentials file

<!-- TODO: Document `path` and `privs`  -->
| Parameter  | Type   | Description               |
| ---------- | ------ | ------------------------- |
| `realm`    | string | Authentication Realm      |
| `username` | string | User name                 |
| `password` | string | User password             |
| `otp`      | string | One-time password for 2FA |

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
