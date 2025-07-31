# HTMX Chat Application

A real-time chat application built with modern web technologies, demonstrating the power of server-side rendering combined with dynamic client-side interactions.

## Overview

This application is a real-time chat platform that allows users to:
- Create and join chat rooms
- Send messages in real-time
- See updates instantly without page refreshes

The application demonstrates how to build a modern, interactive web application with minimal JavaScript by leveraging HTMX for dynamic content updates and WebSockets for real-time communication.

## Major Components

### HTMX

[HTMX](https://htmx.org/) is a lightweight JavaScript library that allows you to access modern browser features directly from HTML, rather than using JavaScript. In this application, HTMX is used for:

- Making AJAX requests to fetch and update content without page reloads
- Triggering updates based on events (like new messages)
- Swapping HTML content dynamically
- Integrating with WebSockets for real-time updates

Key HTMX attributes used in this project:
- `hx-get` - For fetching data from the server
- `hx-post` - For submitting forms
- `hx-trigger` - For defining when to make requests
- `hx-swap` - For specifying how to update the DOM
- `hx-target` - For targeting specific elements to update

### Golang Templates

Go's [html/template](https://pkg.go.dev/html/template) package is used for server-side rendering. The application uses a structured template approach:

- **Layouts**: Base templates that define the overall page structure
- **Partials**: Reusable components and page fragments

Templates are organized in a modular way, allowing for:
- Component reuse
- Conditional rendering
- Dynamic content generation
- Nested templates

### Gin Web Framework

[Gin](https://github.com/gin-gonic/gin) is a high-performance HTTP web framework written in Go. In this application, Gin is used for:

- Routing HTTP requests
- Serving static files
- Rendering HTML templates
- Handling form submissions
- Managing WebSocket connections

Gin's middleware system is leveraged for common web application features, and its context-based request handling makes the code clean and maintainable.

### TailwindCSS

[TailwindCSS](https://tailwindcss.com/) is a utility-first CSS framework that allows for rapid UI development. In this application, Tailwind is used for:

- Responsive layout design
- Component styling
- Consistent theming
- Utility classes for spacing, colors, typography, etc.

The application uses Tailwind via CDN for simplicity, making it easy to get started without a build process.

### WebSockets

[WebSockets](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API) provide a persistent connection between the client and server, allowing for real-time, bidirectional communication. In this application, WebSockets are used for:

- Broadcasting new messages to all connected clients
- Notifying clients when new rooms are created
- Ensuring all users see updates in real-time

The implementation uses the [Gorilla WebSocket](https://github.com/gorilla/websocket) package and follows a hub-based architecture for managing connections and broadcasting messages.

## Installation and Setup

### Prerequisites

- Go 1.24 or later
- Git

### Steps

1. Clone the repository:
   ```
   git clone <repository-url>
   cd htmx
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Run the application:
   ```
   go run main.go
   ```

4. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

## Usage

### Creating a Room

1. Enter a room name in the "Create a new room" form
2. Click "Create"
3. The new room will appear in the sidebar for all connected users

### Joining a Room

1. Click on a room name in the sidebar
2. You'll be taken to the room's chat interface

### Sending Messages

1. Enter your username
2. Type your message
3. Click "Send"
4. Your message will appear in real-time for all users in the room

## Project Structure

```
├── internal/
│   ├── handlers/       # HTTP and WebSocket handlers
│   ├── models/         # Data models and in-memory stores
│   └── templates/      # Go HTML templates
│       ├── layouts/    # Base page layouts
│       └── partials/   # Reusable components
├── static/
│   └── css/            # Custom CSS styles
├── main.go             # Application entry point
└── go.mod              # Go module definition
```

## Development

The application uses an in-memory data store for simplicity, making it easy to get started with development. For a production environment, you would want to replace this with a persistent database.

## License

[MIT License](LICENSE)