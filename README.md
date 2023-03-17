# Fabric Automation Tools
It's a pilot project considering to prototype and build swiss-knife tools necessary to manage SR Linux based fabric.
While brilliant tools exists, like [gnoic](#gnoic) or [gnmic](#gnmic), which are giving to you almost full flexibility.
Some practical use-cases may require additional scripting and coding. 

#Tools

## SR Linux config extractor

`srlce` is tool allowing you to extract info object config from SR Linux device using different ways depending what's available to you as interface: SSH / JSON RPC / gNOI.
On top of that it allows to cleanup [clab](#clab) configuration artifacts related to banner, certificate,... 
Utility does not support JSON config, since it can be done via [gnmic](#gnmic) in more robust way.

```sh
[azyablov@ecartman srlce]$ ./srlce --help
Usage of ./srlce:
  -InsecConn
        TLS insecure connectivity
  -JRPCport int
        JSON RPC port (default 443)
  -SSHport int
        SSH port (default 22)
  -SkipVerify
        skip TLS certificate chain verification
  -cclab
        Clean up clab generated config
  -cert string
        Client certificate file in PEM format
  -d    Enable debug, by default warn
  -gNOIdld
        Use gNOI to download info config
  -gNOIport int
        gNOI port (default 57400)
  -jsonrpc
        Use JSON RPC instead of SSH
  -key string
        Client private key file
  -logFile string
        Log all messages into specified log file instead of stderr
  -logSSH
        Enable SSH debug, by default disabled
  -noSKey
        No SSH key checking (default true)
  -password string
        SSH password (default "NokiaSrl1!")
  -printTree
        Print info object tree
  -rFile string
        Remote file name on target NE (default "myconfig.cfg")
  -rootCA string
        CA certificate file in PEM format
  -target string
        Target hostname
  -timeout duration
        Connection timeout (default 10s)
  -username string
        SSH username (default "admin")
```
A few examples how `srlce` can be used:

```sh
echo ">>>>>> SSH scraping only"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -d -logSSH
--
echo ">>>>>> SSH scraping with gNOI skip verify"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -SkipVerify -d
--
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -d
--
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -cclab -d
--
echo ">>>>>> JSON RPC scraping only"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -d
--
echo ">>>>>> JSON RPC scraping only + clab cleanup"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -cclab -d
--
echo ">>>>>> JSON RPC with gNOI skip verify"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -SkipVerify -cclab -d
--
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate" 
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -d
--
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -cclab -d
--
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -cclab -d
--
echo ">>>>>> JSON RPC with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file"
go run srlce.go -target $TARGET -jsonrpc -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -cclab -d -logFile ./srlce_jsonrpc.log
--
echo ">>>>>> SSH scraping with gNOI and root CA certificate, key and certificate + clab cleanup + debug + log file"
go run srlce.go -target $TARGET -username $USER -password $PASSWORD -gNOIdld -rootCA ${LAB_CA_DIR}/root-ca.pem -key ${LAB_CA_DIR}/srl-key.pem  -cert ${LAB_CA_DIR}/srl.pem -cclab -d -logFile ./srlce_ssh.log -logSSH
```



[gnoic]: https://github.com/karimra/gnoic
[gnmic]: https://github.com/openconfig/gnmic
[clab]: https://containerlab.dev


