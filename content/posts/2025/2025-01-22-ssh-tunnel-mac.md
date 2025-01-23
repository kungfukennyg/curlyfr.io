---
title: "SSH tunnel as a VPN on Mac OS X"
date: 2025-01-22T22:49:18-05:00
draft: false
toc: true
images:
tags:
  - ssh
  - ssh tunnel
  - mac
  - socks
  - proxy
---

This post describes how to setup a local SSH tunnel to a remote server and configure Mac OS X to use that tunnel as a SOCKS proxy. Tested on Mac OS Sequoia 15.2.

## Prerequisites
- Server with SSH access

## Steps

1. Find the network adapter you want to configure the proxy on (I'm on `Wi-Fi`):

    ```shell
    curly@fry:~$ networksetup -listallnetworkservices
    Wi-Fi
    Thunderbolt Bridge
    iPhone USB
    Bluetooth PAN
    ```

2. Configure the adapter to proxy requests through a local port (`1080`):

    ```shell
    curly@fry:~$ networksetup -setsocksfirewallproxy "Wi-Fi" 127.0.0.1 1080
    ```

3. Create an SSH tunnel from chosen local port to our SSH server (named `curlyfr.io` here):

    ```shell
    curly@fry:~$ ssh -N -D 1080 curlyfr.io -p 22
    ```

    Note that this command does not execute in a new process, meaning the shell you run it in will hold the proxy open until you `Ctrl + C`. If you'd rather it run in the background, you can pass the `-f` option to `ssh`.

4. Confirm that we can make connections through our tunnel and that our public-facing IP matches our SSH server:

    ```shell
    curly@fry:~$ curl -x 127.0.0.1:1080 ipinfo.io/ip
    <SSH server IP>
    ```

    You can disable the proxy by terminating the shell holding the SSH tunnel open, and turning off the SOCKS proxy on your network adapter: 

    ```shell
    curly@fry:~$ networksetup -setsocksfirewallproxystate "Wi-Fi" off
    ```

## Script

I packaged this behavior into a simple [fish](https://fishshell.com/) script that configures the proxy, sets some environment variables, and cleans up on exit:

```shell
#!/usr/local/bin/fish

# cleanup turns off the SOCKS proxy and unsets env variables
function cleanup --on-signal INT
    networksetup -setsocksfirewallproxystate "Wi-Fi" off
    set -e HTTP_PROXY
    set -e HTTPS_PROXY
    echo "done"
end

if test (count $argv) -lt 4
    echo "USAGE:
    ./sock.fish 127.0.0.1 1080 Wi-Fi some_ssh_server

DESCRIPTION:
    Configures Mac OS SOCKS proxy config to point to a local port, and creates and binds an SSH Tunnel to that port. Also sets HTTP_PROXY and HTTPS_PROXY. Disables the SOCKS proxy on exit.
"
end

set bind_addr $argv[1]
set port $argv[2]
set network_adapter $argv[3]
set ssh_tunnel_target $argv[4]

networksetup -setsocksfirewallproxy "$network_adapter" "$bind_addr" "$port"
# turn on
networksetup -setsocksfirewallproxystate "Wi-Fi" on
export HTTP_PROXY="$bind_addr:$port"
export HTTPS_PROXY="$bind_addr:$port"

# create ssh tunnel
echo "sock.fish: tunnel $bind_addr:$port->$ssh_tunnel_target on adapter $network_adapter, Ctrl+C to exit and cleanup..."
ssh -N -D "$port" "$ssh_tunnel_target" -p 22
```

You can accomplish the same thing in a bash script with `trap`:

```shell
#!/bin/bash

# cleanup turns off the SOCKS proxy and unsets env variables
cleanup () {
    networksetup -setsocksfirewallproxystate "Wi-Fi" off
    unset HTTP_PROXY
    unset HTTPS_PROXY
    echo "done"
}

bind_addr="$1"
port="$2"
network_adapter="$3"
ssh_tunnel_target="$4"

if [[ -z "$bind_addr" || -z "$port" || -z "$ssh_tunnel_target" 
|| -z "$network_adapter" ]];
then

cat <<EOL
USAGE:
    ./sock.sh 127.0.0.1 1080 Wi-Fi some_ssh_server

DESCRIPTION:
    Configures Mac OS SOCKS proxy config to point to a local port, and creates and binds an SSH Tunnel to that port. Also sets HTTP_PROXY and HTTPS_PROXY. Disables the SOCKS proxy on exit. 
EOL
fi

networksetup -setsocksfirewallproxy "$network_adapter" "$bind_addr" "$port"
# turn on
networksetup -setsocksfirewallproxystate "Wi-Fi" on
export HTTP_PROXY="$bind_addr:$port"
export HTTPS_PROXY="$bind_addr:$port"

trap 'cleanup' INT
# create ssh tunnel
echo "Creating tunnel $bind_addr:$port->$ssh_tunnel_target on adapter $network_adapter, Ctrl+C to exit and cleanup..."
ssh -N -D "$port" "$ssh_tunnel_target" -p 22
```