const sqlite3 = require('sqlite3').verbose();
const path = require('path');

const DB_PATH = process.env.DB_PATH || './database.sqlite';

class Database {
  constructor() {
    this.db = null;
  }

  async connect() {
    return new Promise((resolve, reject) => {
      this.db = new sqlite3.Database(DB_PATH, (err) => {
        if (err) {
          reject(err);
        } else {
          console.log('Connected to SQLite database');
          resolve();
        }
      });
    });
  }

  async run(sql, params = []) {
    return new Promise((resolve, reject) => {
      this.db.run(sql, params, function(err) {
        if (err) {
          reject(err);
        } else {
          resolve({ id: this.lastID, changes: this.changes });
        }
      });
    });
  }

  async get(sql, params = []) {
    return new Promise((resolve, reject) => {
      this.db.get(sql, params, (err, row) => {
        if (err) {
          reject(err);
        } else {
          resolve(row);
        }
      });
    });
  }

  async all(sql, params = []) {
    return new Promise((resolve, reject) => {
      this.db.all(sql, params, (err, rows) => {
        if (err) {
          reject(err);
        } else {
          resolve(rows);
        }
      });
    });
  }

  close() {
    if (this.db) {
      this.db.close();
    }
  }
}

const database = new Database();

async function initializeDatabase() {
  await database.connect();
  
  // Create tables
  await createTables();
  
  // Seed initial data
  await seedData();
}

async function createTables() {
  // Users table
  await database.run(`
    CREATE TABLE IF NOT EXISTS users (
      id TEXT PRIMARY KEY,
      username TEXT UNIQUE NOT NULL,
      email TEXT UNIQUE NOT NULL,
      password_hash TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // Books table
  await database.run(`
    CREATE TABLE IF NOT EXISTS books (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL,
      author TEXT NOT NULL,
      genre TEXT NOT NULL,
      isbn TEXT UNIQUE,
      total_copies INTEGER NOT NULL DEFAULT 1,
      available_copies INTEGER NOT NULL DEFAULT 1,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // Loans table
  await database.run(`
    CREATE TABLE IF NOT EXISTS loans (
      id TEXT PRIMARY KEY,
      user_id TEXT NOT NULL,
      book_id TEXT NOT NULL,
      borrowed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      due_date DATETIME NOT NULL,
      returned_at DATETIME NULL,
      status TEXT NOT NULL DEFAULT 'borrowed',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (user_id) REFERENCES users (id),
      FOREIGN KEY (book_id) REFERENCES books (id)
    )
  `);

  // Create indexes for better performance
  await database.run(`CREATE INDEX IF NOT EXISTS idx_loans_user_id ON loans (user_id)`);
  await database.run(`CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans (book_id)`);
  await database.run(`CREATE INDEX IF NOT EXISTS idx_loans_status ON loans (status)`);
  await database.run(`CREATE INDEX IF NOT EXISTS idx_books_title ON books (title)`);
  await database.run(`CREATE INDEX IF NOT EXISTS idx_books_author ON books (author)`);
  await database.run(`CREATE INDEX IF NOT EXISTS idx_books_genre ON books (genre)`);
}

async function seedData() {
  // Check if books already exist
  const existingBooks = await database.all('SELECT COUNT(*) as count FROM books');
  if (existingBooks[0].count > 0) {
    return; // Data already seeded
  }

  // Sample books data
  const books = [
    { id: '1', title: 'The Great Gatsby', author: 'F. Scott Fitzgerald', genre: 'Classic Literature', isbn: '9780743273565', total_copies: 3, available_copies: 3 },
    { id: '2', title: 'To Kill a Mockingbird', author: 'Harper Lee', genre: 'Classic Literature', isbn: '9780446310789', total_copies: 2, available_copies: 2 },
    { id: '3', title: '1984', author: 'George Orwell', genre: 'Dystopian Fiction', isbn: '9780451524935', total_copies: 4, available_copies: 4 },
    { id: '4', title: 'Pride and Prejudice', author: 'Jane Austen', genre: 'Romance', isbn: '9780141439518', total_copies: 2, available_copies: 2 },
    { id: '5', title: 'The Catcher in the Rye', author: 'J.D. Salinger', genre: 'Coming-of-age Fiction', isbn: '9780316769488', total_copies: 3, available_copies: 3 },
    { id: '6', title: 'Harry Potter and the Sorcerer\'s Stone', author: 'J.K. Rowling', genre: 'Fantasy', isbn: '9780439708180', total_copies: 5, available_copies: 5 },
    { id: '7', title: 'The Lord of the Rings', author: 'J.R.R. Tolkien', genre: 'Fantasy', isbn: '9780544003415', total_copies: 3, available_copies: 3 },
    { id: '8', title: 'Dune', author: 'Frank Herbert', genre: 'Science Fiction', isbn: '9780441172719', total_copies: 2, available_copies: 2 },
    { id: '9', title: 'The Hobbit', author: 'J.R.R. Tolkien', genre: 'Fantasy', isbn: '9780547928227', total_copies: 4, available_copies: 4 },
    { id: '10', title: 'Brave New World', author: 'Aldous Huxley', genre: 'Dystopian Fiction', isbn: '9780060850524', total_copies: 2, available_copies: 2 }
  ];

  for (const book of books) {
    await database.run(
      'INSERT INTO books (id, title, author, genre, isbn, total_copies, available_copies) VALUES (?, ?, ?, ?, ?, ?, ?)',
      [book.id, book.title, book.author, book.genre, book.isbn, book.total_copies, book.available_copies]
    );
  }

  console.log('Database seeded with sample books');
}

module.exports = {
  database,
  initializeDatabase
};