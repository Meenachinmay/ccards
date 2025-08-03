# CCards - Corporate Card Management System

CCards is a Go-based backend service for managing corporate credit cards. It allows companies to register, upload employee data, issue credit cards to employees, and manage card settings like spending limits.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [Testing](#testing)
  - [Running Tests](#running-tests)
  - [HTTP Tests](#http-tests)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Features

- Company registration and authentication
- Employee data upload via CSV
- Credit card issuance for employees
- Card management (view cards, update spending limits)
- JWT-based authentication

## Technologies Used

- **Go**: Main programming language
- **Gin**: Web framework
- **PostgreSQL**: Database for storing application data
- **Redis**: For caching and session management
- **JWT**: For authentication
- **Docker**: For containerization
- **Viper**: For configuration management

## Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.16 or higher
- PostgreSQL 16 or higher
- Redis 7 or higher
- Docker and Docker Compose (optional, for containerized setup)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/ccards.git
cd ccards
```

2. Install dependencies:

```bash
go mod download
```

3. Set up the database and Redis using Docker Compose:

```bash
docker-compose -f docker-local.compose.yml up -d postgres redis
```

## Configuration

The application uses YAML configuration files located in the `config` directory. The base configuration is in `base.yml`, and environment-specific configurations are in files like `local.yml`, `dev.yml`, etc.

You can also use environment variables to override configuration values. The application looks for a `.env` file in the project root.

Key configuration parameters:

- **Database**: Connection details for PostgreSQL
- **Redis**: Connection details for Redis
- **JWT**: Secret and token durations for authentication
- **Server**: Host, port, and timeout settings

## Running the Application

### Using Go directly

```bash
make run
```

or

```bash
go run ./cmd/server/main.go
```

### Using Docker Compose

Uncomment the `app` service in `docker-local.compose.yml` and run:

```bash
docker-compose -f docker-local.compose.yml up
```

The application will be available at http://localhost:8080.

## Testing

### Running Tests

The project includes various test commands in the Makefile:

- Run all tests:
  ```bash
  make test
  ```

- Run unit tests only:
  ```bash
  make test-unit
  ```

- Run integration tests only:
  ```bash
  make test-integration
  ```

- Run tests with coverage:
  ```bash
  make test-coverage
  ```

- Run tests in Docker environment:
  ```bash
  make test-docker
  ```

### HTTP Tests

The project includes HTTP tests that can be run using REST client tools like [REST Client for VS Code](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) or [Postman](https://www.postman.com/).

The HTTP tests are located in `tests/http_tests/api_tests.http`. They demonstrate the API endpoints and how to use them.

To run the HTTP tests:

1. Start the application
2. Open `tests/http_tests/api_tests.http` in a REST client
3. Run the requests in sequence:
   - Health Check
   - Register Company
   - Company Login (this will set the access token for subsequent requests)
   - Get Company Details
   - Upload Card CSV (using the provided `employee.csv` file)
   - Get Cards To Issue
   - Issue New Cards
   - Verify Cards Status After Issuing
   - Get all cards for a company
   - Update Card Spending Limit

## API Documentation

### Authentication

- **POST /auth/company/login**: Authenticate a company and get an access token
  ```json
  {
    "email": "test@example.com",
    "password": "your_password"
  }
  ```

### Admin Endpoints

- **POST /admin/company/register**: Register a new company
  ```json
  {
    "name": "Test Company",
    "email": "test@example.com",
    "address": "123 Test Street, Test City",
    "phone": "+1234567890"
  }
  ```

### Company Endpoints

- **GET /api/company**: Get company details
- **POST /api/company/upload-csv**: Upload employee data via CSV
- **GET /api/company/card-to-issue**: Get cards ready to be issued
- **POST /api/company/issue-cards**: Issue new cards to employees

### Card Endpoints

- **GET /api/cards**: Get all cards for the authenticated company
- **POST /api/cards/update/spending-limit**: Update card spending limit
  ```json
  {
    "spending_limit": 5000
  }
  ```

### Health Check

- **GET /health**: Check if the application is running

## Project Structure

```
ccards/
├── cmd/                    # Application entry points
│   └── server/             # Main server application
├── config/                 # Configuration files
├── db/                     # Database migrations
│   └── migrations/         # SQL migration files
├── internal/               # Internal packages
│   ├── api/                # API request/response models
│   ├── card/               # Card management
│   ├── client/             # Client (company) management
│   ├── notification/       # Notification services
│   ├── router/             # HTTP router setup
│   ├── server/             # Server initialization
│   ├── store/              # Store management
│   └── transaction/        # Transaction management
├── pkg/                    # Shared packages
│   ├── config/             # Configuration loading
│   ├── database/           # Database connection
│   ├── errors/             # Error handling
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Shared data models
│   └── utils/              # Utility functions
└── tests/                  # Tests
    ├── http_tests/         # HTTP test files
    ├── repository/         # Repository tests
    └── service/            # Service tests
```

## Contributing

Contributions are welcome! Here's how you can contribute:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes: `git commit -m 'Add some feature'`
4. Push to the branch: `git push origin feature/your-feature-name`
5. Open a pull request

Please make sure your code follows the project's coding style and includes appropriate tests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
