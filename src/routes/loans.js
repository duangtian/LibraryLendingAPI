const express = require('express');
const { body, param, query } = require('express-validator');
const { v4: uuidv4 } = require('uuid');

const { database } = require('../models/database');
const { authenticateToken } = require('../middleware/auth');
const { validate, paginate, getPaginationMeta } = require('../utils/helpers');
const { NotFoundError, ConflictError, ValidationError } = require('../middleware/errorHandler');

const router = express.Router();

// All loan routes require authentication
router.use(authenticateToken);

// Validation rules
const borrowValidation = [
  body('bookId')
    .notEmpty()
    .withMessage('Book ID is required'),
  body('daysToReturn')
    .optional()
    .isInt({ min: 1, max: 30 })
    .withMessage('Days to return must be between 1 and 30')
];

const returnValidation = [
  param('loanId')
    .notEmpty()
    .withMessage('Loan ID is required')
];

const historyValidation = [
  query('page')
    .optional()
    .isInt({ min: 1 })
    .withMessage('Page must be a positive integer'),
  query('limit')
    .optional()
    .isInt({ min: 1, max: 100 })
    .withMessage('Limit must be between 1 and 100'),
  query('status')
    .optional()
    .isIn(['borrowed', 'returned', 'overdue'])
    .withMessage('Status must be "borrowed", "returned", or "overdue"')
];

// Borrow a book (idempotent operation)
router.post('/borrow', borrowValidation, validate, async (req, res, next) => {
  try {
    const { bookId, daysToReturn = 14 } = req.body;
    const userId = req.user.id;

    // Check if book exists and is available
    const book = await database.get(
      'SELECT id, title, author, available_copies FROM books WHERE id = ?',
      [bookId]
    );

    if (!book) {
      throw new NotFoundError('Book');
    }

    if (book.available_copies <= 0) {
      throw new ConflictError('Book is not available for borrowing');
    }

    // Check if user already has an active loan for this book (idempotent check)
    const existingLoan = await database.get(
      'SELECT id, status, borrowed_at FROM loans WHERE user_id = ? AND book_id = ? AND status = "borrowed"',
      [userId, bookId]
    );

    if (existingLoan) {
      // Return existing loan instead of creating duplicate (idempotent behavior)
      const loanWithDetails = await database.get(`
        SELECT l.*, b.title, b.author 
        FROM loans l 
        JOIN books b ON l.book_id = b.id 
        WHERE l.id = ?
      `, [existingLoan.id]);

      return res.json({
        message: 'Book already borrowed by user',
        loan: loanWithDetails
      });
    }

    // Check if user has reached borrowing limit (e.g., 5 books)
    const activeLoanCount = await database.get(
      'SELECT COUNT(*) as count FROM loans WHERE user_id = ? AND status = "borrowed"',
      [userId]
    );

    if (activeLoanCount.count >= 5) {
      throw new ConflictError('Maximum borrowing limit reached (5 books)');
    }

    // Create loan transaction
    const loanId = uuidv4();
    const borrowedAt = new Date().toISOString();
    const dueDate = new Date(Date.now() + daysToReturn * 24 * 60 * 60 * 1000).toISOString();

    // Begin transaction
    await database.run('BEGIN TRANSACTION');

    try {
      // Create loan record
      await database.run(
        'INSERT INTO loans (id, user_id, book_id, borrowed_at, due_date, status) VALUES (?, ?, ?, ?, ?, ?)',
        [loanId, userId, bookId, borrowedAt, dueDate, 'borrowed']
      );

      // Decrease available copies
      await database.run(
        'UPDATE books SET available_copies = available_copies - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?',
        [bookId]
      );

      await database.run('COMMIT');

      // Fetch complete loan details
      const loanDetails = await database.get(`
        SELECT l.*, b.title, b.author 
        FROM loans l 
        JOIN books b ON l.book_id = b.id 
        WHERE l.id = ?
      `, [loanId]);

      res.status(201).json({
        message: 'Book borrowed successfully',
        loan: loanDetails
      });
    } catch (error) {
      await database.run('ROLLBACK');
      throw error;
    }
  } catch (error) {
    next(error);
  }
});

