# ApiMug

A terminal-based API client for OpenAPI/Swagger specifications with an interactive TUI interface.

## Features

- Browse and explore OpenAPI 3.0 and Swagger 2.0 specifications
- Interactive TUI powered by Bubbletea
- Send HTTP requests directly from the terminal
- Multiple authentication methods (Bearer, API Key, Basic, OAuth2)
- Built-in Swagger UI server
- Live configuration of base URL and server port
- Support for both JSON and YAML formats
- Automatic Swagger 2.0 to OpenAPI 3.0 conversion

## Installation

```bash
go install github.com/doganarif/apimug/cmd/apimug@latest
```

Or build from source:

```bash
git clone https://github.com/doganarif/apimug.git
cd apimug
go build -o apimug ./cmd/apimug
```

## Usage

### Basic Usage

```bash
apimug spec.yaml
```

### With Options

```bash
# Specify custom base URL
apimug spec.yaml --base-url https://api.example.com

# Specify custom Swagger UI port
apimug spec.yaml --port 3000

# Load from URL
apimug https://petstore.swagger.io/v2/swagger.json
```

### Keyboard Shortcuts

**Main List View**
- `↑/↓` or `j/k` - Navigate endpoints
- `Enter` - View endpoint details
- `s` - Configure authentication
- `c` - Open settings
- `q` - Quit

**Endpoint Details**
- `Enter` - Send request
- `Esc` - Back to list
- `q` - Quit

**Request Form**
- `Tab` - Navigate between fields
- `Ctrl+S` - Send request
- `Esc` - Back to details

**Response View**
- `Esc` - Back to request form
- `q` - Quit

**Settings**
- `Tab` - Navigate between fields
- `Ctrl+S` - Save settings
- `Esc` - Cancel

**Authentication**
- `↑/↓` - Select auth scheme
- `Tab` - Navigate between fields
- `Ctrl+S` - Save configuration
- `Esc` - Cancel

## Authentication

ApiMug supports multiple authentication methods:

- **None** - No authentication
- **Bearer Token** - JWT or other bearer tokens
- **API Key** - Header, query, or cookie-based API keys
- **Basic Auth** - Username and password
- **OAuth2** - OAuth2 bearer tokens

Configure authentication by pressing `s` from the main view.

## Settings

Press `c` from the main view to configure:

- **Base URL** - API endpoint base URL
- **Swagger UI Port** - Port for the built-in Swagger UI server

Settings can be changed at runtime without restarting the application.

## Examples

The repository includes example specifications:

- `example.yaml` - OpenAPI 3.0 Pet Store API
- `swagger-example.yaml` - Swagger 2.0 Pet Store API

Try them out:

```bash
apimug example.yaml
apimug swagger-example.yaml
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome. Please feel free to submit a Pull Request.
