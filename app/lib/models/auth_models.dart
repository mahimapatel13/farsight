
class UserLoginRequest {
  final String? username;
  final String? email;
  final String password;

  UserLoginRequest({
    this.username,
    this.email,
    required this.password,
  });

  Map<String, dynamic> toJson() {
    return {
      'username': username,
      'email': email,
      'password': password,
    };
  }
}

class UserSignupRequest {
  final String username;
  final String email;

  UserSignupRequest({
    required this.username,
    required this.email,
  });

  Map<String, dynamic> toJson() {
    return {
      'username': username,
      'email': email,
    };
  }
}

class ForgotPasswordRequest {
  final String email;

  ForgotPasswordRequest({required this.email});

  Map<String, dynamic> toJson() {
    return {
      'email': email,
    };
  }
}

class UserSignupResponse {
  final String username;
  final String email;
  final String message;

  UserSignupResponse({
    required this.username,
    required this.email,
    required this.message,
  });

  factory UserSignupResponse.fromJson(Map<String, dynamic> json) {
    final user = json['data']['user'];
    return UserSignupResponse(
      username: user['username'],
      email: user['email'],
      message: json['message'],
    );
  }
}

class UserInfo {
  final String id;
  final String username;
  final String email;
  final String status;
  final DateTime? lastLogin;

  UserInfo({
    required this.id,
    required this.username,
    required this.email,
    required this.status,
    this.lastLogin,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'],
      username: json['username'],
      email: json['email'],
      status: json['status'],
      lastLogin: json['last_login_at'] != null ? DateTime.parse(json['last_login_at']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      'status': status,
      'last_login_at': lastLogin?.toIso8601String(),
    };
  }
}

class UserLoginResponse {
  final UserInfo user;
  final String accessToken;
  final String refreshToken;
  final int expiresIn;

  UserLoginResponse({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
    required this.expiresIn,
  });

  factory UserLoginResponse.fromJson(Map<String, dynamic> json) {
    return UserLoginResponse(
      user: UserInfo.fromJson(json['user']),
      accessToken: json['access_token'],
      refreshToken: json['refresh_token'],
      expiresIn: json['expires_in'],
    );
  }
}

class AuthResponse {
  final bool success;
  final String? message;

  AuthResponse({
    required this.success,
    this.message,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      success: json['success'] ?? false,
      message: json['message'],
    );
  }
}
