import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

void main() {
  runApp(const TempConvApp());
}

class TempConvApp extends StatelessWidget {
  const TempConvApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'TempConv',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const ConverterPage(),
    );
  }
}

class ConverterPage extends StatefulWidget {
  const ConverterPage({super.key});

  @override
  State<ConverterPage> createState() => _ConverterPageState();
}

class _ConverterPageState extends State<ConverterPage> {
  static String get _apiBase {
    // Allow override via --dart-define=API_BASE=...
    const envBase = bool.hasEnvironment('API_BASE')
        ? String.fromEnvironment('API_BASE')
        : '';
    if (envBase.isNotEmpty) return envBase;
    // Default to Render backend
    return 'https://tempconv-80uz.onrender.com/api';
  }

  final _inputController = TextEditingController();
  double? _result;
  String? _error;
  bool _loading = false;
  bool _celsiusToFahrenheit = true;

  @override
  void dispose() {
    _inputController.dispose();
    super.dispose();
  }

  Future<void> _convert() async {
    final text = _inputController.text.trim();
    if (text.isEmpty) {
      setState(() {
        _error = 'Enter a temperature';
        _result = null;
      });
      return;
    }
    final value = double.tryParse(text);
    if (value == null) {
      setState(() {
        _error = 'Invalid number';
        _result = null;
      });
      return;
    }

    setState(() {
      _error = null;
      _result = null;
      _loading = true;
    });

    try {
      const path = '/convert'; // backend expects /api/convert; _apiBase ends with /api
      final uri = Uri.parse('$_apiBase$path');

      final fromUnit = _celsiusToFahrenheit ? 'CELSIUS' : 'FAHRENHEIT';
      final toUnit = _celsiusToFahrenheit ? 'FAHRENHEIT' : 'CELSIUS';

      final res = await http.post(
        uri,
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'value': value,
          'from_unit': fromUnit,
          'to_unit': toUnit,
        }),
      );
      if (res.statusCode != 200) {
        setState(() {
          _error = 'Server error: ${res.statusCode}';
          _loading = false;
        });
        return;
      }
      final data = jsonDecode(res.body) as Map<String, dynamic>;
      setState(() {
        _result = (data['value'] as num).toDouble();
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _result = null;
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final from = _celsiusToFahrenheit ? 'Celsius' : 'Fahrenheit';
    final to = _celsiusToFahrenheit ? 'Fahrenheit' : 'Celsius';

    return Scaffold(
      appBar: AppBar(
        title: const Text('TempConv'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text(
                  '$from → $to',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 24),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _inputController,
                        keyboardType:
                            const TextInputType.numberWithOptions(decimal: true),
                        inputFormatters: [
                          FilteringTextInputFormatter.allow(RegExp(r'[-0-9.]')), 
                        ],
                        decoration: InputDecoration(
                          labelText: from,
                          border: const OutlineInputBorder(),
                        ),
                        onSubmitted: (_) => _convert(),
                      ),
                    ),
                    const SizedBox(width: 16),
                    FilledButton(
                      onPressed: _loading ? null : _convert,
                      child: _loading
                          ? const SizedBox(
                              width: 24,
                              height: 24,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Text('Convert'),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                SwitchListTile(
                  title: const Text('Celsius → Fahrenheit'),
                  value: _celsiusToFahrenheit,
                  onChanged: (v) {
                    setState(() {
                      _celsiusToFahrenheit = v;
                      _result = null;
                      _error = null;
                    });
                  },
                ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    _error!,
                    style: TextStyle(color: Theme.of(context).colorScheme.error),
                  ),
                ],
                if (_result != null) ...[
                  const SizedBox(height: 24),
                  Text(
                    'Result: ${_result!.toStringAsFixed(2)} °$to',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}