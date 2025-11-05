class ContextaApp {
    constructor() {
        this.baseUrl = 'http://localhost:8888/api';
        this.currentDocument = null;
        this.documents = [];
        this.chatHistory = [];
        this.token = localStorage.getItem('authToken');
        this.userEmail = localStorage.getItem('userEmail');

        this.initializeElements();
        this.setupEventListeners();
        this.checkAuthStatus();
    }

    initializeElements() {
        // Auth elements
        this.loginScreen = document.getElementById('loginScreen');
        this.appScreen = document.getElementById('appScreen');
        this.loginForm = document.getElementById('loginForm');
        this.signupForm = document.getElementById('signupForm');
        this.loginButton = document.getElementById('loginButton');
        this.signupButton = document.getElementById('signupButton');
        this.switchToSignup = document.getElementById('switchToSignup');
        this.switchToLogin = document.getElementById('switchToLogin');
        this.authStatus = document.getElementById('authStatus');
        this.logoutButton = document.getElementById('logoutButton');
        this.userEmailSpan = document.getElementById('userEmail');

        // App elements
        this.uploadForm = document.getElementById('uploadForm');
        this.fileInput = document.getElementById('fileInput');
        this.uploadButton = document.getElementById('uploadButton');
        this.uploadStatus = document.getElementById('uploadStatus');
        
        this.documentsList = document.getElementById('documentsList');
        this.chatHeader = document.getElementById('chatHeader');
        this.chatMessages = document.getElementById('chatMessages');
        this.messageInput = document.getElementById('messageInput');
        this.sendButton = document.getElementById('sendButton');
        this.newChatBtn = document.getElementById('newChatBtn');
    }

    setupEventListeners() {
        // Auth event listeners
        this.loginForm.addEventListener('submit', (e) => this.handleLogin(e));
        this.signupForm.addEventListener('submit', (e) => this.handleSignup(e));
        this.switchToSignup.addEventListener('click', () => this.switchAuthForm('signup'));
        this.switchToLogin.addEventListener('click', () => this.switchAuthForm('login'));
        this.logoutButton.addEventListener('click', () => this.handleLogout());

        // App event listeners
        this.uploadForm.addEventListener('submit', (e) => this.handleUpload(e));
        this.sendButton.addEventListener('click', () => this.sendMessage());
        this.messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendMessage();
            }
        });
        this.newChatBtn.addEventListener('click', () => this.startNewChat());
    }

    checkAuthStatus() {
        if (this.token && this.userEmail) {
            this.showApp();
        } else {
            this.showLogin();
        }
    }

    switchAuthForm(form) {
        if (form === 'signup') {
            this.loginForm.style.display = 'none';
            this.signupForm.style.display = 'block';
        } else {
            this.signupForm.style.display = 'none';
            this.loginForm.style.display = 'block';
        }
        this.authStatus.innerHTML = '';
    }

    async handleLogin(e) {
        e.preventDefault();
        
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        this.loginButton.disabled = true;
        this.loginButton.textContent = 'Logging in...';
        this.authStatus.innerHTML = '<div class="loading">Logging in...</div>';

        try {
            const response = await fetch(`${this.baseUrl}/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Login failed');
            }

            const data = await response.json();
            
            // Store token and user info
            this.token = data.token;
            this.userEmail = email;
            localStorage.setItem('authToken', this.token);
            localStorage.setItem('userEmail', this.userEmail);

            this.showApp();

        } catch (error) {
            this.authStatus.innerHTML = `<div class="error">Login failed: ${error.message}</div>`;
        } finally {
            this.loginButton.disabled = false;
            this.loginButton.textContent = 'Login';
        }
    }

    async handleSignup(e) {
        e.preventDefault();
        
        const email = document.getElementById('signupEmail').value;
        const password = document.getElementById('signupPassword').value;
        const name = document.getElementById('signupName').value;

        this.signupButton.disabled = true;
        this.signupButton.textContent = 'Signing up...';
        this.authStatus.innerHTML = '<div class="loading">Creating account...</div>';

        try {
            const response = await fetch(`${this.baseUrl}/signup`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email, password, name })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Signup failed');
            }

            // Switch to login form after successful signup
            this.switchAuthForm('login');
            this.authStatus.innerHTML = '<div class="success">Account created! Please login.</div>';

        } catch (error) {
            this.authStatus.innerHTML = `<div class="error">Signup failed: ${error.message}</div>`;
        } finally {
            this.signupButton.disabled = false;
            this.signupButton.textContent = 'Sign Up';
        }
    }

    handleLogout() {
        this.token = null;
        this.userEmail = null;
        localStorage.removeItem('authToken');
        localStorage.removeItem('userEmail');
        this.showLogin();
    }

    showApp() {
        this.loginScreen.style.display = 'none';
        this.appScreen.style.display = 'block';
        this.userEmailSpan.textContent = this.userEmail;
        this.loadDocuments();
    }

    showLogin() {
        this.loginScreen.style.display = 'flex';
        this.appScreen.style.display = 'none';
        this.loginForm.reset();
        this.signupForm.reset();
        this.authStatus.innerHTML = '';
    }

    // Helper method for authenticated requests
    async authenticatedFetch(url, options = {}) {
        const headers = {
            'Authorization': `Bearer ${this.token}`,
            ...options.headers
        };

        const response = await fetch(url, {
            ...options,
            headers
        });

        if (response.status === 401) {
            this.handleLogout();
            throw new Error('Session expired. Please login again.');
        }

        return response;
    }

    async handleUpload(e) {
        e.preventDefault();
        
        const file = this.fileInput.files[0];
        if (!file) {
            this.showUploadStatus('Please select a file', 'error');
            return;
        }

        this.uploadButton.disabled = true;
        this.uploadButton.textContent = 'Uploading...';
        this.showUploadStatus('Uploading document...', 'loading');

        try {
            const formData = new FormData();
            formData.append('file', file);

            const response = await this.authenticatedFetch(`${this.baseUrl}/documents/upload`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Upload failed: ${response.status}`);
            }

            const result = await response.json();
            this.showUploadStatus('Document uploaded successfully! Processing...', 'success');
            this.fileInput.value = '';
            
            // Reload documents list
            setTimeout(() => {
                this.loadDocuments();
            }, 1000);

        } catch (error) {
            this.showUploadStatus(`Upload failed: ${error.message}`, 'error');
        } finally {
            this.uploadButton.disabled = false;
            this.uploadButton.textContent = 'Upload';
        }
    }

    showUploadStatus(message, type) {
        this.uploadStatus.innerHTML = `<div class="${type}">${message}</div>`;
        
        if (type === 'success') {
            setTimeout(() => {
                this.uploadStatus.innerHTML = '';
            }, 5000);
        }
    }

    async loadDocuments() {
        try {
            const response = await this.authenticatedFetch(`${this.baseUrl}/documents`);
            if (!response.ok) throw new Error('Failed to load documents');
            
            this.documents = await response.json();
            this.renderDocuments();
        } catch (error) {
            this.documentsList.innerHTML = `<div class="error">Error loading documents: ${error.message}</div>`;
        }
    }

    renderDocuments() {
        if (this.documents.length === 0) {
            this.documentsList.innerHTML = '<div>No documents found. Upload a document to get started.</div>';
            return;
        }

        this.documentsList.innerHTML = this.documents.map(doc => `
            <div class="document-item ${this.currentDocument?.id === doc.id ? 'active' : ''}" 
                 onclick="app.selectDocument('${doc.id}')">
                <strong>${doc.file_name}</strong>
                <span class="document-status status-${doc.status}">${doc.status}</span>
                <br>
                <small>Uploaded: ${new Date(doc.created_at).toLocaleDateString()}</small>
            </div>
        `).join('');
    }

    selectDocument(documentId) {
        this.currentDocument = this.documents.find(d => d.id === documentId);
        this.renderDocuments();
        this.updateChatInterface();
        
        if (this.currentDocument.status === 'ready') {
            this.chatMessages.innerHTML = `
                <div class="message assistant">
                    Ready to ask questions about "<strong>${this.currentDocument.file_name}</strong>". 
                    Start asking questions below!
                </div>
            `;
        } else {
            this.chatMessages.innerHTML = `
                <div class="message assistant">
                    Document "<strong>${this.currentDocument.file_name}</strong>" is still processing. 
                    Status: <span class="status-${this.currentDocument.status}">${this.currentDocument.status}</span>
                    <br>Please wait until processing is complete to ask questions.
                </div>
            `;
        }
    }

    updateChatInterface() {
        const hasDocument = !!this.currentDocument;
        const isReady = hasDocument && this.currentDocument.status === 'ready';
        
        this.messageInput.disabled = !isReady;
        this.sendButton.disabled = !isReady;
        this.newChatBtn.disabled = !isReady;

        if (this.currentDocument) {
            this.chatHeader.innerHTML = `
                <h3>Chat: ${this.currentDocument.file_name}</h3>
                <div class="session-info">Ask questions about this document</div>
            `;
        } else {
            this.chatHeader.innerHTML = `<h3>Select a document to start chatting</h3>`;
        }
    }

    async sendMessage() {
        const message = this.messageInput.value.trim();
        if (!message || !this.currentDocument) return;

        // Add user message to chat
        this.addMessage('user', message);
        this.messageInput.value = '';
        this.setLoading(true);

        try {
            const response = await this.authenticatedFetch(`${this.baseUrl}/chat/query`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    document_id: this.currentDocument.id,
                    query: message
                })
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP ${response.status}`);
            }

            const data = await response.json();
            
            // Add assistant response
            this.addMessage('assistant', data.answer);

        } catch (error) {
            this.showError(`Error sending message: ${error.message}`);
        } finally {
            this.setLoading(false);
        }
    }

    addMessage(role, content) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${role}`;
        messageDiv.innerHTML = this.escapeHtml(content);
        this.chatMessages.appendChild(messageDiv);
        this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
    }

    startNewChat() {
        if (!this.currentDocument) return;
        
        this.chatMessages.innerHTML = `
            <div class="message assistant">
                Started new chat about "<strong>${this.currentDocument.file_name}</strong>". Ask your first question!
            </div>
        `;
    }

    setLoading(loading) {
        this.sendButton.disabled = loading;
        this.sendButton.textContent = loading ? 'Sending...' : 'Send';
        
        if (loading) {
            const loadingDiv = document.createElement('div');
            loadingDiv.className = 'message assistant loading';
            loadingDiv.textContent = 'Thinking...';
            this.chatMessages.appendChild(loadingDiv);
            this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
        } else {
            const loadingMsg = this.chatMessages.querySelector('.loading');
            if (loadingMsg) {
                loadingMsg.remove();
            }
        }
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error';
        errorDiv.textContent = message;
        this.chatMessages.appendChild(errorDiv);
        this.chatMessages.scrollTop = this.chatMessages.scrollHeight;
    }

    escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;")
            .replace(/\n/g, '<br>');
    }
}

// Initialize the app when page loads
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new ContextaApp();
});