// Return a book
router.patch('/return/:loanId', returnValidation, validate, async (req, res, next) => {
  try {
    const { loanId } = req.params;
    const userId = req.user.id;

    // Find the loan
    const loan = await database.get(
      'SELECT * FROM loans WHERE id = ? AND user_id = ? AND status = "borrowed"',
      [loanId, userId]
    );

    if (!loan) {
      throw new NotFoundError('Active loan');
    }

    const returnedAt = new Date().toISOString();

    // Begin transaction
    await database.run('BEGIN TRANSACTION');

    try {
      // Update loan status
      await database.run(
        'UPDATE loans SET status = "returned", returned_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?',
        [returnedAt, loanId]
      );

      // Increase available copies
      await database.run(
        'UPDATE books SET available_copies = available_copies + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?',
        [loan.book_id]
      );

      await database.run('COMMIT');

      // Fetch updated loan details
      const updatedLoan = await database.get(`
        SELECT l.*, b.title, b.author 
        FROM loans l 
        JOIN books b ON l.book_id = b.id 
        WHERE l.id = ?
      `, [loanId]);

      res.json({
        message: 'Book returned successfully',
        loan: updatedLoan
      });
    } catch (error) {
      await database.run('ROLLBACK');
      throw error;
    }
  } catch (error) {
    next(error);
  }
});

// Get user's borrowing history
router.get('/history', historyValidation, validate, async (req, res, next) => {
  try {
    const userId = req.user.id;
    const { page = 1, limit = 10, status } = req.query;
    const { limit: queryLimit, offset } = paginate(parseInt(page), parseInt(limit));

    // Build query with optional status filter
    let whereClause = 'WHERE l.user_id = ?';
    let params = [userId];

    if (status) {
      if (status === 'overdue') {
        whereClause += ' AND l.status = "borrowed" AND l.due_date < datetime("now")';
      } else {
        whereClause += ' AND l.status = ?';
        params.push(status);
      }
    }

    // Get total count
    const countQuery = `SELECT COUNT(*) as total FROM loans l ${whereClause}`;
    const countResult = await database.get(countQuery, params);
    const total = countResult.total;

    // Get loans with book details
    const loansQuery = `
      SELECT 
        l.*,
        b.title,
        b.author,
        b.genre,
        CASE 
          WHEN l.status = 'borrowed' AND l.due_date < datetime('now') THEN 'overdue'
          ELSE l.status
        END as current_status
      FROM loans l
      JOIN books b ON l.book_id = b.id
      ${whereClause}
      ORDER BY l.borrowed_at DESC
      LIMIT ? OFFSET ?
    `;

    const loans = await database.all(loansQuery, [...params, queryLimit, offset]);

    const pagination = getPaginationMeta(total, page, limit);

    res.json({
      loans,
      pagination,
      filters: {
        status: status || null
      }
    });
  } catch (error) {
    next(error);
  }
});

// Get current active loans
router.get('/active', async (req, res, next) => {
  try {
    const userId = req.user.id;

    const activeLoans = await database.all(`
      SELECT 
        l.*,
        b.title,
        b.author,
        b.genre,
        CASE 
          WHEN l.due_date < datetime('now') THEN 'overdue'
          ELSE 'borrowed'
        END as current_status
      FROM loans l
      JOIN books b ON l.book_id = b.id
      WHERE l.user_id = ? AND l.status = 'borrowed'
      ORDER BY l.due_date ASC
    `, [userId]);

    res.json({
      activeLoans,
      count: activeLoans.length
    });
  } catch (error) {
    next(error);
  }
});

// Get loan statistics for user
router.get('/stats', async (req, res, next) => {
  try {
    const userId = req.user.id;

    const stats = await database.get(`
      SELECT 
        COUNT(*) as total_loans,
        COUNT(CASE WHEN status = 'borrowed' THEN 1 END) as active_loans,
        COUNT(CASE WHEN status = 'returned' THEN 1 END) as returned_loans,
        COUNT(CASE WHEN status = 'borrowed' AND due_date < datetime('now') THEN 1 END) as overdue_loans
      FROM loans 
      WHERE user_id = ?
    `, [userId]);

    res.json({ stats });
  } catch (error) {
    next(error);
  }
});

module.exports = router;