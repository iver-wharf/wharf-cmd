# Provisioner Git SSH secret

The wharf-cmd-provisioner (`wharf provisioner` subcommands) creates a
wharf-cmd-worker (`wharf run` subcommand) with a Git cloning command to
bootstrap the wharf-cmd-worker.

## Prerequisites

- `known_hosts` file.
- Password-less SSH private key, e.g a `id_rsa` or `id_ed25519` file.
- `config`, a SSH config file ([ssh_config(5) manual](https://linux.die.net/man/5/ssh_config))

The secret is mounted into `~/.ssh` to be automatically used by Git for SSH
remotes.

### Prerequisites: known_hosts

For security reasons, Wharf will not update the `known_hosts` file
automatically, as there's no safe way to do this from only SSH.

The following commands will fetch the public SSH key from GitHub, store it in
`known_hosts` inside your current working directory, and then print out the
fingerprint to the console:

```console
$ ssh-keyscan -t rsa github.com | tee -a known_hosts | ssh-keygen -lf -
# github.com:22 SSH-2.0-babeld-98453d8a
2048 SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8 github.com (RSA)

$ ssh-keyscan -t rsa gitlab.com | tee -a known_hosts | ssh-keygen -lf -
# gitlab.com:22 SSH-2.0-OpenSSH_8.4p1 Debian-5
2048 SHA256:ROQFvPThGrW4RuWLoL9tq9I9zJ42fK4XywyRtbOz/EQ gitlab.com (RSA)
```

You can run the commands multiple times but with different hostname, and it will
keep adding keys to the existing `known_hosts` file.

Run this command for all the hosts you want wharf-cmd to clone repos from, and
compare the fingerprints with one given by the hosting provider themselves.
For example, here you can find some fingerprints of publicly known forges:

- Azure DevOps: `https://dev.azure.com/${YOUR ORG NAME}/_usersSettings/keys`
- GitHub: <https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/githubs-ssh-key-fingerprints>
- GitLab: <https://gitlab.com/help/instance_configuration>

### Prerequisites: SSH private key

Generate an SSH key without a password and add it to an account, preferably to a
"service account" (an account that's not meant to be used by humans, but only
by other services, that way the account isn't tied to an employee).

:warning: Note: The SSH key must be password-less, meaning it has an empty
password to be able to be used. This is because Wharf does not support supplying
an SSH password.

For GitHub, you would add the SSH key at <https://github.com/settings/ssh/new>.

The file looks something like this:

```rsa
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEA2vQcXP/6o7BVPms+P7mlzjHK6Xe2DqjQh8lLzgmnUNhZeBIqjoQJ
// snip
n9reK9kUsm/uLW8moqRvt2ED9mmfKBkId//NwU2jeCP0c75TzyAiEEtL5PcTITvdnAwhuj
7kA9cVuIvaoMG0cAAAAAAQID
-----END OPENSSH PRIVATE KEY-----
```

### Prerequisites: ssh_config

To set a key with name `id_rsa` to be used on all SSH connections, write the
following `ssh_config` file:

```ssh-config
Host *
    IdentityFile ~/.ssh/id_rsa
```

To use a different key with the name `id_github_rsa` on only `github.com`, you
can write the following `ssh_config` file:

```ssh-config
Host github.com
    IdentityFile ~/.ssh/id_github_rsa

Host *
    IdentityFile ~/.ssh/id_rsa
```

## Kubernetes

A secret named `wharf-cmd-worker-git-ssh` is expected to live in the same
Kubernetes cluster and namespace as where the wharf-cmd-worker pods are
created.

```sh
kubectl create secret generic wharf-cmd-worker-git-ssh \
  --from-file=id_rsa=./id_rsa  \
  --from-file=known_hosts=./known_hosts \
  --from-file=config=./ssh_config
```
