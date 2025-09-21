const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const path = require('path');
require('dotenv').config();

const { initializeDatabase } = require('./models/database');
const { errorHandler, problemDetailsHandler } = require('./middleware/errorHandler');
const rateLimitMiddleware = require('./middleware/rateLimit');

// Import routes
const authRoutes = require('./routes/auth');
const booksRoutes = require('./routes/books');
const loansRoutes = require('./routes/loans');

const app = express();
const PORT = process.env.PORT || 3000;

// Security middleware
app.use(helmet());
app.use(cors());

// Rate limiting
app.use(rateLimitMiddleware);

// Body parsing middleware
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true }));

// API routes
app.use('/api/auth', authRoutes);
app.use('/api/books', booksRoutes);
app.use('/api/loans', loansRoutes);

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'OK', timestamp: new Date().toISOString() });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    type: 'https://httpstatuses.com/404',
    title: 'Not Found',
    status: 404,
    detail: `The requested resource '${req.originalUrl}' was not found`,
    instance: req.originalUrl
  });
});

// Error handling middleware
app.use(problemDetailsHandler);
app.use(errorHandler);

// Initialize database and start server
async function startServer() {
  try {
    await initializeDatabase();
    app.listen(PORT, () => {
      console.log(`Library Lending API server is running on port ${PORT}`);
      console.log(`Environment: ${process.env.NODE_ENV}`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

startServer();

module.exports = app;