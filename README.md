# testssl-webui

The testssl-webui project was inspired by two key sources: the powerful [testssl.sh](https://github.com/testssl/testssl.sh) tool and the convenient [PoodleScan service](https://www.poodlescan.com/testssl.sh/).

While testssl.sh is an extremely capable TLS/SSL configuration analysis tool, its command-line interface isn't user-friendly for non-technical individuals. Online services like PoodleScan, despite offering a convenient web interface, require the target to have a public URL, which is a limitation for internal networks or services that shouldn't be exposed.

To address these pain points, testssl-webui was created.

---

## Project Features

- Lightweight Go Gin HTTP Server: Utilizes the Go Gin framework to provide a lightweight and efficient HTTP server, hosting a simple and intuitive web interface.

- Built on testssl.sh Docker Image: The project's Docker image is directly based on the official testssl.sh Docker image (ghcr.io/testssl/testssl.sh:3.2). This means you don't need to install testssl.sh and all its complex dependencies; simply launch the container.

- Private Network Scanning Capability: Users only need to launch the testssl-webui container within their private network to securely perform testssl.sh scans on targets within private domains, without exposing their services to the public internet.

- Real-time SSE Scan Responses: Leveraging Server-Sent Events (SSE) technology, real-time responses during the scanning process are displayed directly on the web page, allowing users to monitor progress and results instantly.

- PDF Report Export: After the scan is complete, users can choose to export detailed scan results as a PDF report, making it easy for sharing, archiving, or further analysis.


> testssl-webui aims to provide users with a convenient, secure, and fully-featured web-based testssl.sh experience, making TLS/SSL security checks easily accessible.

## How to Use

To get testssl-webui up and running, simply use a Docker command.

To run in the foreground (showing logs):

```sh
docker run -it --rm --platform=linux/amd64 -p 8081:8081 vulcanshen2304/testssl-webui
```

To run in the background (detached mode):

```sh
docker run -d --rm --platform=linux/amd64 -p 8081:8081 vulcanshen2304/testssl-webui
```

### Important Notes:

- If you're using an ARM-based CPU (like Apple M1/M2 Macs, Raspberry Pi, etc.), make sure to include the `--platform=linux/amd64` flag. This tells Docker to perform cross-architecture emulation.

- The `-p 8081:8081` flag maps your host machine's port 8081 to the container's internal port 8081.

- Once the container is running, open your web browser and go to: http://localhost:8081

You can now start using testssl-webui for your TLS/SSL scanning!
