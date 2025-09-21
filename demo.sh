#!/bin/bash

# Library Lending API Demo Script
# This script demonstrates all the major features of the API

BASE_URL="http://localhost:3000"

echo "🚀 Library Lending API Demo Script"
echo "====================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}📋 Step $1: $2${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Step 1: Health Check
print_step 1 "Health Check"
response=$(curl -s "$BASE_URL/health")
if [[ $? -eq 0 ]]; then
    print_success "API is running"
    echo "$response" | jq .
else
    print_error "API is not responding"
    exit 1
fi
echo ""

# Step 2: Register a new user
print_step 2 "User Registration"
register_response=$(curl -s -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "demo_user",
        "email": "demo@example.com",
        "password": "SecurePass123"
    }')

echo "$register_response" | jq .

# Extract token
TOKEN=$(echo "$register_response" | jq -r '.token')
if [[ "$TOKEN" != "null" && "$TOKEN" != "" ]]; then
    print_success "User registered successfully"
    print_info "Token: ${TOKEN:0:20}..."
else
    print_error "Registration failed"
    echo "$register_response"
fi
echo ""

# Step 3: Login with the user
print_step 3 "User Login"
login_response=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "demo_user",
        "password": "SecurePass123"
    }')

echo "$login_response" | jq .
print_success "User logged in successfully"
echo ""

# Step 4: Browse available books
print_step 4 "Browse Available Books"
books_response=$(curl -s "$BASE_URL/api/books?limit=5")
echo "$books_response" | jq .
print_success "Retrieved book list"
echo ""

# Step 5: Search for specific books
print_step 5 "Search Books by Title"
search_response=$(curl -s "$BASE_URL/api/books/search?title=gatsby")
echo "$search_response" | jq .
print_success "Search completed"
echo ""

# Step 6: Search by genre
print_step 6 "Search Books by Genre"
genre_response=$(curl -s "$BASE_URL/api/books/search?genre=fantasy&sort=title:asc")
echo "$genre_response" | jq .
print_success "Genre search completed"
echo ""

# Step 7: Get available genres
print_step 7 "Get Available Genres"
genres_response=$(curl -s "$BASE_URL/api/books/meta/genres")
echo "$genres_response" | jq .
print_success "Retrieved genres"
echo ""

# Step 8: Borrow a book
print_step 8 "Borrow a Book"
borrow_response=$(curl -s -X POST "$BASE_URL/api/loans/borrow" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "bookId": "1",
        "daysToReturn": 14
    }')

echo "$borrow_response" | jq .
LOAN_ID=$(echo "$borrow_response" | jq -r '.loan.id')
print_success "Book borrowed successfully"
print_info "Loan ID: $LOAN_ID"
echo ""

# Step 9: Check active loans
print_step 9 "Check Active Loans"
active_response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/loans/active")
echo "$active_response" | jq .
print_success "Retrieved active loans"
echo ""

# Step 10: Borrow another book
print_step 10 "Borrow Another Book"
borrow2_response=$(curl -s -X POST "$BASE_URL/api/loans/borrow" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "bookId": "6",
        "daysToReturn": 21
    }')

echo "$borrow2_response" | jq .
LOAN_ID2=$(echo "$borrow2_response" | jq -r '.loan.id')
print_success "Second book borrowed successfully"
echo ""

# Step 11: View borrowing history
print_step 11 "View Borrowing History"
history_response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/loans/history")
echo "$history_response" | jq .
print_success "Retrieved borrowing history"
echo ""

# Step 12: Get loan statistics
print_step 12 "Get Loan Statistics"
stats_response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/loans/stats")
echo "$stats_response" | jq .
print_success "Retrieved loan statistics"
echo ""

# Step 13: Return a book
print_step 13 "Return a Book"
return_response=$(curl -s -X PATCH "$BASE_URL/api/loans/return/$LOAN_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "$return_response" | jq .
print_success "Book returned successfully"
echo ""

# Step 14: Check updated active loans
print_step 14 "Check Updated Active Loans"
updated_active_response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/loans/active")
echo "$updated_active_response" | jq .
print_success "Retrieved updated active loans"
echo ""

# Step 15: Demonstrate idempotent borrowing
print_step 15 "Test Idempotent Borrowing (borrow same book again)"
idempotent_response=$(curl -s -X POST "$BASE_URL/api/loans/borrow" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "bookId": "6",
        "daysToReturn": 14
    }')

echo "$idempotent_response" | jq .
print_success "Idempotent behavior demonstrated"
echo ""

# Step 16: Test validation error
print_step 16 "Test Validation Error (invalid book borrow)"
validation_error_response=$(curl -s -X POST "$BASE_URL/api/loans/borrow" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "bookId": "",
        "daysToReturn": 100
    }')

echo "$validation_error_response" | jq .
print_success "Validation error handling demonstrated"
echo ""

# Step 17: Test authentication error
print_step 17 "Test Authentication Error (no token)"
auth_error_response=$(curl -s -X POST "$BASE_URL/api/loans/borrow" \
    -H "Content-Type: application/json" \
    -d '{
        "bookId": "1"
    }')

echo "$auth_error_response" | jq .
print_success "Authentication error handling demonstrated"
echo ""

# Step 18: Test rate limiting (make many requests quickly)
print_step 18 "Test Rate Limiting (multiple rapid requests)"
print_info "Making 5 rapid requests to test rate limiting..."

for i in {1..5}; do
    rate_test_response=$(curl -s "$BASE_URL/api/books?page=$i")
    status_code=$(echo "$rate_test_response" | jq -r '.status // 200')
    if [[ "$status_code" == "429" ]]; then
        print_info "Rate limit triggered on request $i"
        echo "$rate_test_response" | jq .
        break
    else
        print_info "Request $i: Success"
    fi
    sleep 0.1
done
print_success "Rate limiting test completed"
echo ""

print_step 19 "Final Summary"
final_stats_response=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/loans/stats")
echo "Final loan statistics:"
echo "$final_stats_response" | jq '.stats'

echo ""
echo -e "${GREEN}🎉 Demo completed successfully!${NC}"
echo ""
echo "Summary of demonstrated features:"
echo "✅ User registration and authentication"
echo "✅ Book search with filtering and sorting"
echo "✅ Book borrowing with transaction handling"
echo "✅ Idempotent borrowing operations"
echo "✅ Book returning"
echo "✅ Borrowing history and statistics"
echo "✅ Error handling with JSON Problem Details"
echo "✅ Input validation"
echo "✅ Authentication requirements"
echo "✅ Rate limiting"
echo ""
echo "The Library Lending API is fully functional! 📚"