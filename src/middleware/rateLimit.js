const rateLimit = require('express-rate-limit');
const { RateLimitError } = require('./errorHandler');

// Simple in-memory rate limiting
const rateLimitMiddleware = rateLimit({
  windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS) || 60000, // 1 minute
  max: parseInt(process.env.RATE_LIMIT_MAX) || 60, // 60 requests per window
  message: {
    type: 'https://httpstatuses.com/429',
    title: 'Rate Limit Exceeded',
    status: 429,
    detail: 'Too many requests from this IP, please try again later.'
  },
  standardHeaders: true, // Return rate limit info in the `RateLimit-*` headers
  legacyHeaders: false, // Disable the `X-RateLimit-*` headers
  handler: (req, res, next) => {
    const error = new RateLimitError('Too many requests from this IP, please try again later.');
    next(error);
  }
});

module.exports = rateLimitMiddleware;