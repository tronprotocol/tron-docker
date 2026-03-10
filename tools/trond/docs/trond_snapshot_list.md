## trond snapshot list

List available snapshots of target source.

### Synopsis

Refer to the snapshot source domain you input, the available backup snapshots will be showen below.<br>
Note: different domain may have different snapshots that can be downloaded.


```
trond snapshot list [flags]
```

### Examples

```
# List available snapshots of target source domain 34.143.247.77
$ ./trond snapshot list -d 34.143.247.77

```

### Options

```
  -d, --domain string   Domain for target snapshot source (required)
                        Please run command "./trond snapshot source" to get the available snapshot source domains
  -h, --help            help for list
```

### SEE ALSO

* [trond snapshot](trond_snapshot.md)	 - Commands for getting java-tron node snapshots.
