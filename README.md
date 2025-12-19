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

1) Download the latest build
  ```sh
  curl github.com/Supraboy981322/yt-dls/builds/server -O -J
  ```

2) Make it executable (may require `su`)
  ```sh
  chmod a+x server
  ```

3) Create a systemd service file (`yt-dls.service`) with the following contents (typically located in `/etc/systemd/system`, replace `/your/server/path` with the path to your server, may require `su`)
  ```systemd
  [Unit]
  Description=yt-dls
  After=network.target
  
  [Service]
  WorkingDirectory=/your/server/path
  ExecStart=./server
  Restart=always
  User=root
  Group=root
  ```
  Or, your distro's equivalent

3) Enable and start the systemd service (or your distro's equivalent, may require `su`)
  ```sh
  systemctl enable yt-dls.service
  systemctl start yt-dls.service
  ```
  You can ensure that the systemd service is working with this command
  ```sh
  systemd status yt-dls.service
  ```
#### Client

There are two ways to download the client

1) Download the binary from the server (replace `your.server.address` with your server address)
  ```sh
  curl your.server.address/client-dl -o yT
  ```
  Or, get the latest version from GitHub
  ```sh
  curl github.com/Supraboy981322/yt-dls/builds/client/yT -O -J
  ```

2) Make it executable (may require `su`)
  ```sh
  chmod a+x yT
  ```

3) Move to somewhere in your path (eg: `/bin`, may require `su`)
  ```sh
  mv yT /bin/yT
  ```
