import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'dart:io';
import 'package:managment/data/model/add_date.dart';
// import 'package:flutter_tesseract_ocr/flutter_tesseract_ocr.dart';
import 'package:provider/provider.dart';
import 'package:managment/providers/transaction_provider.dart';

import 'package:google_generative_ai/google_generative_ai.dart';

class ReceiptScannerScreen extends StatefulWidget {
  const ReceiptScannerScreen({super.key});

  @override
  State<ReceiptScannerScreen> createState() => _ReceiptScannerScreenState();
}

class _ReceiptScannerScreenState extends State<ReceiptScannerScreen> {
  File? _image;
  List<ReceiptItem> extractedItems = [];
  bool isProcessing = false;
  final ImagePicker _picker = ImagePicker();

  final List<String> _categories = [
    'Food',
    "Transfer",
    "Medicine",
    "Transportation",
    "Education",
    "Grooming",
    "Gifts",
    "Housing"
  ];

  Future<void> pickImage(ImageSource source) async {
    try {
      final XFile? pickedFile = await _picker.pickImage(source: source);
      if (pickedFile != null) {
        setState(() {
          _image = File(pickedFile.path);
          extractedItems = [];
        });
        await extractTextFromImage();
      }
    } catch (e) {
      print('❌ [ReceiptScanner] Error picking image: $e');
    }
  }

  Future<void> extractTextFromImage() async {
    if (_image == null) return;

    setState(() => isProcessing = true);

    try {
      final model = GenerativeModel(
          model: 'gemini-2.5-flash',
          apiKey: 'AIzaSyBDU45fR3pC7naikYhxnMdQTWaSYBJjikA');
      final imageBytes = await _image!.readAsBytes();
      final prompt =
          'You are a receipt scanner AI specialized in Indian e-bills. Extract only purchased item names and their exact amounts. Ignore headers, totals, taxes, addresses, and irrelevant text. Output one item per line in the format: Item Name — Amount. Examples: Shaft — 5310, Brake Disc — 1534. Do not add explanations or extra text.';
      // final prompt = 'You are an OCR and receipt-parsing system. For every receipt image provided, extract only the purchased item names and their corresponding amounts. Output text only. No summaries, no explanations, no formatting other than one item per line. For each line, use the structure: Item Name — Amount. If an amount is unclear, write UNKNOWN. Process every item visible in the image.';
      final content = [
        Content.multi([TextPart(prompt), DataPart('image/jpeg', imageBytes)])
      ];
      final response = await model.generateContent(content);
      String extractedText = response.text ?? '';
      print("📄 [ReceiptScanner] Extracted text:\n$extractedText");
      parseTextForItems(extractedText);
    } catch (e, stackTrace) {
      print('❌ [ReceiptScanner] Error extracting text: $e');
      print('   Stack trace: $stackTrace');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error processing image: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }

    if (!mounted) return;

    setState(() => isProcessing = false);
  }

