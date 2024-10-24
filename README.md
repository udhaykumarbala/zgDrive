# ZgDrive

ZgDrive is a web-based file storage and sharing platform that allows users to upload, store, and share files securely. It provides a user-friendly interface for managing and accessing files from anywhere with an internet connection.

## Features

- **Secure File Storage:** ZgDrive ensures that your files are stored securely and protected from unauthorized access.
- **User-Friendly Interface:** The platform offers an intuitive and easy-to-navigate interface, making it accessible for users of all technical backgrounds.
- **File Sharing:** Users can share files with others by generating download links or by inviting collaborators to access the files.
- **File Management:** ZgDrive allows users to manage their files efficiently, including creating folders, renaming files, and deleting unnecessary files.
- **File Sharing:** Users can share files with others by generating download links or by inviting collaborators to access the files.
- **File Management:** ZgDrive allows users to manage their files efficiently, including creating folders, renaming files, and deleting unnecessary files.

## Technologies Used

- **Frontend:** HTML, JavaScript, Tailwind CSS
- **Backend:** Go, Gin, SQLite
- **Storage:** [0g](https://0g.ai)

## Backend Setup

1. Copy the `.env.example` file to `.env` and set the environment variables. Make sure to add private key for wallet.
2. Run the backend server:

```bash
cp .env.example .env
go run main.go
```

## Frontend Setup

```bash
cd ui/zgclient
npm install
npm run dev
```

## Usage

1. Open the frontend in your browser: `http://localhost:5173`
2. Use the interface to upload, manage, and share your files.

## Swagger API Documentation

- Open the API documentation in your browser: `http://localhost:8080/swagger/index.html`

## 
