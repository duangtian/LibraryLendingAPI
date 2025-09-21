// Error handling middleware with JSON Problem Details (RFC 7807)

class AppError extends Error {
  constructor(message, statusCode = 500, type = null, title = null, detail = null) {
    super(message);
    this.statusCode = statusCode;
    this.type = type;
    this.title = title;
    this.detail = detail || message;
    this.isOperational = true;

    Error.captureStackTrace(this, this.constructor);
  }
}

class ValidationError extends AppError {
  constructor(message, errors = []) {
    super(message, 400, 'https://httpstatuses.com/400', 'Validation Error', message);
    this.errors = errors;
  }
}

class AuthenticationError extends AppError {
  constructor(message = 'Authentication required') {
    super(message, 401, 'https://httpstatuses.com/401', 'Authentication Error', message);
  }
}

class AuthorizationError extends AppError {
  constructor(message = 'Insufficient permissions') {
    super(message, 403, 'https://httpstatuses.com/403', 'Authorization Error', message);
  }
}

class NotFoundError extends AppError {
  constructor(resource = 'Resource') {
    const message = `${resource} not found`;
    super(message, 404, 'https://httpstatuses.com/404', 'Not Found', message);
  }
}

class ConflictError extends AppError {
  constructor(message = 'Resource already exists') {
    super(message, 409, 'https://httpstatuses.com/409', 'Conflict', message);
  }
}

class RateLimitError extends AppError {
  constructor(message = 'Too many requests') {
    super(message, 429, 'https://httpstatuses.com/429', 'Rate Limit Exceeded', message);
  }
}

// JSON Problem Details middleware
const problemDetailsHandler = (err, req, res, next) => {
  if (err.isOperational || err.statusCode) {
    const problemDetails = {
      type: err.type || 'https://httpstatuses.com/500',
      title: err.title || 'Internal Server Error',
      status: err.statusCode || 500,
      detail: err.detail || err.message,
      instance: req.originalUrl
    };

    // Add validation errors if present
    if (err.errors && Array.isArray(err.errors)) {
      problemDetails.errors = err.errors;
    }

    res.status(err.statusCode || 500)
       .type('application/problem+json')
       .json(problemDetails);
  } else {
    next(err);
  }
};

// General error handler
const errorHandler = (err, req, res, next) => {
  console.error('Unhandled error:', err);

  // Don't leak error details in production
  const isDevelopment = process.env.NODE_ENV === 'development';
  
  const problemDetails = {
    type: 'https://httpstatuses.com/500',
    title: 'Internal Server Error',
    status: 500,
    detail: isDevelopment ? err.message : 'An unexpected error occurred',
    instance: req.originalUrl
  };

  if (isDevelopment && err.stack) {
    problemDetails.stack = err.stack;
  }

  res.status(500)
     .type('application/problem+json')
     .json(problemDetails);
};

module.exports = {
  AppError,
  ValidationError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ConflictError,
  RateLimitError,
  problemDetailsHandler,
  errorHandler
};