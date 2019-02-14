# fauxy

Work in progress.

Proxy connections from `:8080` to `localhost:4567` with `tmp/allow_all.json`:

```json
{
    "policies": {
        "allowAll": true
    }
}
```

```console
$ fauxy proxy --from :8888 --to localhost:4567 --config tmp/allow_all.json
...
```

## Request Handling

```plaintext
TCP Connection -> FROM: Fauxy Proxy :TO -> TCP Connection
               ^    ^        |       ^   ^
   Allow/Deny -|    |______Config____|   |
               |_____________|           |
      Timeout -|             |           |- Timeout
               |_____________|___________|
      Hexdump -|             |           |- Hexdump
               |_____________|___________|

```

## Configuration

```json
{
    "from": "192.168.0.2:80",
    "to": "localhost:8080",
    "policies": {
        "allowAll": false,
        "denyAll": false,
        "allow": ["192.168.0.3", "192.168.0.4"],
        "deny": ["192.168.0.5"]
    },
    "hexdump": true,
     "monitor": {
        "bytes_copied": true
    },
    "log": {
        "stdout": true,
        // "file": "fauxy.80.to.8080.log"
    }
}
```