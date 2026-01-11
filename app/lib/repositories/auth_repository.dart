import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user_model.dart';
import '../models/auth_models.dart';

class AuthRepository {
  // TODO: Replace with your actual backend URL
  static const String baseUrl = "http://10.19.238.162:8080/api/v1";

  // Auth endpoints
  static const String loginEndpoint = '$baseUrl/user/signin';
  static const String signupEndpoint = '$baseUrl/user/signup';
  static const String forgotPasswordEndpoint = '$baseUrl/user/forgot-password';

  // SharedPreferences keys
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _expiresInKey = 'expires_in';
  static const String _userKey = 'user_data';

  /// Login with username/email and password
  Future<Map<String, dynamic>> login(UserLoginRequest request) async {
    try {
      print(
          '🔐 [AuthRepository] Attempting login for: ${request.username ?? request.email}');

      final response = await http
          .post(
            Uri.parse(loginEndpoint),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode(request.toJson()),
          )
          .timeout(const Duration(seconds: 30));

      print(
          '📡 [AuthRepository] Login response status: ${response.statusCode}');
      print('📡 [AuthRepository] Login response body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body);
        if (data['success'] == true && data['data'] != null && data['data']['data'] != null) {
          final loginData = data['data']['data'];
          final loginResponse = UserLoginResponse.fromJson(loginData);

          // Save token and user data
          await _saveAuthData(
              loginResponse.accessToken,
              loginResponse.refreshToken,
              loginResponse.expiresIn,
              loginResponse.user);
          print(
              '✅ [AuthRepository] Login successful for: ${loginResponse.user.username}');

          return {'success': true, 'data': loginResponse};
        } else {
          print('❌ [AuthRepository] Invalid response structure');
          return {'success': false, 'message': 'Invalid response from server'};
        }
      } else {
        String errorMessage = 'Login failed';
        try {
          final errorData = jsonDecode(response.body);
          errorMessage = errorData['message'] ?? 'Invalid credentials';
          print('❌ [AuthRepository] Login failed: $errorMessage');
        } catch (e) {
          print('❌ [AuthRepository] Login failed: HTTP ${response.statusCode} - ${response.body}');
        }
        return {'success': false, 'message': errorMessage};
      }
    } catch (e, stackTrace) {
      print('❌ [AuthRepository] Login error: $e');
      print('   Stack trace: $stackTrace');
      return {'success': false, 'message': 'Network error. Please check your connection.'};
    }
  }

  /// Sign up with username, email, and password
  Future<Map<String, dynamic>> signup(UserSignupRequest request) async {
    try {
      print(
          '📝 [AuthRepository] Attempting signup for: ${request.username} (${request.email})');

      final response = await http
          .post(
            Uri.parse(signupEndpoint),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode(request.toJson()),
          )
          .timeout(const Duration(seconds: 30));

      print(
          '📡 [AuthRepository] Signup response status: ${response.statusCode}');
      print('📡 [AuthRepository] Signup response body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body);
        final signupResponse = UserSignupResponse.fromJson(data);

        print(
            '✅ [AuthRepository] Signup successful for: ${signupResponse.username}');

        return {'success': true, 'data': signupResponse};
      } else {
        final errorData = jsonDecode(response.body);
        final errorMessage = errorData['message'] ?? 'Signup failed';
        print('❌ [AuthRepository] Signup failed: $errorMessage');
        return {'success': false, 'message': errorMessage};
      }
    } catch (e, stackTrace) {
      print('❌ [AuthRepository] Signup error: $e');
      print('   Stack trace: $stackTrace');
      return {'success': false, 'message': 'Network error. Please check your connection.'};
    }
  }

  /// Forgot password - sends email with password
  Future<AuthResponse> forgotPassword(ForgotPasswordRequest request) async {
    try {
      print(
          '🔑 [AuthRepository] Requesting password reset for: ${request.email}');

      final response = await http
          .post(
            Uri.parse(forgotPasswordEndpoint),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode(request.toJson()),
          )
          .timeout(const Duration(seconds: 30));

      print(
          '📡 [AuthRepository] Forgot password response status: ${response.statusCode}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body);
        print('✅ [AuthRepository] Password reset email sent successfully');
        return AuthResponse(
          success: true,
          message: data['message'] ?? 'Password reset email sent successfully.',
        );
      } else {
        final errorData = jsonDecode(response.body);
        print(
            '❌ [AuthRepository] Forgot password failed: ${errorData['message'] ?? 'Unknown error'}');
        return AuthResponse(
          success: false,
          message:
              errorData['message'] ?? 'Failed to send password reset email.',
        );
      }
    } catch (e, stackTrace) {
      print('❌ [AuthRepository] Forgot password error: $e');
      print('   Stack trace: $stackTrace');
      return AuthResponse(
        success: false,
        message: 'Network error. Please check your connection.',
      );
    }
  }

  /// Save authentication data to local storage
  Future<void> _saveAuthData(String accessToken, String refreshToken,
      int expiresIn, UserInfo user) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_accessTokenKey, accessToken);
      await prefs.setString(_refreshTokenKey, refreshToken);
      await prefs.setInt(_expiresInKey, expiresIn);
      await prefs.setString(_userKey, jsonEncode(user.toJson()));
      print('💾 [AuthRepository] Auth data saved to local storage');
    } catch (e) {
      print('❌ [AuthRepository] Error saving auth data: $e');
    }
  }

  /// Get saved user data
  Future<User?> getSavedUser() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final userJson = prefs.getString(_userKey);
      if (userJson != null) {
        final userData = jsonDecode(userJson);
        return User.fromJson(userData);
      }
      return null;
    } catch (e) {
      print('❌ [AuthRepository] Error getting saved user: $e');
      return null;
    }
  }

  /// Get saved access token
  Future<String?> getSavedToken() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      return prefs.getString(_accessTokenKey);
    } catch (e) {
      print('❌ [AuthRepository] Error getting saved token: $e');
      return null;
    }
  }

  /// Check if user is logged in
  Future<bool> isLoggedIn() async {
    final token = await getSavedToken();
    return token != null && token.isNotEmpty;
  }

  /// Logout - clear saved data
  Future<void> logout() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(_accessTokenKey);
      await prefs.remove(_refreshTokenKey);
      await prefs.remove(_expiresInKey);
      await prefs.remove(_userKey);
      print('🚪 [AuthRepository] User logged out, data cleared');
    } catch (e) {
      print('❌ [AuthRepository] Error during logout: $e');
    }
  }
}
