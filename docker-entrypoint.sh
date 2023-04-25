#!/usr/bin/dumb-init /bin/sh
set -e

# Modified: https://github.com/runatlantis/atlantis/blob/bbb0ed2f0041844dc4abfddef2d1fe2f25340249/docker-entrypoint.sh

# If arguments are received directly, then pass them to the Terralist command
if [ "$(echo "${1}" | cut -c1)" = "-" ]; then
    set -- terralist "$@"
fi

# If a command is received directly, we should identify if it is a Terralist
# sub-command. To do that, pass it to the `terralist help` command and check
# if we can find the substring in the result (because terralist help 
# subcommand always exit with code 0, even if the subcommand does not exist)
if terralist help "$1" 2>&1 | grep -q "terralist $1"; then
    set -- terralist "$@"
fi

# If the current uid running does not have a user create one in /etc/passwd
if ! whoami > /dev/null 2>&1; then
  if [ -w /etc/passwd ]; then
    echo "${USER_NAME:-default}:x:$(id -u):0:${USER_NAME:-default} user:/home/terralist:/sbin/nologin" >> /etc/passwd
  fi
fi

# If we're running as root and we're trying to execute terralist then we use
# su-exec to step down from root and run as the terralist user.
if [ "$(id -u)" = 0 ] && [ "$1" = 'terralist' ]; then
    # If requested, set the capability to bind to privileged ports before
    # we drop to the non-root user. Note that this doesn't work with all
    # storage drivers (it won't work with AUFS).
    if [ -n "${TERRALIST_ALLOW_PRIVILEGED_PORTS+x}" ]; then
        setcap "cap_net_bind_service=+ep" /usr/local/bin/terralist
    fi

    set -- su-exec terralist "$@"
fi

exec "$@"