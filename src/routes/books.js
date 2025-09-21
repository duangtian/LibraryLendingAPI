const express = require('express');
const { query, param } = require('express-validator');

const { database } = require('../models/database');
const { authenticateToken, optionalAuth } = require('../middleware/auth');
const { validate, paginate, getPaginationMeta, buildBookSearchQuery, buildBookSortQuery } = require('../utils/helpers');
const { NotFoundError } = require('../middleware/errorHandler');

const router = express.Router();

// Validation rules
const searchValidation = [
  query('page')
    .optional()
    .isInt({ min: 1 })
    .withMessage('Page must be a positive integer'),
  query('limit')
    .optional()
    .isInt({ min: 1, max: 100 })
    .withMessage('Limit must be between 1 and 100'),
  query('sort')
    .optional()
    .matches(/^(title|author|genre|created_at):(asc|desc)$/)
    .withMessage('Sort must be in format "field:order" where field is title|author|genre|created_at and order is asc|desc'),
  query('availability')
    .optional()
    .isIn(['available', 'unavailable'])
    .withMessage('Availability must be "available" or "unavailable"')
];

const bookIdValidation = [
  param('id')
    .notEmpty()
    .withMessage('Book ID is required')
];

// Search books with filtering, sorting, and pagination
router.get('/search', searchValidation, validate, async (req, res, next) => {
  try {
    const {
      title,
      author,
      genre,
      availability,
      sort,
      page = 1,
      limit = 10
    } = req.query;

    // Build search query
    const filters = { title, author, genre, availability };
    const { whereClause, params } = buildBookSearchQuery(filters);
    const orderClause = buildBookSortQuery(sort);
    const { limit: queryLimit, offset } = paginate(parseInt(page), parseInt(limit));

    // Get total count
    const countQuery = `SELECT COUNT(*) as total FROM books ${whereClause}`;
    const countResult = await database.get(countQuery, params);
    const total = countResult.total;

    // Get books
    const searchQuery = `
      SELECT id, title, author, genre, isbn, total_copies, available_copies, created_at, updated_at 
      FROM books 
      ${whereClause} 
      ${orderClause} 
      LIMIT ? OFFSET ?
    `;
    
    const books = await database.all(searchQuery, [...params, queryLimit, offset]);

    // Generate pagination metadata
    const pagination = getPaginationMeta(total, page, limit);

    res.json({
      books,
      pagination,
      filters: {
        title: title || null,
        author: author || null,
        genre: genre || null,
        availability: availability || null,
        sort: sort || null
      }
    });
  } catch (error) {
    next(error);
  }
});

// Get all books (simplified endpoint)
router.get('/', async (req, res, next) => {
  try {
    const { page = 1, limit = 10 } = req.query;
    const { limit: queryLimit, offset } = paginate(parseInt(page), parseInt(limit));

    // Get total count
    const countResult = await database.get('SELECT COUNT(*) as total FROM books');
    const total = countResult.total;

    // Get books
    const books = await database.all(
      'SELECT id, title, author, genre, isbn, total_copies, available_copies, created_at, updated_at FROM books ORDER BY title ASC LIMIT ? OFFSET ?',
      [queryLimit, offset]
    );

    const pagination = getPaginationMeta(total, page, limit);

    res.json({
      books,
      pagination
    });
  } catch (error) {
    next(error);
  }
});

// Get book by ID
router.get('/:id', bookIdValidation, validate, async (req, res, next) => {
  try {
    const { id } = req.params;

    const book = await database.get(
      'SELECT id, title, author, genre, isbn, total_copies, available_copies, created_at, updated_at FROM books WHERE id = ?',
      [id]
    );

    if (!book) {
      throw new NotFoundError('Book');
    }

    res.json({ book });
  } catch (error) {
    next(error);
  }
});

// Get available genres
router.get('/meta/genres', async (req, res, next) => {
  try {
    const genres = await database.all('SELECT DISTINCT genre FROM books ORDER BY genre ASC');
    res.json({
      genres: genres.map(g => g.genre)
    });
  } catch (error) {
    next(error);
  }
});

// Get available authors
router.get('/meta/authors', async (req, res, next) => {
  try {
    const authors = await database.all('SELECT DISTINCT author FROM books ORDER BY author ASC');
    res.json({
      authors: authors.map(a => a.author)
    });
  } catch (error) {
    next(error);
  }
});

module.exports = router;