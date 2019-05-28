# provenlogs

Prove logs on Blockchain with ProvenDB

![ProvenLogs Architecture](docs/architecture.png)

## Usage

### Run ProvenLogs Service

You can run `provenlogs` to continuously prove the logs from an app. Currently, [Zap](https://github.com/uber-go/zap) production log format is supported.

1. Build `provenlogs`: `make`
2. Initialize a RSA key pair in current working directory (optional, you can provide your own with `-i` option flag): `./provenlogs init`
3. Build the demo app `zapper`: `make build-zapper`
4. Run all together:

   ```bash
   ./zapper 2>&1 | ./provenlogs -u mongodb://yourusername:yourpassword@yourservicename.provendb.io/yourservicename?ssl=true --batch.size 10 --batch.time 10s
   ```

   Then you will get some output like this:

   ```bash
   {"level":"info","ts":1559002658.639124,"caller":"zapper/zapper.go:29","msg":"test 0","delay":1201}
   {"level":"info","ts":1559002659.844852,"caller":"zapper/zapper.go:29","msg":"test 1","delay":1277}
   {"level":"info","ts":1559002661.123698,"caller":"zapper/zapper.go:29","msg":"test 2","delay":1367}
   {"level":"info","ts":1559002662.494548,"caller":"zapper/zapper.go:29","msg":"test 3","delay":1599}
   {"level":"info","ts":1559002664.0970318,"caller":"zapper/zapper.go:29","msg":"test 4","delay":511}
   {"level":"info","ts":1559002664.608268,"caller":"zapper/zapper.go:29","msg":"test 5","delay":1088}
   2019-05-28T10:17:44.608+1000    DEBUG   provenlogs/batcher.go:147       finalizing batch
   2019-05-28T10:17:44.609+1000    DEBUG   provenlogs/batcher.go:149       sign batch
   2019-05-28T10:17:44.612+1000    DEBUG   provenlogs/batcher.go:155       submit to ProvenDB
   {"level":"info","ts":1559002665.698405,"caller":"zapper/zapper.go:29","msg":"test 6","delay":745}
   2019-05-28T10:17:47.779+1000    DEBUG   provenlogs/batcher.go:162       finalized batch
   {"level":"info","ts":1559002666.444567,"caller":"zapper/zapper.go:29","msg":"test 7","delay":1600}
   {"level":"info","ts":1559002668.045363,"caller":"zapper/zapper.go:29","msg":"test 8","delay":896}
   ```

   Please note that `provenlogs` will echo back the log output from the source app (`zapper`) immediately upon receiving them, though some of them are still being queued up to be proven with ProvenDB. Therefore, if you use `ctrl-c` to kill the pipeline, which will in turn kill both `zapper` and `provenlogs`, some latest logs might be lost and not be proven. In order to prevent this, only kill the `zapper` process using `kill`, then `provenlogs` will make sure all the logs from `zapper` are proven before terminating.

### Verify a Log Entry

Suppose the raw log entry we are interested and are going to be verified is one similar to the logs from the above demo output:

```bash
{"level":"info","ts":1558935983.776685,"caller":"zapper/zapper.go:29","msg":"test 6","delay":745}
```

Then we can simply verify it using:

```bash
./provenlogs verify -u mongodb://yourusername:yourpassword@yourservicename.provendb.io/yourservicename?ssl=true -l '{"level":"info","ts":1558935983.776685,"caller":"zapper/zapper.go:29","msg":"test 6","delay":745}'
```

which will output something like this when the raw log entry is verified:

```bash
The raw log entry:
	{"level":"info","ts":1558935983.776685,"caller":"zapper/zapper.go:29","msg":"test 6","delay":745}
is found in the batch with ProvenDB version:
	4
is stored in ProvenDB as:
	{"_provendb_metadata":{"_id":{"$oid":"5ceb79b37126e5440ff8d099"},"_mongoId":{"$oid":"5ceb79b37126e5440ff8d099"},"minVersion":{"$numberLong":"4"},"hash":"d4c15a75b742ea85c707c5bc911469d6e712df08b1ebfbf16137ab2bf55cc0d9","maxVersion":{"$numberLong":"9223372036854775807"}},"timestamp":{"$date":{"$numberLong":"1558935983777"}},"level":"info","message":"test 6","data":{"caller":"zapper/zapper.go:29","delay":{"$numberDouble":"745.0"}}}
has its batch RSA signature stored in the last log entry of the batch with hash:
	4fb59e6e96f6e6d04b1d3311a1e0f34f1a3e2887887144684689a610e88e5c7f
has a valid batch RSA signature that has been verified using the batch data stored in ProvenDB and with the public key:
	/Users/guiguan/Documents/Projects/SBS/provenlogs/key-pub.pem
has its batch RSA signature proven to be existed on Bitcoin with this ProvenDB document proof (just verified):
	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
	which is confirmed at: 2019-05-27 16:07:47 +1000 AEST
	which has transaction ID: 274d8fc382988d3eaa3201a66f0d912ee93d79243e50691acaeb04b67a0e3855
	which resides in block: 577974
is verified
```
