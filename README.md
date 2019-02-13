# fauxy

Work in progress.

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
               |

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
        "deny": ["192.168.0.5"],
        "rateLimit": {
            "second": {
                "3": ["192.168.0.3", "192.168.0.4"]
            }
        }
    },
    "hexdump": true,
    "replace": {
        "*": "hello world!"
    },
    "monitor": {
        "from": true,
        "to": true,
    },
    "logStdout": true,
    "logFile": "fauxy.80.to.8080.log"
}
```