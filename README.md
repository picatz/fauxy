# fauxy

Proxy local connections from port `8888` to `7777`:

```console
$ fauxy proxy --from 127.0.0.1:8888 --to 127.0.0.1:7777
...
```

Proxy local connections from port `8888` to `7777`, with the following config `config.json`:

```json
{
    "allow": ["*"],
    "deny": ["192.168.0.2"]
}
```

```console
$ fauxy proxy --from 127.0.0.1:8888 --to 127.0.0.1:7777 --confg config.json
...
```