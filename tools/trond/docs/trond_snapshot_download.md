## trond snapshot download

Download target backup snapshot to current directory

### Synopsis

Refer to the snapshot source domain and backup name you input, the available backup snapshot will be downloaded to the local directory.<br>

Note:
- because some snapshot sources have multiple snapshot types, you need to specify the type(full, lite) of snapshot you want to download.<br>
- the snapshot is large, it may need a long time to finish the download, depends on your network performance. You could add 'nohup' to make it continue running even after you log out of your terminal session.


```
trond snapshot download [flags]
```

### Examples

```
# Download target backup snapshot (backup20250205 in 34.143.247.77) to current directory.
$ nohup ./trond snapshot download -d 34.143.247.77 -b backup20250205 -t lite &

```

### Options

```
  -b, --backup string   Backup name(required).
                        Please run command "./trond snapshot list" to get the available backup name under target source domains.
  -d, --domain string   Domain for target snapshot source(required).
                        Please run command "./trond snapshot source" to get the available snapshot source domains.
  -h, --help            help for download
  -t, --type string     Node type of the snapshot(required, available: full, lite).
```

### SEE ALSO

* [trond snapshot](trond_snapshot.md)	 - Commands for getting java-tron node snapshots.
* [trond snapshot download default-main](trond_snapshot_download_default-main.md)	 - Download latest mainnet lite fullnode snapshot from default source to current directory
* [trond snapshot download default-nile](trond_snapshot_download_default-nile.md)	 - Download latest nile testnet lite fullnode snapshot from default source to local current directory
