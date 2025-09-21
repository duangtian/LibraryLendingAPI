const { validationResult } = require('express-validator');
const { ValidationError } = require('../middleware/errorHandler');

// Validation middleware
const validate = (req, res, next) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    const validationErrors = errors.array().map(error => ({
      field: error.path,
      message: error.msg,
      value: error.value
    }));
    
    throw new ValidationError('Validation failed', validationErrors);
  }
  next();
};

// Pagination helper
const paginate = (page = 1, limit = 10) => {
  const offset = (page - 1) * limit;
  return { limit, offset };
};

// Generate pagination metadata
const getPaginationMeta = (total, page, limit) => {
  const totalPages = Math.ceil(total / limit);
  const hasNext = page < totalPages;
  const hasPrev = page > 1;
  
  return {
    page: parseInt(page),
    limit: parseInt(limit),
    total,
    totalPages,
    hasNext,
    hasPrev
  };
};

// Build WHERE clause for book search
const buildBookSearchQuery = (filters) => {
  const conditions = [];
  const params = [];
  
  if (filters.title) {
    conditions.push('title LIKE ?');
    params.push(`%${filters.title}%`);
  }
  
  if (filters.author) {
    conditions.push('author LIKE ?');
    params.push(`%${filters.author}%`);
  }
  
  if (filters.genre) {
    conditions.push('genre LIKE ?');
    params.push(`%${filters.genre}%`);
  }
  
  if (filters.availability === 'available') {
    conditions.push('available_copies > 0');
  } else if (filters.availability === 'unavailable') {
    conditions.push('available_copies = 0');
  }
  
  const whereClause = conditions.length > 0 ? `WHERE ${conditions.join(' AND ')}` : '';
  
  return { whereClause, params };
};

// Build ORDER BY clause for book search
const buildBookSortQuery = (sort) => {
  const validSortFields = ['title', 'author', 'genre', 'created_at'];
  const validSortOrders = ['asc', 'desc'];
  
  if (!sort) {
    return 'ORDER BY title ASC';
  }
  
  const [field, order = 'asc'] = sort.split(':');
  
  if (!validSortFields.includes(field) || !validSortOrders.includes(order.toLowerCase())) {
    return 'ORDER BY title ASC';
  }
  
  return `ORDER BY ${field} ${order.toUpperCase()}`;
};

module.exports = {
  validate,
  paginate,
  getPaginationMeta,
  buildBookSearchQuery,
  buildBookSortQuery
};