import 'package:flutter/foundation.dart';
import '../models/user_model.dart';
import '../models/auth_models.dart';
import '../repositories/auth_repository.dart';

class AuthProvider extends ChangeNotifier {
  final AuthRepository _authRepository = AuthRepository();

  User? _user;
  bool _isLoading = false;
  String? _errorMessage;

  User? get user => _user;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;
  bool get isAuthenticated => _user != null;

  AuthProvider() {
    _loadSavedUser();
  }

  Future<void> _loadSavedUser() async {
    try {
      print('🔄 [AuthProvider] Loading saved user...');
      _isLoading = true;
      notifyListeners();

      _user = await _authRepository.getSavedUser();
      if (_user != null) {
        print('✅ [AuthProvider] Saved user loaded: ${_user!.username}');
      } else {
        print('ℹ️ [AuthProvider] No saved user found');
      }

      _isLoading = false;
      notifyListeners();
    } catch (e) {
      print('❌ [AuthProvider] Error loading saved user: $e');
      _isLoading = false;
      notifyListeners();
    }
  }

  /// Login
  Future<bool> login(String usernameOrEmail, String password) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      print('🔐 [AuthProvider] Starting login process...');
      final request = UserLoginRequest(
        username: usernameOrEmail.contains('@') ? null : usernameOrEmail,
        email: usernameOrEmail.contains('@') ? usernameOrEmail : null,
        password: password,
      );

      final result = await _authRepository.login(request);

      _isLoading = false;

      if (result['success']) {
        final response = result['data'] as UserLoginResponse;
        _user = User.fromJson(response.user.toJson());
        _errorMessage = null;
        print('✅ [AuthProvider] Login successful');
        notifyListeners();
        return true;
      } else {
        _errorMessage = result['message'];
        print('❌ [AuthProvider] Login failed: $_errorMessage');
        notifyListeners();
        return false;
      }
    } catch (e, stackTrace) {
      _isLoading = false;
      _errorMessage = 'An error occurred during login';
      print('❌ [AuthProvider] Login error: $e');
      print('   Stack trace: $stackTrace');
      notifyListeners();
      return false;
    }
  }

  /// Signup
  Future<bool> signup(String username, String email) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      print('📝 [AuthProvider] Starting signup process...');
      final request = UserSignupRequest(
        username: username,
        email: email,
      );

      final result = await _authRepository.signup(request);

      _isLoading = false;

      if (result['success']) {
        _errorMessage = null;
        final response = result['data'] as UserSignupResponse;
        print('✅ [AuthProvider] Signup successful: ${response.message}');
        notifyListeners();
        return true;
      } else {
        _errorMessage = result['message'];
        print('❌ [AuthProvider] Signup failed: $_errorMessage');
        notifyListeners();
        return false;
      }
    } catch (e, stackTrace) {
      _isLoading = false;
      _errorMessage = 'An error occurred during signup';
      print('❌ [AuthProvider] Signup error: $e');
      print('   Stack trace: $stackTrace');
      notifyListeners();
      return false;
    }
  }

  /// Forgot password
  Future<bool> forgotPassword(String email) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      print('🔑 [AuthProvider] Starting forgot password process...');
      final request = ForgotPasswordRequest(email: email);

      final response = await _authRepository.forgotPassword(request);

      _isLoading = false;

      if (response.success) {
        _errorMessage = null;
        print('✅ [AuthProvider] Forgot password email sent');
        notifyListeners();
        return true;
      } else {
        _errorMessage =
            response.message ?? 'Failed to send password reset email';
        print('❌ [AuthProvider] Forgot password failed: $_errorMessage');
        notifyListeners();
        return false;
      }
    } catch (e, stackTrace) {
      _isLoading = false;
      _errorMessage = 'An error occurred';
      print('❌ [AuthProvider] Forgot password error: $e');
      print('   Stack trace: $stackTrace');
      notifyListeners();
      return false;
    }
  }

  /// Logout
  Future<void> logout() async {
    try {
      print('🚪 [AuthProvider] Logging out...');
      await _authRepository.logout();
      _user = null;
      _errorMessage = null;
      notifyListeners();
      print('✅ [AuthProvider] Logout successful');
    } catch (e) {
      print('❌ [AuthProvider] Logout error: $e');
    }
  }

  /// Clear error message
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}
