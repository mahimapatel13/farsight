class User {
  final String id;
  final String username;
  final String email;
  final String status;
  final DateTime? lastLogin;

  User({
    required this.id,
    required this.username,
    required this.email,
    required this.status,
    this.lastLogin,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? json['_id'] ?? '',
      username: json['username'] ?? '',
      email: json['email'] ?? '',
      status: json['status'] ?? '',
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

  User copyWith({
    String? id,
    String? username,
    String? email,
    String? status,
    DateTime? lastLogin,
  }) {
    return User(
      id: id ?? this.id,
      username: username ?? this.username,
      email: email ?? this.email,
      status: status ?? this.status,
      lastLogin: lastLogin ?? this.lastLogin,
    );
  }
}

