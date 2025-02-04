# Book Renting App

This is a **Book Renting App** built to learn **Golang** with **Hasura GraphQL**. The backend is developed using **Golang**, **PostgreSQL**, and **Hasura GraphQL**, and runs with **Docker**.

## Features
- **User Authentication**
  - Custom authentication using **JWT** and **bcrypt**
  - OAuth authentication via **Google** and **GitHub**
- **GraphQL API with Hasura**
- **Book Renting System**
- **Image Uploading with Cloudinary**
- **Email Verification & Password Reset via Google SMTP**
- **Payments Integration with Chapa**
- **Go GraphQL Client for Hasura**
- **Gin Framework for API Routing**

## Getting Started

### Prerequisites
- Docker
- Golang
- Hasura CLI
- PostgreSQL

### Installation
1. Clone the repository:
   ```sh
   git clone [https://github.com/LidoHon/book-it-server.git](https://github.com/LidoHon/Book-it-server.git)
   cd Book-it-server
   ```
2. Set up your environment variables in a `.env` file:
   ```ini
   PORT=
   POSTGRES_USER=
   POSTGRES_PASSWORD=
   POSTGRES_DB=
   HASURA_GRAPHQL_PORT=
   ACTION_BASE_URL=
   HASURA_GRAPHQL_ADMIN_SECRET=
   HASURA_ACTION_SECRET=
   HASURA_GRAPHQL_DATABASE_URL=
   HASURA_GRAPHQL_API_ENDPOINT=
   HASURA_GRAPHQL_ENDPOINT=
   DB_HOST=
   DB_PORT=
   DB_USER=
   DB_PASSWORD=
   DB_NAME=
   JWT_SECRET_KEY=
   RESET_PASS_URL=
   CHAPA_RETURN_URL=
   EMAIL_HOST=
   EMAIL_USERNAME=
   GOOGLE_EMAIL_PASSWORD=
   EMAIL_PASSWORD=
   FROM_EMAIL=
   SERVICE=
   GMAIL_HOST=
   GMAIL_PORT=
   EMAIL_PORT=
   CLOUD_NAME=
   CLOUDINARY_API_KEY=
   CLOUDINARY_API_SECRET=
   CLOUDINARY_URL=
   CHAPA_SECRET_KEY=
   CHAPA_PUBLIC_KEY=
   CHAPA_PAYMENT_ENDPOINT=
   CHAPA_BANKS_ENDPOINT=
   CHAPA_TRANSACTION_VERIFICATION_ENDPOINT=
   CHAPA_CALLBACK_URL=
   BASE_URL=
   CLIENT_LOGIN_URL=
   CLIENT_HOMEPAGE_URL=
   GOOGLE_CLIENT_ID=
   GOOGLE_CLIENT_SECRET=
   GITHUB_CLIENT_ID=
   GITHUB_CLIENT_SECRET=
   SESSION_SECRET=
   ```

3. Start the services using Docker:
   ```sh
   docker-compose up -d
   ```
4. Run the Go Backend
```sh
go run main.go
```
or 
```sh 
air
```

## API Documentation
The API is powered by **Hasura GraphQL**, and the endpoints can be accessed via:
```
http://localhost:8080/v1/graphql
```

## Authentication
### Custom Authentication
- **JWT-based authentication** using bcrypt for password hashing.
- Secure user registration and login system.

### OAuth Authentication
- **Google OAuth**
- **GitHub OAuth**

## Payment Integration
Payments are handled using **Chapa**, allowing secure and seamless transactions.

## Deployment
To deploy the application, ensure all environment variables are correctly set and run:
```sh
docker-compose up --build -d
```

## Contributing
Feel free to submit issues and pull requests to improve the project!

## License
This project is licensed under the **MIT License**.
