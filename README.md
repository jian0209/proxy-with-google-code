## If you are looking for a reverse proxy server that included 2fa authentication, this is the right proxy server you can use.

## how to build
```sh
git clone https://github.com/jian0209/proxy-with-google-code.git
cd proxy-with-google-code
go build
./proxy-google-code --help
```

## build with docker

## how to use
- config.json file is required in the root directory of the project. (Can name it yourself, more information please use --help)
- can refer to the config.json.example file for the configuration file format.
- run the commmand `./proxy_google_code --help` for more information.

## how it works
- The proxy server will listen to the port you set in the config.json file. (default is 8080)
- Proxy server check the code from header named `x-google-code` to verify the 2fa code.
- If the code is correct, the proxy server will forward the request to the target server.
- Can set up the number of failed attempts before the server will block the ip, using redis to store, limit 5 minutes of blocking.
- User cannot use the same code to access the server again, the code will be invalid after using it once, using redis to store the code.

## config.json.example
```json
{
  "authenticated": true, // if true, the proxy server will require 2fa code to access the server
  "number_of_failed": 5, // the number of failed attempts before the server will block the ip, 0 means no limit
  "server_port": 9000, // the port that the proxy server will listen to (default is 8080)
  "redis": {
    "host": "127.0.0.1",
    "port": "6379",
    "auth": "", // the password to access the redis server, if no password, leave it empty
    "db": 0
  },
  "username": "jian0209", // the username to register the 2fa code
  "pass_key": "", // the pass key to register the 2fa code, it will be generated when you use -g
  "proxy_url": [ // an array of the target server that the proxy server will forward the request to
    {
      "id": "1",
      "name": "w", // the name of url, can be anything (eg: http://127.0.0.1:9000/w)
      "url": "http://127.0.0.1:8080" // the url of the target server
    }
  ]
}
```
