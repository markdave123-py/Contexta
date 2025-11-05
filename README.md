Contexta - Document Intelligence Platform
https://img.shields.io/badge/Contexta-Document%2520Intelligence-blue
https://img.shields.io/badge/Go-1.21%252B-blue
https://img.shields.io/badge/PostgreSQL-15%252B-blue
https://img.shields.io/badge/Frontend-Vanilla%2520JS-yellow

Contexta is a powerful document intelligence platform that allows users to upload documents and ask questions about their content using AI-powered chat. Built with Go, PostgreSQL, and modern web technologies.

🚀 Features
Document Upload: Support for PDF, TXT, and DOCX files

AI-Powered Chat: Ask questions about your documents using Gemini AI

Smart Retrieval: Vector-based semantic search for accurate answers

User Authentication: Secure JWT-based authentication system

Real-time Processing: Background processing of uploaded documents

Responsive UI: Clean, modern interface for seamless user experience

🛠 Tech Stack
Backend
Go - High-performance backend server

Chi Router - Lightweight HTTP router

PostgreSQL - Primary database with pgvector extension

AWS S3 - Cloud storage for documents

JWT - Secure authentication

AI/ML
Google Gemini - LLM for generating answers

Vector Embeddings - Semantic search using document chunks

RAG Architecture - Retrieval Augmented Generation

Frontend
Vanilla JavaScript - No framework dependencies

Modern CSS - Responsive design with Flexbox/Grid

Local Storage - Token persistence

📦 Installation
Prerequisites
Go 1.21+

PostgreSQL 15+ with pgvector extension

AWS S3 bucket (for file storage)

Google Gemini API key

1. Clone the Repository
bash
git clone https://github.com/your-username/contexta.git
cd contexta
2. Environment Configuration
Create a .env file:

env
DATABASE_URL=postgres://user:password@localhost:5432/contexta?sslmode=disable
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_S3_BUCKET=your-bucket-name
AWS_REGION=us-east-1
GEMINI_API_KEY=your_gemini_api_key
JWT_SECRET=your_jwt_secret_key
PORT=8888

4. Build and Run
bash
# Install dependencies
go mod tidy

# Run the application
go run ./cmd/api

# Or build and run
go build -o contexta ./cmd/api
./contexta
The application will be available at http://localhost:8888

🎯 Usage
1. Authentication
Register a new account or login with existing credentials

JWT tokens are automatically stored and used for API calls

2. Document Upload
Click "Upload Document" and select a PDF, TXT, or DOCX file

Wait for processing (status will change from "processing" to "ready")

Documents are automatically chunked and embedded for search

3. Chat with Documents
Select a processed document from the sidebar

Ask questions about the document content

Receive AI-generated answers based on the document content

Start new chat sessions to clear conversation history

🔧 API Endpoints
Authentication
POST /api/signup - Create new user account

POST /api/login - User login

Documents
GET /api/documents - Get user's documents

POST /api/documents/upload - Upload new document

Chat
POST /api/chat/query - Ask questions about a document

🏗 Architecture
Document Processing Pipeline
Upload → Document uploaded to S3

Extraction → Text extracted using docconv

Chunking → Text split into semantic chunks

Embedding → Chunks converted to vector embeddings

Storage → Chunks and embeddings stored in PostgreSQL

Retrieval Augmented Generation (RAG)
Query Embedding → User question converted to vector

Semantic Search → Find most relevant document chunks

Context Building → Relevant chunks form context

AI Generation → LLM generates answer from context

📁 Project Structure
text
contexta/
├── cmd/
│   └── api/                 # Application entry point
├── internal/
│   ├── app/                 # Application setup and server
│   ├── config/              # Configuration management
│   ├── core/                # Core interfaces and models
│   │   ├── database/        # Database client and operations
│   │   ├── handlers/        # HTTP request handlers
│   │   ├── ingestion_engine/ # Document processing pipeline
│   │   ├── llm/            # AI model integrations
│   │   └── object-client/   # S3 client for file storage
│   └── middleware/          # HTTP middleware (auth, CORS, etc.)
├── web/                     # Frontend static files
│   ├── index.html          # Main application UI
│   └── app.js              # Frontend JavaScript
└── Makefile                # Build and development tasks



🔒 Security Features
JWT-based authentication

Document ownership verification

Secure file upload validation

CORS configuration

Input sanitization

SQL injection prevention

🚀 Performance Optimizations
Vector indexing for fast similarity search

Batch processing for document embedding

Connection pooling for database

Background job processing

Efficient chunking strategies

🧪 Development
Running Tests
bash
go test ./...
Code Formatting
bash
go fmt ./...
Building for Production
bash
make build
🤝 Contributing
Fork the repository

Create a feature branch (git checkout -b feature/amazing-feature)

Commit your changes (git commit -m 'Add amazing feature')

Push to the branch (git push origin feature/amazing-feature)

Open a Pull Request

📄 License
This project is licensed under the MIT License - see the LICENSE file for details.

🙏 Acknowledgments
Google Gemini for AI capabilities

Poppler for PDF text extraction

pgvector for PostgreSQL vector operations

Chi Router for HTTP routing

📞 Support
For support and questions:

Create an issue on GitHub

Email: support@contexta.com

Documentation: docs.contexta.com