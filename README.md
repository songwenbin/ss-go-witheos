# shadowsock go with eos pay

## create envirement

```
mkdir -p $GOPATH/src/github.com
cd  $GOPATH/src/github.com
git clone https://github.com/songwenbin/ss-go-witheos.git
```

## compile

```
cd  $GOPATH/src/github.com/ss-go-witheos
make
```

*note*: if depenceny package is missing, use go get command to download package

## config

filename: config.json
```
{
    "server":"127.0.0.1",
    "server_port":8388,
    "local_port":1080,
    "local_address":"127.0.0.1",
    "password":"barfoo!",
    "method": "aes-128-cfb",
    "timeout":600,
    "contract": {
        "address":"0xdeadbeef",
        "url":"http://ec2-54-95-158-74.ap-northeast-1.compute.amazonaws.com:8888",
        "scope":"incomering1",
        "code":"incomering1",
        "table":"purchase"
    }
}
```

## start server
```
nohup $GOPATH/bin/shadowsock-server -c config.json &
```

## view log 
```
tail -f nohup.out
```