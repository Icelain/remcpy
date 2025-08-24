# remcpy

A minimal HTTP file transfer service for temporary file sharing.

## Features

- Upload files via HTTP POST
- Download files via HTTP GET
- Automatic file cleanup after 1 hour
- 5GB file size limit
- Simple identifier-based URLs

## Usage

### Try at [remcpy.ice.computer](remcpy.ice.computer)

### Start the server

```bash
go build
# If no port is specified, remcpy defaults to 5000. If no display host is provided, remcpy defaults to remcpy.ice.computer
./remcpy --port 8080 --display-host "remcpy.ice.computer"
```

### Upload a file

```bash
curl -X POST -F "file=@example.txt" http://localhost:5000/@myfile
```

### Download a file

```bash
curl -X GET http://localhost:5000/@myfile -o downloaded.txt
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | API documentation page |
| `POST` | `/@{id}` | Upload file with identifier |
| `GET` | `/@{id}` | Download file by identifier |

## Configuration

- **Port**: `-port` flag (default: 5000)
- **Storage**: Files stored in `./store/` directory
- **Cleanup**: Files automatically deleted after 1 hour
- **Size limit**: 5GB maximum file size

## Requirements

- Go 1.21+
- Write permissions in current directory
