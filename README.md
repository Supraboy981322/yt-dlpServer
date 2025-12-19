# yt-dlpServer

A stripped-down version of [the toolbox thingy](https://github.com/Supraboy981322/toolbox) with just yt-dlp and a dedicated client cli tool


## Usage

#### Official client

- Video URL
  ```sh
  yT https://youtu.be/some-id
  ```

- File format (replace `mp4` with your desired format)
  ```sh
  yT https://youtu.be/some-id -f mp4
  ```

- Output file (replace `foo` with your filename, extension is appened to end, so no need to add it here)
  ```sh
  yT https://youtu.be/some-id -o foo
  ```

- Over-ride server address set in config (replace `your.server.address`)
  ```sh
  yT https://youtu.be/some-id -s your.server.address
  ```

- Additional yt-dlp args (replace `some random args` with your args)
  ```sh
  yT https://youtu.be/some-id -a some random args
  ```

- Verbose
  ```sh
  yT https://youtu.be/some-id -v
  ```

- Help
  ```sh
  yT -h
  ```

#### cURL

TODO: write cURL usage


## Installation

#### Server

1) Go install
  ```sh
  go install github.com/Supraboy981322/yt-dlpServer/yt-dlpServer
  ```
2) Create a config file named (`config.gomn` in the working directory for the server)
  ```gomn
  ["log level"] := "debug"
  ["port"] := 4895
  ["use web ui"] := false //TODO: web ui 
  ```
3) Create a systemd service file (`yt-dlpServer.service`) with the following contents (typically located in `/etc/systemd/system`, replace `/your/server/directory` with the directory for your server config, and `/your/gobin/path` with path that `yt-dlpServer` is installed to, may require `su`)
  ```systemd
  [Unit]
  Description=yt-dlpServer
  After=network.target
  
  [Service]
  WorkingDirectory=/your/server/directory
  ExecStart=/your/gobin/path/yt-dlpServer
  Restart=always
  User=root
  Group=root

  [Install]
  WantedBy=multi-user.target
  Alias=yt-dlpServer.service
  ```
  Or, your distro's equivalent

4) Enable and start the systemd service (or your distro's equivalent, may require `su`)
  ```sh
  systemctl enable yt-dlpServer.service
  systemctl start yt-dlpServer.service
  ```
  You can ensure that the systemd service is working with this command
  ```sh
  systemd status yt-dlpServer.service
  ```

#### Client

1) Using Go install
  ```sh
  go install github.com/Supraboy981322/yt-dlpServer/yT
  ```
2) Run the client command, to create a configuration file (if it doesn't already exist)
  ```sh
  yT
  ```
