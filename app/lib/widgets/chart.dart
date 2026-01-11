import 'package:flutter/material.dart';
import 'package:managment/data/model/add_date.dart';
import 'package:managment/data/utlity.dart';
import 'package:syncfusion_flutter_charts/charts.dart';

class Chart extends StatefulWidget {
  int indexx;
  Chart({Key? key, required this.indexx}) : super(key: key);

  @override
  State<Chart> createState() => _ChartState();
}

class _ChartState extends State<Chart> {
  List<Add_data>? a;
  Map<String, int> categoryExpenses = {};
  int totalIncome = 0;
  int totalExpenses = 0;

  @override
  Widget build(BuildContext context) {
    switch (widget.indexx) {
      case 0:
        a = today();
        break;
      case 1:
        a = week();
        break;
      case 2:
        a = month();
        break;
      case 3:
        a = year();
        break;
      default:
        a = [];
    }

    _calculateData();

    return Column(
      children: [
        Container(
          width: double.infinity,
          height: 300,
          child: SfCartesianChart(
            primaryXAxis: CategoryAxis(
              labelRotation: 45,
            ),
            primaryYAxis: NumericAxis(),
            series: <BarSeries<CategoryExpense, String>>[
              BarSeries<CategoryExpense, String>(
                dataSource: categoryExpenses.entries.map((entry) {
                  return CategoryExpense(entry.key, entry.value);
                }).toList(),
                xValueMapper: (CategoryExpense data, _) => data.category,
                yValueMapper: (CategoryExpense data, _) => data.expense,
                color: Color.fromARGB(255, 101, 76, 116),
                dataLabelSettings: DataLabelSettings(isVisible: true),
              )
            ],
          ),
        ),
        SizedBox(height: 20),
        Text(
          'Savings: \$${totalIncome - totalExpenses}',
          style: TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: (totalIncome - totalExpenses) >= 0 ? Colors.green : Colors.red,
          ),
        ),
      ],
    );
  }

  void _calculateData() {
    categoryExpenses.clear();
    totalIncome = 0;
    totalExpenses = 0;

    if (a == null) return;

    for (var transaction in a!) {
      String cleanAmountStr = transaction.amount.replaceAll(RegExp(r'[^0-9]'), '');
      if (cleanAmountStr.isEmpty) cleanAmountStr = '0';
      int amount = int.parse(cleanAmountStr);

      if (transaction.IN == 'Income') {
        totalIncome += amount;
      } else {
        totalExpenses += amount;
        String category = transaction.name;
        categoryExpenses[category] = (categoryExpenses[category] ?? 0) + amount;
      }
    }
  }
}

class CategoryExpense {
  final String category;
  final int expense;

  CategoryExpense(this.category, this.expense);
}

class SalesData {
  SalesData(this.year, this.sales);
  final String year;
  final int sales;
}
