# ssh_force_command

[![][license-svg]][license-url]

## Considerations

[*ssh*](https://en.wikipedia.org/wiki/Secure_Shell) is an angular stone in *nix administration.

**Allowing remote access while restricting the actions** permitted in the server is a very common scenario, and  `force command`  is just the *ingredient* needed for this purpose.

Requires _Public key authentication_:

1. > The file **~/.ssh/authorized_keys** lists the public keys that are permitted for logging in.  When the user logs in, the ssh program tells the server which key pair it would like to use for authentication.  The client proves that it has access to the private key and the server checks that the corresponding public key is authorized to accept the account.

2. > **ForceCommand** Forces the execution of the command specified in ~/.ssh/authorized_key , ignoring any command supplied by the client and ~/.ssh/rc if present. The command is invoked by using the user's login shell with the -c option. This applies to shell, command, or subsystem execution. It is most useful inside a Match block.

3. > **SSH_ORIGINAL_COMMAND**  This variable contains the original command line if a forced command is executed.  It can be used to extract the original arguments.

## Purpose

This software is based in two elements, and **executable** and a **config file** written in [*yaml*](https://en.wikipedia.org/wiki/YAML).

The binary should be set as the forced command in `~/.ssh/authorized_keys`,  and `~/.ssh/authorized_forced_commands.yml` must store the allowed actions for this particular public. key  

Goals:

1. Improve external access **safety**.

  The launcher is compiled and the configuration file is checked using the same security patterns as the rest of the ssh machinery.

2. Simplify the **management** of allowed commands.

  By standardize the configuration yaml file, way more easy and flexible than a random piped separated text file  

## Installation

Obviously you need [`go`](https://golang.org/doc/install) installed in your machine.

Then just:

```bash
go get -v github.com/klashxx/ssh_force_command
````

And the executable will be compiled and placed in your `$GOPATH/bin` directory.

## Configuration

A config file `authorized_forced_commands.yml`  must be placed in the `~/.ssh` directory.

Safety rules:

- Should not be accesible by *others* group.
- Owner must be ssh user.
- Group should be ssh user group.

It's written in `yaml` and the format is pretty self explanatory:

```yaml
tag: my_tag
commands:
  - path: command1
    description: my first desc
    env: null
  - path: /path/to/command2 arg1 arg2
    description: my second desc
    env:
      - VAR1=/var1/value
      - VAR2=value2
```

**NOTE**:  `ssh_force_command` uses the current process's environment, if  `env` is  *NOT* null listed variables will be appended before execution 

## Example

### In the remote box

1. Place `ssh_force_command` binary **and** this *test script* in your **HOME** dir:

```bash
#!/bin/bash
echo "just a simple test"
echo "parameters: $@"
echo "VAR1: ${VAR1:-not_set}"
echo "VAR2: ${VAR2:-not_set}"
exit 0
```

2. Create the configuration file `~/.ssh/authorized_forced_commands.yml` with the appropriate permissions:

```yaml
tag: test
commands:
    - path: ~/test_ssh_force_command.sh
      description: very dummy test
      env: null
```

3. Set the forced command for the corresponding key in  `~/.ssh/authorized_keys`, example:

```text
command="/home/user/ssh_force_command",no-pty ssh-rsa ZZZZB3NzaC1yc2EAAAABIwAAAQEAqxekXWvfwc74bSZxyzTxPpWaogaeMCKlXE8tgEAN/jS8+28x2h/PGzI4ij9H3aZHLayjL7PY1Uj3SETG913+NOTGONNAWORK+r9vPzyRwbJLh3dkbvYdsC0drbsqIN+3K7mGIT8U/Aw9i5oZpNZ/mpEO+dT2ymMLvLJL+sizNK7Aw10x1YWOBTEVKf6C5E/dtmWYWKyx14tpBxlh6wxiofb2hDO9i6TU/N3PKNZ/xToIDTGMpOO9mbPT6v3DRof0fIgBF3rPNaIPLUWKuwjmP4JbAiP76L93DM+Mwhc1cw7H6+oOljpTSRxmTQi20iohqVQonAhlY1w== dummy@server.int
```

### In the local machine

Just execute the ssh command:

```bash
ssh user@remote_server "~/test_ssh_force_command.sh arg1 arg2"
```

The output should be:

```text
just a simple test
parameters: arg1 arg2
VAR1: not_set
VAR2: not_set
```

### Back to the remote box

Let's add some *env* variables to the config file, and another command:

```yaml
tag: test
commands:
  - path: ~/test_ssh_force_command.sh arg1 arg2
    description: adding env vars
    env:
      - VAR1=/var1/value
      - VAR2=foo
  - path: ls
    description: a simple ls
    env: null
```

Now, **from the local machine**, the ssh execution output must be:

```text
just a simple test
parameters: arg1 arg2
VAR1: /var1/value
VAR2: foo
```

And you should be able to list the content of any remote dir where the exec user has permissions.

[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://opensource.org/licenses/MIT