  // ===================== RECEIPT PARSING LOGIC ===================== //
  void parseTextForItems(String text) {
    final List<ReceiptItem> items = [];
    final lines = text.split('\n');

    // City / area blacklist to prevent selecting addresses
    final blockedKeywords = [
    ];

    bool isAddress(String name) {
      for (final k in blockedKeywords) {
        if (name.toLowerCase().contains(k)) return true;
      }
      return false;
    }

    for (var rawLine in lines) {
      // Normalize
      String line = rawLine.trim().replaceAll(RegExp(r'\s+'), ' ');
      if (line.isEmpty) continue;

      // Assume format: Item Name — Price
      final parts = line.split(' — ');
      if (parts.length == 2) {
        String itemName = parts[0].trim();
        String priceStr = parts[1].trim().replaceAll(',', '');

        double? price = double.tryParse(priceStr);
        if (price != null &&
            price >= 5 &&
            price <= 2000000 &&
            priceStr.replaceAll('.', '').length < 10) {
          itemName = itemName
              .replaceAll(RegExp(r'[^\w\s-]'), ' ')
              .replaceAll(RegExp(r'\s+'), ' ')
              .trim();

          if (_isValidItemName(itemName) && !isAddress(itemName)) {
            items.add(ReceiptItem(
              name: itemName,
              amount: price.toStringAsFixed(2),
              category: "Food",
            ));
          }
        }
      }
    }

    // Debug summary
    print("\n📋 ===== READY =====");
    print("📄 Lines: ${lines.length}");
    print("🛍 Items Detected: ${items.length}");
    for (int i = 0; i < items.length; i++) {
      print(" ${i + 1}. ${items[i].name} → ₹${items[i].amount}");
    }
    print("================================\n");

    setState(() => extractedItems = items);

    // Popup
    Future.delayed(const Duration(milliseconds: 200), () {
      if (!mounted) return;
      showDialog(
        context: context,
        builder: (context) {
          final text = items.isEmpty
              ? "❌ No valid items detected!"
              : "${items.length} items found!";
          return AlertDialog(
            title: const Text("📊 Scan Complete"),
            content: Text(text),
            actions: [
              TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text("OK")),
            ],
          );
        },
      );
    });
  }

  // ---------- Helper rules ---------- //

  bool _isValidItemName(String name) {
    if (name.isEmpty) return false;

    // Must contain at least one letter
    if (!RegExp(r'[A-Za-z]').hasMatch(name)) return false;

    // Pure numbers / money are not item names
    if (RegExp(r'^\d+([.,]\d+)?$').hasMatch(name)) return false;

    // Overly long descriptions are usually whole sentences / notes
    if (name.split(' ').length > 12) return false;

    // Drop very generic meta words
    if (RegExp(
            r'(total|sub\s*total|discount|amount|tax|igst|cgst|sgst|balance|due|rate|gst|qty|pcs|oth|taxable|value|phone|mobile|email|date|time|bill|receipt|invoice|store|customer|contact|header|footer|summary)',
            caseSensitive: false)
        .hasMatch(name)) {
      return false;
    }

    return true;
  }
  // =================== rest of your original code =================== //

  void updateItem(int index, ReceiptItem item) {
    print('✏️ [ReceiptScanner] Updating item at index $index: ${item.name}');
    setState(() {
      extractedItems[index] = item;
    });
    print('✅ [ReceiptScanner] Item updated successfully');
  }

  void removeItem(int index) {
    if (index >= 0 && index < extractedItems.length) {
      print(
          '🗑️ [ReceiptScanner] Removing item at index $index: ${extractedItems[index].name}');
      setState(() {
        extractedItems.removeAt(index);
      });
      print('✅ [ReceiptScanner] Item removed successfully');
    } else {
      print('⚠️ [ReceiptScanner] Invalid index for removal: $index');
    }
  }

  void saveAllItems() async {
    try {
      print('💾 [ReceiptScanner] Save all items button tapped');
      print('💾 [ReceiptScanner] Items to save: ${extractedItems.length}');

      if (extractedItems.isEmpty) {
        print('⚠️ [ReceiptScanner] No items to save');
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('No items to save'),
              backgroundColor: Colors.orange,
            ),
          );
        }
        return;
      }

      int savedCount = 0;
      int errorCount = 0;
      List<String> errors = [];

      print(
          '🔄 [ReceiptScanner] Starting to save ${extractedItems.length} items...');

      for (int i = 0; i < extractedItems.length; i++) {
        ReceiptItem item = extractedItems[i];
        try {
          print(
              '📦 [ReceiptScanner] Processing item ${i + 1}/${extractedItems.length}: ${item.name}');

          if (item.name.isEmpty || item.amount.isEmpty) {
            print(
                '⚠️ [ReceiptScanner] Skipping item ${i + 1}: empty name or amount');
            errorCount++;
            errors.add('Item ${i + 1}: Empty name or amount');
            continue;
          }

          // Validate amount
          if (double.tryParse(item.amount) == null) {
            print(
                '❌ [ReceiptScanner] Invalid amount for item ${i + 1}: ${item.amount}');
            errors.add('${item.name}: Invalid amount (${item.amount})');
            errorCount++;
            continue;
          }

          print(
              '🔨 [ReceiptScanner] Creating Add_data for: ${item.name} (₹${item.amount})');
          final add = Add_data(
            'Expense',
            item.amount,
            DateTime.now(),
            item.name,
            item.category,
          );

          print(
              '💾 [ReceiptScanner] Calling TransactionProvider.addTransaction...');
          context.read<TransactionProvider>().addTransaction(add);
          savedCount++;
          print(
              '✅ [ReceiptScanner] Item ${i + 1} (${item.name}) saved successfully');
        } catch (e, stackTrace) {
          print(
              '❌ [ReceiptScanner] Error saving item ${i + 1} (${item.name}):');
          print('   Error: $e');
          print('   Stack trace: $stackTrace');
          errorCount++;
          errors.add('${item.name}: ${e.toString()}');
        }
      }

      print(
          '📊 [ReceiptScanner] Save summary: $savedCount saved, $errorCount errors');

      // Show result message
      if (!mounted) return;

      String message;
      Color backgroundColor;

      if (savedCount == 0 && errorCount > 0) {
        // All failed
        message =
            'Failed to save all $errorCount items. Check logs for details.';
        backgroundColor = Colors.red;
        print('❌ [ReceiptScanner] All items failed to save');
      } else if (errorCount > 0) {
        // Partial success
        message =
            '✅ $savedCount saved, ❌ $errorCount failed. Check logs for details.';
        backgroundColor = Colors.orange;
        print(
            '⚠️ [ReceiptScanner] Partial success: $savedCount/${extractedItems.length} saved');
      } else {
        // All succeeded
        message = '✅ All $savedCount items saved successfully!';
        backgroundColor = Colors.green;
        print('🎉 [ReceiptScanner] All items saved successfully!');
      }

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: backgroundColor,
          duration: Duration(seconds: savedCount > 0 ? 2 : 4),
        ),
      );

      // Only navigate away if at least one item was saved successfully
      // Wait a bit for the snackbar to show, then navigate
      if (savedCount > 0 && mounted) {
        await Future.delayed(const Duration(milliseconds: 500));
        if (mounted && Navigator.of(context).canPop()) {
          print('🚪 [ReceiptScanner] Navigating back...');
          Navigator.of(context).pop();
        }
      }
    } catch (e, stackTrace) {
      print('❌ [ReceiptScanner] Critical error in saveAllItems:');
      print('   Error: $e');
      print('   Error type: ${e.runtimeType}');
      print('   Stack trace: $stackTrace');

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error saving items: ${e.toString()}'),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 4),
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      body: SafeArea(
        child: Stack(
          alignment: AlignmentDirectional.center,
          children: [
            backgroundContainer(context),
            if (_image == null)
              Positioned(
                top: 120,
                child: noImageContainer(),
              )
            else
              Positioned(
                top: 120,
                child: imageAndItemsContainer(),
              ),
          ],
        ),
      ),
    );
  }

  Container noImageContainer() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        color: Colors.white,
      ),
      height: 500,
      width: 340,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.receipt, size: 80, color: Colors.grey.shade400),
          const SizedBox(height: 20),
          Text(
            'Scan Receipt',
            style: TextStyle(
              fontSize: 22,
              fontWeight: FontWeight.w600,
              color: Colors.grey.shade700,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            'Take a photo of your receipt to extract expenses',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 14,
              color: Colors.grey.shade500,
            ),
          ),
          const SizedBox(height: 50),
          GestureDetector(
            onTap: () => pickImage(ImageSource.camera),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 15),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                color: Color.fromARGB(255, 101, 76, 116),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.camera_alt, color: Colors.white),
                  SizedBox(width: 10),
                  Text(
                    'Take Photo',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 15),
          GestureDetector(
            onTap: () => pickImage(ImageSource.gallery),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 30, vertical: 15),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: Color.fromARGB(255, 101, 76, 116),
                  width: 2,
                ),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.image, color: Color.fromARGB(255, 101, 76, 116)),
                  SizedBox(width: 10),
                  Text(
                    'Choose from Gallery',
                    style: TextStyle(
                      color: Color.fromARGB(255, 101, 76, 116),
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Container imageAndItemsContainer() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        color: Colors.white,
      ),
      height: 550,
      width: 340,
      child: Column(
        children: [
          const SizedBox(height: 15),
          ClipRRect(
            borderRadius: BorderRadius.circular(15),
            child: Image.file(
              _image!,
              height: 100,
              width: 100,
              fit: BoxFit.cover,
            ),
          ),
          const SizedBox(height: 10),
          if (isProcessing)
            const Padding(
              padding: EdgeInsets.all(20),
              child: CircularProgressIndicator(
                color: Color.fromARGB(255, 101, 76, 116),
              ),
            )
          else if (extractedItems.isEmpty)
            Padding(
              padding: const EdgeInsets.all(20),
              child: Text(
                'No items detected. Try adjusting the image.',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey.shade600,
                ),
              ),
            )
          else
            Expanded(
              child: ListView.builder(
                itemCount: extractedItems.length,
                padding:
                    const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
                itemBuilder: (context, index) {
                  return ItemEditCard(
                    item: extractedItems[index],
                    categories: _categories,
                    onUpdate: (item) => updateItem(index, item),
                    onRemove: () => removeItem(index),
                  );
                },
              ),
            ),
          const SizedBox(height: 15),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              GestureDetector(
                onTap: () {
                  setState(() {
                    _image = null;
                    extractedItems = [];
                  });
                },
                child: Container(
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: Color.fromARGB(255, 101, 76, 116),
                      width: 2,
                    ),
                  ),
                  width: 100,
                  height: 45,
                  child: const Text(
                    'Retake',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: Color.fromARGB(255, 101, 76, 116),
                      fontSize: 16,
                    ),
                  ),
                ),
              ),
              GestureDetector(
                onTap: saveAllItems,
                child: Container(
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    color: Color.fromARGB(255, 101, 76, 116),
                  ),
                  width: 100,
                  height: 45,
                  child: const Text(
                    'Save All',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: Colors.white,
                      fontSize: 16,
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
        ],
      ),
    );
  }

  Column backgroundContainer(BuildContext context) {
    return Column(
      children: [
        Container(
          width: double.infinity,
          height: 240,
          decoration: const BoxDecoration(
            color: Color.fromARGB(255, 101, 76, 116),
            borderRadius: BorderRadius.only(
              bottomLeft: Radius.circular(20),
              bottomRight: Radius.circular(20),
            ),
          ),
          child: Column(
            children: [
              const SizedBox(height: 40),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 15),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    GestureDetector(
                      onTap: () {
                        Navigator.of(context).pop();
                      },
                      child: const Icon(Icons.arrow_back, color: Colors.white),
                    ),
                    const Text(
                      'Receipt Scanner',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                    const Icon(
                      Icons.receipt_long,
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

class ReceiptItem {
  String name;
  String amount;
  String category;

  ReceiptItem({
    required this.name,
    required this.amount,
    required this.category,
  });
}

class ItemEditCard extends StatefulWidget {
  final ReceiptItem item;
  final List<String> categories;
  final Function(ReceiptItem) onUpdate;
  final VoidCallback onRemove;

  const ItemEditCard({
    super.key,
    required this.item,
    required this.categories,
    required this.onUpdate,
    required this.onRemove,
  });

  @override
  State<ItemEditCard> createState() => _ItemEditCardState();
}

class _ItemEditCardState extends State<ItemEditCard> {
  late TextEditingController nameController;
  late TextEditingController amountController;
  late String selectedCategory;

  @override
  void initState() {
    super.initState();
    nameController = TextEditingController(text: widget.item.name);
    amountController = TextEditingController(text: widget.item.amount);
    selectedCategory = widget.item.category;
  }

  @override
  void dispose() {
    nameController.dispose();
    amountController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: TextField(
                    controller: nameController,
                    decoration: InputDecoration(
                      labelText: 'Item Name',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                    ),
                    onChanged: (_) {
                      widget.onUpdate(
                        ReceiptItem(
                          name: nameController.text,
                          amount: amountController.text,
                          category: selectedCategory,
                        ),
                      );
                    },
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 80,
                  child: TextField(
                    controller: amountController,
                    decoration: InputDecoration(
                      labelText: 'Amount',
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 8,
                      ),
                    ),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                    onChanged: (_) {
                      widget.onUpdate(
                        ReceiptItem(
                          name: nameController.text,
                          amount: amountController.text,
                          category: selectedCategory,
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: DropdownButton<String>(
                    value: selectedCategory,
                    isExpanded: true,
                    items: widget.categories
                        .map(
                          (category) => DropdownMenuItem(
                            value: category,
                            child: Text(category),
                          ),
                        )
                        .toList(),
                    onChanged: (value) {
                      if (value != null) {
                        setState(() {
                          selectedCategory = value;
                        });
                        widget.onUpdate(
                          ReceiptItem(
                            name: nameController.text,
                            amount: amountController.text,
                            category: selectedCategory,
                          ),
                        );
                      }
                    },
                  ),
                ),
                const SizedBox(width: 10),
                IconButton(
                  icon: const Icon(Icons.delete, color: Colors.red),
                  onPressed: widget.onRemove,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
