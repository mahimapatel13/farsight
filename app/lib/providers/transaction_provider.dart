import 'package:flutter/foundation.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:managment/data/model/add_date.dart';

class TransactionProvider extends ChangeNotifier {
  final Box<Add_data> _box = Hive.box<Add_data>('data');

  Box<Add_data> get box => _box;

  int get total => _calculateTotal();
  int get income => _calculateIncome();
  int get expenses => _calculateExpenses();

  List<Add_data> get today => _getToday();
  List<Add_data> get week => _getWeek();
  List<Add_data> get month => _getMonth();
  List<Add_data> get year => _getYear();

  /// Safely parses an amount string to an integer, handling decimals and invalid values
  int _safeParseAmount(String amountStr, {int index = -1}) {
    try {
      if (amountStr.isEmpty || amountStr.trim().isEmpty) {
        print(
            '⚠️ [TransactionProvider] Empty amount string at index $index, using 0');
        return 0;
      }

      // Try parsing as double first (handles decimals), then convert to int
      double? parsedDouble = double.tryParse(amountStr.trim());
      if (parsedDouble == null) {
        print(
            '❌ [TransactionProvider] Invalid amount format: "$amountStr" at index $index, using 0');
        return 0;
      }

      // Round to nearest integer
      int result = parsedDouble.round();
      return result;
    } catch (e) {
      print(
          '❌ [TransactionProvider] Error parsing amount "$amountStr" at index $index: $e, using 0');
      return 0;
    }
  }

  int _calculateTotal() {
    try {
      var history = _box.values.toList();
      int total = 0;

      for (var i = 0; i < history.length; i++) {
        try {
          int amount = _safeParseAmount(history[i].amount, index: i);
          if (history[i].IN == 'Income') {
            total += amount;
          } else {
            total -= amount;
          }
        } catch (e) {
          print(
              '❌ [TransactionProvider] Error processing transaction $i in _calculateTotal: $e');
          print(
              '   Transaction: IN=${history[i].IN}, Amount=${history[i].amount}');
        }
      }

      return total;
    } catch (e, stackTrace) {
      print('❌ [TransactionProvider] Critical error in _calculateTotal:');
      print('   Error: $e');
      print('   Stack trace: $stackTrace');
      return 0; // Return 0 on error to prevent app crash
    }
  }

  int _calculateIncome() {
    try {
      var history = _box.values.toList();
      int total = 0;

      for (var i = 0; i < history.length; i++) {
        try {
          if (history[i].IN == 'Income') {
            int amount = _safeParseAmount(history[i].amount, index: i);
            total += amount;
          }
        } catch (e) {
          print(
              '❌ [TransactionProvider] Error processing transaction $i in _calculateIncome: $e');
          print(
              '   Transaction: IN=${history[i].IN}, Amount=${history[i].amount}');
        }
      }

      return total;
    } catch (e, stackTrace) {
      print('❌ [TransactionProvider] Critical error in _calculateIncome:');
      print('   Error: $e');
      print('   Stack trace: $stackTrace');
      return 0; // Return 0 on error to prevent app crash
    }
  }

  int _calculateExpenses() {
    try {
      var history = _box.values.toList();
      int total = 0;

      for (var i = 0; i < history.length; i++) {
        try {
          if (history[i].IN != 'Income') {
            int amount = _safeParseAmount(history[i].amount, index: i);
            total += amount;
          }
        } catch (e) {
          print(
              '❌ [TransactionProvider] Error processing transaction $i in _calculateExpenses: $e');
          print(
              '   Transaction: IN=${history[i].IN}, Amount=${history[i].amount}');
        }
      }

      return total;
    } catch (e, stackTrace) {
      print('❌ [TransactionProvider] Critical error in _calculateExpenses:');
      print('   Error: $e');
      print('   Stack trace: $stackTrace');
      return 0; // Return 0 on error to prevent app crash
    }
  }

  List<Add_data> _getToday() {
    List<Add_data> a = [];
    var history = _box.values.toList();
    DateTime date = DateTime.now();
    for (var i = 0; i < history.length; i++) {
      if (history[i].datetime.day == date.day &&
          history[i].datetime.month == date.month &&
          history[i].datetime.year == date.year) {
        a.add(history[i]);
      }
    }
    return a;
  }

  List<Add_data> _getWeek() {
    List<Add_data> a = [];
    DateTime date = DateTime.now();
    var history = _box.values.toList();
    for (var i = 0; i < history.length; i++) {
      if (date.difference(history[i].datetime).inDays <= 7) {
        a.add(history[i]);
      }
    }
    return a;
  }

  List<Add_data> _getMonth() {
    List<Add_data> a = [];
    var history = _box.values.toList();
    DateTime date = DateTime.now();
    for (var i = 0; i < history.length; i++) {
      if (history[i].datetime.month == date.month &&
          history[i].datetime.year == date.year) {
        a.add(history[i]);
      }
    }
    return a;
  }

  List<Add_data> _getYear() {
    List<Add_data> a = [];
    var history = _box.values.toList();
    DateTime date = DateTime.now();
    for (var i = 0; i < history.length; i++) {
      if (history[i].datetime.year == date.year) {
        a.add(history[i]);
      }
    }
    return a;
  }

  void addTransaction(Add_data transaction) {
    try {
      print('📝 [TransactionProvider] Starting to add transaction...');
      print('📝 [TransactionProvider] Transaction details:');
      print('   - IN: ${transaction.IN}');
      print('   - Amount: ${transaction.amount}');
      print('   - Date: ${transaction.datetime}');
      print('   - Explain: ${transaction.explain}');
      print('   - Name: ${transaction.name}');

      // Validate transaction data
      if (transaction.IN.isEmpty) {
        throw Exception('Transaction type (IN) cannot be empty');
      }
      if (transaction.amount.isEmpty) {
        throw Exception('Amount cannot be empty');
      }
      if (double.tryParse(transaction.amount) == null) {
        throw Exception('Amount must be a valid number: ${transaction.amount}');
      }
      if (transaction.name.isEmpty) {
        throw Exception('Name cannot be empty');
      }
      if (transaction.explain.isEmpty) {
        throw Exception('Explain cannot be empty');
      }

      // Check if box is open
      if (!_box.isOpen) {
        throw Exception('Hive box is not open');
      }

      print('📝 [TransactionProvider] Validation passed, adding to box...');
      _box.add(transaction);
      print('✅ [TransactionProvider] Transaction added successfully');

      notifyListeners();
      print('✅ [TransactionProvider] Listeners notified');
    } catch (e, stackTrace) {
      print('❌ [TransactionProvider] Error adding transaction:');
      print('   Error: $e');
      print('   Stack trace: $stackTrace');
      print(
          '   Transaction data: IN=${transaction.IN}, Amount=${transaction.amount}, Name=${transaction.name}');
      rethrow; // Re-throw to let the UI handle it
    }
  }

  void deleteTransaction(int index) {
    try {
      print('🗑️ [TransactionProvider] Deleting transaction at index $index');
      _box.deleteAt(index);
      notifyListeners();
      print('✅ [TransactionProvider] Transaction deleted successfully');
    } catch (e, stackTrace) {
      print(
          '❌ [TransactionProvider] Error deleting transaction at index $index:');
      print('   Error: $e');
      print('   Stack trace: $stackTrace');
      rethrow;
    }
  }
}
