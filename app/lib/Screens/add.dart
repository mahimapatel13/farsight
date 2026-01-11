import 'package:flutter/material.dart';
import 'package:managment/data/model/add_date.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:provider/provider.dart';
import 'package:managment/providers/transaction_provider.dart';

class Add_Screen extends StatefulWidget {
  const Add_Screen({super.key});

  @override
  State<Add_Screen> createState() => _Add_ScreenState();
}

class _Add_ScreenState extends State<Add_Screen> {
  final box = Hive.box<Add_data>('data');
  DateTime date = new DateTime.now();
  String? selctedItem;
  String? selctedItemi;
  final TextEditingController expalin_C = TextEditingController();
  FocusNode ex = FocusNode();
  final TextEditingController amount_c = TextEditingController();
  FocusNode amount_ = FocusNode();
  final List<String> _item = [
    'Food',
    "Transfer",
    "Medicine",
    "Transportation",
    "Education",
    "Grooming",
    "Gifts",
    "Housing"
  ];
  final List<String> _itemei = [
    'Income',
    "Expense",
  ];
  @override
  void initState() {
    // TODO: implement initState
    super.initState();
    ex.addListener(() {
      setState(() {});
    });
    amount_.addListener(() {
      setState(() {});
    });
  }

  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      body: SafeArea(
        child: Stack(
          alignment: AlignmentDirectional.center,
          children: [
            background_container(context),
            Positioned(
              top: 120,
              child: main_container(),
            ),
          ],
        ),
      ),
    );
  }

  Container main_container() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        color: Colors.white,
      ),
      height: 550,
      width: 340,
      child: Column(
        children: [
          SizedBox(height: 50),
          name(),
          SizedBox(height: 30),
          explain(),
          SizedBox(height: 30),
          amount(),
          SizedBox(height: 30),
          How(),
          SizedBox(height: 30),
          date_time(),
          Spacer(),
          save(),
          SizedBox(height: 25),
        ],
      ),
    );
  }

  GestureDetector save() {
    return GestureDetector(
      onTap: () async {
        try {
          print('💾 [AddScreen] Save button tapped');
          print('💾 [AddScreen] Form data:');
          print('   - selctedItemi: $selctedItemi');
          print('   - selctedItem: $selctedItem');
          print('   - amount_c.text: ${amount_c.text}');
          print('   - expalin_C.text: ${expalin_C.text}');
          print('   - date: $date');

          // Validate required fields
          if (selctedItemi == null || selctedItemi!.isEmpty) {
            print(
                '❌ [AddScreen] Validation failed: Income/Expense not selected');
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Please select Income or Expense'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }

          if (selctedItem == null || selctedItem!.isEmpty) {
            print('❌ [AddScreen] Validation failed: Category not selected');
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Please select a category'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }

          if (amount_c.text.isEmpty) {
            print('❌ [AddScreen] Validation failed: Amount is empty');
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Please enter an amount'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }

          // Validate amount is a number
          if (double.tryParse(amount_c.text) == null) {
            print(
                '❌ [AddScreen] Validation failed: Invalid amount format: ${amount_c.text}');
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Please enter a valid amount'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }

          if (expalin_C.text.isEmpty) {
            print('❌ [AddScreen] Validation failed: Explanation is empty');
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Please enter an explanation'),
                backgroundColor: Colors.red,
              ),
            );
            return;
          }

          print('✅ [AddScreen] Validation passed, creating Add_data object...');
          var add = Add_data(
            selctedItemi!,
            amount_c.text,
            date,
            expalin_C.text,
            selctedItem!,
          );

          print('💾 [AddScreen] Calling TransactionProvider.addTransaction...');
          context.read<TransactionProvider>().addTransaction(add);
          print('✅ [AddScreen] Transaction added successfully');

          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Transaction saved successfully!'),
              backgroundColor: Colors.green,
            ),
          );

          Navigator.of(context).pop();
        } catch (e, stackTrace) {
          print('❌ [AddScreen] Error saving transaction:');
          print('   Error: $e');
          print('   Stack trace: $stackTrace');
          print('   Error type: ${e.runtimeType}');

          String errorMessage = 'Failed to save transaction';
          if (e.toString().contains('null')) {
            errorMessage = 'Please fill in all required fields';
          } else if (e.toString().contains('Hive')) {
            errorMessage = 'Database error: ${e.toString()}';
          } else {
            errorMessage = 'Error: ${e.toString()}';
          }

          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(errorMessage),
                backgroundColor: Colors.red,
                duration: const Duration(seconds: 4),
              ),
            );
          }
        }
      },
      child: Container(
        alignment: Alignment.center,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(15),
          color: Color.fromARGB(255, 101, 76, 116),
        ),
        width: 120,
        height: 50,
        child: Text(
          'Save',
          style: TextStyle(
            fontFamily: 'f',
            fontWeight: FontWeight.w600,
            color: Colors.white,
            fontSize: 17,
          ),
        ),
      ),
    );
  }

  Widget date_time() {
    return Container(
      alignment: Alignment.bottomLeft,
      decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          border: Border.all(width: 2, color: Color(0xffC5C5C5))),
      width: 300,
      child: TextButton(
        onPressed: () async {
          DateTime? newDate = await showDatePicker(
              context: context,
              initialDate: date,
              firstDate: DateTime(2020),
              lastDate: DateTime(2100));
          if (newDate == Null) return;
          setState(() {
            date = newDate!;
          });
        },
        child: Text(
          'Date : ${date.year} / ${date.day} / ${date.month}',
          style: TextStyle(
            fontSize: 15,
            color: Colors.black,
          ),
        ),
      ),
    );
  }

  Padding How() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 15),
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: 15),
        width: 300,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            width: 2,
            color: Color(0xffC5C5C5),
          ),
        ),
        child: DropdownButton<String>(
          value: selctedItemi,
          onChanged: ((value) {
            setState(() {
              selctedItemi = value!;
            });
          }),
          items: _itemei
              .map((e) => DropdownMenuItem(
                    child: Container(
                      alignment: Alignment.center,
                      child: Row(
                        children: [
                          Text(
                            e,
                            style: TextStyle(fontSize: 18),
                          )
                        ],
                      ),
                    ),
                    value: e,
                  ))
              .toList(),
          selectedItemBuilder: (BuildContext context) => _itemei
              .map((e) => Row(
                    children: [Text(e)],
                  ))
              .toList(),
          hint: Padding(
            padding: const EdgeInsets.only(top: 12),
            child: Text(
              'How',
              style: TextStyle(color: Colors.grey),
            ),
          ),
          dropdownColor: Colors.white,
          isExpanded: true,
          underline: Container(),
        ),
      ),
    );
  }

  Padding amount() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: TextField(
        keyboardType: TextInputType.number,
        focusNode: amount_,
        controller: amount_c,
        decoration: InputDecoration(
          contentPadding: EdgeInsets.symmetric(horizontal: 15, vertical: 15),
          labelText: 'Amount',
          labelStyle: TextStyle(fontSize: 17, color: Colors.grey.shade500),
          enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(width: 2, color: Color(0xffC5C5C5))),
          focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(width: 2, color: Color(0xff7b1fa2))),
        ),
      ),
    );
  }

  Padding explain() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: TextField(
        focusNode: ex,
        controller: expalin_C,
        decoration: InputDecoration(
          contentPadding: EdgeInsets.symmetric(horizontal: 15, vertical: 15),
          labelText: 'Description',
          labelStyle: TextStyle(fontSize: 17, color: Colors.grey.shade500),
          enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(width: 2, color: Color(0xffC5C5C5))),
          focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(10),
              borderSide: BorderSide(width: 2, color: Color(0xff7b1fa2))),
        ),
      ),
    );
  }

  Padding name() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 15),
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: 15),
        width: 300,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          border: Border.all(
            width: 2,
            color: Color(0xffC5C5C5),
          ),
        ),
        child: DropdownButton<String>(
          value: selctedItem,
          onChanged: ((value) {
            setState(() {
              selctedItem = value!;
            });
          }),
          items: _item
              .map((e) => DropdownMenuItem(
                    child: Container(
                      alignment: Alignment.center,
                      child: Row(
                        children: [
                          Container(
                            width: 40,
                            child: Image.asset('images/${e}.png'),
                          ),
                          SizedBox(width: 10),
                          Text(
                            e,
                            style: TextStyle(fontSize: 18),
                          )
                        ],
                      ),
                    ),
                    value: e,
                  ))
              .toList(),
          selectedItemBuilder: (BuildContext context) => _item
              .map((e) => Row(
                    children: [
                      Container(
                        width: 42,
                        child: Image.asset('images/${e}.png'),
                      ),
                      SizedBox(width: 5),
                      Text(e)
                    ],
                  ))
              .toList(),
          hint: Padding(
            padding: const EdgeInsets.only(top: 12),
            child: Text(
              'Item Type',
              style: TextStyle(color: Colors.grey),
            ),
          ),
          dropdownColor: Colors.white,
          isExpanded: true,
          underline: Container(),
        ),
      ),
    );
  }

  Column background_container(BuildContext context) {
    return Column(
      children: [
        Container(
          width: double.infinity,
          height: 240,
          decoration: BoxDecoration(
            color: Color.fromARGB(255, 101, 76, 116),
            borderRadius: BorderRadius.only(
              bottomLeft: Radius.circular(20),
              bottomRight: Radius.circular(20),
            ),
          ),
          child: Column(
            children: [
              SizedBox(height: 40),
              Padding(
                padding: EdgeInsets.symmetric(horizontal: 15),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    GestureDetector(
                      onTap: () {
                        Navigator.of(context).pop();
                      },
                      child: Icon(Icons.arrow_back, color: Colors.white),
                    ),
                    Text(
                      'Adding',
                      style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w600,
                          color: Colors.white),
                    ),
                    Icon(
                      Icons.attach_file_outlined,
                      color: Colors.white,
                    )
                  ],
                ),
              )
            ],
          ),
        ),
      ],
    );
  }
}
