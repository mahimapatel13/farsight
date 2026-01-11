# Authentication & Backend Integration Setup

## ✅ What's Been Created

### 1. **Authentication Models** (`lib/models/`)
   - `user_model.dart` - User model with id, username, email, token
   - `auth_models.dart` - LoginRequest, SignupRequest, ForgotPasswordRequest, AuthResponse

### 2. **Authentication Repository** (`lib/repositories/auth_repository.dart`)
   - Login with username/email + password
   - Signup with username, email, password
   - Forgot password (sends email)
   - Token and user data persistence using SharedPreferences
   - All methods include emoji logging for debugging

### 3. **Authentication Provider** (`lib/providers/auth_provider.dart`)
   - State management for authentication
   - Auto-loads saved user on app start
   - Login, signup, forgot password methods
   - Error handling and loading states

### 4. **UI Screens**
   - **Login Screen** (`lib/Screens/login_screen.dart`)
     - Username/Email + Password fields
     - Forgot password button (opens dialog)
     - Navigation to signup screen
     - Beautiful modern UI matching app theme
   
   - **Signup Screen** (`lib/Screens/signup_screen.dart`)
     - Username, Email, Password, Confirm Password fields
     - Form validation
     - Navigation to login screen
     - Beautiful modern UI matching app theme

### 5. **Transaction Service** (`lib/services/transaction_service.dart`)
   - Placeholder service for backend integration
   - Methods ready for:
     - Create transaction
     - Create multiple transactions (batch)
     - Get transactions
     - Update transaction
     - Delete transaction
   - All methods include emoji logging
   - Ready to be updated with your DTO and API routes

### 6. **Updated Files**
   - `lib/main.dart` - Now shows login screen first, then home if authenticated
   - `lib/Screens/home.dart` - Shows username from AuthProvider instead of hardcoded name
   - `pubspec.yaml` - Added `http` and `shared_preferences` packages

## 🔧 Configuration Needed

### 1. **Update Backend URLs**
   In `lib/repositories/auth_repository.dart`:
   ```dart
   static const String baseUrl = 'https://your-backend-api.com/api';
   ```
   Replace with your actual backend URL.

### 2. **Update Transaction Service URLs**
   In `lib/services/transaction_service.dart`:
   ```dart
   static const String baseUrl = 'https://your-backend-api.com/api';
   ```
   Replace with your actual backend URL.

### 3. **Update API Endpoints**
   The following endpoints are expected:
   - `POST /api/auth/login` - Login endpoint
   - `POST /api/auth/signup` - Signup endpoint
   - `POST /api/auth/forgot-password` - Forgot password endpoint
   - `POST /api/transactions/create` - Create transaction
   - `GET /api/transactions` - Get transactions
   - `PUT /api/transactions/:id` - Update transaction
   - `DELETE /api/transactions/:id` - Delete transaction

### 4. **Update Transaction Service with Your DTO**
   When you provide the DTO structure:
   1. Update `createTransaction()` method to map `Add_data` to your DTO
   2. Update `getTransactions()` method to parse your response DTO
   3. Update request/response structures as needed

## 📝 Expected API Response Formats

### Login/Signup Response
```json
{
  "success": true,
  "message": "Login successful",
  "user": {
    "id": "user_id",
    "username": "john_doe",
    "email": "john@example.com"
  },
  "token": "jwt_token_here"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message here"
}
```

## 🚀 How It Works

1. **App Start**: 
   - Checks if user is logged in (saved token/user)
   - Shows login screen if not authenticated
   - Shows home screen if authenticated

2. **Login Flow**:
   - User enters username/email + password
   - Calls backend API
   - Saves token and user data locally
   - Navigates to home screen

3. **Signup Flow**:
   - User enters username, email, password
   - Backend creates account and sends email with password
   - User is automatically logged in
   - Navigates to home screen

4. **Forgot Password**:
   - User enters email
   - Backend sends email with password
   - Shows success message

5. **Home Screen**:
   - Displays username from AuthProvider
   - Shows "User" if no username available

## 🔐 Security Notes

- Tokens are stored in SharedPreferences (consider encryption for production)
- Passwords are sent over HTTPS (ensure your backend uses HTTPS)
- Token is included in Authorization header for authenticated requests

## 📦 Dependencies Added

- `http: ^1.1.0` - For API calls
- `shared_preferences: ^2.2.2` - For local storage

## 🎨 UI Features

- Modern, clean design matching app theme (Color(0xff368983))
- Form validation
- Loading states
- Error messages via SnackBar
- Password visibility toggle
- Smooth navigation

## 🐛 Debugging

All methods include emoji logging:
- 🔐 Login operations
- 📝 Signup operations
- 🔑 Password reset
- 💾 Transaction operations
- ✅ Success messages
- ❌ Error messages

Check console logs for detailed debugging information.

## 📋 Next Steps

1. Update backend URLs in `auth_repository.dart` and `transaction_service.dart`
2. Test login/signup with your backend
3. Provide DTO structure for transactions
4. Update `TransactionService` with actual DTO mapping
5. Integrate `TransactionService` with `TransactionProvider` to sync with backend
6. Test end-to-end flow

## 💡 Integration with TransactionProvider

To sync transactions with backend, update `TransactionProvider.addTransaction()`:

```dart
void addTransaction(Add_data transaction) async {
  // Save to local Hive (existing)
  _box.add(transaction);
  notifyListeners();
  
  // Also save to backend
  final token = await _authRepository.getSavedToken();
  await _transactionService.createTransaction(transaction, token: token);
}
```

