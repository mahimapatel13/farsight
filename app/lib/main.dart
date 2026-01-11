import 'package:flutter/material.dart';
import 'package:managment/widgets/bottomnavigationbar.dart';
import 'package:managment/Screens/login_screen.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:provider/provider.dart';

import 'data/model/add_date.dart';
import 'providers/transaction_provider.dart';
import 'providers/auth_provider.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Hive.initFlutter();
  Hive.registerAdapter(AdddataAdapter());
  await Hive.openBox<Add_data>('data');
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (context) => AuthProvider()),
        ChangeNotifierProvider(create: (context) => TransactionProvider()),
      ],
      child: MaterialApp(
        debugShowCheckedModeBanner: false,
        home: const AuthWrapper(),
      ),
    );
  }
}

class AuthWrapper extends StatelessWidget {
  const AuthWrapper({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, authProvider, child) {
        // Show loading while checking auth state
        if (authProvider.isLoading) {
          return const Scaffold(
            body: Center(
              child: CircularProgressIndicator(
                color: Color(0xff7b1fa2),
              ),
            ),
          );
        }

        // Show login screen if not authenticated, otherwise show home
        if (authProvider.isAuthenticated) {
          return const Bottom();
        } else {
          return const LoginScreen();
        }
      },
    );
  }
}
