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

### 📤 Uploading and 📥 Downloading Files in ZGDrive

#### Uploading a File
Navigate to the Cloud Tab. This is where you can upload files to the decentralized 0G storage network.

![CleanShot 2024-10-31 at 21 01 43@2x](https://github.com/user-attachments/assets/0f70701a-619f-4189-a03f-5d4a9ecee634)


#### Uploading a File
Choose the file you want to upload. Once you initiate the upload, ZGDrive breaks down your file into chunks and securely distributes them across 0G’s decentralized nodes. During this process, you’ll see the file in the upload queue.

![CleanShot 2024-10-31 at 21 03 29@2x](https://github.com/user-attachments/assets/c08c7e20-0ca2-44c6-ba1a-36afb58f92b0)

![CleanShot 2024-10-31 at 21 03 35@2x](https://github.com/user-attachments/assets/e2843785-2407-4af9-b108-699f050f8b23)


#### Downloading a File
File in Download Queue
Once you request to download a file from the cloud, it will appear in the download queue. ZGDrive retrieves the file from 0G’s decentralized nodes, reassembling the chunks using the Merkle root hash to ensure data integrity.

![CleanShot 2024-10-31 at 21 05 14@2x](https://github.com/user-attachments/assets/0b7675d7-dbd2-419e-b03d-5c219be1802b)

![CleanShot 2024-10-31 at 21 05 46@2x](https://github.com/user-attachments/assets/1f124934-9064-4b42-b68e-3a6361db371e)


#### File in Local Tab After Download Completion
After the download is complete, the file will appear in the Local tab. This means it’s fully downloaded and ready for you to access offline. ZGDrive ensures your data is intact and accessible, thanks to the decentralized storage network. The file will be available in local till the expiry time (Default: 1 Hour) and then the background process will clean up the local files to save space. Your file is safe on the 0g storage network to get on your command

![CleanShot 2024-10-31 at 21 06 40@2x](https://github.com/user-attachments/assets/7fe5aa11-c5a4-4099-afae-de2dcc1acb54)


#### End-to-End Decentralized Storage Workflow
With ZGDrive, the upload and download process is fully decentralized, ensuring that your data is securely stored and retrievable without a centralized intermediary. By splitting and distributing files across multiple nodes and retrieving them as needed, ZGDrive delivers privacy, control, and resilience for your data.

## Swagger API Documentation

- Open the API documentation in your browser: `http://localhost:8080/swagger/index.html`
